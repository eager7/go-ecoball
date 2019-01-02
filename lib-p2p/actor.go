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
	"github.com/ecoball/go-ecoball/lib-p2p/net"
	"github.com/ecoball/go-ecoball/sharding/common"
	"reflect"
	"sync"
)

type BroadcastMessage struct {
	ShardId uint32
	Message message.EcoMessage
}

type SingleMessage struct {
	PublicKey string
	Address   string
	Port      string
	Message   message.EcoMessage
}

type netActor struct {
	ctx              context.Context
	pid              *actor.PID
	instance         *net.Instance
	singleMessage    chan SingleMessage
	broadcastMessage chan BroadcastMessage
	exit             chan struct{}
}

func (n *netActor) initialize() {
	n.singleMessage = make(chan SingleMessage, 100)
	n.broadcastMessage = make(chan BroadcastMessage, 100)
	n.exit = make(chan struct{})
}

func (n *netActor) finished() {
	close(n.singleMessage)
	close(n.broadcastMessage)
}

func NewNetActor(n *netActor) (err error) {
	n.initialize()
	props := actor.FromProducer(func() actor.Actor {
		return n
	})
	n.pid, err = actor.SpawnNamed(props, "netActor")
	if err != nil {
		return err
	}
	event.RegisterActor(event.ActorP2P, n.pid)
	go n.Engine()
	log.Debug("start net actor:", n.pid)

	return nil
}

func (n *netActor) Receive(ctx actor.Context) {
	log.Debug("Actor receive msg:", reflect.TypeOf(ctx.Message()))

	switch msg := ctx.Message().(type) {
	case *actor.Started:
		log.Debug("NetActor started")
	case *actor.Stop, *actor.Stopped, *actor.Stopping:
		n.exit <- struct{}{}
		n.pid.Stop()
	case message.Transaction:
		n.broadcastMessage <- BroadcastMessage{ShardId: msg.ShardID, Message: msg.Tx}
	case *common.ShardingTopo:
		go n.ConnectToShardingPeers(msg)
	case message.NetPacket:
		n.singleMessage <- SingleMessage{PublicKey: msg.PublicKey, Address: msg.Address, Port: msg.Port, Message: msg.Message}
	default:
		log.Error("unknown message type:", reflect.TypeOf(ctx.Message()))
	}

}

//连接本shard内的节点, 跳过自身，并且和其他节点保持长连接
func (n *netActor) ConnectToShardingPeers(info *common.ShardingTopo) {
	n.instance.ShardInfo.Upgrading(info)
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

func (n *netActor) Engine() {
	log.Debug("start message engine to send message")
	defer n.finished()
	for {
		select {
		case msg := <-n.singleMessage:
			if err := n.instance.SendMessage(msg.PublicKey, msg.Address, msg.Port, msg.Message); err != nil {
				log.Error("send message error:", err)
			}
		case msg := <-n.broadcastMessage:
			if err := n.instance.BroadcastToShard(msg.ShardId, msg.Message); err != nil {
				log.Error("broadcast message error:", err)
			}
		case <-n.ctx.Done():
			log.Warn("lib p2p actor receive quit signal")
			return
		case <-n.exit:
			return
		}
	}
}
