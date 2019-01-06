package pcep

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

//SRRROSub str
type SRRROSub struct {
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

// https://tools.ietf.org/html/draft-ietf-pce-segment-routing-14#section-5.4
func parseRRO(data []byte) ([]*SRRROSub, error) {
	// fmt.Printf("After Int %08b \n", data)
	eros := make([]*SRRROSub, 0)
	var offset int
	for (len(data) - offset) > 4 {
		var (
			e   SRRROSub
			err error
		)
		if data[offset] != 36 {
			return nil, fmt.Errorf("wrong ero type %d", uint8(data[offset]))
		}
		e.NT = data[offset+2] >> 4
		e.NoNAI, err = uintToBool(readBits(data[offset+3], 3))
		if err != nil {
			return nil, err
		}
		e.NoSID, err = uintToBool(readBits(data[offset+3], 2))
		if err != nil {
			return nil, err
		}
		e.CBit, err = uintToBool(readBits(data[offset+3], 1))
		if err != nil {
			return nil, err
		}
		e.MBit, err = uintToBool(readBits(data[offset+3], 0))
		if err != nil {
			return nil, err
		}
		if e.NoSID {
			err = parseRRONAI(data[offset+4:], &e)
			if err != nil {
				return nil, err
			}
		}
		sid := binary.BigEndian.Uint32(data[offset+4 : offset+8])
		if e.MBit {
			sid = sid >> 12
		}
		err = parseRRONAI(data[offset+8:], &e)
		if err != nil {
			return nil, err
		}
		eros = append(eros, &e)
		offset = offset + int(data[offset+1])
	}
	return eros, nil
}

func parseRRONAI(data []byte, ero *SRRROSub) error {
	switch ero.NT {
	case 1:
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, binary.BigEndian.Uint32(data[:4]))
		ero.IPv4NodeID = ip.String()
	case 3:
		localIP := make(net.IP, 4)
		binary.BigEndian.PutUint32(localIP, binary.BigEndian.Uint32(data[:4]))
		ero.IPv4Adjacency = make([]string, 2)
		ero.IPv4Adjacency[0] = localIP.String()
		remoteIP := make(net.IP, 4)
		binary.BigEndian.PutUint32(remoteIP, binary.BigEndian.Uint32(data[4:8]))
		ero.IPv4Adjacency[1] = remoteIP.String()
	default:
		return errors.New("NAI type not implemented yet")
	}
	return nil
}
