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

package net

import (
	"bytes"
	"os"
	"fmt"
	"context"

	"github.com/ipfs/go-ipfs/core"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/QmXScvRbYh9X9okLuX9YMnz1HR4WgRTU2hocjBs15nmCNG/go-libp2p-floodsub"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/util"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/net/p2p"
	"github.com/ecoball/go-ecoball/net/ipfs"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/net/dispatcher"
)

type NetCtrl struct {
	IpfsCtrl   *ipfs.IpfsCtrl
	NetNode    *NetNode
	network    p2p.EcoballNetwork
	actor      *NetActor
}

var log = elog.NewLogger("net", elog.DebugLog)

//TODO move to config
var ecoballChainId uint32 = 1

var netCtrl *NetCtrl

type NetNode struct {
	ctx          context.Context
	ipfsNode     *core.IpfsNode
	self         peer.ID
	network      p2p.EcoballNetwork
	broadCastCh  chan message.EcoBallNetMsg
	handlers 	 map[uint32]message.HandlerFunc
	actorId      *actor.PID
	pubSub       *floodsub.PubSub

	//TODO cache check
	//netMsgCache  *lru.Cache
}

func New(parent context.Context, ipfs *core.IpfsNode, network p2p.EcoballNetwork) *NetNode {
	netNode := &NetNode{
		ctx: parent,
		ipfsNode: ipfs,
		self: ipfs.Identity,
		network: network,
		broadCastCh: make(chan message.EcoBallNetMsg, 4 * 1024),//TODO move to config
		handlers: message.MakeHandlers(),
		pubSub: ipfs.Floodsub,
	}
	netNode.broadcastLoop()
	netNode.subTxLoop()
	network.SetDelegate(netNode)
	return netNode
}

func (node *NetNode) SendMsg2RandomPeers(peerCounts int, msg message.EcoBallNetMsg){
	peers := node.SelectRandomPeers(peerCounts)
	for _, pid := range peers {
		err := node.network.SendMessage(context.Background(), pid, msg)
		if err != nil {
			log.Error("send msg to ", pid.Pretty(), err.Error())
		}
	}
}

func (node *NetNode) SendMsg2Peer(pid peer.ID, msg message.EcoBallNetMsg) error{
	err := node.network.SendMessage(context.Background(), pid, msg)
	if err != nil {
		log.Error("send msg to ", pid.Pretty(), err.Error())
	}
	return err
}

func (node *NetNode) SendBroadcastMsg(msg message.EcoBallNetMsg) {
	node.broadCastCh <- msg
}

func (node *NetNode) broadcastLoop() {
	go func() {
		for {
			select {
			case msg := <-node.broadCastCh:
				//TODO cache check
				//node.netMsgCache.Add(msg.DataSum, msg.Size)
				node.broadcastMessage(msg)
			}
		}
	}()
}

func (node *NetNode) broadcastMessage(msg message.EcoBallNetMsg) {
	peers := node.connectedPeerIds()
	for _, pid := range peers {
		err := node.network.SendMessage(context.Background(), pid, msg)
		if err != nil {
			log.Error("send msg to ", pid.Pretty(), err.Error())
		}
	}
}

func (node *NetNode)subTxLoop()  {
	go func() {
		sub, err := node.pubSub.Subscribe("transaction")
		if err != nil {
			return
		}
		self := []byte(node.self)
		for {
			msg, err := sub.Next(context.Background())
			if err != nil {
				return
			}
			if !bytes.Equal(msg.From, self) {
				message.HdTransactionMsg(msg.Data)
			}
		}
	}()
}

func (node *NetNode) connectedPeerIds() []peer.ID  {
	peers := []peer.ID{}
	host := node.ipfsNode.PeerHost
	if host == nil {
		return peers
	}
	conns := host.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		peers = append(peers, pid)
	}
	return peers
}

// select randomly k peers from remote peers and returns them.
func (node *NetNode) SelectRandomPeers(k int) []peer.ID {
	host := node.ipfsNode.PeerHost
	if host == nil {
		return []peer.ID{}
	}

	conns := host.Network().Conns()
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

func (bs *NetNode) ReceiveMessage(ctx context.Context, p peer.ID, incoming message.EcoBallNetMsg) {
	log.Debug("receive msg:",incoming.Type(), "from ", p.Pretty())
	if incoming.Type() > message.APP_MSG_MAX {
		log.Error("receive a invalid message ", incoming.Type())
		return
	}
	handler, ok := bs.handlers[incoming.Type()]
	if ok {
		err := handler(incoming.Data())
		if err != nil {
			log.Error(err.Error())
			return
		}
		dispatcher.Publish(incoming)
	} else {
		dispatcher.Publish(incoming)
		log.Error("publish msg ", incoming.Type())
		return 
	}
}


func (bs *NetNode) ReceiveError(err error) {
	// TODO log the network error
	// TODO bubble the network error up to the parent context/error logger
}

// Connected/Disconnected warns net about peer connections
func (bs *NetNode) PeerConnected(p peer.ID) {
//TODO
}

// Connected/Disconnected warns bitswap about peer connections
func (bs *NetNode) PeerDisconnected(p peer.ID) {
//TODO
}
func (node *NetNode) SelfId() string {
	return node.self.Pretty()
}

func (node *NetNode) SelfRawId() peer.ID {
	return node.self
}

func (node *NetNode) Nbrs() []string  {
	peers := []string{}
	host := node.ipfsNode.PeerHost
	if host == nil {
		return peers
	}
	conns := host.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		peers = append(peers, pid.Pretty())
	}
	return peers
}

func (node *NetNode) SetActorPid(pid *actor.PID) {
	node.actorId = pid
}

func (node *NetNode) GetActorPid() *actor.PID {
	return node.actorId
}

func SetChainId(id uint32)  {
	ecoballChainId = id
}

func GetChainId() uint32 {
	return ecoballChainId
}

func InitNetWork()  {
	//TODO load config
	//configFile, err := ioutil.ReadFile(ConfigFile)
	//if err != nil {
	//
	//}
	//TODO move to config file
	//InitIpfsConfig(path)
	var path = "/tmp/store"

	ipfsCtrl, err := ipfs.InitIpfs(path)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	ipfsNode := ipfsCtrl.IpfsNode
	network := p2p.NewFromIpfsHost(ipfsNode.PeerHost, ipfsNode.Routing)
	netNode := New(context.Background(), ipfsNode, network)
	dispatcher.InitMsgDispatcher(netNode)

	netCtrl = & NetCtrl{
		IpfsCtrl:ipfsCtrl,
		NetNode:netNode,
		network:network,
	}

	fmt.Printf("i am %s \n", netNode.SelfId())
}

func StartNetWork()  {

	netActor := NewNetActor(netCtrl.NetNode)
	// gossiper.Start()
	actorId, _ := netActor.Start()
	netCtrl.NetNode.SetActorPid(actorId)

	//start store repo stat engine
	netCtrl.IpfsCtrl.RepoStat.Start()

	fmt.Printf("node %s is running.\n", netCtrl.NetNode.SelfId())
}
