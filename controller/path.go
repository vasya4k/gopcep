package controller

import "sync"

type Path struct {
	Src   string
	Dst   string
	Cost  int
	Links []*Link
}

type Link struct {
	LocalNode       string
	RemoteNode      string
	IntIP           string
	NeighbourIP     string
	DefaultTEMetric uint32
	IGPMetric       uint32
	BW              float32
	ReservableBW    float32
	UnreservedBW    float32
	SRAdjacencySID  uint32
}

type Paths struct {
	sync.Map
}

func (p *Paths) StorePath(key string, value []*Path) {
	p.Store(key, value)
}

func (p *Paths) GetPath(key string) ([]*Path, bool) {
	v, ok := p.Load(key)
	if ok {
		return v.([]*Path), ok
	}
	return nil, ok
}

func (p *Paths) DelPath(key string) {
	p.Delete(key)
}

func (p *Paths) RangePath(f func(key interface{}, value interface{}) bool) {
	p.Range(f)
}
