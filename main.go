package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
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
	NAIToSID, err := uintToBool(readBits(data[6], 6))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	NoMSDLimit, err := uintToBool(readBits(data[6], 7))
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
	UFlag, err := uintToBool(readBits(data[7], 0))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Printf("UFlag: %08b \n", readBits(data[7], 0))
	sCap := &StatefulPCECapability{
		Type:   binary.BigEndian.Uint16(data[:2]),
		Length: binary.BigEndian.Uint16(data[2:4]),
		Flags:  binary.BigEndian.Uint32(data[4:8]),
		UFlag:  UFlag,
	}
	printAsJSON(sCap)
	return sCap
}

func (p *PCEPSession) initLSP() {

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

func handleTCPConn(conn net.Conn) {
	var p PCEPSession
	log.Println("Got connection", conn.RemoteAddr())
	defer conn.Close()
	buff := make([]byte, 1024)
	for {
		l, err := conn.Read(buff)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}
		if l < 4 {
			continue
		}
		p.handleMsg(buff[:l], conn)
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
		go handleTCPConn(conn)
	}
}
