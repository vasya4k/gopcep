package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/bits"
	"net"
	"time"
)

//PCEPSession ssae
type PCEPSession struct {
	State     int
	Conn      net.Conn
	ID        uint8
	RemoteOK  bool
	Keepalive int
	LSPs      map[uint32]string
	IDCounter uint32
}

func newSRPObject(flags, SRPID uint32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, flags)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, SRPID)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// https://tools.ietf.org/html/rfc8231#section-7.3.2
func newPathName(name string) ([]byte, error) {
	var objType uint16 = 17
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, objType)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, uint16(len(name)))
	if err != nil {
		return nil, err
	}
	b, err := padBytes(append(buf.Bytes(), []byte(name)...), 4)
	if err != nil {
		return nil, err
	}
	return b, nil
}

//SREROSub str
type SREROSub struct {
	LooseHop      bool
	NT            uint8
	MBit          bool
	CBit          bool
	NoSID         bool
	NoNAI         bool
	SID           uint32
	IPv4NodeID    string
	IPv4Adjacency []string
	UnnuV4Adj     UnnuAdjIPv4NodeIDs
}

func (s *SREROSub) validateSREROSub() error {
	if s.NoNAI && s.NT > 0 {
		return errors.New("NoNAI flag is set but NT is not zero ")
	}
	if !s.NoNAI && s.NT == 0 {
		return errors.New("NoNAI flag is not set but NT is zero ")
	}
	if s.MBit && !s.CBit {
		return errors.New("M bit is zero then the C bit MUST be zero")
	}
	if s.NoSID && (!s.MBit || !s.CBit) {
		return errors.New("M and C bits MUST be set to zero")
	}
	if s.NoSID && s.NoNAI {
		return errors.New("The S and F bits MUST NOT both be set to 1")
	}
	return nil
}

type UnnuAdjIPv4NodeIDs struct {
	LocalNodeID       string
	LocalInterfaceID  string
	RemoteNodeID      string
	RemoteInterfaceID string
}

// https://tools.ietf.org/html/draft-ietf-pce-segment-routing-14#section-5.3.1
func newSREROSubObject(ero SREROSub) ([]byte, error) {
	err := ero.validateSREROSub()
	if err != nil {
		return nil, err
	}
	var objType uint8 = 36
	if ero.LooseHop {
		objType |= (1 << 7)
	}
	var flags uint8
	if ero.MBit {
		flags |= (1 << 0)
	}
	if ero.CBit {
		flags |= (1 << 1)
	}
	if ero.NoSID {
		flags |= (1 << 2)
	}
	if ero.NoNAI {
		flags |= (1 << 3)
	}
	byteERO := []byte{
		0: objType,
		1: 0,
		2: bits.RotateLeft8(ero.NT, 4),
		3: flags,
	}
	if !ero.NoSID {
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.BigEndian, ero.SID)
		if err != nil {
			return nil, err
		}
		byteERO = append(byteERO, buf.Bytes()...)
	}
	switch ero.NT {
	case 0:
		byteERO[1] = uint8(len(byteERO))
		return byteERO, nil
	case 1:
		nodeID, err := ipToUnit32(ero.IPv4NodeID)
		if err != nil {
			return nil, err
		}
		buf := new(bytes.Buffer)
		err = binary.Write(buf, binary.BigEndian, nodeID)
		if err != nil {
			return nil, err
		}
		byteERO = append(byteERO, buf.Bytes()...)
		byteERO[1] = uint8(len(byteERO))
		return byteERO, nil
	case 2:
		return nil, errors.New("IPv6 Node ID not implemented yet")
	case 3:
		if len(ero.IPv4Adjacency) != 2 {
			return nil, errors.New("malformed IPv4 Adjacency specified")
		}
		local, err := ipToUnit32(ero.IPv4Adjacency[0])
		if err != nil {
			return nil, err
		}
		remote, err := ipToUnit32(ero.IPv4Adjacency[1])
		if err != nil {
			return nil, err
		}
		buf := new(bytes.Buffer)
		err = binary.Write(buf, binary.BigEndian, local)
		if err != nil {
			return nil, err
		}
		err = binary.Write(buf, binary.BigEndian, remote)
		if err != nil {
			return nil, err
		}
		byteERO = append(byteERO, buf.Bytes()...)
		byteERO[1] = uint8(len(byteERO))
		return byteERO, nil
	case 4:
		return nil, errors.New("IPv6 Adjacency not implemented yet")
	case 5:
		return nil, errors.New("Unnumbered Adjacency with IPv4 NodeIDs not implemented yet")
	default:
		return nil, errors.New(" NAI Type not defined in RFC")
	}
}

// https://tools.ietf.org/html/rfc5440#section-7.9
func newSREroObj(subEROs []SREROSub) ([]byte, error) {
	ero := make([]byte, 0)
	for _, subERO := range subEROs {
		subEROBytes, err := newSREROSubObject(subERO)
		if err != nil {
			return nil, err
		}
		ero = append(ero, subEROBytes...)
	}
	headerERO, err := newCommonObjHeader(7, 1, true, uint16(len(ero)))
	if err != nil {
		return nil, err
	}
	return headerERO, nil
}

// https://tools.ietf.org/html/rfc5440#section-7.11
func newLSPAObject(setupPrio, holdPrio uint8, localProtect bool) ([]byte, error) {
	lspa := []byte{
		0: setupPrio,
		1: holdPrio,
		2: func() uint8 {
			var flags uint8
			if localProtect {
				flags |= (1 << 0)
				return flags
			}
			return flags
		}(),
		3: 0,
	}
	//first 12 bytes set to zero
	lspa = append(make([]byte, 12), lspa...)
	headerLSPA, err := newCommonObjHeader(9, 1, true, uint16(len(lspa)))
	if err != nil {
		return nil, err
	}
	return headerLSPA, nil
}

// https://tools.ietf.org/html/rfc5440#section-7.7
func newBandwidthObj(objType uint8, bandwidth uint32) ([]byte, error) {
	if objType != 1 || objType != 2 {
		return nil, errors.New("Object-Type values can only be 1 or 2")
	}
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, bandwidth)
	if err != nil {
		return nil, err
	}
	headerBW, err := newCommonObjHeader(5, objType, true, uint16(len(buf.Bytes())))
	if err != nil {
		return nil, err
	}
	return headerBW, nil
}

//https://tools.ietf.org/html/rfc5440#section-7.6
func newEndpointsPbj(srcStr, dstStr string) ([]byte, error) {
	src, err := ipToUnit32(srcStr)
	if err != nil {
		return nil, err
	}
	dst, err := ipToUnit32(dstStr)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, src)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, dst)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// https://tools.ietf.org/html/rfc8231#section-7.3
// Operational - 3 bits on PCC will always be set to zero
// so not accepting it as a param
func (p *PCEPSession) newLSPObj(delegate, sync, remove, admin bool) ([]byte, error) {
	p.IDCounter++
	// 2 ** 20 - 1 = 1048575 checking for overflow of 20bits
	if p.IDCounter > 1048575 {
		return nil, errors.New("session id limit reached > 1048575 exiting")
	}
	body := bits.RotateLeft32(p.IDCounter, 4)
	if delegate {
		// setting delegate flag at possition 1
		body |= (1 << 0)
	}
	if sync {
		// setting sync flag at possition 1
		body |= (1 << 1)
	}
	if remove {
		// setting remove flag at possition 1
		body |= (1 << 2)
	}
	if admin {
		// setting remove flag at possition 1
		body |= (1 << 3)
	}
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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

func (p *PCEPSession) handleMsg(data []byte, conn net.Conn) {
	p.Conn = conn
	fmt.Printf("Whole MSG: %08b \n", data)
	ch := parseCommonHeader(data[:4])
	coh := parseCommonObjectHeader(data[4:8])
	switch {
	case ch.MessageType == 1:
		if len(data) < 12 {
			log.Println("OPEN msg is too short")
		}
		go p.RcvSessionOpen(coh, data)
	case ch.MessageType == 2:
		log.Printf("recv keepalive from %s peer", p.Conn.RemoteAddr().String())
		if p.State == 2 {
			p.initLSP()
		}
	case ch.MessageType == 3:
		log.Println("recv path computation request ")
	case ch.MessageType == 4:
		log.Println("recv path computation reply ")
	case ch.MessageType == 5:
		log.Println("recv notification ")
	case ch.MessageType == 6:
		if len(data) < 12 {
			log.Println("ERR msg is too short")
		}
		parseErr(data[8:])
	case ch.MessageType == 7:
		log.Println("recv close msg")
	case ch.MessageType == 10:
		log.Println("recv PCC State report ")
	case ch.MessageType == 11:
		log.Println("pcc update msg recved")
	default:
		log.Println("Unknown msg recived")
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
	p.State = 2
	p.SendKeepAlive()
}
