package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

const msgTooShortErr = "recived msg is too short to parse common header and common object header out of it"

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
	State     int
	Conn      net.Conn
	ID        uint8
	RemoteOK  bool
	Keepalive int
}

//RcvSessionOpen recive msg handler
func (p *PCEPSession) RcvSessionOpen(coh *CommonObjectHeader, data []byte) {
	if coh.ObjectClass != 1 && coh.ObjectType != 1 {
		log.Printf("Remote IP: %s, object class and object type do not mathc OPEN msg RFC defenitions", p.Conn.RemoteAddr())
		return
	}
	oo := parseOpenObject(data[8:12])
	p.ID = oo.SID
	p.Keepalive = int(oo.Keepalive)
	p.RemoteOK = true
	p.State = 1
	parseStatefulPCECap(data[12:20])
	parseSRCap(data[20:28])
	p.SendSessionOpen()
}

//SendKeepAlive start sending keep alive msgs
func (p *PCEPSession) SendKeepAlive() {
	commH := []byte{
		0: 32,
		1: 2,
		2: 0,
		3: 4,
	}
	var firstSent bool
	for {
		if firstSent {
			time.Sleep(time.Second * time.Duration(p.Keepalive))
		}
		i, err := p.Conn.Write(commH)
		if err != nil {
			log.Println(err)
		}
		log.Printf("keep alive sent %d bytes", i)
		firstSent = true
	}

}

//SendSessionOpen send OPNE msg handler
func (p *PCEPSession) SendSessionOpen() {
	//[00100000 00000001 00000000 00011100]
	commH := []byte{
		0: 32,
		1: 1,
		2: 0,
		3: 28,
	}
	// 00000001 00010000 00000000 00011000
	commObjH := []byte{
		0: 1,
		1: 16,
		2: 0,
		3: 24,
	}
	// 00100000 00011110 01111000 00000001
	open := []byte{
		0: 32,
		1: 30,
		2: 120,
		3: p.ID,
	}
	stCap := []byte{
		0: 0,
		1: 16,
		2: 0,
		3: 4,
		4: 0,
		5: 0,
		6: 0,
		7: 5,
	}
	srCap := []byte{
		0: 0,
		1: 26,
		2: 0,
		3: 4,
		4: 0,
		5: 0,
		6: 0,
		7: 5,
	}
	packet := append(commH, commObjH...)
	packet = append(packet, open...)
	packet = append(packet, stCap...)
	packet = append(packet, srCap...)
	fmt.Printf("Sending open %08b \n", packet)
	i, err := p.Conn.Write(packet)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Sent: %d byte", i)
	p.SendKeepAlive()
}

//ErrMsg  rfc8231#section-7.15
type ErrMsg struct {
	Reserved   uint8
	Flags      uint8
	ErrorType  uint8
	ErrorValue uint8
}

func parseErr(data []byte) {
	e := &ErrMsg{
		Reserved:   data[0],
		Flags:      data[1],
		ErrorType:  data[2],
		ErrorValue: data[3],
	}
	printAsJSON(e)
}

type SRPObject struct {
	Flags       uint32
	SRPIDNumber uint32
}

func parseSRP(data []byte) {
	// sr := &SRPObject{
	// 	Flags:
	// }
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

		if len(data) < 4 {
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
			go pSession.RcvSessionOpen(coh, data)
		case ch.MessageType == 6:
			if len(data) < 12 {
				log.Println("ERR msg is too short")
				continue
			}
			parseErr(data[8:])
		case ch.MessageType == 2:
			log.Printf("recv keepalive from %s peer", conn.RemoteAddr().String())
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
