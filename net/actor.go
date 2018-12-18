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
	"reflect"
	"github.com/AsynkronIT/protoactor-go/actor"
	eactor "github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/rpc"
	"github.com/ecoball/go-ecoball/net/message/pb"
)

type NetActor struct {
	props    *actor.Props
	node   	 *NetNode
}

func NewNetActor(node *NetNode) *NetActor {
	return &NetActor{
		node:     node,
	}
}

func (n *NetActor) Start() (*actor.PID, error) {
	n.props = actor.FromProducer(func() actor.Actor { return n })
	netPid, err := actor.SpawnNamed(n.props, "net")
	eactor.RegisterActor(eactor.ActorP2P, netPid)
	return netPid, err
}

func (n *NetActor) Receive(ctx actor.Context) {
	var buffer []byte
	var msgType pb.MsgType
	msg := ctx.Message()
	switch msg.(type) {
	case *actor.Started:
		log.Debug("NetActor started")
	case *types.Transaction:
		msgType = pb.MsgType_APP_MSG_TRN
		buffer, _ = msg.(*types.Transaction).Serialize()
		netMsg := message.New(msgType, buffer)
		log.Debug("new transactions")
		n.node.broadCastCh <- netMsg
	case *rpc.ListMyIdReq:
		id := n.node.SelfId()
		ctx.Sender().Request(&rpc.ListMyIdRsp{Id:id}, ctx.Self())
	case *rpc.ListPeersReq:
		peers := n.node.Neighbors()
		log.Info(peers)
		ctx.Sender().Request(&rpc.ListPeersRsp{Peer: peers}, ctx.Self())
	case pb.SignaturePreBlockA:
		// broadcast the signature for the previous block
		info,_ := msg.(pb.SignaturePreBlockA)
		msgType = pb.MsgType_APP_MSG_SIGNPRE
		buffer, _ = info.Serialize()
		netMsg := message.New(msgType, buffer)
		n.node.broadCastCh <- netMsg
	case pb.BlockFirstRound:
		// broadcast the first round block
		info,_ := msg.(pb.BlockFirstRound)
		msgType = pb.MsgType_APP_MSG_BLKF
		buffer, _ = info.BlockFirst.Serialize()
		netMsg := message.New(msgType, buffer)
		n.node.broadCastCh <- netMsg
	case pb.REQSynA:
		// broadcast the synchronization request to update the ledger
		info,_ := msg.(pb.REQSynA)
		msgType = pb.MsgType_APP_MSG_REQSYN
		buffer, _ = info.Serialize()
		netMsg := message.New(msgType, buffer)
		n.node.broadCastCh <- netMsg
	case pb.REQSynSolo:
		// broadcast the synchronization request to update the ledger
		info,_ := msg.(pb.REQSynSolo)
		msgType = pb.MsgType_APP_MSG_REQSYNSOLO
		buffer, _ = info.Serialize()
		netMsg := message.New(msgType, buffer)
		n.node.broadCastCh <- netMsg
	case pb.TimeoutMsg:
		info,_ := msg.(pb.TimeoutMsg)
		msgType = pb.MsgType_APP_MSG_TIMEOUT
		// buffer, _ = msg.(*ababft.TimeoutMsg).Serialize()
		buffer, _ = info.Serialize()
		netMsg := message.New(msgType, buffer)
		n.node.broadCastCh <- netMsg
	case pb.SignatureBlkFA:
		// broadcast the signature for the first-round block
		info,_ := msg.(pb.SignatureBlkFA)
		msgType = pb.MsgType_APP_MSG_SIGNBLKF
		buffer, _ = info.Serialize()
		netMsg := message.New(msgType, buffer)
		n.node.broadCastCh <- netMsg
	//case ababft.BlockSecondRound:
	case *types.Block:
		// broadcast the first round block
		msgType = pb.MsgType_APP_MSG_BLKS
		// buffer, _ = msg.(*ababft.BlockSecondRound).blockSecond.Serialize()
		buffer, _ = msg.(*types.Block).Serialize()
		netMsg := message.New(msgType, buffer)
		n.node.broadCastCh <- netMsg
	case pb.BlockSynA:
		// broadcast the block according to the synchronization request
		info,_ := msg.(pb.BlockSynA)
		msgType = pb.MsgType_APP_MSG_BLKSYN
		buffer, _ = info.Serialize()
		netMsg := message.New(msgType, buffer)
		n.node.broadCastCh <- netMsg
	default:
		log.Error("Error Xmit message ", reflect.TypeOf(ctx.Message()))
	}

	log.Debug("Actor receive msg ", reflect.TypeOf(ctx.Message()))
}
