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
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/event"
	commonMsg "github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"reflect"
	"github.com/ecoball/go-ecoball/sharding/common"
)

type netActor struct {
	props *actor.Props
	node  *Node
}

func NewNetActor(node *Node) *netActor {
	return &netActor{
		node: node,
	}
}

func (n *netActor) Start() (*actor.PID, error) {
	n.props = actor.FromProducer(func() actor.Actor { return n })
	netPid, err := actor.SpawnNamed(n.props, "net")
	event.RegisterActor(event.ActorP2P, netPid)
	return netPid, err
}

func (n *netActor) Receive(ctx actor.Context) {
	log.Debug("Actor receive msg:", reflect.TypeOf(ctx.Message()))

	switch msg := ctx.Message().(type) {
	case *actor.Started:
		log.Debug("NetActor started")
	case commonMsg.Transaction:
		msgType := pb.MsgType_APP_MSG_TRN
		buffer, _ := msg.Tx.Serialize()
		if msg.ShardID == n.node.shardInfo.GetLocalId() {
			netMsg := message.New(msgType, buffer)
			log.Debug("send transactions in shard")
			n.node.broadCastCh <- netMsg
		} else {
			m := message.New(msgType, buffer)
			peerMap := n.node.shardInfo.GetShardNodes(msg.ShardID)
			if peerMap == nil {
				log.Error(fmt.Sprintf("can't find shard[%d] nodes", msg.ShardID))
				return
			}
			var peers []peer.ID
			for node := range peerMap.Iterator() {
				peers = append(peers, node.PeerInfo.ID)
			}
			log.Debug("send transaction to ", peers)
			n.node.network.SendMsgToPeersWithId(peers, m)
		}
	case *common.ShardingTopo:
		go n.node.updateShardingInfo(msg)
	case *types.Block: //not shard block
		msgType := pb.MsgType_APP_MSG_BLKS
		buffer, _ := msg.Serialize()
		netMsg := message.New(msgType, buffer)
		n.node.broadCastCh <- netMsg

	default:
		log.Error("unknown message ", reflect.TypeOf(ctx.Message()))
	}

}
