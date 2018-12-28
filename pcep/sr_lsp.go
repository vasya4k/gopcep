package pcep

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/bits"
)

// SRLSP represents a Segment routing LSP
type SRLSP struct {
	Delegate     bool
	Sync         bool
	Remove       bool
	Admin        bool
	Name         string
	Src          string
	Dst          string
	EROList      []SREROSub
	SetupPrio    uint8
	HoldPrio     uint8
	LocalProtect bool
	BW           uint32
}

// InitLSP aaaa
func (s *Session) InitSRLSP(l *SRLSP) error {
	sro, err := s.newSRPObject()
	if err != nil {
		return err
	}
	// fmt.Printf("SRO %d bin string %08b \n", len(sro), sro)
	lsp, err := s.newLSPObj(l.Delegate, l.Sync, l.Remove, l.Admin, l.Name)
	if err != nil {
		return err
	}
	// fmt.Printf("lsp %d bin string %08b \n", len(lsp), lsp)
	ep, err := newEndpointsObj(l.Src, l.Dst)
	if err != nil {
		return err
	}
	fmt.Printf("ep %d bin string %08b \n", len(ep), ep)
	ero, err := newSRERObj(l.EROList)
	if err != nil {
		return err
	}
	// fmt.Printf("ero %d bin string %08b \n", len(ero), ero)
	lspa, err := newLSPAObject(l.SetupPrio, l.HoldPrio, l.LocalProtect)
	if err != nil {
		return err
	}
	// fmt.Printf("lspa %d bin string %08b \n", len(lspa), lspa)
	// bw, err := newBandwidthObj(1, l.BW)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("bw %d bin string %08b \n", len(bw), bw)
	msg := append(sro, lsp...)
	// fmt.Printf("msg %d bin string %08b \n", len(msg), msg)
	msg = append(msg, ep...)
	// fmt.Printf("msg ep %d bin string %08b \n", len(msg), msg)
	msg = append(msg, ero...)
	msg = append(msg, lspa...)
	// msg = append(msg, bw...)

	ch, err := newCommonHeader(12, uint16(len(msg)))
	if err != nil {
		return err
	}
	ch = append(ch, msg...)
	// fmt.Printf("Len %d bin string %08b \n", len(ch), ch)
	i, err := s.Conn.Write(ch)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Sent LSP Initiate Request: %d byte", i)
	// parseCommonHeader(ch[:4])
	// parseCommonObjectHeader(ch[4:8])
	// parseCommonObjectHeader(ch[16:20])
	// parseCommonObjectHeader(ch[36:40])
	// parseCommonObjectHeader(ch[48:52])
	// parseCommonObjectHeader(ch[64:68])
	// parseCommonObjectHeader(ch[84:88])
	return nil
}

// https://tools.ietf.org/html/rfc8231#section-7.2
// Stateful PCE Request Parameters
// Flags (32 bits): None defined yet.
// SRP Object-Class is 33.
// SRP Object-Type is 1.
func (s *Session) newSRPObject() ([]byte, error) {
	var flags uint32
	if s.SRPID == 4294967295 {
		s.SRPID = 0
	}
	s.SRPID++
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, flags)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, s.SRPID)
	if err != nil {
		return nil, err
	}
	ps, err := newPathSetupObj()
	if err != nil {
		return nil, err
	}
	sro, err := newCommonObjHeader(33, 1, true, append(buf.Bytes(), ps...))
	if err != nil {
		return nil, err
	}
	// sro, err := newCommonObjHeader(33, 1, true, buf.Bytes())
	// if err != nil {
	// 	return nil, err
	// }
	return sro, nil
}

// https://tools.ietf.org/html/rfc8408#section-4
func newPathSetupObj() ([]byte, error) {
	var (
		objType uint16 = 28
		length  uint16 = 4
		pst     uint32 = 1
	)
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, objType)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, length)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, pst)
	if err != nil {
		return nil, err
	}
	fmt.Printf("msg newPathSetupObj %d bin string %08b \n", len(buf.Bytes()), buf.Bytes())
	return buf.Bytes(), nil
}

// https://tools.ietf.org/html/rfc8231#section-7.3.2
// Type (16 bits): the type is 17.
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
	headerLSPA, err := newCommonObjHeader(9, 1, true, lspa)
	if err != nil {
		return nil, err
	}
	return headerLSPA, nil
}

// https://tools.ietf.org/html/rfc5440#section-7.7
func newBandwidthObj(objType uint8, bandwidth uint32) ([]byte, error) {
	if objType == 0 || objType > 2 {
		return nil, errors.New("Object-Type values can only be 1 or 2")
	}
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, bandwidth)
	if err != nil {
		return nil, err
	}
	headerBW, err := newCommonObjHeader(5, objType, true, buf.Bytes())
	if err != nil {
		return nil, err
	}
	return headerBW, nil
}

//https://tools.ietf.org/html/rfc5440#section-7.6
//currently IPv4 only
func newEndpointsObj(srcStr, dstStr string) ([]byte, error) {
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
	headerEP, err := newCommonObjHeader(4, 1, true, buf.Bytes())
	if err != nil {
		return nil, err
	}
	return headerEP, nil
}

// https://tools.ietf.org/html/rfc8231#section-7.3
// Operational - 3 bits on PCC will always be set to zero
// so not accepting it as a param
//    LSP Object-Class is 32.
//    LSP Object-Type is 1.
func (s *Session) newLSPObj(delegate, sync, remove, admin bool, name string) ([]byte, error) {
	// s.IDCounter++
	// 2 ** 20 - 1 = 1048575 checking for overflow of 20bits
	// if s.IDCounter > 1048575 {
	// 	return nil, errors.New("session id limit reached > 1048575 exiting")
	// }
	body := bits.RotateLeft32(s.IDCounter, 4)
	if delegate {
		// setting delegate flag at possition 0
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
	pn, err := newPathName(name)
	if err != nil {
		return nil, err
	}
	lspWH, err := newCommonObjHeader(32, 1, true, append(buf.Bytes(), pn...))
	if err != nil {
		return nil, err
	}
	return lspWH, nil
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

// https://tools.ietf.org/html/rfc5440#section-7.9
func newSRERObj(subEROs []SREROSub) ([]byte, error) {
	ero := make([]byte, 0)
	for _, subERO := range subEROs {
		subEROBytes, err := newSREROSubObject(subERO)
		if err != nil {
			return nil, err
		}
		ero = append(ero, subEROBytes...)
	}
	headerERO, err := newCommonObjHeader(7, 1, true, ero)
	if err != nil {
		return nil, err
	}
	return headerERO, nil
}

func (s *SREROSub) validateSREROSub() error {
	if s.NoNAI && s.NT > 0 {
		return errors.New("NoNAI flag is set but NT is not zero ")
	}
	if !s.NoNAI && s.NT == 0 {
		return errors.New("NoNAI flag is not set but NT is zero ")
	}
	if !s.MBit && s.CBit {
		return errors.New("M bit is zero then the C bit MUST be zero")
	}
	if s.NoSID && (s.MBit || s.CBit) {
		return fmt.Errorf("M and C bits MUST be set to zero m bit is %t and c bit is %t", s.MBit, s.CBit)
	}
	if s.NoSID && s.NoNAI {
		return errors.New("The S and F bits MUST NOT both be set to 1")
	}
	return nil
}

//UnnuAdjIPv4NodeIDs aaa
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
		err = binary.Write(buf, binary.BigEndian, ero.SID)
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
		fmt.Printf("SUB ERO %d bin string %08b \n", len(byteERO), byteERO)

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
