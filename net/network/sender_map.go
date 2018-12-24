package network

import (
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"sync"
)

type SenderMap struct {
	Senders map[peer.ID]*messageSender
	lock    sync.RWMutex
}

func (p *SenderMap) Initialize() SenderMap {
	p.Senders = make(map[peer.ID]*messageSender)
	return *p
}

func (p *SenderMap) Len() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return len(p.Senders)
}

func (p *SenderMap) Add(key peer.ID, Value *messageSender) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Senders[key] = Value
}

func (p *SenderMap) Get(key peer.ID) *messageSender {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if k, ok := p.Senders[key]; ok {
		return k
	} else {
		return nil
	}
}

func (p *SenderMap) Contains(key peer.ID) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if _, ok := p.Senders[key]; ok {
		return true
	} else {
		return false
	}
}

func (p *SenderMap) Del(key peer.ID) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.Senders[key]; ok {
		delete(p.Senders, key)
	}
}

func (p *SenderMap) Purge() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for k := range p.Senders {
		delete(p.Senders, k)
	}
}

func (p *SenderMap) Clone() SenderMap {
	n := SenderMap{}
	n.Initialize()
	for k, v := range p.Senders {
		n.Senders[k] = v
	}
	return n
}

func (p *SenderMap) Iterator() <-chan messageSender {
	channel := make(chan messageSender)
	go func() {
		p.lock.RLock()
		defer p.lock.RUnlock()
		for _, v := range p.Senders {
			channel <- *v
		}
		close(channel)
	}()
	return channel
}
