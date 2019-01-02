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
	"context"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/lib-p2p/net"
	"github.com/ecoball/go-ecoball/sharding/common"
	"reflect"
	"sync"
)

type netActor struct {
	ctx      context.Context
	pid      *actor.PID
	instance *net.Instance
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
	case message.Transaction:
		n.instance.BroadcastToShard(msg.ShardID, msg.Tx)
	case *common.ShardingTopo:
		go n.UpdateShardingInfo(msg)

	case message.NetPacket:
		//n.node.network.SendMsgToPeer(msg.Address, msg.Port, msg.PublicKey, msg.Message)
		return //TODO
		data, err := msg.Message.Serialize()
		if err != nil {
			log.Error(err)
		}
		if err := NodeNetWork.SendMessage(msg.PublicKey, msg.Address, msg.Port, &mpb.Message{Identify: msg.Message.Identify(), Payload: data}); err != nil {
			log.Error("send message failed:", err)
		}
	default:
		log.Error("unknown message ", reflect.TypeOf(ctx.Message()))
	}

}

func (n *netActor) UpdateShardingInfo(info *common.ShardingTopo) {
	n.instance.ShardInfo.Purge()
	n.instance.ShardInfo.SetLocalId(uint32(info.ShardId))
	n.instance.ShardInfo.SetLocalPub(info.Pubkey)
	for sid, shard := range info.ShardingInfo {
		for _, member := range shard {
			/*id, err := network.IdFromConfigEncodePublicKey(member.Pubkey)
			if err != nil {
				log.Error("error for getting id from public key")
				continue
			}
			if id == n.instance.Host.ID() {
				n.instance.ShardInfo.SetLocalId(uint32(sid))
			}*/
			n.instance.ShardInfo.AddShardNode(uint32(sid), member.Pubkey, member.Address, member.Port)
		}
	}
	log.Info("the shard info is :", n.instance.ShardInfo.JsonString())
	n.ConnectToShardingPeers()
}

//连接本shard内的节点, 跳过自身，并且和其他节点保持长连接
func (n *netActor) ConnectToShardingPeers() {
	peerMap := n.instance.ShardInfo.GetShardNodes(n.instance.ShardInfo.GetLocalId())
	if peerMap == nil {
		return
	}
	var wg sync.WaitGroup
	for node := range peerMap.Iterator() {
		if node.Pubkey == n.instance.ShardInfo.GetLocalPub() {
			log.Debug("skip self")
			continue
		}
		wg.Add(1)
		go func() {
			log.Debug("connect to peer:", node.Address, node.Port)
			if err := n.instance.Connect(node.Pubkey, node.Address, node.Port); err != nil {
				log.Error("connect to peer:", node.Address, node.Port, "error:", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	log.Debug("finish connecting to sharding peers exit...")
}
