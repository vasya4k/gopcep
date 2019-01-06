package pcep

import (
	"encoding/binary"
	"fmt"

	"github.com/sirupsen/logrus"
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

//StatefulPCECapability  rfc8231#section-7.1.1
type StatefulPCECapability struct {
	Type   uint16
	Length uint16
	Flags  uint32
	UFlag  bool
}

// https://tools.ietf.org/html/rfc8231#section-7.1.1
func parseStatefulPCECap(data []byte) (*StatefulPCECapability, error) {
	UFlag, err := uintToBool(readBits(data[7], 0))
	if err != nil {
		return nil, err
	}
	return &StatefulPCECapability{
		Type:   binary.BigEndian.Uint16(data[:2]),
		Length: binary.BigEndian.Uint16(data[2:4]),
		Flags:  binary.BigEndian.Uint32(data[4:8]),
		UFlag:  UFlag,
	}, nil
}

//SRPCECap  https://tools.ietf.org/html/draft-ietf-pce-segment-routing-14#section-5.1
type SRPCECap struct {
	Type       uint16
	Length     uint16
	Reserved   uint16
	NAIToSID   bool
	NoMSDLimit bool
	MSD        uint8
}

// https://tools.ietf.org/html/draft-ietf-pce-segment-routing-14#section-5
func parseSRCap(data []byte) (*SRPCECap, error) {
	NAIToSID, err := uintToBool(readBits(data[6], 6))
	if err != nil {
		return nil, err
	}
	NoMSDLimit, err := uintToBool(readBits(data[6], 7))
	if err != nil {
		return nil, err
	}
	srCap := &SRPCECap{
		Type:       binary.BigEndian.Uint16(data[:2]),
		Length:     binary.BigEndian.Uint16(data[2:4]),
		Reserved:   binary.BigEndian.Uint16(data[4:6]),
		NAIToSID:   NAIToSID,
		NoMSDLimit: NoMSDLimit,
		MSD:        data[7],
	}
	logrus.WithFields(logrus.Fields{
		"type":       srCap.Type,
		"length":     srCap.Length,
		"reserved":   srCap.Reserved,
		"naitosid":   srCap.NAIToSID,
		"nomsdlimit": srCap.NoMSDLimit,
		"msd":        srCap.MSD,
	}).Info("parsed sr capability obj")
	return srCap, nil
}
