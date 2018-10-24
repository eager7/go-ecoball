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
package network

import (
	"time"
	"fmt"
	"io"
	"context"
	"sync"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/common/config"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	ggio "gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/io"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p/p2p/discovery"
)

const (
	sendWorkerCount        = 4

	sendMessageTimeout     = time.Minute * 10

	discoveryConnTimeout   = time.Second * 30

	gossipMsgTTL           = time.Second * 100

	ServiceTag                 = "_net-discovery._udp"

	ProtocolP2pV1  protocol.ID = "/ecoball/app/1.0.0"
)

var (
	log = elog.NewLogger("network", elog.DebugLog)
	netImpl *NetImpl
)

func NewNetwork(ctx context.Context, host host.Host) EcoballNetwork {
	if netImpl != nil {
		return netImpl
	}
	netImpl = &NetImpl{
		ctx:          ctx,
		host:         host,
		engine:       NewMsgEngine(ctx, host.ID()),
		strmap:       make(map[peer.ID]*messageSender),
		gossipStore:  NewMsgStore(ctx, gossipMsgTTL),
		quitsendJb:   make(chan bool, 1),
	}

	netImpl.routingTable = NewRouteTable(netImpl)

	host.SetStreamHandler(ProtocolP2pV1, netImpl.handleNewStream)
	host.Network().Notify((*netNotifiee)(netImpl))

	return netImpl
}

func GetNetInstance() EcoballNetwork {
	return netImpl
}

// impl transforms the network interface, which sends and receives
// NetMessage objects, into the ecoball network interface.
type NetImpl struct {
	ctx          context.Context
	host         host.Host

	// inbound messages from the network are forwarded to the receiver
	receiver     Receiver
	// outbound message engine
	engine       *MsgEngine
	quitsendJb   chan bool	
	strmap       map[peer.ID]*messageSender
	strmlk       sync.Mutex

	gossipStore  MsgStore

	mdnsService  discovery.Service
	bootstrapper *BootStrapper

	routingTable *NetRouteTable
}

func (net *NetImpl)GetPeerID() (peer.ID, error) {
	return net.host.ID(), nil
}

func (net *NetImpl) Host() host.Host {
	return net.host
}

func (net *NetImpl) SetDelegate(r Receiver) {
	net.receiver = r
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
				log.Error(fmt.Sprintf("error from %s, %s", s.Conn().RemotePeer(), err))
			}
			return
		}

		p := s.Conn().RemotePeer()
		ctx := context.Background()
		net.routingTable.update(p)
		if received.Type() == message.APP_MSG_GOSSIP {
			net.preHandleGossipMsg(received, p)
			msg, err := net.unwarpGossipMsg(received)
			if err != nil {
				log.Error(err)
			} else {
				net.receiver.ReceiveMessage(ctx, p, msg)
			}
		} else {
			net.receiver.ReceiveMessage(ctx, p, received)
		}
	}
}

func (net *NetImpl) preHandleGossipMsg(gmsg message.EcoBallNetMsg, sender peer.ID) {
	log.Debug(fmt.Sprintf("receive a gossip msg(id=%d) from peer %s", gmsg.Type(), sender.Pretty()))

	peers := net.getRandomPeers(GossipPeerCount, net.receiver.IsNotMyShard)

	var fwPeers []peer.ID
	for _, peer := range peers {
		if peer != sender {
			fwPeers = append(fwPeers, peer)
		}
	}

	// targets is null or there is a same gossip message in the store
	if len(fwPeers) == 0 || !net.gossipStore.Add(gmsg) {
		log.Debug("terminate a gossip message")
		return
	}

	net.forwardMsg(gmsg, fwPeers)
}

func (net *NetImpl) AddMsgJob(job *SendMsgJob) {
	net.engine.PushJob(job)
}

func (net *NetImpl) startSendWorkers() {
	for i:=0; i<sendWorkerCount; i++ {
		i := i
		go net.sendWorker(i)
	}
}

func (net *NetImpl) sendWorker(id int) {
	//log.Debug("network send message worker ", id, " start.")
	defer log.Debug("network send message worker ", id, " shutting down.")
	for {
		select {
		case nextWrapper := <-net.engine.Outbox():
			select {
			case wriapper, ok := <-nextWrapper:
				if !ok {
					continue
				}
				//log.Debug(fmt.Sprintf("worker %d is going to send a message to %s", id, wriapper.pi.ID.Pretty()))
				if err:= net.sendMessage(wriapper.pi, wriapper.emsg); err != nil {
					log.Error("send message to ", wriapper.pi.ID.Pretty(), err)
				}
			case <-net.ctx.Done():
				return
			}
		case <-net.ctx.Done():
			return
		}
	}
}

func (net *NetImpl) StartLocalDiscovery() (discovery.Service, error) {
	service, err := discovery.NewMdnsService(net.ctx, net.host, 10*time.Second, ServiceTag)
	if err != nil {
		return nil, fmt.Errorf("net discovery error,", err)
	}
	service.RegisterNotifee((*netNotifiee)(net))

	return service, nil
}

// Start network
func (net *NetImpl) Start() {
	if config.DisableSharding {
		net.routingTable.Start()

		if config.EnableLocalDiscovery {
			var err error
			net.mdnsService, err = net.StartLocalDiscovery()
			if err != nil {
				log.Error("start p2p local discovery",err)
			} else {
				log.Debug("start p2p local discovery")
			}
		}
	}

	net.bootstrapper = net.bootstrap(config.SwarmConfig.BootStrapAddr)

	net.startSendWorkers()
}

// Stop network
func (net *NetImpl) Stop() {
	if config.DisableSharding {
		net.routingTable.Stop()

		if net.mdnsService != nil  {
			net.mdnsService.Close()
		}
	}


	if net.bootstrap != nil {
		net.bootstrapper.closer.Close()
	}

	net.host.Network().StopNotify((*netNotifiee)(net))

	net.engine.Stop()

	net.quitsendJb <- true
}