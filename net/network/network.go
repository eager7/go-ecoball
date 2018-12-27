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
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/net/dispatcher"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p/p2p/discovery"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"time"
)

const (
	sendWorkerCount                  = 4
	sendMessageTimeout               = time.Minute * 10
	discoveryConnTimeout             = time.Second * 30
	gossipMsgTTL                     = time.Second * 100
	ServiceTag                       = "_net-discovery._udp"
	ProtocolP2pV1        protocol.ID = "/ecoball/app/1.0.0"
)

var (
	log      = elog.NewLogger("network", elog.DebugLog)
	instance *NetWork
)

type NetWork struct {
	ctx          context.Context
	host         host.Host
	ShardInfo    *ShardInfo
	BroadCastCh  chan message.EcoBallNetMsg
	engine       *MsgEngine // outbound message engine
	SenderMap    SenderMap
	gossipStore  MsgStore
	mdnsService  discovery.Service
	bootStrapper *BootStrapper
	routingTable *NetRouteTable
}

func NewNetwork(ctx context.Context, host host.Host) *NetWork {
	if instance != nil {
		return instance
	}
	instance = &NetWork{
		ctx:          ctx,
		host:         host,
		ShardInfo:    new(ShardInfo).Initialize(),
		BroadCastCh:  make(chan message.EcoBallNetMsg, 4*1024),
		engine:       NewMsgEngine(ctx, host.ID()),
		SenderMap:    new(SenderMap).Initialize(),
		gossipStore:  NewMsgStore(ctx, gossipMsgTTL),
		mdnsService:  nil,
		bootStrapper: nil,
		routingTable: nil,
	}
	instance.routingTable = NewRouteTable(instance)
	instance.bootStrapper = instance.bootstrap(config.SwarmConfig.BootStrapAddr)

	host.SetStreamHandler(ProtocolP2pV1, instance.NetWorkHandler)
	host.Network().Notify(instance)

	return instance
}

func GetNetInstance() (EcoballNetwork, error) {
	if instance == nil {
		return nil, fmt.Errorf("network has not been initialized")
	}
	return instance, nil
}
func (net *NetWork) Start() {
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
	net.startSendWorkers()
	net.nativeMessageLoop()
}
func (net *NetWork) Stop() {
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

func (net *NetWork) NetWorkHandler(s net.Stream) {
	id := s.Conn().RemotePeer()
	addresses := s.Conn().RemoteMultiaddr()
	log.Info("receive connect peer from:", id.Pretty(), addresses.String())
	go net.HandleNewStream(s)
}
func (net *NetWork) HandleNewStream(s net.Stream) {
	log.Info("start stream handler:", s.Conn().RemotePeer().Pretty(), s.Conn().RemoteMultiaddr().String())
	defer s.Close()
	reader := message.NewReader(s)
	for {
		if received, err := message.FromPBReader(reader); err != nil {
			err := errors.New(fmt.Sprintf("error from %s, %s", s.Conn().RemotePeer(), err))
			log.Error(err)
			s.Reset()
			return
		} else {
			net.routingTable.update(s.Conn().RemotePeer())
			if received.Type() == pb.MsgType_APP_MSG_GOSSIP {
				net.preHandleGossipMsg(received, s.Conn().RemotePeer())
				msg, err := net.unWarpGossipMsg(received)
				if err != nil {
					log.Error(err)
				} else {
					net.ReceiveMessage(msg)
				}
			} else {
				net.ReceiveMessage(received)
			}
		}
	}
}

func (net *NetWork) ReceiveMessage(incoming message.EcoBallNetMsg) error {
	log.Debug(fmt.Sprintf("receive msg type: %s", incoming.Type().String()))
	if incoming.Type() >= pb.MsgType_APP_MSG_UNDEFINED {
		log.Error()
		return errors.New(fmt.Sprintf("receive a invalid message:%s", incoming.Type().String()))
	}

	if err := dispatcher.Publish(incoming); err != nil {
		return err
	}
	return nil
}

func (net *NetWork) preHandleGossipMsg(msg message.EcoBallNetMsg, sender peer.ID) {
	log.Debug(fmt.Sprintf("receive a gossip msg(id=%d) from peer %s", msg.Type(), sender.Pretty()))

	peers := net.getRandomPeers(GossipPeerCount, net.IsNotMyShard)

	var fwPeers []peer.ID
	for _, p := range peers {
		if p != sender {
			fwPeers = append(fwPeers, p)
		}
	}

	// targets is null or there is a same gossip message in the store
	if len(fwPeers) == 0 || !net.gossipStore.Add(msg) {
		log.Debug("terminate a gossip message")
		return
	}

	net.forwardMsg(msg, fwPeers)
}

func (net *NetWork) Host() host.Host {
	return net.host
}
func (net *NetWork) IsNotMyShard(p peer.ID) bool {
	if peerMap := net.ShardInfo.GetShardNodes(net.ShardInfo.GetLocalId()); peerMap == nil {
		return true
	} else {
		return !peerMap.Contains(p)
	}
}
func (net *NetWork) IsValidRemotePeer(p peer.ID) bool {
	return net.ShardInfo.IsValidRemotePeer(p)
}
func (net *NetWork) StartLocalDiscovery() (discovery.Service, error) {
	service, err := discovery.NewMdnsService(net.ctx, net.host, 10*time.Second, ServiceTag)
	if err != nil {
		return nil, fmt.Errorf("net discovery error, %s", err)
	}
	service.RegisterNotifee(net)

	return service, nil
}
func (net *NetWork) Neighbors() (peers []*peerstore.PeerInfo) {
	for _, c := range net.Host().Network().Conns() {
		pid := c.RemotePeer()
		if !net.IsNotMyShard(pid) {
			peers = append(peers, &peerstore.PeerInfo{ID: pid})
		}
	}
	return peers
}
func (net *NetWork) SelectRandomPeers(peerCount uint16) []peer.ID {
	return net.getRandomPeers(int(peerCount), net.IsNotMyShard)
}
