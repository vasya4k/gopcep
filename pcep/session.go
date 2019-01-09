package pcep

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

//Session holds everything for a PCEP session
type Session struct {
	ID           uint8
	MsgCount     int64
	State        int
	Conn         net.Conn
	RemoteOK     bool
	LocalOK      bool
	Keepalive    uint8
	DeadTimer    uint8
	PLSPIDToName map[uint32]string
	LSPs         map[string]*LSP
	IDCounter    uint32
	SRPID        uint32
	StopKA       chan struct{} `json:"-"`
	SRCap        *SRPCECap
	StatefulCap  *StatefulPCECapability
	Open         *OpenObject
}

func (s *Session) saveUpdLSP(lsp *LSP) {
	if lsp.Name != "" {
		s.LSPs[lsp.Name] = lsp
		s.PLSPIDToName[lsp.PLSPID] = lsp.Name

	} else {
		s.LSPs[s.PLSPIDToName[lsp.PLSPID]] = lsp
	}
}

func (s *Session) delLSP(lsp *LSP) {
	if lsp.Name != "" {
		delete(s.LSPs, lsp.Name)
		delete(s.PLSPIDToName, lsp.PLSPID)

	} else {
		delete(s.LSPs, s.PLSPIDToName[lsp.PLSPID])
		delete(s.PLSPIDToName, lsp.PLSPID)
	}
}

func (s *Session) getLSP(lsp *LSP) *LSP {
	if lsp.Name != "" {
		return s.LSPs[lsp.Name]
	}
	return s.LSPs[s.PLSPIDToName[lsp.PLSPID]]
}

//RcvSessionOpen recive msg handler
func (s *Session) RcvSessionOpen(data []byte) {
	h, err := parseCommonObjectHeader(data)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type": "err",
			"func": "parseCommonObjectHeader",
		}).Error(err)
		return
	}
	if h.ObjectClass != 1 && h.ObjectType != 1 {
		log.Printf("Remote IP: %s, object class and object type do not mathc OPEN msg RFC defenitions", s.Conn.RemoteAddr())
		return
	}
	s.Open, err = parseOpenObject(data[4:8])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type": "err",
			"func": "parseOpenObject",
		}).Error(err)
		return
	}
	s.ID = s.Open.SID
	s.StatefulCap, err = parseStatefulPCECap(data[8:16])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type": "err",
			"func": "parseStatefulPCECap",
		}).Error(err)
		return
	}
	s.SRCap, err = parseSRCap(data[16:24])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type": "err",
			"func": "parseSRCap",
		}).Error(err)
		return
	}
	s.SendSessionOpen()
}

//SendKeepAlive start sending keep alive msgs
func (s *Session) SendKeepAlive() {
	commH := []byte{
		0: 32,
		1: 2,
		2: 0,
		3: 4,
	}
	var firstSent bool
	for {
		if firstSent {
			time.Sleep(time.Second * time.Duration(s.Keepalive))
		}
		select {
		case <-s.StopKA: // triggered when the stop channel is closed
			logrus.WithFields(logrus.Fields{
				"peer": s.Conn.RemoteAddr().String(),
			}).Info("stopping keepalive")
			return
		default:

			_, err := s.Conn.Write(commH)
			if err != nil {
				log.Println("SendKeepAlive", err)
				return
			}
			// log.Printf("keep alive sent %d bytes", i)
			firstSent = true
		}
	}

}

//Configure aa
func (s *Session) Configure() {
	s.PLSPIDToName = make(map[uint32]string)
	s.LSPs = make(map[string]*LSP)
}

// HandleNewMsg handles incoming data
func (s *Session) HandleNewMsg(data []byte) {
	var (
		offset    uint16
		newOffset uint16
	)
	for (len(data) - int(newOffset)) > 4 {
		offset = newOffset
		ch, err := parseCommonHeader(data[newOffset : newOffset+4])
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": "err",
				"func": "parseCommonHeader",
			}).Error(err)
			return
		}
		newOffset = newOffset + ch.MessageLength
		switch {
		case ch.MessageType == 1:
			logrus.WithFields(logrus.Fields{
				"type": "open",
				"peer": s.Conn.RemoteAddr().String(),
			}).Info("new msg")
			go s.RcvSessionOpen(data[offset+4:])
		case ch.MessageType == 2:
			logrus.WithFields(logrus.Fields{
				"type": "keepalive",
				// "peer": s.Conn.RemoteAddr().String(),
			}).Info("new msg")
			// if s.State == 2 {
			// 	if strings.Split(s.Conn.RemoteAddr().String(), ":")[0] == "10.0.0.10" {
			// 		logrus.WithFields(logrus.Fields{
			// 			"type": "before",
			// 			"func": "InitSRLSP",
			// 		}).Info("new msg")
			// 		lsp := &SRLSP{}
			// 		err := s.InitSRLSP(lsp)
			// 		if err != nil {
			// 			fmt.Println(err)
			// 		}
			// 		logrus.WithFields(logrus.Fields{
			// 			"type": "after",
			// 			"func": "InitSRLSP",
			// 		}).Info("new msg")
			// 		s.State = 3
			// 	}
			// }
		case ch.MessageType == 3:
			log.Println("recv path computation request ")
		case ch.MessageType == 4:
			log.Println("recv path computation reply ")
		case ch.MessageType == 5:
			log.Println("recv notification ")
		case ch.MessageType == 6:
			fmt.Printf("len %d Whole ERR MSG: %08b \n", ch.MessageLength, data[:ch.MessageLength])
			s.handleErrObj(data[4:])

		case ch.MessageType == 7:
			logrus.WithFields(logrus.Fields{
				"type": "close",
				"peer": s.Conn.RemoteAddr().String(),
				"msg":  parseClose(data[offset+8 : offset+12]),
			}).Info("new msg")
			err := s.Conn.Close()
			if err != nil {
				log.Println(err)
			}
		case ch.MessageType == 10:
			logrus.WithFields(logrus.Fields{
				"type": "path computation lsp state report",
				// "peer": s.Conn.RemoteAddr().String(),
				"len": ch.MessageLength,
			}).Info("new msg")
			// err := ioutil.WriteFile("/home/egorz/go/src/gopcep/dat1", data[:ch.MessageLength], 0644)
			// if err != nil {
			// 	log.Println(err)
			// }
			s.HandlePCRpt(data[offset+4 : offset+ch.MessageLength])
			// fmt.Printf("%s %+v\n", "LSP", lsp)
			// fmt.Printf("%s %+v\n", "LSPIdentifiers", parseLSPIdentifiers(data[32:52]))
		case ch.MessageType == 11:
			log.Println("Path Computation LSP Update Request")
		case ch.MessageType == 12:
			log.Println("LSP Initiate Request")

		default:
			log.Println("Unknown msg recived")
		}
	}
	printAsJSON(s)
}

//SendSessionOpen send OPNE msg handler
func (s *Session) SendSessionOpen() {
	//[00100000 00000001 00000000 00011100]
	commH := []byte{
		0: 32,
		1: 1,
		2: 0,
		3: 28,
	}
	// 00000001 00010000 00000000 00011000
	commObjH := []byte{
		0: 1,
		1: 16,
		2: 0,
		3: 24,
	}
	// 00100000 00011110 01111000 00000001
	open := []byte{
		0: 32,
		1: 30,
		2: 120,
		3: s.ID,
	}
	stCap := []byte{
		0: 0,
		1: 16,
		2: 0,
		3: 4,
		4: 0,
		5: 0,
		6: 0,
		7: 5,
	}
	srCap := []byte{
		0: 0,
		1: 26,
		2: 0,
		3: 4,
		4: 0,
		5: 0,
		6: 0,
		7: 5,
	}
	packet := append(commH, commObjH...)
	packet = append(packet, open...)
	packet = append(packet, stCap...)
	packet = append(packet, srCap...)
	// fmt.Printf("Sending open %08b \n", packet)
	i, err := s.Conn.Write(packet)
	if err != nil {
		log.Println(err)
	}
	logrus.WithFields(logrus.Fields{
		"type":  "info",
		"event": "open",
	}).Info(fmt.Sprintf("Sent Open: %d byte", i))
	s.State = 2
	s.SendKeepAlive()
}
