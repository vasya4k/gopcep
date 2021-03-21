package pcep

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"

	"github.com/sirupsen/logrus"
)

// InitLSP aaaa
func (s *Session) InitLSP(l *LSP) error {
	sro, err := s.newSRPObject()
	if err != nil {
		return err
	}
	lsp, err := s.newLSPObj(l.Delegate, l.Sync, l.Remove, l.Admin, l.Name)
	if err != nil {
		return err
	}
	ep, err := newEndpointsObj(l.Src, l.Dst)
	if err != nil {
		return err
	}
	ero, err := newERObj(l.EROList)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"type":  "session",
		"event": "create_ero",
	}).Infof("ero %d bin string %08b \n", len(ero), ero)

	lspa, err := newLSPAObject(l.SetupPrio, l.HoldPrio, l.LocalProtect)
	if err != nil {
		return err
	}
	msg := append(sro, lsp...)
	msg = append(msg, ep...)
	msg = append(msg, ero...)
	msg = append(msg, lspa...)

	ch, err := newCommonHeader(12, uint16(len(msg)))
	if err != nil {
		return err
	}
	ch = append(ch, msg...)
	i, err := s.Conn.Write(ch)
	if err != nil {
		log.Println(err)
	}
	logrus.WithFields(logrus.Fields{
		"type":  "session",
		"event": "sent lsp init request",
	}).Infof("sent LSP initiate Request: %d byte", i)
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
	SREROList    []*SREROSub
	SRRROList    []*SRRROSub
	SetupPrio    uint8
	HoldPrio     uint8
	LocalProtect bool
	BW           uint32
	PLSPID       uint32
	LSPID        uint16
	IPv4ID       *LSPIPv4Identifiers
	IPv6ID       *LSPIPv6Identifiers
	SRPID        uint32
	ExcludeAny   uint32
	IncludeAny   uint32
	IncludeAll   uint32
}

//https://tools.ietf.org/html/rfc8231#section-7.3
func (l *LSP) parseLSPObj(data []byte) error {
	var err error
	l.Delegate, err = uintToBool(readBits(data[3], 0))
	if err != nil {
		return err
	}
	l.Sync, err = uintToBool(readBits(data[3], 1))
	if err != nil {
		return err
	}
	l.Remove, err = uintToBool(readBits(data[3], 2))
	if err != nil {
		return err
	}
	l.Admin, err = uintToBool(readBits(data[3], 3))
	if err != nil {
		return err
	}
	//shift right to get rid of d,s,r,a flags
	// then shift left to get rid remaining one bit
	// then shit right again to get the a clean value
	// there is a better solution but i do not have time right now
	l.Oper = ((data[3] >> 4) << 5) >> 5
	l.PLSPID = binary.BigEndian.Uint32(data[:4]) >> 12
	// if l.PLSPID == 0 {
	// 	return fmt.Errorf("PLSPID has a 0 value which mus not be used")
	// }
	l.parseLSPSubObj(data[4:])
	return nil
}

func (l *LSP) parseLSPSubObj(data []byte) error {
	var (
		offset  uint16
		err     error
		counter int
	)
	// +4 is needed because obj header is not included into length
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

// https://tools.ietf.org/html/rfc5440#section-7.11
func (l *LSP) parseLSPAObj(data []byte) error {
	if len(data) < 16 {
		return fmt.Errorf("data len is %d but should be 16", len(data))
	}
	l.ExcludeAny = binary.BigEndian.Uint32(data[:4])
	l.IncludeAny = binary.BigEndian.Uint32(data[4:8])
	l.IncludeAll = binary.BigEndian.Uint32(data[8:12])
	l.SetupPrio = data[12]
	l.HoldPrio = data[13]
	var err error
	l.LocalProtect, err = uintToBool(readBits(data[14], 0))
	if err != nil {
		return err
	}
	return nil
}

// LSPMetric s
type LSPMetric struct {
	CFlaf  bool
	BFlag  bool
	Metric float32
}

func parseMetric(data []byte) (*LSPMetric, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data len is %d but should be 8", len(data))
	}
	var (
		m   LSPMetric
		err error
	)
	m.CFlaf, err = uintToBool(readBits(data[2], 1))
	if err != nil {
		return nil, err
	}
	m.BFlag, err = uintToBool(readBits(data[2], 0))
	if err != nil {
		return nil, err
	}
	u := binary.BigEndian.Uint32(data[4:8])
	m.Metric = math.Float32frombits(u)
	return &m, nil
}

//SRPObject https://tools.ietf.org/html/rfc8231#section-7.2
type SRPObject struct {
	Flags       uint32
	SRPIDNumber uint32
}

// https://tools.ietf.org/html/rfc8231#section-7.2
func parseSRP(data []byte) *SRPObject {
	return &SRPObject{
		Flags:       binary.BigEndian.Uint32(data[0:4]),
		SRPIDNumber: binary.BigEndian.Uint32(data[4:8]),
	}
}

//PathSetupType https://tools.ietf.org/html/rfc8408#section-4
type PathSetupType struct {
	Type   uint16
	Length uint16
	PST    uint8
}

// https://tools.ietf.org/html/rfc8408#section-4
func parsePathSetupType(data []byte) *PathSetupType {
	return &PathSetupType{
		Type:   binary.BigEndian.Uint16(data[:2]),
		Length: binary.BigEndian.Uint16(data[2:4]),
		PST:    data[7],
	}

}
