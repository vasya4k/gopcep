package pcep

import (
	"encoding/binary"

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
func parseOpenObject(data []byte) *OpenObject {
	open := &OpenObject{
		Version:   data[0] >> 5,
		Flags:     data[0] & (32 - 1),
		Keepalive: data[1],
		DeadTimer: data[2],
		SID:       data[3],
	}
	return open
}

//StatefulPCECapability  rfc8231#section-7.1.1
type StatefulPCECapability struct {
	Type   uint16
	Length uint16
	Flags  uint32
	UFlag  bool
}

// https://tools.ietf.org/html/rfc8231#section-7.1.1
func parseStatefulPCECap(data []byte) *StatefulPCECapability {
	UFlag, err := uintToBool(readBits(data[7], 0))
	if err != nil {
		return nil
	}
	sCap := &StatefulPCECapability{
		Type:   binary.BigEndian.Uint16(data[:2]),
		Length: binary.BigEndian.Uint16(data[2:4]),
		Flags:  binary.BigEndian.Uint32(data[4:8]),
		UFlag:  UFlag,
	}
	logrus.WithFields(logrus.Fields{
		"type":   sCap.Type,
		"length": sCap.Length,
		"flags":  sCap.Flags,
		"uflag":  sCap.UFlag,
	}).Info("parsed stateful capability obj")
	return sCap
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
func parseSRCap(data []byte) *SRPCECap {
	NAIToSID, err := uintToBool(readBits(data[6], 6))
	if err != nil {
		return nil
	}
	NoMSDLimit, err := uintToBool(readBits(data[6], 7))
	if err != nil {
		return nil
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
	return srCap
}
