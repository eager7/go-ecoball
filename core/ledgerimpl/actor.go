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

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/consensus/dpos"
	"github.com/ecoball/go-ecoball/core/types"
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
	case *types.Transaction:
		//log.Info("Receive Transaction:", msg.Hash.HexString())
		err := l.ledger.ChainTx.CheckTransaction(msg)
		if err != nil {
			ctx.Sender().Tell(err)
			break
		}
		ret, cpu, net, err := l.ledger.ChainTx.HandleTransaction(
			l.ledger.ChainTx.TempStateDB,
			msg,
			msg.TimeStamp,
			l.ledger.ChainTx.CurrentHeader.Receipt.BlockCpu,
			l.ledger.ChainTx.CurrentHeader.Receipt.BlockNet)
		log.Debug(ret, cpu, net, err)
		ctx.Sender().Tell(err)
	case message.GetTransaction:
		tx, err := l.ledger.ChainTx.GetTransaction(msg.Key)
		if err != nil {
			log.Error("Get Transaction Failed:", err)
		} else {
			ctx.Sender().Tell(tx)
		}
	case *types.Block:
		if err := l.ledger.ChainTx.SaveBlock(msg); err != nil {
			log.Error("save block error:", err)
			break
		}
	case *dpos.DposBlock:
		//TODO

		if err := event.Send(event.ActorLedger, event.ActorTxPool, msg.Block); err != nil {
			log.Error("send block to tx pool error:", err)
		}
	default:
		log.Warn("unknown type message:", msg, "type", reflect.TypeOf(msg))
	}
}
