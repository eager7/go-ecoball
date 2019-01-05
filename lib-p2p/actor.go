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
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/lib-p2p/net"
	"github.com/ecoball/go-ecoball/mobsync"
	"reflect"
)

type BroadcastMessage struct {
	Message types.EcoMessage
}

type SingleMessage struct {
	PublicKey string
	Address   string
	Port      string
	Message   types.EcoMessage
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
	if err := event.RegisterActor(event.ActorP2P, n.pid); err != nil {
		return err
	}
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
	case *types.Transaction:
		n.broadcastMessage <- BroadcastMessage{Message: msg}
	case message.NetPacket:
		n.singleMessage <- SingleMessage{PublicKey: msg.PublicKey, Address: msg.Address, Port: msg.Port, Message: msg.Message}
	case *types.Block:
		n.broadcastMessage <- BroadcastMessage{Message: msg}
	case *mobsync.BlockRequest:
		log.Debug(msg.String())
		n.broadcastMessage <- BroadcastMessage{Message: msg}
	case *mobsync.BlockResponse:
		log.Debug(msg.String())
		n.broadcastMessage <- BroadcastMessage{Message: msg}
	default:
		log.Error("unknown message type:", reflect.TypeOf(ctx.Message()))
	}
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
			if err := n.instance.BroadcastToNeighbors(msg.Message); err != nil {
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
