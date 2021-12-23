package controller

import (
	"gopcep/pcep"
	"sync"
)

type LSPs struct {
	sync.Map
}

func (l *LSPs) StoreLSP(key string, value *pcep.SRLSP) {
	l.Store(key, value)
}

func (l *LSPs) GetLSP(key string) (*pcep.SRLSP, bool) {
	v, ok := l.Load(key)
	if ok {
		return v.(*pcep.SRLSP), ok
	}
	return nil, ok
}

func (l *LSPs) DelLSP(key string) {
	l.Delete(key)
}

func (l *LSPs) RangeLSPs(f func(key interface{}, value interface{}) bool) {
	l.Range(f)
}
