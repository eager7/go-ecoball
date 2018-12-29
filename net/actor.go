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
	"context"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/event"
	commonMsg "github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/net/network"
	"github.com/ecoball/go-ecoball/net/util"
	"github.com/ecoball/go-ecoball/sharding/common"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"reflect"
	"sync"
)

type netActor struct {
	ctx  context.Context
	pid  *actor.PID
	node *Node
}

func NewNetActor(n *netActor) (err error) {
	props := actor.FromProducer(func() actor.Actor {
		return n
	})
	n.pid, err = actor.SpawnNamed(props, "netActor")
	if err != nil {
		return err
	}
	event.RegisterActor(event.ActorP2P, n.pid)
	log.Debug("start net actor:", n.pid)

	return nil
}

func (n *netActor) Receive(ctx actor.Context) {
	log.Debug("Actor receive msg:", reflect.TypeOf(ctx.Message()))

	switch msg := ctx.Message().(type) {
	case *actor.Started:
		log.Debug("NetActor started")
	case commonMsg.Transaction:
		msgType := pb.MsgType_APP_MSG_TRN
		buffer, err := msg.Tx.Serialize()
		if err != nil {
			log.Error(err)
			return
		}
		log.Debug("broadcast transactions to shard:", msg.ShardID)
		if msg.ShardID == n.node.network.ShardInfo.GetLocalId() {
			netMsg := message.New(msgType, buffer)
			n.node.network.BroadCastCh <- netMsg
		} else {
			m := message.New(msgType, buffer)
			peerMap := n.node.network.ShardInfo.GetShardNodes(msg.ShardID)
			if peerMap == nil {
				log.Error(fmt.Sprintf("can't find shard[%d] nodes", msg.ShardID))
				return
			}
			var peers []peer.ID
			for node := range peerMap.Iterator() {
				peers = append(peers, node.PeerInfo.ID)
			}
			log.Debug("send transaction to ", peers)
			for _, p := range peers {
				log.Debug(n.node.network.Host().Peerstore().Addrs(p))
			}
			n.node.network.SendMsgToPeersWithId(peers, m)
		}
	case *common.ShardingTopo:
		go n.UpdateShardingInfo(msg)
	case *types.Block: //not shard block
		msgType := pb.MsgType_APP_MSG_BLKS
		buffer, _ := msg.Serialize()
		netMsg := message.New(msgType, buffer)
		n.node.network.BroadCastCh <- netMsg

	default:
		log.Error("unknown message ", reflect.TypeOf(ctx.Message()))
	}

}

func (n *netActor) UpdateShardingInfo(info *common.ShardingTopo) {
	n.node.network.ShardInfo.Purge()
	for sid, shard := range info.ShardingInfo {
		for _, member := range shard {
			id, err := network.IdFromConfigEncodePublicKey(member.Pubkey)
			if err != nil {
				log.Error("error for getting id from public key")
				continue
			}
			if id == n.node.network.Host().ID() {
				n.node.network.ShardInfo.SetLocalId(uint32(sid))
			}
			addInfo := util.ConstructAddrInfo(member.Address, member.Port)
			addr, err := multiaddr.NewMultiaddr(addInfo)
			if err != nil {
				log.Error("error for create ip addr from member Info", err)
				continue
			}
			n.node.network.ShardInfo.AddShardNode(uint32(sid), id, addr)
		}
	}
	n.ConnectToShardingPeers()
	log.Info("the shard info is :", n.node.network.ShardInfo.JsonString())
}

//连接本shard内的节点, 跳过自身，并且和其他节点保持长连接
func (n *netActor) ConnectToShardingPeers() {
	peerMap := n.node.network.ShardInfo.GetShardNodes(n.node.network.ShardInfo.GetLocalId())
	if peerMap == nil {
		return
	}
	h := n.node.network.Host()
	var wg sync.WaitGroup
	for node := range peerMap.Iterator() {
		if node.PeerInfo.ID == h.ID() {
			continue
		}
		wg.Add(1)
		go func(p peer.ID, addr []multiaddr.Multiaddr) {
			log.Info("start host connect thread:", p, addr)
			defer wg.Done()
			h.Peerstore().AddAddrs(p, addr, peerstore.PermanentAddrTTL)
			pi := peerstore.PeerInfo{ID: p, Addrs: addr}
			if err := h.Connect(n.ctx, pi); err != nil {
				log.Error("failed to connect peer ", pi, err)
			} else {
				log.Debug("succeed to connect peer ", pi)
			}
		}(node.PeerInfo.ID, node.PeerInfo.Addrs)
	}
	wg.Wait()
	log.Debug("finish connecting to sharding peers exit...")
}
