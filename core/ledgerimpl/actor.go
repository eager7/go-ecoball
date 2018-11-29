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
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/consensus/dpos"
	"github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/txpool"
	"time"
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
	case *actor.Stop, *actor.Stopped, *actor.Stopping:
		l.pid.Stop()
	case *actor.Restarting:
	case *types.Block:
		log.Info("receive a block:", msg.Hash.HexString())
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
		log.Info("receive a ", shard.HeaderType(msg.Type()).String(), "block:", msg.Hash().HexString(), "height:", msg.GetHeight())
		chain, ok := l.ledger.ChainTxs[msg.GetChainID()]
		if !ok {
			log.Error(fmt.Sprintf("the chain:%s is not existed", msg.GetChainID().HexString()))
			return
		}
		begin := time.Now().UnixNano()
		if err := chain.SaveShardBlock(msg); err != nil {
			log.Error("save block["+msg.GetChainID().HexString()+"] error:", err)
			break
		}
		end := time.Now().UnixNano()
		t := (end-begin)/1000
		log.Info("save ", shard.HeaderType(msg.Type()).String(), "block["+msg.Hash().HexString()+"]:", t, "us")
		if t > 50000 {
			log.Error("save block maybe trouble:", t)
			//os.Exit(-1)
		}
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
			if txpool.T == nil {
				ctx.Sender().Tell(errors.New(log, "create minor block err the txPool is nil"))
				return
			}
			txs, err := txpool.T.GetTxsList(msg.ChainID)
			if err != nil {
				ctx.Sender().Tell(errors.New(log, fmt.Sprintf("create final block err:%s", err.Error())))
				return
			}
			PACKAGE:
			minorBlock, txs, err := l.ledger.NewMinorBlock(msg.ChainID, txs, time.Now().UnixNano())
			if err != nil {
				log.Warn(errors.New(log, fmt.Sprintf("create minor block err:%s", err.Error())))
				if txs != nil {
					goto PACKAGE
				} else {
					ctx.Sender().Tell(err)
				}
			}
			ctx.Sender().Tell(minorBlock)
		case shard.HeCmBlock:
			log.Warn("the minor block nonsupport create by actor")
		case shard.HeFinalBlock:
			block, err := l.ledger.NewFinalBlock(msg.ChainID, time.Now().UnixNano(), msg.Hashes)
			if err != nil {
				ctx.Sender().Tell(errors.New(log, fmt.Sprintf("create final block err:%s", err.Error())))
				return
			}
			ctx.Sender().Tell(block)
		default:
			log.Error("unknown type:", msg.Type.String())
		}
	case message.CheckBlock:
		result := message.CheckBlock{
			Block:  msg.Block,
			Result: nil,
		}
		switch msg.Block.Type() {
		case uint32(shard.HeMinorBlock):

		case uint32(shard.HeCmBlock):

		case uint32(shard.HeFinalBlock):

		default:
			result.Result = errors.New(log, "unknown header type")
		}
		ctx.Sender().Tell(result)
	default:
		log.Warn("unknown type message:", msg, "type", reflect.TypeOf(msg))
	}
}
