package controller

import (
	"fmt"
	"gopcep/pcep"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Controller represents TE controller
type Controller struct {
	sync.RWMutex
	PCEPSessions map[string]*pcep.Session
	NewSession   chan *pcep.Session
	TopoView     *TopoView
}

// LoadPSessions aa
func (c *Controller) LoadPSessions(key string) (*pcep.Session, bool) {
	c.RLock()
	result, ok := c.PCEPSessions[key]
	c.RUnlock()
	return result, ok
}

// DeletePSession aa
func (c *Controller) DeletePSession(key string) {
	c.Lock()
	delete(c.PCEPSessions, key)
	c.Unlock()
}

// StorePSessions aa
func (c *Controller) StorePSessions(key string, value *pcep.Session) *pcep.Session {
	c.Lock()
	c.PCEPSessions[key] = value
	c.NewSession <- value
	c.Unlock()
	return value
}

// SessionStart aa
func (c *Controller) SessionStart(value *pcep.Session) {
	c.StorePSessions(value.Conn.RemoteAddr().String(), value)
}

// SessionEnd aa
func (c *Controller) SessionEnd(key string) {
	c.DeletePSession(key)
}

// Start  aa
func Start() *Controller {
	c := &Controller{
		PCEPSessions: make(map[string]*pcep.Session),
		NewSession:   make(chan *pcep.Session),
		TopoView:     NewTopoView(),
	}

	go startBGPLS(c.TopoView)

	go func() {
		for {
			select {
			case s := <-c.NewSession:
				logrus.WithFields(logrus.Fields{
					"type":           "session",
					"event":          "created",
					"router_address": s.Conn.RemoteAddr().String(),
				}).Info("new session created")
				go c.watchSession(s)
			}
		}
	}()

	return c
}

func (c *Controller) watchSession(session *pcep.Session) {
	for {
		select {
		case <-session.SessionReady:
			logrus.WithFields(logrus.Fields{
				"type":           "session",
				"event":          "ready",
				"router_address": session.Conn.RemoteAddr().String(),
			}).Info("new session is ready")
			c.InitSRLSPs(session)
		case <-session.SessionClosed:
			logrus.WithFields(logrus.Fields{
				"type":           "session",
				"event":          "closed",
				"router_address": session.Conn.RemoteAddr().String(),
			}).Info("session closed")
		}
	}
}

// InitSRLSPs aaaa
func (c *Controller) InitSRLSPs(session *pcep.Session) {
	start := time.Now()
	for srcNode := range c.TopoView.NodesByIGPRouteID {
		for dstNode := range c.TopoView.NodesByIGPRouteID {
			if srcNode != dstNode {
				c.TopoView.FindAllPaths(srcNode, dstNode)
			}
		}
	}
	logrus.Printf("topo calc took %s", time.Since(start))
	srcAddr := strings.Split(session.Conn.RemoteAddr().String(), ":")[0]
	allDestinations := []string{"0192.0168.0014", "0192.0168.0011"}

	for _, dst := range allDestinations {
		bestPath := c.TopoView.findBestPath(0, srcAddr, dst)
		if bestPath == nil {
			continue
		}
		lsp, err := c.TopoView.createSRLSP(100, bestPath)
		if err != nil {
			fmt.Println(err)
		}
		err = session.InitSRLSP(lsp)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type":  "session",
				"event": "lsp_init",
			}).Error(err)
		}
		logrus.WithFields(logrus.Fields{
			"type": "lsp_provision",
			"func": "InitSRLSP",
			"src":  lsp.Src,
			"dst":  lsp.Dst,
		}).Info("new lsp provisioned")
	}

	logrus.WithFields(logrus.Fields{
		"type": "after",
		"func": "InitSRLSP",
	}).Info("new msg")

}

// InitSRLSPs aaaa
func (c *Controller) InitSRLSPsDebug(session *pcep.Session) error {
	if strings.Split(session.Conn.RemoteAddr().String(), ":")[0] == "10.0.0.10" {
		logrus.WithFields(logrus.Fields{
			"type": "before",
			"func": "InitSRLSP",
		}).Info("new msg")
		for _, lsp := range getLSPS() {
			err := session.InitSRLSP(lsp)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"type":  "session",
					"event": "lsp_init",
				}).Error(err)
			}
			logrus.WithFields(logrus.Fields{
				"type": "lsp_provision",
				"func": "InitSRLSP",
				"src":  lsp.Src,
				"dst":  lsp.Dst,
			}).Info("new lsp provisioned")
		}

		logrus.WithFields(logrus.Fields{
			"type": "after",
			"func": "InitSRLSP",
		}).Info("new msg")
	}
	return nil
}

func getLSPS() []*pcep.SRLSP {
	return []*pcep.SRLSP{
		{
			Delegate: true,
			Sync:     false,
			Remove:   false,
			Admin:    true,
			Name:     "LSP-55",
			Src:      "10.10.10.10",
			Dst:      "14.14.14.14",
			EROList: []pcep.SREROSub{
				{
					LooseHop:   false,
					MBit:       true,
					NT:         3,
					IPv4NodeID: "",
					SID:        402011,
					NoSID:      false,
					IPv4Adjacency: []string{
						0: "10.1.0.1",
						1: "10.1.0.0",
					},
				},
				{
					LooseHop:   false,
					MBit:       true,
					NT:         1,
					IPv4NodeID: "15.15.15.15",
					SID:        402015,
					NoSID:      false,
				},
				{
					LooseHop:   false,
					MBit:       true,
					NT:         1,
					IPv4NodeID: "14.14.14.14",
					SID:        402014,
					NoSID:      false,
				},
			},
			SetupPrio:    7,
			HoldPrio:     7,
			LocalProtect: false,
			BW:           100,
		},
		{
			Delegate: true,
			Sync:     false,
			Remove:   false,
			Admin:    true,
			Name:     "LSP-66",
			Src:      "10.10.10.10",
			Dst:      "13.13.13.13",
			EROList: []pcep.SREROSub{
				{
					LooseHop:   false,
					MBit:       true,
					NT:         3,
					IPv4NodeID: "",
					SID:        402011,
					NoSID:      false,
					IPv4Adjacency: []string{
						0: "10.1.0.1",
						1: "10.1.0.0",
					},
				},
				{
					LooseHop:   false,
					MBit:       true,
					NT:         1,
					IPv4NodeID: "15.15.15.15",
					SID:        402015,
					NoSID:      false,
				},
				{
					LooseHop:   false,
					MBit:       true,
					NT:         1,
					IPv4NodeID: "14.14.14.14",
					SID:        402014,
					NoSID:      false,
				},
				{
					LooseHop:   false,
					MBit:       true,
					NT:         1,
					IPv4NodeID: "13.13.13.13",
					SID:        402013,
					NoSID:      false,
				},
			},
			SetupPrio:    7,
			HoldPrio:     7,
			LocalProtect: false,
			BW:           100,
		},
	}
}
