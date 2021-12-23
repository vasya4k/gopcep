package controller

import "sync"

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

type BGPLSPeer struct {
	NeighborAddress     string
	PeerAs              int
	EbgpMultihopEnabled bool
	EBGPMultihopTtl     int
}

type Routers struct {
	sync.Map
}

func (r *Routers) StoreRouter(key string, value *Router) {
	r.Store(key, value)
}

func (r *Routers) GetRouter(key string) (*Router, bool) {
	v, ok := r.Load(key)
	if ok {
		return v.(*Router), ok
	}
	return nil, ok
}

func (r *Routers) DelRouter(key string) {
	r.Delete(key)
}

func (r *Routers) RangeRouters(f func(key interface{}, value interface{}) bool) {
	r.Range(f)
}
