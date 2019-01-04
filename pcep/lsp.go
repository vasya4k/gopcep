package pcep

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/big"

	"github.com/sirupsen/logrus"
)

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

//LSP represents a Segment routing LSP  https://tools.ietf.org/html/rfc8231#section-7.3
type LSP struct {
	Delegate     bool
	Sync         bool
	Remove       bool
	Admin        bool
	Oper         uint8
	Name         string
	Src          string
	Dst          string
	EROList      []EROSub
	SREROList    []SREROSub
	SetupPrio    uint8
	HoldPrio     uint8
	LocalProtect bool
	BW           uint32
	PLSPID       uint32
	LSPID        uint16
	IPv4ID       *LSPIPv4Identifiers
	IPv6ID       *LSPIPv6Identifiers
}

//https://tools.ietf.org/html/rfc8231#section-7.3
func parseLSPObj(data []byte) (*LSP, error) {
	d, err := uintToBool(readBits(data[3], 0))
	if err != nil {

		return nil, err
	}
	s, err := uintToBool(readBits(data[3], 1))
	if err != nil {
		return nil, err
	}
	r, err := uintToBool(readBits(data[3], 2))
	if err != nil {
		return nil, err
	}
	a, err := uintToBool(readBits(data[3], 3))
	if err != nil {
		return nil, err
	}
	lsp := &LSP{
		Delegate: d,
		Sync:     s,
		Remove:   r,
		Admin:    a,
		//shift right to get rid of d,s,r,a flags
		// then shift left to get rid remaining one bit
		// then shit right again to get the a clean value
		// there is a better solution but i do not have time right now
		Oper:   ((data[3] >> 4) << 5) >> 5,
		PLSPID: binary.BigEndian.Uint32(data[:4]) >> 12,
	}
	// fmt.Printf("After Int %08b \n", data)
	lsp.parseLSPSubObj(data[4:])
	printAsJSON(lsp)
	return lsp, nil
}

func (l *LSP) parseLSPSubObj(data []byte) error {
	fmt.Printf("After Int %08b \n", data)
	var (
		offset  uint16
		err     error
		counter int
	)
	// +4 need because obj header is not included in length
	for (len(data) - int(offset)) > 4 {
		counter++
		switch binary.BigEndian.Uint16(data[offset : offset+2]) {
		case 18:
			l.IPv4ID, err = parseLSPIPv4Identifiers(data[offset:])
			if err != nil {
				return err
			}
			offset = offset + binary.BigEndian.Uint16(data[offset+2:offset+4]) + 4
			continue
		case 19:
			l.IPv6ID, err = parseLSPIPv6Identifiers(data[offset:])
			if err != nil {
				return err
			}
			offset = offset + binary.BigEndian.Uint16(data[offset+2:offset+4]) + 4
			continue
		// https://tools.ietf.org/html/rfc8231#section-7.3.2
		case 17:
			length := binary.BigEndian.Uint16(data[offset+2 : offset+4])
			l.Name = string(data[offset+4 : offset+4+length])
			offset = offset + (length + (4 - (length % 4))) + 4

			logrus.WithFields(logrus.Fields{
				"type":     binary.BigEndian.Uint16(data[offset : offset+2]),
				"length":   length,
				"lsp_name": l.Name,
				"offset":   offset,
				"counter":  counter,
			}).Info("unknown object type")
			continue
		default:
			logrus.WithFields(logrus.Fields{
				"type":   binary.BigEndian.Uint16(data[offset : offset+2]),
				"length": binary.BigEndian.Uint16(data[offset+2 : offset+4]),
			}).Info("unknown object type")
			offset = offset + binary.BigEndian.Uint16(data[offset+2:offset+4]) + 4
		}
	}
	return nil
}

// LSPIPv4Identifiers https://tools.ietf.org/html/rfc8231#section-7.3.1
type LSPIPv4Identifiers struct {
	Type             uint16
	Length           uint16
	LSPID            uint16
	TunnelID         uint16
	SenderAddr       uint32
	ExtendedTunnelID uint32
	EndpointAddr     uint32
}

// https://tools.ietf.org/html/rfc8231#section-7.3.1
func parseLSPIPv4Identifiers(data []byte) (*LSPIPv4Identifiers, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("data len is %d but should be 20", len(data))
	}
	return &LSPIPv4Identifiers{
		Type:             binary.BigEndian.Uint16(data[:2]),
		Length:           binary.BigEndian.Uint16(data[2:4]),
		LSPID:            binary.BigEndian.Uint16(data[8:10]),
		TunnelID:         binary.BigEndian.Uint16(data[10:12]),
		SenderAddr:       binary.BigEndian.Uint32(data[4:8]),
		ExtendedTunnelID: binary.BigEndian.Uint32(data[12:16]),
		EndpointAddr:     binary.BigEndian.Uint32(data[16:20]),
	}, nil
}

// LSPIPv6Identifiers https://tools.ietf.org/html/rfc8231#section-7.3.1
type LSPIPv6Identifiers struct {
	Type             uint16
	Length           uint16
	LSPID            uint16
	TunnelID         uint16
	SenderAddr       *big.Int
	ExtendedTunnelID *big.Int
	EndpointAddr     *big.Int
}

// https://tools.ietf.org/html/rfc8231#section-7.3.1
func parseLSPIPv6Identifiers(data []byte) (*LSPIPv6Identifiers, error) {
	if len(data) < 52 {
		return nil, fmt.Errorf("data len is %d but should be 52", len(data))
	}
	return &LSPIPv6Identifiers{
		Type:             binary.BigEndian.Uint16(data[:2]),
		Length:           binary.BigEndian.Uint16(data[2:4]),
		LSPID:            binary.BigEndian.Uint16(data[20:22]),
		TunnelID:         binary.BigEndian.Uint16(data[22:24]),
		SenderAddr:       new(big.Int).SetBytes(data[4:20]),
		ExtendedTunnelID: new(big.Int).SetBytes(data[24:40]),
		EndpointAddr:     new(big.Int).SetBytes(data[40:56]),
	}, nil
}

func parseLSPAObj(data []byte) error {

	return nil
}
