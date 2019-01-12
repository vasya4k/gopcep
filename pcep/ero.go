package pcep

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

// https://tools.ietf.org/html/draft-ietf-pce-segment-routing-14#section-5.3.1
func parseERO(data []byte) ([]*SREROSub, error) {
	eros := make([]*SREROSub, 0)
	var offset int
	for (len(data) - offset) > 4 {
		var (
			e   SREROSub
			err error
		)
		e.LooseHop, err = uintToBool(uint(data[offset]) >> 7)
		if err != nil {
			return nil, err
		}
		// now set loose hop bit to zero so we can determine type
		data[offset] |= (0 << 7)
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
			err = parseNAI(data[offset+4:], &e)
			if err != nil {
				return nil, err
			}
		}
		sid := binary.BigEndian.Uint32(data[offset+4 : offset+8])
		if e.MBit {
			sid = sid >> 12
		}
		err = parseNAI(data[offset+8:], &e)
		if err != nil {
			return nil, err
		}
		eros = append(eros, &e)
		offset = offset + int(data[offset+1])
	}
	return eros, nil
}

func parseNAI(data []byte, ero *SREROSub) error {
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

//EROSub str
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
