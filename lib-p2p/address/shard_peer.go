package address

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common/errors"
	"sync"
)

type ShardNode struct {
	Pubkey  string
	Address string
	Port    string
}

type PeerMap struct {
	Peers map[string]ShardNode
	lock  sync.RWMutex
}

func (p *PeerMap) Initialize() PeerMap {
	p.Peers = make(map[string]ShardNode, 1)
	return *p
}

func (p *PeerMap) Add(b64Pub, addr, port string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.Peers[b64Pub]; ok {
		return
	}
	p.Peers[b64Pub] = ShardNode{Pubkey: b64Pub, Address: addr, Port: port}
}

func (p *PeerMap) Del(b64Pub string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.Peers[b64Pub]; ok {
		delete(p.Peers, b64Pub)
		return nil
	}
	return errors.New(fmt.Sprintf("can't find stream by id:%s", b64Pub))
}

func (p *PeerMap) Get(b64Pub string) *ShardNode {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if info, ok := p.Peers[b64Pub]; ok { //copy value
		return &info
	}
	return nil
}

func (p *PeerMap) Contains(b64Pub string) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if _, ok := p.Peers[b64Pub]; ok {
		return true
	}
	return false
}

func (p *PeerMap) Iterator() <-chan ShardNode {
	channel := make(chan ShardNode)
	go func() {
		p.lock.RLock()
		defer p.lock.RUnlock()
		for _, v := range p.Peers {
			channel <- v
		}
		close(channel)
	}()
	return channel
}

func (p *PeerMap) Clone() *PeerMap {
	p.lock.RLock()
	defer p.lock.RUnlock()
	np := new(PeerMap).Initialize()
	for k, v := range p.Peers {
		np.Add(k, v.Address, v.Port)
	}
	return &np
}
