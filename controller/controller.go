package controller

import (
	"encoding/json"
	"errors"
	"gopcep/pcep"
	"strings"
	"sync"
	"time"

	gobgp "github.com/osrg/gobgp/pkg/server"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// Controller represents TE controller
type Controller struct {
	sync.RWMutex
	PCEPSessions map[string]*pcep.Session
	NewSession   chan *pcep.Session
	TopoView     *TopoView
	// The LSP list is maintained by the controller and
	// inside PCEP libriry as well. If I just use one list in
	// PCEP then the controller does not know it created an LSP
	// and can try to create the same one before it recives RPT
	// from the router. Below are the LSPs initiated by us.
	// In theory there can be other LSPs delegated to us and they are stored
	// inside the PCEP Session struct
	LSPs      map[string]*pcep.SRLSP
	StopBGP   chan bool
	db        *bolt.DB
	bgpServer *gobgp.BgpServer
	Routers   map[string]*Router
}
type BGPLSPeer struct {
	NeighborAddress     string
	PeerAs              int
	EbgpMultihopEnabled bool
	EBGPMultihopTtl     int
}

type Router struct {
	Name              string
	ID                string
	MgmtIP            string
	ISOAddr           string
	LoopbackIP        string
	BGPLSPeer         bool
	BGPLSPeerCfg      BGPLSPeer
	IncludeInFullMesh bool
	PCEPSessionSrcIP  string
}

func (c *Controller) GetLSPs() []*pcep.LSP {
	defer c.RUnlock()

	c.RLock()
	var lsps []*pcep.LSP
	for _, session := range c.PCEPSessions {
		for _, lsp := range session.LSPs {
			lsps = append(lsps, lsp)
		}
	}
	return lsps
}

// AddRouter aa
func (c *Controller) CreateUpdRouter(router *Router) error {
	defer c.Unlock()
	c.Lock()
	err := c.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("routers"))
		if err != nil {
			return err
		}
		data, err := json.Marshal(router)
		if err != nil {
			return err
		}
		_, ok := c.Routers[router.ID]
		if ok {
			id, err := uuid.FromString(router.ID)
			if err != nil {
				return err
			}
			return b.Put(id.Bytes(), data)
		}
		id := uuid.NewV4()
		router.ID = id.String()
		return b.Put(id.Bytes(), data)
	})
	c.Routers[router.ID] = router
	if err != nil {
		return err
	}
	return nil
}

// DeleteRouter aa
func (c *Controller) DeleteRouter(id string) error {
	defer c.Unlock()
	c.Lock()
	err := c.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("routers"))
		if err != nil {
			return err
		}

		uid, err := uuid.FromString(id)
		if err != nil {
			return err
		}

		var r *Router

		for _, v := range c.Routers {
			if v.ID == uid.String() {
				r = v
			}
		}
		delete(c.Routers, r.ID)
		return b.Delete(uid.Bytes())
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) GetRouters() ([]*Router, error) {
	defer c.RUnlock()
	c.RLock()
	routers := make([]*Router, 0)
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("routers"))
		if b != nil {
			err := b.ForEach(func(k, v []byte) error {
				var r Router
				err := json.Unmarshal(v, &r)
				if err != nil {
					return err
				}
				routers = append(routers, &r)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return routers, nil
}

func (c *Controller) LoadRouters() error {
	defer c.Unlock()
	routers, err := c.GetRouters()
	if err != nil {
		return err
	}
	c.Lock()
	for _, r := range routers {
		c.Routers[r.ID] = r
	}
	return nil
}

func (c *Controller) GetAllLSPDestinations() []string {
	defer c.RUnlock()

	var destinations []string

	c.RLock()
	for _, r := range c.Routers {
		if !r.IncludeInFullMesh {
			continue
		}
		destinations = append(destinations, r.ISOAddr)
	}
	return destinations
}

func (c *Controller) GetRouterISOAddr(pcepSrcIP string) (string, error) {
	defer c.RUnlock()

	c.RLock()
	for _, r := range c.Routers {
		if !r.IncludeInFullMesh {
			continue
		}
		if r.PCEPSessionSrcIP == pcepSrcIP {
			return r.ISOAddr, nil
		}
	}
	return "", errors.New("router with a given PCEP src session addr not found")
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

// Start  controller
func Start(db *bolt.DB) *Controller {
	c := &Controller{
		PCEPSessions: make(map[string]*pcep.Session),
		NewSession:   make(chan *pcep.Session),
		TopoView:     NewTopoView(),
		LSPs:         make(map[string]*pcep.SRLSP),
		StopBGP:      make(chan bool),
		Routers:      make(map[string]*Router),
		db:           db,
	}

	err := c.LoadRouters()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "controller",
			"event": "load_routers",
		}).Fatal(err)
	}
	go c.StartBGPLS()

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
			case <-c.TopoView.TopologyUpdate:
				logrus.WithFields(logrus.Fields{
					"type":  "topology",
					"event": "update",
				}).Info("new topology update running LSP optimisation")
				for _, session := range c.PCEPSessions {
					c.InitSRLSPs(session)
				}
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

func getSrcAddrFromSession(session *pcep.Session) string {
	return strings.Split(session.Conn.RemoteAddr().String(), ":")[0]
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
	logrus.WithFields(logrus.Fields{
		"type":      "lsp_init",
		"event":     "topo calc done",
		"time_took": time.Since(start),
	}).Info("find all path done")

	// sessionSrcAddr := getSrcAddrFromSession(session)
	// allDestinations := []string{"0192.0168.0014", "0192.0168.0011"}
	// pcepSrctoIGPSrcMapping := map[string]string{"10.0.0.10": "0100.1001.0010"}
	// srcAddr := pcepSrctoIGPSrcMapping[sessionSrcAddr]

	allDestinations := c.GetAllLSPDestinations()
	srcAddr, err := c.GetRouterISOAddr(getSrcAddrFromSession(session))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "lsp_init",
			"event": "GetRouterISOAddr",
		}).Error(err)
		return
	}

	logrus.WithFields(logrus.Fields{
		"type":        "lsp_init",
		"event":       "best_path_calc",
		"src_address": srcAddr,
		"dsts":        allDestinations,
		"paths": func() []string {
			keys := make([]string, len(c.TopoView.Paths))
			i := 0
			for k := range c.TopoView.Paths {
				keys[i] = k
				i++
			}
			return keys
		}(),
	}).Info("looking for best paths for all destinations")

	for _, dst := range allDestinations {
		bestPath := c.TopoView.findBestPath(0, srcAddr, dst)
		if bestPath == nil {
			continue
		}
		logrus.WithFields(logrus.Fields{
			"type":        "lsp_init",
			"event":       "best_path_found",
			"src_address": srcAddr,
			"dss":         dst,
			"path":        bestPath,
		}).Info("looking for best paths for all destinations")

		lsp, err := c.TopoView.createSRLSP(100, bestPath)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type":  "session",
				"event": "create_lsp",
			}).Error(err)
			continue
		}
		// this needs to be turned into proper comparation of LSPs
		// so if the new LSP is the same no point touching it
		_, ok := c.LSPs[lsp.Name]
		if ok {
			continue
		}

		logrus.WithFields(logrus.Fields{
			"type":        "lsp_init",
			"event":       "lsp_created",
			"src_address": srcAddr,
			"dss":         dst,
			"lsp":         lsp,
		}).Info("lsp created now running pcep init")
		err = session.InitSRLSP(lsp)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type":  "session",
				"event": "lsp_init",
			}).Error(err)
		}
		c.LSPs[lsp.Name] = lsp
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
