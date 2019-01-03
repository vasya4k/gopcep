package pcep

import (
	"encoding/binary"
	"errors"
	"net"

	"github.com/sirupsen/logrus"
)

//CommonHeader to store PCEP CommonHeader
type CommonHeader struct {
	Version       uint8
	Flags         uint8
	MessageType   uint8
	MessageLength uint16
}

// https://tools.ietf.org/html/rfc5440#section-6.1
func parseCommonHeader(data []byte) *CommonHeader {
	header := &CommonHeader{
		Version:       data[0] >> 5,
		Flags:         data[0] & (32 - 1),
		MessageType:   data[1],
		MessageLength: binary.BigEndian.Uint16(data[2:4]),
	}
	return header
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
func parseCommonObjectHeader(data []byte) *CommonObjectHeader {
	obj := &CommonObjectHeader{
		ObjectClass:  data[0],
		ObjectType:   data[1] >> 4,
		ObjectLength: binary.BigEndian.Uint16(data[2:4]),
	}
	return obj
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

//LSPObj https://tools.ietf.org/html/rfc8231#section-7.3
type LSPObj struct {
	Delegate bool
	Sync     bool
	Remove   bool
	Admin    bool
	Oper     uint8
	PLSPID   uint32
}

//https://tools.ietf.org/html/rfc8231#section-7.3
func parseLSPObj(data []byte) (*LSPObj, error) {
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
	return &LSPObj{
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
	}, nil
}

// LSPIdentifiers https://tools.ietf.org/html/rfc8231#section-7.3.1
type LSPIdentifiers struct {
	Type             uint16
	Length           uint16
	LSPID            uint16
	TunnelID         uint16
	SenderAddr       uint32
	ExtendedTunnelID uint32
	EndpointAddr     uint32
}

// https://tools.ietf.org/html/rfc8231#section-7.3.1
func parseLSPIdentifiers(data []byte) *LSPIdentifiers {
	return &LSPIdentifiers{
		Type:             binary.BigEndian.Uint16(data[:2]),
		Length:           binary.BigEndian.Uint16(data[2:4]),
		LSPID:            binary.BigEndian.Uint16(data[8:10]),
		TunnelID:         binary.BigEndian.Uint16(data[10:12]),
		SenderAddr:       binary.BigEndian.Uint32(data[4:8]),
		ExtendedTunnelID: binary.BigEndian.Uint32(data[12:16]),
		EndpointAddr:     binary.BigEndian.Uint32(data[16:20]),
	}
}

// https://tools.ietf.org/html/rfc5440#section-7.17
func parseClose(data []byte) uint8 {
	return uint8(data[3])
}

//ErrObj https://tools.ietf.org/html/rfc5440#section-7.15
type ErrObj struct {
	Reserved    uint8
	Flags       uint8
	ErrType     uint8
	ErrValue    uint8
	ErrValueStr string
}

func parseErrObj(data []byte) (*ErrObj, error) {
	coh := parseCommonObjectHeader(data[:4])
	if coh.ObjectClass != 13 {
		return nil, errors.New("Object Class is not 13 ")
	}
	et := [][]string{
		1: []string{
			0: "PCEP session establishment failure undefined",
			1: "PCEP session establishment failure reception of an invalid Open message or a non Open message. ",
			2: "PCEP session establishment failure no Open message received before the expiration of the OpenWait timer ",
			3: "PCEP session establishment failure unacceptable and non-negotiable session characteristics ",
			4: "PCEP session establishment failure unacceptable but negotiable session characteristics ",
			5: "PCEP session establishment failure reception of a second Open message with still unacceptable session characteristics",
			6: "PCEP session establishment failure reception of a PCErr message proposing unacceptable session characteristics ",
			7: "PCEP session establishment failure No Keepalive or PCErr message received before the expiration of the KeepWait timer",
		},
		2: []string{
			0: "Capability not supported",
		},
		3: []string{
			1: "Unknown Object Unrecognized object class",
			2: "Unknown Object Unrecognized object type",
		},
		4: []string{
			1: "Not supported object  class",
			2: "Not supported object  type",
		},
		5: []string{
			1: "Policy violation C bit of the METRIC object set (request rejected)",
			2: "Policy violation O bit of the RP object set (request rejected)",
		},
		6: []string{
			1:  "Mandatory Object missing RP object missing",
			2:  "Mandatory Object missing RRO object missing for a reoptimization request (R bit of the RP object set) when bandwidth is not equal to 0.",
			3:  "Mandatory Object missing END-POINTS object missing",
			8:  "Mandatory Object missing LSP object missing",
			9:  "Mandatory Object missing ERO object missing",
			10: "Mandatory Object missing SRP object missing",
			11: "Mandatory Object missing LSP-IDENTIFIERS TLV missing",
		},
		7: []string{
			0: "Synchronized path computation request missing",
		},
		8: []string{
			0: "Unknown request reference",
		},
		9: []string{
			0: "Attempt to establish a second PCEP session",
		},
		10: []string{
			1:  "Reception of an invalid object reception of an object with P flag not set although the P flag must be set according to this specification.",
			2:  "Reception of an invalid object Bad label value ",
			3:  "Reception of an invalid object Unsupported number of SR-ERO subobjects",
			4:  "Reception of an invalid object Bad label format ",
			5:  "Reception of an invalid object ERO mixes SR-ERO  subobjects with other subobject types",
			6:  "Reception of an invalid object Both SID and NAI are absent in SR-ERO subobject",
			7:  "Reception of an invalid object Both SID and NAI are absent in SR-RRO subobject",
			8:  "Reception of an invalid object SYMBOLIC-PATH-NAME TLV missing",
			9:  "Reception of an invalid object MSD exceeds the default for the PCEP session",
			10: "Reception of an invalid object RRO mixes SR-RRO subobjects with other subobject types",
			11: "Reception of an invalid object Malformed object",
		},
		19: []string{
			1:  "Invalid Operation Attempted LSP Update Request for a non-delegated  LSP.  The PCEP-ERROR object is followed by the LSP object that identifies the LSP.",
			2:  "Invalid Operation Attempted LSP Update Request if the stateful PCE  capability was not advertised.",
			3:  "Invalid Operation Attempted LSP Update Request for an LSP identified  by an unknown PLSP-ID.",
			5:  "Invalid Operation Attempted LSP State Report if stateful PCE  capability was not advertised.",
			6:  "Invalid Operation Attempted LSP  PCE-initiated LSP limit reached",
			7:  "Invalid Operation Attempted LSP  Delegation for PCE-initiated LSP cannot be revoked",
			8:  "Invalid Operation Attempted LSP  Non-zero PLSP-ID in LSP Initiate Request",
			9:  "Invalid Operation Attempted LSP  LSP is not PCE initiated",
			10: "Invalid Operation Attempted LSP PCE-initiated operation-frequency limit reached",
		},
		20: []string{
			1: "A PCE indicates to a PCC that it cannot process (an otherwise valid) LSP State Report.  The PCEP-ERROR object is followed by the LSP object that identifies the LSP.",
			5: "A PCC indicates to a PCE that it cannot complete the State Synchronization.",
		},
		21: []string{
			0: "Invalid traffic engineering path setup type Unassigned                   RFC 8408",
			1: "Invalid traffic engineering path setup type Unsupported path setup type  RFC 8408",
			2: "Invalid traffic engineering path setup type Mismatched path setup type   RFC 8408",
		},
		23: []string{
			1: "Bad parameter value SYMBOLIC-PATH-NAME in use",
			2: "Bad parameter value Speaker identity included for an LSP that is not PCE initiated ",
		},
		24: []string{
			1: "LSP instantiation error Unacceptable instantiation parameters	",
			2: "LSP instantiation error Internal error",
			3: "LSP instantiation error Signaling error",
		},
		// pcep_obj_trace: ERROR object: type: 24, value: 1
	}
	return &ErrObj{
		Reserved:    data[4],
		Flags:       data[5],
		ErrType:     data[6],
		ErrValue:    data[7],
		ErrValueStr: et[data[6]][data[7]],
	}, nil
}

//https://tools.ietf.org/html/rfc5440#section-7.15
func (s Session) handleErrObj(data []byte) {
	var offset uint16

	for (len(data) - int(offset)) > 4 {

		coh := parseCommonObjectHeader(data[offset : offset+4])

		if coh.ObjectClass != 13 {
			offset = coh.ObjectLength
			logrus.WithFields(logrus.Fields{
				"type":          coh.ObjectType,
				"peer":          s.Conn.RemoteAddr().String(),
				"class":         coh.ObjectClass,
				"process_rules": coh.ProcessingRule,
				"length":        coh.ObjectLength,
				"ignore":        coh.Ignore,
				"reserved":      coh.Reservedfield,
			}).Info("found obj in err msg")
			continue
		}
		errObj, err := parseErrObj(data[offset:])
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": "err",
				"func": "parseErrObj",
			}).Error(err)
		} else {
			logrus.WithFields(logrus.Fields{
				"type":        "err",
				"peer":        s.Conn.RemoteAddr().String(),
				"reserved":    errObj.Reserved,
				"flags":       errObj.Flags,
				"errtype":     errObj.ErrType,
				"errvalue":    errObj.ErrValue,
				"errvaluestr": errObj.ErrValueStr,
			}).Error("new err msg")
		}
		offset = offset + coh.ObjectLength
	}
}

// https://tools.ietf.org/html/rfc8231#section-6.1
func (s Session) HandlePCRpt(data []byte) {
	var offset uint16

	for (len(data) - int(offset)) > 4 {

		coh := parseCommonObjectHeader(data[offset : offset+4])

		if coh.ObjectClass == 33 && coh.ObjectType == 1 {

			srp := parseSRP(data[offset+4:])
			logrus.WithFields(logrus.Fields{
				"type": coh.ObjectType,
				// "peer":          s.Conn.RemoteAddr().String(),
				"class":         coh.ObjectClass,
				"process_rules": coh.ProcessingRule,
				"length":        coh.ObjectLength,
				"ignore":        coh.Ignore,
				"reserved":      coh.Reservedfield,
				"flags":         srp.Flags,
				"id":            srp.SRPIDNumber,
			}).Info("found obj in report msg")
			offset = offset + coh.ObjectLength
			continue
		}
		if coh.ObjectClass == 32 && coh.ObjectType == 1 {

			lsp, err := parseLSPObj(data[offset+4:])
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseErrObj",
				}).Error(err)
				offset = offset + coh.ObjectLength
				continue
			}
			logrus.WithFields(logrus.Fields{
				"type": coh.ObjectType,
				// "peer":          s.Conn.RemoteAddr().String(),
				"class":         coh.ObjectClass,
				"process_rules": coh.ProcessingRule,
				"length":        coh.ObjectLength,
				"ignore":        coh.Ignore,
				"reserved":      coh.Reservedfield,
				"admin":         lsp.Admin,
				"delegate":      lsp.Delegate,
				"operational":   lsp.Oper,
				"plsp_id":       lsp.PLSPID,
				"remove":        lsp.Remove,
				"sync":          lsp.Sync,
			}).Info("found obj in report msg")
			offset = offset + coh.ObjectLength
			continue
		}
		if coh.ObjectClass == 7 && coh.ObjectType == 1 {
			eros, err := parseERO(data[offset+4 : coh.ObjectLength])
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseErrObj",
				}).Error(err)
				offset = offset + coh.ObjectLength
				continue
			}
			printAsJSON(eros)
			offset = offset + coh.ObjectLength
			continue
		}
		offset = offset + coh.ObjectLength
	}
}

func parseERO(data []byte) ([]*SREROSub, error) {
	eros := make([]*SREROSub, 0)
	var offset uint8
	for (len(data) - int(offset)) > 4 {
		var (
			e   SREROSub
			err error
		)
		e.LooseHop, err = uintToBool(uint(data[0]) >> 7)
		if err != nil {
			return nil, err
		}
		// checking obj type
		data[0] |= (1 << 7)
		if data[0] != 36 {
			return nil, errors.New("wrong ero type")
		}
		e.NT = data[2] >> 4
		e.NoNAI, err = uintToBool(readBits(data[3], 3))
		if err != nil {
			return nil, err
		}
		e.NoSID, err = uintToBool(readBits(data[3], 2))
		if err != nil {
			return nil, err
		}
		e.CBit, err = uintToBool(readBits(data[3], 1))
		if err != nil {
			return nil, err
		}
		e.MBit, err = uintToBool(readBits(data[3], 0))
		if err != nil {
			return nil, err
		}
		if e.NoSID {
			err = parseNAI(data[4:], &e)
			if err != nil {
				return nil, err
			}
		}
		sid := binary.BigEndian.Uint32(data[4:8])
		if e.MBit {
			sid = sid >> 12
		}
		err = parseNAI(data[8:], &e)
		if err != nil {
			return nil, err
		}
		eros = append(eros, &e)
		offset = offset + data[1]
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
