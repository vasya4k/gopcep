package pcep

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

//ErrObj https://tools.ietf.org/html/rfc5440#section-7.15
type ErrObj struct {
	Reserved    uint8
	Flags       uint8
	ErrType     uint8
	ErrValue    uint8
	ErrValueStr string
}

func parseErrObj(data []byte) (*ErrObj, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data is too short len: %d but should be at least 8", len(data))
	}
	coh, err := parseCommonObjectHeader(data[:4])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type": "err",
			"func": "parseERO",
		}).Error(err)
		return nil, err
	}
	if coh.ObjectClass != 13 {
		return nil, errors.New("Object Class is not 13 ")
	}
	et := map[int]map[int]string{
		1: {
			0: "PCEP session establishment failure undefined",
			1: "PCEP session establishment failure reception of an invalid Open message or a non Open message. ",
			2: "PCEP session establishment failure no Open message received before the expiration of the OpenWait timer ",
			3: "PCEP session establishment failure unacceptable and non-negotiable session characteristics ",
			4: "PCEP session establishment failure unacceptable but negotiable session characteristics ",
			5: "PCEP session establishment failure reception of a second Open message with still unacceptable session characteristics",
			6: "PCEP session establishment failure reception of a PCErr message proposing unacceptable session characteristics ",
			7: "PCEP session establishment failure No Keepalive or PCErr message received before the expiration of the KeepWait timer",
		},
		2: {
			0: "Capability not supported",
		},
		3: {
			1: "Unknown Object Unrecognized object class",
			2: "Unknown Object Unrecognized object type",
		},
		4: {
			1: "Not supported object  class",
			2: "Not supported object  type",
		},
		5: {
			1: "Policy violation C bit of the METRIC object set (request rejected)",
			2: "Policy violation O bit of the RP object set (request rejected)",
		},
		6: {
			1:  "Mandatory Object missing RP object missing",
			2:  "Mandatory Object missing RRO object missing for a reoptimization request (R bit of the RP object set) when bandwidth is not equal to 0.",
			3:  "Mandatory Object missing END-POINTS object missing",
			8:  "Mandatory Object missing LSP object missing",
			9:  "Mandatory Object missing ERO object missing",
			10: "Mandatory Object missing SRP object missing",
			11: "Mandatory Object missing LSP-IDENTIFIERS TLV missing",
		},
		7: {
			0: "Synchronized path computation request missing",
		},
		8: {
			0: "Unknown request reference",
		},
		9: {
			0: "Attempt to establish a second PCEP session",
		},
		10: {
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
		19: {
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
		20: {
			1: "A PCE indicates to a PCC that it cannot process (an otherwise valid) LSP State Report.  The PCEP-ERROR object is followed by the LSP object that identifies the LSP.",
			5: "A PCC indicates to a PCE that it cannot complete the State Synchronization.",
		},
		21: {
			0: "Invalid traffic engineering path setup type Unassigned                   RFC 8408",
			1: "Invalid traffic engineering path setup type Unsupported path setup type  RFC 8408",
			2: "Invalid traffic engineering path setup type Mismatched path setup type   RFC 8408",
		},
		23: {
			1: "Bad parameter value SYMBOLIC-PATH-NAME in use",
			2: "Bad parameter value Speaker identity included for an LSP that is not PCE initiated ",
		},
		24: {
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
		ErrValueStr: et[int(data[6])][int(data[7])],
	}, nil
}

//https://tools.ietf.org/html/rfc5440#section-7.15
func (s *Session) handleErrObj(data []byte) {
	var offset uint16
	for (len(data) - int(offset)) > 4 {
		coh, err := parseCommonObjectHeader(data[offset : offset+4])
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": "err",
				"func": "parseErrObj",
			}).Error(err)
			return
		}
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
