package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
)

// only needed below for sample processing

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

func handleRequest(conn net.Conn) {
	// rd := bufio.NewReader(conn)
	log.Println("Got connection", conn.RemoteAddr())
	defer conn.Close()
	buff := make([]byte, 1024)
	for {
		l, err := conn.Read(buff)

		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}
		data := buff[:l]

		parseCommonHeader(data[:4])

		parseCommonObjectHeader(data[4:8])

		parseOpenObject(data[8:12])

		parseStatefulPCECap(data[12:20])

	}
}

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
	fmt.Printf("StatefulPCECapability: %08b \n", data)
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

func uintToBool(i uint) (bool, error) {
	if i == 0 {
		return false, nil
	} else if i == 1 {
		return true, nil
	}
	return false, errors.New("Bool value is not 1 or zero")
}

func bits(by byte, subset ...uint) (r uint) {
	b := uint(by)
	i := uint(0)
	for _, v := range subset {
		if b&(1<<v) > 0 {
			r = r | 1<<uint(i)
		}
		i++
	}
	return
}
