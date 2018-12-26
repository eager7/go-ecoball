package network

import (
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"sync"
)

type SenderMap struct {
	senders map[peer.ID]*messageSender
	lock    sync.RWMutex
}

func (p *SenderMap) Initialize() SenderMap {
	p.senders = make(map[peer.ID]*messageSender)
	return *p
}

func (p *SenderMap) Len() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return len(p.senders)
}

func (p *SenderMap) Add(key peer.ID, Value *messageSender) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.senders[key] = Value
}

func (p *SenderMap) Get(key peer.ID) *messageSender {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if k, ok := p.senders[key]; ok {
		return k
	} else {
		return nil
	}
}

func (p *SenderMap) Contains(key peer.ID) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if _, ok := p.senders[key]; ok {
		return true
	} else {
		return false
	}
}

func (p *SenderMap) Del(key peer.ID) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.senders[key]; ok {
		delete(p.senders, key)
	}
}

func (p *SenderMap) Purge() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for k := range p.senders {
		delete(p.senders, k)
	}
}

func (p *SenderMap) Clone() SenderMap {
	n := SenderMap{}
	n.Initialize()
	for k, v := range p.senders {
		n.senders[k] = v
	}
	return n
}

func (p *SenderMap) Iterator() <-chan messageSender {
	channel := make(chan messageSender)
	go func() {
		p.lock.RLock()
		defer p.lock.RUnlock()
		for _, v := range p.senders {
			channel <- *v
		}
		close(channel)
	}()
	return channel
}
