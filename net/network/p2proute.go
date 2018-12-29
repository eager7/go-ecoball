// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball.
//
// The go-ecoball is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.

// Implement the p2p peer routing sync function
package network

import (
	"context"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/net/util"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmZ383TySJVeZWzGnWui6pRcKyYZk9VkKTuW7tmKRWk5au/go-libp2p-routing"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	kb "gx/ipfs/QmesQqwonP618R7cJZoFfA4ioYhhMKnDmtUxcAvvxEEGnw/go-libp2p-kbucket"
	"sync"
	"time"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message/mpb"
)

const (
	// K is the maximum number of requests to perform before returning failure.
	KValue = 20
	// Alpha is the concurrency factor for asynchronous requests.
	AlphaValue = 3

	RTSyncAckMaxCount = 10
	RTSyncInterval    = 30 * time.Second
)

type NetRouteTable struct {
	net      *NetWork
	rt       *kb.RoutingTable
	rtLock   sync.Mutex
	msgSubCh <-chan interface{}
	stop     chan bool

	routing.PeerRouting
}

func NewRouteTable(n *NetWork) *NetRouteTable {
	table := &NetRouteTable{
		net:         n,
		rt:          initRoutingTable(n.host),
		rtLock:      sync.Mutex{},
		msgSubCh:    nil,
		stop:        nil,
		PeerRouting: nil,
	}

	return table
}

func initRoutingTable(host host.Host) (table *kb.RoutingTable) {
	peerID := kb.ConvertPeerID(host.ID())

	rt := kb.NewRoutingTable(KValue, peerID, time.Minute, host.Peerstore())
	manager := host.ConnManager()
	rt.PeerAdded = func(p peer.ID) {
		manager.TagPeer(p, "kbucket", 5)
	}
	rt.PeerRemoved = func(p peer.ID) {
		manager.UntagPeer(p, "kbucket")
	}

	return rt
}

func (nrt *NetRouteTable) SyncRouteTable() {
	syncedPeers := make(map[peer.ID]bool)
	//sync with bootstrap peer
	if nrt.net.bootStrapper != nil {
		for _, bsp := range nrt.net.bootStrapper.bsPeers {
			if nrt.net.host.Network().Connectedness(bsp.ID()) == inet.Connected {
				nrt.SyncWithPeer(bsp.ID())
				syncedPeers[bsp.ID()] = true
			}
		}
	}
	selfID := kb.ConvertPeerID(nrt.net.host.ID())
	nPeers := nrt.rt.NearestPeers(selfID, AlphaValue)
	for _, p := range nPeers {
		if syncedPeers[p] {
			continue
		}
		nrt.SyncWithPeer(p)
		syncedPeers[p] = true
	}
}

func (nrt *NetRouteTable) SyncWithPeer(id peer.ID) {
	syncMsg := new(types.P2PRTSynMsg)
	syncMsg.Req = nrt.net.host.ID()
	data, err := syncMsg.Serialize()
	if err != nil {
		log.Error(err)
		return
	}
	msg := message.New(pb.MsgType_APP_MSG_P2PRTSYN, data)

	nrt.net.SendMsgToPeerWithId(id, msg)
}

func (nrt *NetRouteTable) OnSyncRoute(msg message.EcoBallNetMsg) {
	syncMsg := new(types.P2PRTSynMsg)
	err := syncMsg.Deserialize(msg.Data())
	if err != nil {
		log.Error(err)
		return
	}
	remotePeer := syncMsg.Req

	rtPeers := nrt.rt.ListPeers()
	var ackPeers []peer.ID
	if len(rtPeers) > RTSyncAckMaxCount {
		ackPeers = getRandomPeers(RTSyncAckMaxCount, rtPeers)
	}
	ackPeers = rtPeers

	syncPA := make([]*types.PeerAddress, len(ackPeers))
	for i, p := range ackPeers {
		pi := nrt.net.host.Peerstore().PeerInfo(p)
		addr, err := pstore.InfoToP2pAddrs(&pi)
		if err != nil {
			continue
		}

		pa := &types.PeerAddress{
			Id:     pi.ID,
			Ipport: addr[0].String(), //only get the first address
		}
		syncPA[i] = pa
	}

	syncAck := types.P2PRTSynAckMsg{
		Resp:  nrt.net.host.ID(),
		PAddr: syncPA,
	}

	data, err := syncAck.Serialize()
	if err != nil {
		log.Error(err)
		return
	}

	ackMsg := message.New(pb.MsgType_APP_MSG_P2PRTSYNACK, data)
	nrt.net.SendMsgToPeerWithId(remotePeer, ackMsg)
}

func (nrt *NetRouteTable) OnSyncRouteAck(msg message.EcoBallNetMsg) {
	ackMsg := new(types.P2PRTSynAckMsg)
	err := ackMsg.Deserialize(msg.Data())
	if err != nil {
		log.Error(err)
		return
	}

	for _, pa := range ackMsg.PAddr {
		addr, err := ma.NewMultiaddr(pa.Ipport)
		if err != nil {
			log.Error("failed to create address from ip and port", err)
			continue
		}
		nrt.net.host.Peerstore().AddAddr(pa.Id, addr, pstore.PermanentAddrTTL)
		pi := &pstore.PeerInfo{
			ID:    pa.Id,
			Addrs: []ma.Multiaddr{addr},
		}
		connectedness := nrt.net.host.Network().Connectedness(pi.ID)
		if (pi.ID != nrt.net.host.ID()) && !(connectedness == inet.Connected || connectedness == inet.CanConnect) {
			go func(pi *pstore.PeerInfo) {
				if pi.ID == nrt.net.host.ID() {
					return
				}
				if err := nrt.net.host.Connect(nrt.net.ctx, *pi); err != nil {
					log.Error(err)
					return
				}
				nrt.update(pi.ID)
			}(pi)
		}
	}
}

func (nrt *NetRouteTable) Start() {
	var err error
	msg := []mpb.Identify{
		mpb.Identify_APP_MSG_P2PRTSYN,
		mpb.Identify_APP_MSG_P2PRTSYNACK,
	}
	nrt.msgSubCh, err = event.Subscribe(msg...)
	if err != nil {
		log.Error(err)
		return
	}

	go nrt.SyncRouting(nrt.net.ctx)
}

func (nrt *NetRouteTable) Stop() {
	nrt.stop <- true
}

func (nrt *NetRouteTable) SyncRouting(ctx context.Context) {
	log.Debug("start to sync routing table")
	defer log.Debug("sync routing table was shutting down")

	nrt.SyncRouteTable()

	ticker := time.NewTicker(RTSyncInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-nrt.stop:
			return
		case <-ticker.C:
			nrt.SyncRouteTable()
		case sm := <-nrt.msgSubCh:
			msg, ok := sm.(message.EcoBallNetMsg)
			if !ok {
				continue
			}
			if msg.Type() == pb.MsgType_APP_MSG_P2PRTSYN {
				nrt.OnSyncRoute(msg)
			}
			if msg.Type() == pb.MsgType_APP_MSG_P2PRTSYNACK {
				nrt.OnSyncRouteAck(msg)
			}
		}
	}
}

func (nrt *NetRouteTable) FindPeer(ctx context.Context, id peer.ID) (pstore.PeerInfo, error) {
	// Check if were already connected to them
	if pi := nrt.findLocal(id); pi.ID != "" {
		return pi, nil
	}

	peers := nrt.rt.NearestPeers(kb.ConvertPeerID(id), AlphaValue)
	if len(peers) == 0 {
		return pstore.PeerInfo{}, kb.ErrLookupFailure
	}

	for _, p := range peers {
		if p == id {
			log.Debug("found target peer in list of closest peers...")
			return nrt.net.host.Peerstore().PeerInfo(p), nil
		}
	}

	return pstore.PeerInfo{}, kb.ErrLookupFailure
}

func (nrt *NetRouteTable) findLocal(id peer.ID) pstore.PeerInfo {
	switch nrt.net.host.Network().Connectedness(id) {
	case inet.Connected, inet.CanConnect:
		return nrt.net.host.Peerstore().PeerInfo(id)
	default:
		return pstore.PeerInfo{}
	}
}

func (nrt *NetRouteTable) update(p peer.ID) {
	nrt.rtLock.Lock()
	defer nrt.rtLock.Unlock()
	nrt.rt.Update(p)
}

func (nrt *NetRouteTable) remove(p peer.ID) {
	nrt.rtLock.Lock()
	defer nrt.rtLock.Unlock()
	nrt.rt.Remove(p)
}

func getRandomPeers(k int, peers []peer.ID) []peer.ID {

	if len(peers) < k {
		k = len(peers)
	}
	indices := util.GetRandomIndices(k, len(peers)-1)
	result := make([]peer.ID, len(indices))
	for i, j := range indices {
		result[i] = peers[j]
	}

	return result
}
