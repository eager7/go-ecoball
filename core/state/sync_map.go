package state

import (
	"fmt"
	"sync"
	"errors"
)

type ParamsMap struct {
	Params map[string]uint64
	lock   sync.RWMutex
}

func (p *ParamsMap) Initialize() {
	p.Params = make(map[string]uint64)
}

func (p *ParamsMap) Add(key string, value uint64) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Params[key] = value
}

func (p *ParamsMap) Get(key string) (uint64, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if k, ok := p.Params[key]; ok {
		return k, nil
	} else {
		return 0, errors.New(fmt.Sprintf("can't find the param:%s", key))
	}
}

func (p *ParamsMap) Contains(key string) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if _, ok := p.Params[key]; ok {
		return true
	} else {
		return false
	}
}

func (p *ParamsMap) Del(key string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.Params[key]; ok {
		delete(p.Params, key)
	}
}

func (p *ParamsMap) Purge() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for k := range p.Params {
		delete(p.Params, k)
	}
}

func (p *ParamsMap) Clone() ParamsMap {
	n := ParamsMap{}
	n.Initialize()
	for k, v := range p.Params {
		n.Params[k] = v
	}
	return n
}
