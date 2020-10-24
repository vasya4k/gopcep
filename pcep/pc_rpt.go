package pcep

import (
	"encoding/binary"
	"fmt"

	"github.com/sirupsen/logrus"
)

//HandlePCRpt https://tools.ietf.org/html/rfc8231#section-6.1
// A Path Computation LSP State Report message
func (s Session) HandlePCRpt(data []byte) {
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
		// printCommonObjHdr(coh, "found obj in report msg")
		switch coh.ObjectClass {
		case 5:
			if coh.ObjectType != 1 {
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "bandwidth",
				}).Error(fmt.Errorf("unknown obj type %d", coh.ObjectType))
				return

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
				logrus.WithFields(logrus.Fields{
					"type": "err",
					"func": "parseLSPObj",
				}).Error(fmt.Errorf("unknown obj AAAA type %d", coh.ObjectType))
				return
			}
			err := lsp.parseLSPObj(data[offset+4 : offset+coh.ObjectLength])
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
	if lsp.PLSPID == 0 && lsp.Name == "" {
		logrus.WithFields(logrus.Fields{
			"event": "empty lsp name and zero plspid in pcrpt",
		}).Info("found lsp with no id skipping")
		return
	}
	printAsJSON(lsp)
	logrus.WithFields(logrus.Fields{
		"type": "after",
		"func": "printAsJSON",
	}).Info("new msg")
	s.saveUpdLSP(&lsp)
}
