package controller

import (
	"encoding/json"
	"errors"
	"fmt"
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
	PCEPSessions           map[string]*pcep.Session
	PCEPSessionsByLoopback map[string]*pcep.Session
	NewSession             chan *pcep.Session
	TopoView               *TopoView
	// The LSP list is maintained by the controller and
	// inside PCEP libriry as well. If I just use one list in
	// PCEP then the controller does not know it created an LSP
	// and can try to create the same one before it recives RPT
	// from the router. Below are the LSPs initiated by us.
	// In theory there can be other LSPs delegated to us and they are stored
	// inside the PCEP Session struct so we need two lists.
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

func (c *Controller) GetSRLSPs() []*pcep.SRLSP {
	defer c.RUnlock()

	c.RLock()
	var lsps []*pcep.SRLSP
	for _, lsp := range c.LSPs {
		lsps = append(lsps, lsp)
	}
	return lsps
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

func (c *Controller) DelSRLSP(name string) error {
	defer c.Lock()

	c.Unlock()

	lsp := c.LSPs[name]

	session := c.PCEPSessionsByLoopback[lsp.Src]
	if session == nil {
		return fmt.Errorf("no session found for %s looback address", lsp.Src)
	}
	err := session.InitSRLSP(lsp)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "session",
			"event": "lsp_init",
		}).Error(err)
		return fmt.Errorf("failed to delete %s LSP got err: %s", lsp.Name, err.Error())
	}
	delete(c.LSPs, lsp.Name)

	err = c.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("lsps"))
		if err != nil {
			return err
		}
		delete(c.LSPs, lsp.Name)
		return b.Delete([]byte(lsp.Name))
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) CreateUpdSRLSP(lsp *pcep.SRLSP) error {
	defer c.Unlock()

	c.Lock()

	session := c.PCEPSessionsByLoopback[lsp.Src]
	if session == nil {
		return fmt.Errorf("no session found for %s looback address \n", lsp.Src)
	}

	err := session.InitSRLSP(lsp)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "session",
			"event": "lsp_init",
		}).Error(err)
		return fmt.Errorf("failed to update %s LSP got err: %s", lsp.Name, err.Error())
	}

	err = c.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("lsps"))
		if err != nil {
			return err
		}
		data, err := json.Marshal(lsp)
		if err != nil {
			return err
		}
		return b.Put([]byte(lsp.Name), data)
	})
	if err != nil {
		return err
	}
	c.LSPs[lsp.Name] = lsp

	return nil
}

// LoadLSPs retrive all LSP stored in Bolt DB used to init
func (c *Controller) LoadLSPs() (map[string]*pcep.SRLSP, error) {
	defer c.RUnlock()
	c.RLock()

	lsps := make(map[string]*pcep.SRLSP)

	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("lsps"))
		if b != nil {
			err := b.ForEach(func(k, v []byte) error {

				var LSP pcep.SRLSP
				err := json.Unmarshal(v, &LSP)
				if err != nil {
					return err
				}
				lsps[string(k)] = &LSP
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
	return lsps, nil
}

// CreateUpdRouter aa
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

func (c *Controller) GetRouterByPCEPSessionSrcIP(srcIP string) *Router {
	defer c.RUnlock()
	c.RLock()
	for _, router := range c.Routers {
		if router.PCEPSessionSrcIP == srcIP {
			return router
		}
	}
	return nil
}

func (c *Controller) GetPCEPSessionByLoopback(loopback string) *Router {
	defer c.RUnlock()
	c.RLock()

	for _, router := range c.Routers {
		if router.PCEPSessionSrcIP == loopback {
			return router
		}
	}

	return nil
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
	router := c.GetRouterByPCEPSessionSrcIP(key)

	c.Lock()
	delete(c.PCEPSessions, key)
	delete(c.PCEPSessionsByLoopback, router.LoopbackIP)
	c.Unlock()
}

// StorePSessions aa
func (c *Controller) StorePSessions(key string, value *pcep.Session) *pcep.Session {
	router := c.GetRouterByPCEPSessionSrcIP(value.GetSrcAddrFromSession())

	c.Lock()
	c.PCEPSessions[key] = value
	c.PCEPSessionsByLoopback[router.LoopbackIP] = value
	c.Unlock()

	c.NewSession <- value

	return value
}

// SessionStart aa
func (c *Controller) SessionStart(session *pcep.Session) {
	c.StorePSessions(session.GetSrcAddrFromSession(), session)
}

// SessionEnd aa
func (c *Controller) SessionEnd(key string) {
	c.DeletePSession(key)
}

// Start  controller
func Start(db *bolt.DB) *Controller {
	c := &Controller{
		PCEPSessions:           make(map[string]*pcep.Session),
		PCEPSessionsByLoopback: make(map[string]*pcep.Session),
		NewSession:             make(chan *pcep.Session),
		TopoView:               NewTopoView(),
		LSPs:                   make(map[string]*pcep.SRLSP),
		StopBGP:                make(chan bool),
		Routers:                make(map[string]*Router),
		db:                     db,
	}

	err := c.LoadRouters()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "controller",
			"event": "load_routers",
		}).Fatal(err)
	}

	c.LSPs, err = c.LoadLSPs()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "controller",
			"event": "load_lsp",
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

func (c *Controller) InitSRLSPs(session *pcep.Session) {

	// get details of the router using session src address
	router := c.GetRouterByPCEPSessionSrcIP(session.GetSrcAddrFromSession())
	if router == nil {
		return
	}

	for _, lsp := range c.LSPs {
		// we are only going to provision lsps
		// for a router which coresponds to a given session
		if router.LoopbackIP != lsp.Src {
			continue
		}

		err := session.InitSRLSP(lsp)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type":  "session",
				"event": "lsp_init",
			}).Error(err)
		}
	}
	// Init full mesh after we init all the manually created ones
	// this way you can overide any automaticaly created LSPs
	// as the below methods would not touch lsp which are already loaded
	c.InitSRLSPFullMesh(session)
}

// InitSRLSPs aaaa
func (c *Controller) InitSRLSPFullMesh(session *pcep.Session) {

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

	destinations := c.GetAllLSPDestinations()

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
		"dsts":        destinations,
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

	for _, dst := range destinations {
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
		// need to copare ERO list and other options to decide if we need to update
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
		"type":      "after",
		"func":      "InitSRLSP",
		"time_took": time.Since(start),
	}).Info("LSP init done")

}
