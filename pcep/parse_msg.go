package pcep

import (
	"encoding/binary"
	"fmt"

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

//HandlePCRpt https://tools.ietf.org/html/rfc8231#section-6.1
func (s Session) HandlePCRpt(data []byte) {
	// fmt.Printf("Int %08b \n", data)
	var (
		offset    uint16
		newOffset uint16
		lsp       LSP
	)
	for (len(data) - int(newOffset)) > 4 {
		offset = newOffset
		coh, err := parseCommonObjectHeader(data[newOffset : newOffset+4])
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": "err",
				"func": "parseERO",
			}).Error(err)
			return
		}
		newOffset = newOffset + coh.ObjectLength
		printCommonObjHdr(coh, "found obj in report msg")
		switch coh.ObjectClass {
		case 5:
			if coh.ObjectType != 1 {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "bandwidth",
				}).Error(fmt.Errorf("unknown obj type %d", coh.ObjectType))
			}
			lsp.BW = binary.BigEndian.Uint32(data[offset+4 : offset+8])
			continue
		case 6:
			if coh.ObjectType != 1 {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseMetric",
				}).Error(fmt.Errorf("unknown obj type %d", coh.ObjectType))
				return
			}
			_, err := parseMetric(data[offset+4 : offset+coh.ObjectLength])
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseERO",
				}).Error(err)
				return
			}
			continue
		case 7:
			if coh.ObjectType != 1 {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseERO",
				}).Error(fmt.Errorf("unknown obj type %d", coh.ObjectType))
				return
			}
			lsp.SREROList, err = parseERO(data[offset+4 : offset+coh.ObjectLength])
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseERO",
				}).Error(err)
				return
			}
			continue
		case 8:
			if coh.ObjectType != 1 {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseMetric",
				}).Error(fmt.Errorf("unknown obj type %d", coh.ObjectType))
				return
			}
			lsp.SRRROList, err = parseRRO(data[offset+4 : offset+coh.ObjectLength])
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseERO",
				}).Error(err)
				return
			}
			continue
		case 9:
			if coh.ObjectType != 1 {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseLSPAObj",
				}).Error(fmt.Errorf("unknown obj type %d", coh.ObjectType))
				return
			}
			err := lsp.parseLSPAObj(data[offset+4 : offset+coh.ObjectLength])
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseERO",
				}).Error(err)
				return
			}
			continue
		case 32:
			if coh.ObjectType != 1 {
				fmt.Printf("Int %08b \n", data[offset:offset+4+coh.ObjectLength])
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseLSPObj",
				}).Error(fmt.Errorf("unknown obj type %d", coh.ObjectType))
				return
			}
			err := lsp.parseLSPObj(data[offset+4 : offset+4+coh.ObjectLength])
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseLSPObj",
				}).Error(err)
				return
			}
			continue
		case 33:
			if coh.ObjectType != 1 {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseSRP",
				}).Error(fmt.Errorf("unknown obj type %d", coh.ObjectType))
				return
			}
			srp := parseSRP(data[offset+4:])
			lsp.SRPID = srp.SRPIDNumber
			continue
		default:
			printCommonObjHdr(coh, "found unknown obj in report msg")
		}
	}
	printAsJSON(lsp)
	logrus.WithFields(logrus.Fields{
		"type": "after",
		"func": "printAsJSON",
	}).Info("new msg")
}

func printCommonObjHdr(coh *CommonObjectHeader, msg string) {
	logrus.WithFields(logrus.Fields{
		"type": coh.ObjectType,
		// "peer":          s.Conn.RemoteAddr().String(),
		"class":         coh.ObjectClass,
		"process_rules": coh.ProcessingRule,
		"length":        coh.ObjectLength,
		"ignore":        coh.Ignore,
		"reserved":      coh.Reservedfield,
	}).Info(msg)
}
