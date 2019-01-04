package pcep

import (
	"encoding/binary"

	"github.com/sirupsen/logrus"
)

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

// https://tools.ietf.org/html/rfc5440#section-7.17
func parseClose(data []byte) uint8 {
	return uint8(data[3])
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

//HandlePCRpt https://tools.ietf.org/html/rfc8231#section-6.1
func (s Session) HandlePCRpt(data []byte) {
	var offset uint16
	for (len(data) - int(offset)) > 4 {
		coh := parseCommonObjectHeader(data[offset : offset+4])
		switch coh.ObjectClass {
		case 33:
			if coh.ObjectType == 1 {
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
					"srp_id":        srp.SRPIDNumber,
				}).Info("found obj in report msg")
				offset = offset + coh.ObjectLength
				continue
			}
		case 32:
			if coh.ObjectType == 1 {
				lsp, err := parseLSPObj(data[offset+4 : offset+4+coh.ObjectLength])
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
		case 7:
			if coh.ObjectType == 1 {
				eros, err := parseERO(data[offset+4 : offset+coh.ObjectLength])
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
		// case 9:
		// 	if coh.ObjectType == 1 {

		// 	}
		default:
			logrus.WithFields(logrus.Fields{
				"type": coh.ObjectType,
				// "peer":          s.Conn.RemoteAddr().String(),
				"class":         coh.ObjectClass,
				"process_rules": coh.ProcessingRule,
				"length":        coh.ObjectLength,
				"ignore":        coh.Ignore,
				"reserved":      coh.Reservedfield,
			}).Info("found obj in report msg")
			offset = offset + coh.ObjectLength
		}
	}
}
