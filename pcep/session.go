package pcep

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

//Session holds everything for a PCEP session
type Session struct {
	*sync.RWMutex
	ID                uint8
	MsgCount          uint64
	State             int
	Conn              net.Conn
	RemoteOK          bool
	LocalOK           bool
	Keepalive         uint8
	DeadTimer         uint8
	PLSPIDToName      map[uint32]string
	LSPs              map[string]*LSP
	IDCounter         uint32
	SRPID             uint32
	StopKA            chan struct{} `json:"-"`
	RcvKA             chan bool     `json:"-"`
	SRCap             *SRPCECap
	StatefulCap       *StatefulPCECapability
	Open              *OpenObject
	SessionReady      chan bool `json:"-"`
	SessionClosed     chan bool `json:"-"`
	SessionErrRecived chan bool `json:"-"`
}

//NewSession creates a new session with defaults
func NewSession(conn net.Conn) *Session {
	return &Session{
		Conn:         conn,
		StopKA:       make(chan struct{}),
		RcvKA:        make(chan bool),
		Keepalive:    30,
		PLSPIDToName: make(map[uint32]string),
		LSPs:         make(map[string]*LSP),
		SessionReady: make(chan bool),
		RWMutex:      &sync.RWMutex{},
		SRCap:        &SRPCECap{},
		StatefulCap:  &StatefulPCECapability{},
		Open:         &OpenObject{},
	}
}

// ExportableSession is used as a copy of the sessions
// for exemple when you need to serialise it as JSON
// it is easier and probably faster to copy rather than
// hold the lock while we marshal. Alos the copy can be used to matshal
// into anything not only JSON
type ExportableSession struct {
	ID          uint8
	MsgCount    uint64
	State       int
	Conn        net.Conn
	RemoteOK    bool
	LocalOK     bool
	Keepalive   uint8
	DeadTimer   uint8
	IDCounter   uint32
	SRPID       uint32
	SRCap       SRPCECap
	StatefulCap StatefulPCECapability
	Open        OpenObject
}

func (s *Session) CopyToExportableSession() *ExportableSession {
	defer s.RUnlock()

	s.RLock()

	return &ExportableSession{
		ID:          s.ID,
		MsgCount:    s.MsgCount,
		State:       s.State,
		Conn:        s.Conn,
		RemoteOK:    s.RemoteOK,
		LocalOK:     s.LocalOK,
		Keepalive:   s.Keepalive,
		DeadTimer:   s.DeadTimer,
		IDCounter:   s.IDCounter,
		SRPID:       s.SRPID,
		SRCap:       *s.SRCap,
		StatefulCap: *s.StatefulCap,
		Open:        *s.Open,
	}
}

func (s *Session) GetSrcAddrFromSession() string {
	defer s.RUnlock()
	s.RLock()
	return strings.Split(s.Conn.RemoteAddr().String(), ":")[0]
}

func (s *Session) saveUpdLSP(lsp *LSP) {
	defer s.Unlock()

	s.Lock()

	if lsp.Name != "" {
		s.LSPs[lsp.Name] = lsp
		s.PLSPIDToName[lsp.PLSPID] = lsp.Name

	} else {
		s.LSPs[s.PLSPIDToName[lsp.PLSPID]] = lsp
	}
}

func (s *Session) delLSP(lsp *LSP) {
	defer s.Unlock()

	s.Lock()

	if lsp.Name != "" {
		delete(s.LSPs, lsp.Name)
		delete(s.PLSPIDToName, lsp.PLSPID)

	} else {
		delete(s.LSPs, s.PLSPIDToName[lsp.PLSPID])
		delete(s.PLSPIDToName, lsp.PLSPID)
	}
}

//ProcessOpen recive msg handler
func (s *Session) ProcessOpen(data []byte) {
	defer s.Unlock()

	s.Lock()

	logrus.WithFields(logrus.Fields{
		"type": "open",
		"peer": s.Conn.RemoteAddr().String(),
	}).Info("new msg")
	h, err := parseCommonObjectHeader(data)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":   "err",
			"caller": "RcvSessionOpen",
			"func":   "parseCommonObjectHeader",
		}).Error(err)
		return
	}
	if h.ObjectClass != 1 && h.ObjectType != 1 {
		log.Printf("Remote IP: %s, object class and object type do not match OPEN msg RFC definitions", s.Conn.RemoteAddr())
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

	logrus.WithFields(logrus.Fields{
		"topic":    "open",
		"peer":     s.Conn.RemoteAddr().String(),
		"open_obj": fmt.Sprintf("%+v", s.Open),
	}).Info("parsed open obj")

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

}

//SendSessionOpen send OPEN msg handler
func (s *Session) SendSessionOpen() {
	defer s.Unlock()

	s.Lock()
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

	i, err := s.Conn.Write(packet)
	if err != nil {
		log.Println(err)
	}
	logrus.WithFields(logrus.Fields{
		"type":  "info",
		"event": "open",
	}).Info(fmt.Sprintf("Sent Open: %d byte", i))
}

//StartKeepAlive start sending keep alive msgs
func (s *Session) StartKeepAlive() {
	// message-Type (8 bits): set to 2 means keepalive
	// length is set to zero as only common header itself is going to be sent
	commH, err := newCommonHeader(2, 0)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": "common header creation failure",
			"peer":  s.Conn.RemoteAddr().String(),
		}).Error(err)
		return
	}

	var firstSent bool

	s.RLock()
	k := s.Keepalive
	s.RUnlock()

	for {
		logrus.WithFields(logrus.Fields{
			"peer":      s.Conn.RemoteAddr().String(),
			"keepalive": s.Keepalive,
		}).Info("sent keepalive")
		if firstSent {
			time.Sleep(time.Second * time.Duration(k))
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
				logrus.WithFields(logrus.Fields{
					"topic": "tcp write failure",
					"peer":  s.Conn.RemoteAddr().String(),
				}).Error(err)
				return
			}
			// log.Printf("keep alive sent %d bytes", i)
			firstSent = true
		}
	}
}

//HandleDeadTimer start dead timer and wait for keepalive
func (s *Session) HandleDeadTimer() {
	for {
		select {
		case res := <-s.RcvKA:
			logrus.WithFields(logrus.Fields{
				"topic": "recieved keepalive",
				"peer":  s.Conn.RemoteAddr().String(),
			}).Info(res)
		case <-time.After(time.Duration(s.Open.DeadTimer) * time.Second):
			logrus.WithFields(logrus.Fields{
				"topic": "dead timer expiried",
				"peer":  s.Conn.RemoteAddr().String(),
			}).Info("sending close")

			// https://tools.ietf.org/html/rfc5440#section-7.17
			// 2          DeadTimer expired
			msg, err := newCloseMsg(2)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"topic": "close msg creation failure",
					"peer":  s.Conn.RemoteAddr().String(),
				}).Error(err)
				return
			}
			_, err = s.Conn.Write(msg)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"topic": "tcp write failure",
					"peer":  s.Conn.RemoteAddr().String(),
				}).Error(err)
				return
			}
			err = s.Conn.Close()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"topic": "tcp conn close failure",
					"peer":  s.Conn.RemoteAddr().String(),
				}).Error(err)
			}
		}
	}
}

// HandleNewMsg handles incoming data
func (s *Session) HandleNewMsg(data []byte) {
	var (
		offset    uint16
		newOffset uint16
	)
	for (len(data) - int(newOffset)) >= 4 {
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
			s.ProcessOpen(data[offset+4:])
			s.SendSessionOpen()

			s.Lock()
			s.State = 2
			s.Unlock()

			go s.StartKeepAlive()
			go s.HandleDeadTimer()

		case ch.MessageType == 2:
			logrus.WithFields(logrus.Fields{
				"type": "keepalive",
				"peer": s.Conn.RemoteAddr().String(),
			}).Info("rcv new msg")

			s.RcvKA <- true

			s.Lock()
			if s.State == 2 {
				s.SessionReady <- true
			}
			s.State = 3
			s.Unlock()

		case ch.MessageType == 3:
			log.Println("recv path computation request ")
		case ch.MessageType == 4:
			log.Println("recv path computation reply ")
		case ch.MessageType == 5:
			log.Println("recv notification ")
		case ch.MessageType == 6:
			fmt.Printf("len %d Whole ERR MSG: %08b \n", ch.MessageLength, data[:ch.MessageLength])
			s.handleErrObj(data[offset+4 : offset+ch.MessageLength])

		case ch.MessageType == 7:
			logrus.WithFields(logrus.Fields{
				"type": "close",
				"peer": s.Conn.RemoteAddr().String(),
				"msg":  parseClose(data[offset+8 : offset+12]),
			}).Info("new msg")

			err := s.Conn.Close()
			if err != nil {
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"topic": "tcp conn close failure",
						"peer":  s.Conn.RemoteAddr().String(),
					}).Error(err)
				}
			}
		case ch.MessageType == 10:
			logrus.WithFields(logrus.Fields{
				"type": "path computation lsp state report",
				"peer": s.Conn.RemoteAddr().String(),
				"len":  ch.MessageLength,
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
			log.Println("Unknown msg received")
		}
	}
}

type Controller interface {
	SessionStart(*Session)
	SessionEnd(string)
	GetClients() []string
}

func startPCEPSession(conn net.Conn, controller Controller) {
	session := NewSession(conn)
	controller.SessionStart(session)

	defer func() {
		err := conn.Close()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"topic":       "closing connection",
				"remote_addr": conn.RemoteAddr().String(),
			}).Error(err)
		}
		controller.SessionEnd(strings.Split(conn.RemoteAddr().String(), ":")[0])
	}()

	for {
		buff := make([]byte, 1024)
		l, err := conn.Read(buff)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"topic":       "conn read error",
				"remote_addr": conn.RemoteAddr().String(),
			}).Error(err)
			close(session.StopKA)
			return
		}
		logrus.WithFields(logrus.Fields{
			"remote_addr": conn.RemoteAddr().String(),
			"len":         l,
		}).Info("got some data ")
		session.HandleNewMsg(buff[:l])
	}
}

func clientNotInConfig(conn net.Conn, controller Controller) bool {
	for _, ip := range controller.GetClients() {
		if strings.Split(conn.RemoteAddr().String(), ":")[0] == ip {
			return false
		}
	}
	return true
}

type Cfg struct {
	ListenAddr string
	ListenPort string
	Keepalive  uint8
}

func ListenForNewSession(controller Controller, cfg *Cfg) error {
	ln, err := net.Listen("tcp", cfg.ListenAddr+":"+cfg.ListenPort)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic":     "listen",
			"event":     "net listen error",
			"addr_port": cfg.ListenAddr + ":" + cfg.ListenPort,
		}).Error(err)
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"topic":     "accept",
				"event":     "accept error",
				"addr_port": cfg.ListenAddr + ":" + cfg.ListenPort,
			}).Error(err)
			return err
		}
		if clientNotInConfig(conn, controller) {
			continue
		}
		logrus.WithFields(logrus.Fields{
			"remote_addr": conn.RemoteAddr().String(),
		}).Info("new connection")

		go startPCEPSession(conn, controller)
	}

}
