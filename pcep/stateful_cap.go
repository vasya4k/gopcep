package pcep

import (
	"encoding/binary"
	"fmt"
)

// STATEFUL-PCE-CAPABILITY TLV Flag Field
// Registration Procedure(s)
// Standards Action
// Reference
// [RFC8231]
// Note
// Bits are numbered from bit 0 as the most significant bit.
// Available Formats

// CSV
// Value 	Description 	Reference
// 0-20	Unassigned
// 21	PD-LSP-CAPABILITY (PD-bit)	[RFC8934]
// 22	LSP-SCHEDULING-CAPABILITY (B-bit)	[RFC8934]
// 23	P2MP-LSP-INSTANTIATION-CAPABILITY	[RFC8623]
// 24	P2MP-LSP-UPDATE-CAPABILITY	[RFC8623]
// 25	P2MP-CAPABILITY	[RFC8623]
// 26	TRIGGERED-INITIAL-SYNC	[RFC8232]
// 27	DELTA-LSP-SYNC-CAPABILITY	[RFC8232]
// 28	TRIGGERED-RESYNC	[RFC8232]
// 29	LSP-INSTANTIATION-CAPABILITY (I)	[RFC8281]
// 30	INCLUDE-DB-VERSION	[RFC8232]
// 31	LSP-UPDATE-CAPABILITY	[RFC8231]

//StatefulPCECapability  Stateful PCE Capability object
type StatefulPCECapability struct {
	Type                 uint16
	Length               uint16
	Flags                uint32
	TriggeredInitialSync bool
	DeltaLSPSyncCap      bool
	TriggeredResync      bool
	LSPInitCap           bool
	IncludeDBVersion     bool
	UPDFlag              bool
}

// https://tools.ietf.org/html/rfc8231#section-7.1.1
func parseStatefulPCECap(data []byte) (*StatefulPCECapability, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data len is %d but should be 8", len(data))
	}
	UFlag, err := uintToBool(readBits(data[7], 0))
	if err != nil {
		return nil, err
	}
	IncludeDBVersion, err := uintToBool(readBits(data[7], 1))
	if err != nil {
		return nil, err
	}
	LSPInitCap, err := uintToBool(readBits(data[7], 2))
	if err != nil {
		return nil, err
	}
	TriggeredResync, err := uintToBool(readBits(data[7], 3))
	if err != nil {
		return nil, err
	}
	DeltaLSPSyncCap, err := uintToBool(readBits(data[7], 4))
	if err != nil {
		return nil, err
	}
	TriggeredInitialSync, err := uintToBool(readBits(data[7], 5))
	if err != nil {
		return nil, err
	}
	return &StatefulPCECapability{
		Type:                 binary.BigEndian.Uint16(data[:2]),
		Length:               binary.BigEndian.Uint16(data[2:4]),
		Flags:                binary.BigEndian.Uint32(data[4:8]),
		TriggeredInitialSync: TriggeredInitialSync,
		DeltaLSPSyncCap:      DeltaLSPSyncCap,
		TriggeredResync:      TriggeredResync,
		LSPInitCap:           LSPInitCap,
		IncludeDBVersion:     IncludeDBVersion,
		UPDFlag:              UFlag,
	}, nil
}
