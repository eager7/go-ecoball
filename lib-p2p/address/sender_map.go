package address

import (
	"gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"sync"
)

type Sender struct {
	ID       peer.ID
	Stream   net.Stream
	PeerInfo peerstore.PeerInfo //id, address, port
}

type SenderMap struct {
	Peers map[peer.ID]Sender
	P     sync.Map
	lock  sync.RWMutex
}

func (p *SenderMap) Initialize() {
	p.Peers = make(map[peer.ID]Sender)
}

func (p *SenderMap) Add(id peer.ID, s net.Stream, addr multiaddr.Multiaddr) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.Peers[id]; ok {
		return
	}
	peerInfo := peerstore.PeerInfo{ID: id, Addrs: []multiaddr.Multiaddr{addr}}
	p.Peers[id] = Sender{ID: id, Stream: s, PeerInfo: peerInfo}
}

func (p *SenderMap) Del(id peer.ID) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.Peers[id]; ok {
		delete(p.Peers, id)
	}
}

func (p *SenderMap) Get(id peer.ID) *Sender {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if info, ok := p.Peers[id]; ok {
		return &info
	}
	return nil
}

func (p *SenderMap) Iterator() <-chan Sender {
	channel := make(chan Sender)
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
