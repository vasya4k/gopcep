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
	// accept connection on port

	// run loop forever (or until ctrl-c)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		// will listen for message to process ending in newline (\n)
		// message, err := bufio.NewReader(conn).ReadString('\n')
		// if err != nil {
		// 	log.Fatalln(err)
		// }
		// // output message received
		// fmt.Print("Message Received:", string(message))
		// sample process for string received
		// newmessage := strings.ToUpper(message)
		// send new string back to client
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

// Common Header

//      0                   1                   2                   3
//      0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//     | Ver |  Flags  |  Message-Type |       Message-Length          |
//     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

//                 Figure 7: PCEP Message Common Header

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

// Common Object Header

//    A PCEP object carried within a PCEP message consists of one or more
//    32-bit words with a common header that has the following format:

//     0                   1                   2                   3
//     0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//    | Object-Class  |   OT  |Res|P|I|   Object Length (bytes)       |
//    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//    |                                                               |
//    //                        (Object body)                        //
//    |                                                               |
//    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

//                   Figure 8: PCEP Common Object Header

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

// OPEN Object-Class is 1.

//    OPEN Object-Type is 1.

//    The format of the OPEN object body is as follows:

//     0                   1                   2                   3
//     0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//    | Ver |   Flags |   Keepalive   |  DeadTimer    |      SID      |
//    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//    |                                                               |
//    //                       Optional TLVs                         //
//    |                                                               |
//    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

//                     Figure 9: OPEN Object Format
// As of today 02.12.2018 and testing against Juniper vMX JunOS 17.2R1.13
// If you see Common Object Header length is 24 Bytes 4 bytes is the CommonObjectHeader
// next 4 bytes is OPEN Object so it is 24-4-4 = 16. The remainig 16 are Optional TLVs and can be found
// In PCEP extensions described in https://tools.ietf.org/html/rfc8231#section-7.1.1
// Path Computation Element Communication Protocol (PCEP) Extensions for Stateful PCE

// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |               Type=16         |            Length=4           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                             Flags                           |U|
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

// 		 Figure 9: STATEFUL-PCE-CAPABILITY TLV Format

// The type (16 bits) of the TLV is 16.  The length field is 16 bits
// long and has a fixed value of 4
// The value comprises a single field -- Flags (32 bits):

//    U (LSP-UPDATE-CAPABILITY - 1 bit):  if set to 1 by a PCC, the U flag
//       indicates that the PCC allows modification of LSP parameters; if
//       set to 1 by a PCE, the U flag indicates that the PCE is capable of

// That gives us another 2+2+4 = 8 bytes so 16-8 = 8 bytes remaining and
// we need to look into another rfc draft https://tools.ietf.org/html/draft-ietf-pce-segment-routing-14#section-5.1
// 0                   1                   2                   3
//        0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//       |            Type=26            |            Length=4           |
//       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//       |         Reserved              |   Flags   |N|L|      MSD      |
//       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

//                 Figure 1: SR-PCE-CAPABILITY sub-TLV format
// The code point for the TLV type is 26.  The TLV length is 4 octets.
// The type (16 bits) The length field is 16 bits.
// The 32-bit value is formatted as follows.

//    Reserved:  MUST be set to zero by the sender and MUST be ignored by
//       the receiver.

//    Flags:  This document defines the following flag bits.  The other
//       bits MUST be set to zero by the sender and MUST be ignored by the
//       receiver.

//       *  N: A PCC sets this bit to 1 to indicate that it is capable of
//          resolving a Node or Adjacency Identifier (NAI) to a SID.

//       *  L: A PCC sets this bit to 1 to indicate that it does not impose
//          any limit on the MSD.

//    Maximum SID Depth (MSD):  specifies the maximum number of SIDs (MPLS
//       label stack depth in the context of this document) that a PCC is
//       capable of imposing on a packet.  Section 6.1 explains the
//       relationship between this field and the L bit.
// The above object is  2+2+4 = 8 bytes and now we can see this adds up
// if we test agains a juniper device.

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

// fmt.Println("len", binary.Size(buff))
// cmd, err := rd.  ('\n')
// size, err := rd.ReadF
// if err != nil {
// 	log.Println(err)
// 	return
// }
// buff := make([]byte, int(size))
// packetLen, err := io.ReadFull(rd, buff)
// if err != nil {
// 	log.Println(err)
// 	return
// }

// data, err := ioutil.ReadAll(conn)
// if err != nil {
// 	log.Println(err)
// }

// fmt.Printf("%v len: %d \n", data, len(data))
// x1 := binary.BigEndian.Uint16(data[:0])
// x2 := binary.BigEndian.Uint16(data[0:1])
// x3 := binary.BigEndian.Uint16(data[1:2])
// x4 := binary.BigEndian.Uint16(data[2:3])
// fmt.Printf("%d %d %d  %d ", x1, x2, x3, x4)
// }

// err = binary.Read(bytes.NewReader(data), binary.BigEndian, &data)
// if err != nil {
// 	log.Println(err)
// 	break
// }
// fmt.Printf("%d %d %d %d ",data)

// Make a buffer to hold incoming data.

// Read the incoming connection into the buffer.

// Send a response back to person contacting us.
// conn.Write([]byte("Message received."))
