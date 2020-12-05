package pcep

import (
	"fmt"
)

//OpenObject to store PCEP OPEN Object
type OpenObject struct {
	Version   uint8
	Flags     uint8
	Keepalive uint8
	DeadTimer uint8
	SID       uint8
}

// https://tools.ietf.org/html/rfc5440#section-7.3
func parseOpenObject(data []byte) (*OpenObject, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data len is %d but should be 4", len(data))
	}
	open := &OpenObject{
		Version:   data[0] >> 5,
		Flags:     data[0] & (32 - 1),
		Keepalive: data[1],
		DeadTimer: data[2],
		SID:       data[3],
	}
	if open.Version != 1 {
		return nil, fmt.Errorf("unknown version %d but must be 1", open.Version)
	}
	return open, nil
}
