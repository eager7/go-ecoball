package address

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common/errors"
	"gx/ipfs/QmYj8wdn5sZEHX2XMDWGBvcXJNdzVbaVpHmXvhHBVZepen/go-libp2p-net"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"sync"
)

type Peer struct {
	s         net.Stream
	PeerInfo  peerstore.PeerInfo
	PublicKey string
}

type PeerMap struct {
	Peers map[peer.ID]Peer
	lock  sync.RWMutex
}

func (p *PeerMap) Initialize() PeerMap {
	p.Peers = make(map[peer.ID]Peer, 1)
	return *p
}

func (p *PeerMap) Add(id peer.ID, s net.Stream, addr []multiaddr.Multiaddr, b64Pub string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.Peers[id]; ok {
		return
	}
	peerInfo := peerstore.PeerInfo{ID: id, Addrs: addr}
	p.Peers[id] = Peer{s: s, PeerInfo: peerInfo, PublicKey: b64Pub}
}

func (p *PeerMap) Del(id peer.ID) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.Peers[id]; ok {
		delete(p.Peers, id)
		return nil
	}
	return errors.New(fmt.Sprintf("can't find stream by id:%s", id))
}

func (p *PeerMap) Get(id peer.ID) *Peer {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if info, ok := p.Peers[id]; ok {
		return &info
	}
	return nil
}

func (p *PeerMap) Contains(id peer.ID) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if _, ok := p.Peers[id]; ok {
		return true
	}
	return false
}

func (p *PeerMap) Iterator() <-chan Peer {
	channel := make(chan Peer)
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
		np.Add(k, v.s, v.PeerInfo.Addrs, v.PublicKey)
	}
	return &np
}
