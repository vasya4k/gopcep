package pcep

import (
	"encoding/binary"
	"fmt"

	"github.com/sirupsen/logrus"
)

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
	if len(data) < 8 {
		return nil, fmt.Errorf("data len is %d but should be 8", len(data))
	}
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
