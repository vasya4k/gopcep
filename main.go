package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

const msgTooShortErr = "recived msg is too short to parse common header and common object header out of it"

// fmt.Printf("Data: %08b \n", data[:4])
func printAsJSON(i interface{}) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(b))
}

//CommonHeader to store PCEP CommonHeader
type CommonHeader struct {
	Version       uint8
	Flags         uint8
	MessageType   uint8
	MessageLength uint16
}

func parseCommonHeader(data []byte) *CommonHeader {
	header := &CommonHeader{
		Version:       data[0] >> 5,
		Flags:         data[0] & (32 - 1),
		MessageType:   data[1],
		MessageLength: binary.BigEndian.Uint16(data[2:4]),
	}
	printAsJSON(header)
	return header
}

//CommonObjectHeader to store PCEP CommonObjectHeader
type CommonObjectHeader struct {
	ObjectClass    uint8
	ObjectType     uint8
	Reservedfield  uint8
	ProcessingRule bool
	Ignore         bool
	ObjectLength   uint16
}

func parseCommonObjectHeader(data []byte) *CommonObjectHeader {
	obj := &CommonObjectHeader{
		ObjectClass:  data[0],
		ObjectType:   data[1] >> 4,
		ObjectLength: binary.BigEndian.Uint16(data[2:4]),
	}
	printAsJSON(obj)
	return obj
}

//SRPCECap  https://tools.ietf.org/html/draft-ietf-pce-segment-routing-14#section-5.1
type SRPCECap struct {
	Type       uint16
	Length     uint16
	Reserved   uint16
	NAIToSID   bool
	NoMSDLimit bool
	MSD        uint8
}

func parseSRCap(data []byte) *SRPCECap {
	fmt.Printf("SR Cap: %08b \n", data)
	NAIToSID, err := uintToBool(bits(data[6], 6))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	NoMSDLimit, err := uintToBool(bits(data[6], 7))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	srCap := &SRPCECap{
		Type:       binary.BigEndian.Uint16(data[:2]),
		Length:     binary.BigEndian.Uint16(data[2:4]),
		Reserved:   binary.BigEndian.Uint16(data[4:6]),
		NAIToSID:   NAIToSID,
		NoMSDLimit: NoMSDLimit,
		MSD:        data[7],
	}
	printAsJSON(srCap)
	return srCap
}

//OpenObject to store PCEP OPEN Object
type OpenObject struct {
	Version   uint8
	Flags     uint8
	Keepalive uint8
	DeadTimer uint8
	SID       uint8
}

func parseOpenObject(data []byte) *OpenObject {
	open := &OpenObject{
		Version:   data[0] >> 5,
		Flags:     data[0] & (32 - 1),
		Keepalive: data[1],
		DeadTimer: data[2],
		SID:       data[3],
	}
	printAsJSON(open)
	return open
}

//StatefulPCECapability  rfc8231#section-7.1.1
type StatefulPCECapability struct {
	Type   uint16
	Length uint16
	Flags  uint32
	UFlag  bool
}

func parseStatefulPCECap(data []byte) *StatefulPCECapability {
	UFlag, err := uintToBool(bits(data[7], 0))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Printf("UFlag: %08b \n", bits(data[7], 0))
	sCap := &StatefulPCECapability{
		Type:   binary.BigEndian.Uint16(data[:2]),
		Length: binary.BigEndian.Uint16(data[2:4]),
		Flags:  binary.BigEndian.Uint32(data[4:8]),
		UFlag:  UFlag,
	}
	printAsJSON(sCap)
	return sCap
}

type PCEPSession struct {
	State    int
	Conn     net.Conn
	ID       int
	RemoteOK bool
}

//RcvSessionOpen recive msg handler
func (p *PCEPSession) RcvSessionOpen(coh *CommonObjectHeader, data []byte) {
	if coh.ObjectClass != 1 && coh.ObjectType != 1 {
		log.Printf("Remote IP: %s, object class and object type do not mathc OPEN msg RFC defenitions", p.Conn.RemoteAddr())
		return
	}
	oo := parseOpenObject(data[8:12])
	p.ID = int(oo.SID)
	p.RemoteOK = true
	p.State = 1
	parseStatefulPCECap(data[12:20])
	parseSRCap(data[20:28])
}

//SendSessionOpen send OPNE msg handler
func (p *PCEPSession) SendSessionOpen() {
	var a uint8
	a = 1
	b := []byte{
		0: 1,
	}
	b[0] = a
}

func handleRequest(conn net.Conn) {
	// rd := bufio.NewReader(conn)
	log.Println("Got connection", conn.RemoteAddr())
	defer conn.Close()
	buff := make([]byte, 1024)
	pSession := PCEPSession{
		Conn: conn,
	}
	for {
		l, err := conn.Read(buff)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}
		data := buff[:l]

		if len(data) < 8 {
			log.Println(msgTooShortErr)
			continue
		}
		fmt.Printf("Whole MSG: %08b \n", data)
		ch := parseCommonHeader(data[:4])
		coh := parseCommonObjectHeader(data[4:8])

		switch {
		case ch.MessageType == 1:
			if len(data) < 12 {
				log.Println("OPEN msg is too short")
				continue
			}
			pSession.RcvSessionOpen(coh, data)
		default:
			log.Println("Unknown msg recived")
		}

	}
}

func main() {
	// listen on all interfaces
	ln, err := net.Listen("tcp", "10.0.0.1:4189")
	if err != nil {
		log.Fatalln(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		// conn.Write([]byte(newmessage + "\n"))
		go handleRequest(conn)
	}
}
