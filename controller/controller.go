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
	*sync.RWMutex
	NewSession chan *pcep.Session
	TopoView   *TopoView
	StopBGP    chan bool
	db         *bolt.DB
	bgpServer  *gobgp.BgpServer
	BGPLSCfg   *BGPGlobalCfg
	// The LSP list is maintained by the controller and
	// inside PCEP libriry as well. If I just use one list in
	// PCEP then the controller does not know it created an LSP
	// and can try to create the same one before it recives RPT
	// from the router. Below are the LSPs initiated by us.
	// In theory there can be other LSPs delegated to us and they are stored
	// inside the PCEP Session struct so we need two lists.
	// lsps                   map[string]*pcep.SRLSP
	PCEPSessions           map[string]*pcep.Session
	PCEPSessionsByLoopback map[string]*pcep.Session
	Routers
	LSPs
}

func (c *Controller) GetSRLSPs() []*pcep.SRLSP {
	var lsps []*pcep.SRLSP

	c.RangeLSPs(func(key, value interface{}) bool {
		lsps = append(lsps, value.(*pcep.SRLSP))
		return true
	})

	return lsps
}

func (c *Controller) GetLSPs() []*pcep.LSP {
	defer c.RUnlock()

	c.RLock()
	var lsps []*pcep.LSP
	for _, session := range c.PCEPSessions {
		session.RLock()
		for _, lsp := range session.LSPs {
			lsps = append(lsps, lsp)
		}
		session.RUnlock()
	}
	return lsps
}

func (c *Controller) DelSRLSP(name string) error {
	defer c.Unlock()

	c.Lock()

	ctrLSP, ok := c.GetLSP(name)
	if !ok {
		return fmt.Errorf("no LSP named: %s found", name)
	}

	session, ok := c.PCEPSessionsByLoopback[ctrLSP.Src]

	lsp := session.GetLSP(name)

	ctrLSP.SRPRemove = true
	ctrLSP.PLSPID = lsp.PLSPID
	// if sesstion exists we delete the LSP
	// if not we just delete it from the DB and map
	if ok {
		err := session.InitSRLSP(ctrLSP)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type":  "session",
				"event": "lsp_init",
			}).Error(err)
			return fmt.Errorf("failed to delete %s LSP got err: %s", lsp.Name, err.Error())
		}
	}

	err := c.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("lsps"))
		if err != nil {
			return err
		}
		c.DelLSP(lsp.Name)
		return b.Delete([]byte(lsp.Name))
	})
	if err != nil {
		return err
	}
	c.DelLSP(lsp.Name)

	return nil
}

func (c *Controller) CreateUpdSRLSP(lsp *pcep.SRLSP) error {
	defer c.Unlock()

	c.Lock()

	session, ok := c.PCEPSessionsByLoopback[lsp.Src]
	// if sesstion exists we init the LSP
	// if not we just save it to use once we get session esteblished
	if ok {
		err := session.InitSRLSP(lsp)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type":  "session",
				"event": "lsp_init",
			}).Error(err)
			return fmt.Errorf("failed to update %s LSP got err: %s", lsp.Name, err.Error())
		}
	}

	err := c.db.Update(func(tx *bolt.Tx) error {
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
	c.StoreLSP(lsp.Name, lsp)

	return nil
}

// LoadLSPs retrive all LSP stored in Bolt DB used to init
func (c *Controller) LoadLSPs() error {

	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("lsps"))
		if b != nil {
			err := b.ForEach(func(k, v []byte) error {

				var LSP pcep.SRLSP
				err := json.Unmarshal(v, &LSP)
				if err != nil {
					return err
				}
				c.StoreLSP(string(k), &LSP)
				return nil

			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
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

		if router.ID == "" {
			id := uuid.NewV4()
			router.ID = id.String()

			data, err := json.Marshal(router)
			if err != nil {
				return err
			}
			return b.Put([]byte(router.ID), data)
		}

		id, err := uuid.FromString(router.ID)
		if err != nil {
			return err
		}
		data, err := json.Marshal(router)
		if err != nil {
			return err
		}
		_, ok := c.GetRouter(router.ID)
		if ok {
			return b.Put(id.Bytes(), data)
		}

		return fmt.Errorf("no router with id: %s found", router.ID)
	})
	if err != nil {
		return err
	}
	c.StoreRouter(router.ID, router)
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

		c.DelRouter(uid.String())
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

func (c *Controller) GetClients() []string {
	clients := make([]string, 0)

	routers, err := c.GetRouters()
	if err != nil {
		return clients
	}
	for _, router := range routers {
		clients = append(clients, router.PCEPSessionSrcIP)
	}

	return clients
}

func (c *Controller) GetRouterByPCEPSessionSrcIP(srcIP string) *Router {
	var r *Router

	c.RangeRouters(func(key, value interface{}) bool {
		if value.(*Router).PCEPSessionSrcIP == srcIP {
			r = value.(*Router)
			return false
		}
		return true
	})

	return r
}

func (c *Controller) LoadRouters() error {
	routers, err := c.GetRouters()
	if err != nil {
		return err
	}
	for _, r := range routers {
		c.StoreRouter(r.ID, r)
	}
	return nil
}

func (c *Controller) GetAllLSPDestinations() []string {
	var destinations []string

	c.RangeRouters(func(key, value interface{}) bool {
		r := value.(*Router)

		if r.IncludeInFullMesh {
			destinations = append(destinations, r.ISOAddr)
		}

		return true
	})

	return destinations
}

func (c *Controller) GetRouterISOAddr(pcepSrcIP string) (string, error) {
	var r *Router

	c.RangeRouters(func(key, value interface{}) bool {
		if value.(*Router).PCEPSessionSrcIP == pcepSrcIP {
			r = value.(*Router)
			return false
		}
		return true
	})

	if r == nil {
		return "", errors.New("router with a given PCEP src session addr not found")
	}
	return r.ISOAddr, nil
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
func (c *Controller) StorePSessions(key string, value *pcep.Session) (*pcep.Session, error) {

	router := c.GetRouterByPCEPSessionSrcIP(value.GetSrcAddrFromSession())
	if router == nil {
		return value, fmt.Errorf("router with PCEP SRC addr: %s not found", value.GetSrcAddrFromSession())
	}

	c.Lock()
	c.PCEPSessions[key] = value
	c.PCEPSessionsByLoopback[router.LoopbackIP] = value
	c.Unlock()

	c.NewSession <- value

	return value, nil
}

// SessionStart aa
func (c *Controller) SessionStart(session *pcep.Session) error {
	_, err := c.StorePSessions(session.GetSrcAddrFromSession(), session)
	return err
}

// SessionEnd aa
func (c *Controller) SessionEnd(key string) {
	c.DeletePSession(key)
}

// Start  controller
func Start(db *bolt.DB, bgpcfg *BGPGlobalCfg) *Controller {
	c := &Controller{
		PCEPSessions:           make(map[string]*pcep.Session),
		PCEPSessionsByLoopback: make(map[string]*pcep.Session),
		NewSession:             make(chan *pcep.Session),
		TopoView:               NewTopoView(),
		StopBGP:                make(chan bool),
		Routers:                Routers{},
		LSPs:                   LSPs{},
		RWMutex:                &sync.RWMutex{},
		db:                     db,
		BGPLSCfg:               bgpcfg,
	}

	err := c.LoadRouters()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type":  "controller",
			"event": "load_routers",
		}).Fatal(err)
	}

	err = c.LoadLSPs()
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
				c.RLock()
				for _, session := range c.PCEPSessions {
					c.InitSRLSPs(session)
				}
				c.RUnlock()

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

	for _, lsp := range c.GetSRLSPs() {
		// we are only going to provision lsps
		// for a router which coresponds to a given session
		if router.LoopbackIP != lsp.Src {
			continue
		}

		// Check if an LSP with the same name exists already
		// and if so we do not init again
		// LSP names are unique so to create one with the same name
		// first we need to remove the existing one
		sessionLSP := session.GetLSP(lsp.Name)
		if sessionLSP != nil && sessionLSP.Oper == 2 {
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

	c.TopoView.FindPathsForAllSrcDstPairs()

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
	}).Info("looking for best paths for all destinations")

	for _, dst := range destinations {

		bestPath := c.TopoView.findBestPath(0, srcAddr, dst)
		if bestPath == nil {
			logrus.WithFields(logrus.Fields{
				"type":        "lsp_init",
				"event":       "no_best_path_found",
				"src_address": srcAddr,
				"dst":         dst,
			}).Info("failed to find any path to destination")
			continue
		}
		logrus.WithFields(logrus.Fields{
			"type":        "lsp_init",
			"event":       "best_path_found",
			"src_address": srcAddr,
			"dst":         dst,
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
		_, ok := c.GetLSP(lsp.Name)
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

		c.StoreLSP(lsp.Name, lsp)

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
