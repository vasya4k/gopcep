package pcep

import (
	"encoding/binary"
	"fmt"
)

//CommonHeader to store PCEP CommonHeader
type CommonHeader struct {
	Version       uint8
	Flags         uint8
	MessageType   uint8
	MessageLength uint16
}

// https://tools.ietf.org/html/rfc5440#section-6.1
func parseCommonHeader(data []byte) (*CommonHeader, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data len is %d but should be 4", len(data))
	}
	h := &CommonHeader{
		Version:       data[0] >> 5,
		Flags:         data[0] & (32 - 1),
		MessageType:   data[1],
		MessageLength: binary.BigEndian.Uint16(data[2:4]),
	}
	if h.Version != 1 {
		return nil, fmt.Errorf("unknown version %d but must be 1", h.Version)
	}
	return h, nil
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

// https://tools.ietf.org/html/rfc5440#section-7.2
func parseCommonObjectHeader(data []byte) (*CommonObjectHeader, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data len is %d but should be 4", len(data))
	}
	obj := &CommonObjectHeader{
		ObjectClass:  data[0],
		ObjectType:   data[1] >> 4,
		ObjectLength: binary.BigEndian.Uint16(data[2:4]),
	}
	return obj, nil
}
