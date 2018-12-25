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
	"context"
	"fmt"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p/p2p/discovery"
	ggio "gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/io"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"io"
	"time"
)

const (
	sendWorkerCount = 4

	sendMessageTimeout = time.Minute * 10

	discoveryConnTimeout = time.Second * 30

	gossipMsgTTL = time.Second * 100

	ServiceTag = "_net-discovery._udp"

	ProtocolP2pV1 protocol.ID = "/ecoball/app/1.0.0"
)

var (
	log     = elog.NewLogger("network", elog.DebugLog)
	netImpl *NetImpl
)

// impl transforms the network interface, which sends and receives
// NetMessage objects, into the ecoball network interface.
type NetImpl struct {
	ctx  context.Context
	host host.Host

	// inbound messages from the network are forwarded to the receiver
	receiver Receiver
	// outbound message engine
	engine *MsgEngine

	SenderMap SenderMap

	gossipStore MsgStore

	mdnsService  discovery.Service
	bootStrapper *BootStrapper

	routingTable *NetRouteTable
}

func NewNetwork(ctx context.Context, host host.Host, r Receiver) EcoballNetwork {
	if netImpl != nil {
		return netImpl
	}
	netImpl = &NetImpl{
		ctx:          ctx,
		host:         host,
		receiver:     r,
		engine:       NewMsgEngine(ctx, host.ID()),
		SenderMap:    new(SenderMap).Initialize(),
		gossipStore:  NewMsgStore(ctx, gossipMsgTTL),
		mdnsService:  nil,
		bootStrapper: nil,
		routingTable: nil,
	}
	netImpl.routingTable = NewRouteTable(netImpl)

	host.SetStreamHandler(ProtocolP2pV1, netImpl.handleNewStream)
	host.Network().Notify(netImpl)

	return netImpl
}

func GetNetInstance() (EcoballNetwork, error) {
	if netImpl == nil {
		return nil, fmt.Errorf("network has not been initialized")
	}
	return netImpl, nil
}

func (net *NetImpl) Host() host.Host {
	return net.host
}

func (net *NetImpl) SelectRandomPeers(peerCount uint16) []peer.ID {
	return net.getRandomPeers(int(peerCount), net.receiver.IsNotMyShard)
}

func (net *NetImpl) SetDelegate(r Receiver) {
	net.receiver = r
}

func (net *NetImpl) sendMessage(p pstore.PeerInfo, outgoing message.EcoBallNetMsg) error {
	log.Info("send message to", p.ID.Pretty(), p.Addrs[0].String())
	sender, err := net.NewMessageSender(p)
	if err != nil {
		return err
	}
	err = sender.SendMessage(net.ctx, outgoing)

	return err
}

func (net *NetImpl) handleNewStream(s inet.Stream) {
	go net.handleNewStreamMsg(s)
}

func (net *NetImpl) handleNewStreamMsg(s inet.Stream) {
	log.Info("handleNewStreamMsg:", s.Conn().RemotePeer().Pretty(), s.Conn().RemoteMultiaddr().String())
	defer s.Close()
	if net.receiver == nil {
		log.Error("reset stream")
		s.Reset()
		return
	}

	reader := ggio.NewDelimitedReader(s, inet.MessageSizeMax)
	for {
		received, err := message.FromPBReader(reader)
		if err != nil {
			if err != io.EOF {
				log.Warn("reset stream of ", s)
				log.Error("reset stream")
				s.Reset()
				go net.receiver.ReceiveError(err)
				log.Error(fmt.Sprintf("error from %s, %s", s.Conn().RemotePeer(), err))
			}
			return
		}

		p := s.Conn().RemotePeer()
		ctx := context.Background()
		net.routingTable.update(p)
		if received.Type() == pb.MsgType_APP_MSG_GOSSIP {
			net.preHandleGossipMsg(received, p)
			msg, err := net.unWarpGossipMsg(received)
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
	for _, p := range peers {
		if p != sender {
			fwPeers = append(fwPeers, p)
		}
	}

	// targets is null or there is a same gossip message in the store
	if len(fwPeers) == 0 || !net.gossipStore.Add(gmsg) {
		log.Debug("terminate a gossip message")
		return
	}

	net.forwardMsg(gmsg, fwPeers)
}

func (net *NetImpl) StartLocalDiscovery() (discovery.Service, error) {
	service, err := discovery.NewMdnsService(net.ctx, net.host, 10*time.Second, ServiceTag)
	if err != nil {
		return nil, fmt.Errorf("net discovery error, %s", err)
	}
	service.RegisterNotifee(net)

	return service, nil
}

func (net *NetImpl) Start() {
	if config.DisableSharding {
		net.routingTable.Start()
		if config.EnableLocalDiscovery {
			var err error
			net.mdnsService, err = net.StartLocalDiscovery()
			if err != nil {
				log.Error("start p2p local discovery", err)
			} else {
				log.Debug("start p2p local discovery")
			}
		}
	}

	net.bootStrapper = net.bootstrap(config.SwarmConfig.BootStrapAddr)
	net.startSendWorkers()
}

func (net *NetImpl) Stop() {
	if config.DisableSharding {
		net.routingTable.Stop()
		if net.mdnsService != nil {
			net.mdnsService.Close()
		}
	}
	if net.bootstrap != nil {
		net.bootStrapper.closer.Close()
	}
	net.host.Network().StopNotify(net)
	net.engine.Stop()
	net.gossipStore.Stop()
}
