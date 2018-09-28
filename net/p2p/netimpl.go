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
package p2p

import (
	"time"
	"fmt"
	"io"
	"context"
	"sync"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/util"
	"github.com/ecoball/go-ecoball/common/config"
	kb "gx/ipfs/QmesQqwonP618R7cJZoFfA4ioYhhMKnDmtUxcAvvxEEGnw/go-libp2p-kbucket"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	ggio "gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/io"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p/p2p/discovery"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

const (
	// K is the maximum number of requests to perform before returning failure.
	KValue                 = 20
	// Alpha is the concurrency factor for asynchronous requests.
	AlphaValue             = 3

	sendMessageChanBuff    = 1024
	sendMessageTimeout     = time.Minute * 10

	discoveryConnTimeout   = time.Second * 30

	ServiceTag                 = "_net-discovery._udp"

	ProtocolP2pV1  protocol.ID = "/ecoball/app/1.0.0"
)

var (
	log = elog.NewLogger("p2p", elog.DebugLog)
	netImpl *NetImpl
)

func NewNetwork(ctx context.Context, host host.Host) EcoballNetwork {
	if netImpl != nil {
		return netImpl
	}
	netImpl = &NetImpl{
		ctx:          ctx,
		host:         host,
		strmap:       make(map[peer.ID]*messageSender),
		quitsendJb:   make(chan bool, 1),
		sendJbQueue:  make(chan interface{}, sendMessageChanBuff),
	}
	netImpl.routingTable = initRoutingTable(host)

	host.SetStreamHandler(ProtocolP2pV1, netImpl.handleNewStream)
	host.Network().Notify((*netNotifiee)(netImpl))
	// TODO: StopNotify.
	return netImpl
}

func GetNetInstance() EcoballNetwork {
	return netImpl
}

func initRoutingTable(host host.Host) (table *kb.RoutingTable) {
	peerID := kb.ConvertPeerID(host.ID())

	rt := kb.NewRoutingTable(
		KValue,
		peerID,
		time.Minute,   //TOD, should come from config file
		host.Peerstore())
	cmgr := host.ConnManager()
	rt.PeerAdded = func(p peer.ID) {
		cmgr.TagPeer(p, "kbucket", 5)
	}
	rt.PeerRemoved = func(p peer.ID) {
		cmgr.UntagPeer(p, "kbucket")
	}

	return rt
}

// impl transforms the network interface, which sends and receives
// NetMessage objects, into the ecoball network interface.
type NetImpl struct {
	ctx          context.Context
	host         host.Host

	// inbound messages from the network are forwarded to the receiver
	receiver     Receiver

	sendJbQueue  chan interface{}
	quitsendJb   chan bool

	routingTable *kb.RoutingTable
	rtLock       sync.Mutex

	strmap       map[peer.ID]*messageSender
	strmlk       sync.Mutex

	mdnsService  discovery.Service
	bootstrapper io.Closer
}

func (net *NetImpl)GetPeerID() (peer.ID, error) {
	return net.host.ID(), nil
}

func (net *NetImpl)GetRandomPeers(k int) []peer.ID {
	return net.selectRandomPeers(k)
}

func (net *NetImpl) Host() host.Host {
	return net.host
}

func (net *NetImpl) sendMessage(p pstore.PeerInfo, outgoing message.EcoBallNetMsg) error {
	ms, err := net.messageSenderToPeer(p)
	if err != nil {
		return err
	}

	err = ms.SendMsg(net.ctx, outgoing)

	return err
}

func (net *NetImpl) messageSenderToPeer(p pstore.PeerInfo) (*messageSender, error) {
	net.strmlk.Lock()
	ms, ok := net.strmap[p.ID]
	if ok {
		net.strmlk.Unlock()
		return ms, nil
	}
	ms = NewMsgSender(p, net)
	net.strmap[p.ID] = ms
	net.strmlk.Unlock()

	if err := ms.prepOrInvalidate(); err != nil {
		net.strmlk.Lock()
		defer net.strmlk.Unlock()

		if msCur, ok := net.strmap[p.ID]; ok {
			// Changed. Use the new one, old one is invalid and
			// not in the map so we can just throw it away.
			if ms != msCur {
				return msCur, nil
			}
			// Not changed, remove the now invalid stream from the
			// map.
			delete(net.strmap, p.ID)
		}
		// Invalid but not in map. Must have been removed by a disconnect.
		return nil, err
	}
	// All ready to go.
	return ms, nil
}

func (net *NetImpl) SetDelegate(r Receiver) {
	net.receiver = r
}

func (net *NetImpl) FindPeer(ctx context.Context, id peer.ID) (pstore.PeerInfo, error) {
	// Check if were already connected to them
	if pi := net.findLocal(id); pi.ID != "" {
		return pi, nil
	}

	peers := net.routingTable.NearestPeers(kb.ConvertPeerID(id), AlphaValue)
	if len(peers) == 0 {
		return pstore.PeerInfo{}, kb.ErrLookupFailure
	}

	for _, p := range peers {
		if p == id {
			log.Debug("found target peer in list of closest peers...")
			return net.host.Peerstore().PeerInfo(p), nil
		}
	}

	return pstore.PeerInfo{}, kb.ErrLookupFailure
}

func (net *NetImpl) findLocal(id peer.ID) pstore.PeerInfo {
	switch net.host.Network().Connectedness(id) {
	case inet.Connected, inet.CanConnect:
		return net.host.Peerstore().PeerInfo(id)
	default:
		return pstore.PeerInfo{}
	}
}

// select randomly k peers from remote peers and returns them.
func (net *NetImpl) selectRandomPeers(k int) []peer.ID {
	conns := net.host.Network().Conns()
	if len(conns) < k {
		k = len(conns)
	}
	indices := util.GetRandomIndices(k, len(conns)-1)
	peers := make([]peer.ID, len(indices))
	for i, j := range indices {
		pid := conns[j].RemotePeer()
		peers[i] = pid
	}

	return peers
}

func (net *NetImpl) update(p peer.ID) {
	net.rtLock.Lock()
	defer net.rtLock.Unlock()
	net.routingTable.Update(p)
}

func (net *NetImpl) remove(p peer.ID) {
	net.rtLock.Lock()
	defer net.rtLock.Unlock()
	net.routingTable.Remove(p)
}

func (net *NetImpl) nearestPeersToQuery(id peer.ID, count int) []peer.ID {
	closer := net.routingTable.NearestPeers(kb.ConvertKey(id.String()), count)
	return closer
}

func (net *NetImpl) handleNewStream(s inet.Stream) {
	go net.handleNewStreamMsg(s)
}

func (net *NetImpl) handleNewStreamMsg(s inet.Stream) {
	defer s.Close()
	if net.receiver == nil {
		s.Reset()
		return
	}

	reader := ggio.NewDelimitedReader(s, inet.MessageSizeMax)
	for {
		received, err := message.FromPBReader(reader)
		if err != nil {
			if err != io.EOF {
				s.Reset()
				go net.receiver.ReceiveError(err)
				log.Error(fmt.Sprintf("p2p net from %s %s", s.Conn().RemotePeer(), err))
			}
			return
		}

		p := s.Conn().RemotePeer()
		ctx := context.Background()
		log.Debug("p2p net handleNewStream from ", s.Conn().RemotePeer())
		net.update(p)
		net.receiver.ReceiveMessage(ctx, p, received)
	}
}

func (net *NetImpl) SendMsgJob(job *message.SendMsgJob) {
	net.sendJbQueue <- job
}

func (net *NetImpl) handleSendJob() {
	go func() {
		for {
			select {
			case <-net.quitsendJb:
				return
			case job, ok := <- net.sendJbQueue:
				if !ok {
					log.Error("chan for sending job queue was closed")
					return
				}
				sendJb, ok := job.(*message.SendMsgJob)
				if ok {
					for _, pi := range sendJb.Peers {
						if pi.ID == net.host.ID() {
							continue
						}
/*
						addr := net.host.Peerstore().Addrs(pi.ID)
						if len(addr) == 0 && len(pi.Addrs) >0 {
							if err := net.host.Connect(net.ctx, *pi); err != nil {
								log.Error(err)
								continue
							}
						}
*/
						if err:= net.sendMessage(*pi, sendJb.Msg); err != nil {
							log.Error("send message to ", pi.ID.Pretty(), err)
						}
					}
				}
			}
		}
	}()
}

func (net *NetImpl) startLocalDiscovery() (discovery.Service, error) {
	service, err := discovery.NewMdnsService(net.ctx, net.host, 10*time.Second, ServiceTag)
	if err != nil {
		return nil, fmt.Errorf("net discovery error,", err)
	}
	service.RegisterNotifee((*netNotifiee)(net))

	return service, nil
}

// Start network
func (bsnet *NetImpl) Start() {
	// it is up to the requirement of network sharding,
	if config.EnableLocalDiscovery {
		var err error
		bsnet.mdnsService, err = bsnet.startLocalDiscovery()
		if err != nil {
			log.Error(err)
		}
		log.Debug("start p2p local discovery")
	}

	bsnet.bootstrapper = bsnet.bootstrap(config.SwarmConfig.BootStrapAddr)

	bsnet.handleSendJob()
}

// Stop network
func (net *NetImpl) Stop() {
	if net.mdnsService != nil  {
		net.mdnsService.Close()
	}

	if net.bootstrap != nil {
		net.bootstrapper.Close()
	}

	net.host.Network().StopNotify((*netNotifiee)(net))

	net.quitsendJb <- true
}