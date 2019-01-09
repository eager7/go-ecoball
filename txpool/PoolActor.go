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

package txpool

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/types"
	"reflect"
	"sync"
)

var counter = 0

type PoolActor struct {
	txPool *TxPool
	wg     sync.WaitGroup
	worker map[string]Worker
}

func NewTxPoolActor(pool *TxPool, n uint8) (pid *actor.PID, err error) {
	props := actor.FromProducer(func() actor.Actor { return &PoolActor{txPool: pool} })
	if pid, err = actor.SpawnNamed(props, "TxPoolActor"); nil != err {
		return nil, err
	}
	if err := event.RegisterActor(event.ActorTxPool, pid); err != nil {
		return nil, err
	}
	return
}

func (p *PoolActor) Receive(ctx actor.Context) {
	log.Notice("tx pool receive message:", reflect.TypeOf(ctx.Message()))
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Restarting:
	case *types.Transaction:
		log.Debug("receive tx:", counter, "type:", msg.Type.String(), "Hash:", msg.Hash.HexString())
		counter++
		if err := p.HandleTransaction(msg); err != nil {
			log.Error("pre handle transaction in tx pool failed:", err)
			log.Warn(msg.String())
		}
	case *types.Block:
		log.Debug("new block delete transactions")
		p.handleNewBlock(msg)
	case *message.RegChain:
		log.Debug("Add New TxList:", msg.ChainID.HexString())
		p.txPool.AddTxsList(msg.ChainID)
	case message.DeleteTx:
		p.txPool.Delete(msg.ChainID, msg.Hash)
	default:
		log.Warn("unknown type message:", msg, "type", reflect.TypeOf(msg))
	}
}

func (p *PoolActor) HandleTransaction(tx *types.Transaction) error {
	if p.txPool.txsCache.Contains(tx.Hash) {
		log.Info("transaction already in the tx pool" + tx.Hash.HexString())
		return nil
	}
	var retString string
	defer func() {
		if err := event.PublishCustom(string(retString), tx.Hash.String()); err != nil {
			log.Warn("publish transaction failed:", err)
		}
	}()
	/*if tx.Receipt.IsBeSet() {
		retString = fmt.Sprintf("the trx's receipt is not empty:%s", tx.Receipt.String())
		return errors.New(retString)
	}*/
	if txClone, err := tx.Clone(); err != nil {
		retString = err.Error()
		return err
	} else {
		if ret, err := p.preHandleTransaction(txClone); err != nil {
			retString = err.Error()
			return err
		} else {
			retString = string(ret)
		}
	}
	if err := p.txPool.Push(tx.ChainID, tx); err != nil {
		retString = err.Error()
		return err
	}
	p.txPool.txsCache.Add(tx.Hash, nil)

	if err := event.Send(event.ActorNil, event.ActorP2P, tx); nil != err {
		log.Warn("broadcast transaction failed:", err.Error(), tx.Hash.HexString())
	}

	return nil
}

func (p *PoolActor) handleNewBlock(block *types.Block) {
	for _, v := range block.Transactions {
		log.Debug("Delete tx:", v.Hash.HexString())
		p.txPool.Delete(block.ChainID, v.Hash)
	}

}

func (p *PoolActor) preHandleTransaction(tx *types.Transaction) (ret []byte, err error) {
	s, ok := p.txPool.StateDB[tx.ChainID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("can't find the chain:%s", tx.ChainID.HexString()))
	}
	if ret, _, _, err = p.txPool.ledger.PreHandleTransaction(tx.ChainID, s, tx, tx.TimeStamp); err != nil {
		return nil, err
	}
	return ret, nil
}
