// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package ledgerimpl

import (
	"reflect"

	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/consensus/dpos"
	"github.com/ecoball/go-ecoball/core/types"
	"time"
	"github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/common/errors"
)

type LedActor struct {
	ledger *LedgerImpl

	pid *actor.PID
}

func NewLedgerActor(l *LedActor) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor {
		return l
	})
	pid, err := actor.SpawnNamed(props, "LedgerActor")
	if err != nil {
		return nil, err
	}
	event.RegisterActor(event.ActorLedger, pid)
	log.Debug("start ledger actor:", pid)

	return pid, nil
}

func (l *LedActor) SetLedger(ledger *LedgerImpl) {
	l.ledger = ledger
}

func (l *LedActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Stop:
		l.pid.Stop()
	case *actor.Restarting:
	case *types.Block:
		chain, ok := l.ledger.ChainTxs[msg.ChainID]
		if !ok {
			log.Error(fmt.Sprintf("the chain:%s is not existed", msg.ChainID.HexString()))
			return
		}
		begin := time.Now().UnixNano()
		if err := chain.SaveBlock(msg); err != nil {
			log.Error("save block["+msg.ChainID.HexString()+"] error:", err)
			break
		}
		end := time.Now().UnixNano()
		log.Info("save block["+msg.ChainID.HexString()+"block hash:"+msg.Hash.HexString()+"]:", (end-begin)/1000, "us")
	case shard.BlockInterface:
		chain, ok := l.ledger.ChainTxs[msg.GetChainID()]
		if !ok {
			log.Error(fmt.Sprintf("the chain:%s is not existed", msg.GetChainID().HexString()))
			return
		}
		begin := time.Now().UnixNano()
		if err := chain.SaveShardBlock(0, msg); err != nil {
			log.Error("save block["+msg.GetChainID().HexString()+"] error:", err)
			break
		}
		end := time.Now().UnixNano()
		log.Info("save block["+msg.GetChainID().HexString()+"]:", (end-begin)/1000, "us")
	case *dpos.DposBlock:
		//TODO
		if err := event.Send(event.ActorLedger, event.ActorTxPool, msg.Block); err != nil {
			log.Error("send block to tx pool error:", err)
		}
	case *message.RegChain:
		log.Info("add new chain:", msg.ChainID.HexString())
		if err := l.ledger.NewTxChain(msg.ChainID, msg.Address, false); err != nil {
			log.Error(err)
		}
	case message.ProducerBlock:
		log.Debug("receive create block request")
		switch msg.Type {
		case shard.HeMinorBlock:

		case shard.HeCmBlock:
			log.Warn("the minor block nonsupport create by actor")
		case shard.HeFinalBlock:
			block, err := l.ledger.CreateFinalBlock(msg.ChainID, time.Now().UnixNano())
			if err != nil {
				ctx.Sender().Tell(errors.New(log, fmt.Sprintf("create final block err:%s", err.Error())))
			}
			ctx.Sender().Tell(block)
		default:
			log.Error("unknown type:", msg.Type.String())
		}
	default:
		log.Warn("unknown type message:", msg, "type", reflect.TypeOf(msg))
	}
}
