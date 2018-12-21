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
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"reflect"
)

type netActor struct {
	props *actor.Props
	node  *netNode
}

func NewNetActor(node *netNode) *netActor {
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
	case *types.Transaction:
		msgType := pb.MsgType_APP_MSG_TRN
		buffer, _ := msg.Serialize()
		netMsg := message.New(msgType, buffer)
		log.Debug("new transactions")
		n.node.broadCastCh <- netMsg

	case *types.Block: //not shard block
		msgType := pb.MsgType_APP_MSG_BLKS
		buffer, _ := msg.Serialize()
		netMsg := message.New(msgType, buffer)
		n.node.broadCastCh <- netMsg

	default:
		log.Error("unknown message ", reflect.TypeOf(ctx.Message()))
	}

}
