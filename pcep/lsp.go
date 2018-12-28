package pcep

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

// LSP represents a Segment routing LSP
type LSP struct {
	Delegate     bool
	Sync         bool
	Remove       bool
	Admin        bool
	Name         string
	Src          string
	Dst          string
	EROList      []EROSub
	SetupPrio    uint8
	HoldPrio     uint8
	LocalProtect bool
	BW           uint32
}

// InitLSP aaaa
func (s *Session) InitLSP(l *LSP) error {
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
	// fmt.Printf("ep %d bin string %08b \n", len(ep), ep)
	ero, err := newERObj(l.EROList)
	if err != nil {
		return err
	}
	fmt.Printf("ero %d bin string %08b \n", len(ero), ero)
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

// https://tools.ietf.org/html/rfc3209#section-4.3.3
// https://tools.ietf.org/html/rfc5440#section-7.9
func newERObj(subEROs []EROSub) ([]byte, error) {
	ero := make([]byte, 0)
	for _, subERO := range subEROs {
		subEROBytes, err := newEROSubObject(subERO)
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

//SREROSub str
type EROSub struct {
	LooseHop bool
	IPv4Addr string
	Mask     uint8
	Type     uint8
}

// https://tools.ietf.org/html/rfc3209#section-4.3.3
func newEROSubObject(ero EROSub) ([]byte, error) {
	var objType uint8 = ero.Type
	if ero.LooseHop {
		objType |= (1 << 7)
	}
	ipv4 := []byte{
		0: objType,
		1: 8,
	}
	ip, err := ipToUnit32(ero.IPv4Addr)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, ip)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, ero.Mask)
	if err != nil {
		return nil, err
	}
	var reserved uint8
	err = binary.Write(buf, binary.BigEndian, reserved)
	if err != nil {
		return nil, err
	}
	return append(ipv4, buf.Bytes()...), nil
}
