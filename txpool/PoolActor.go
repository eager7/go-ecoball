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
	"reflect"

	"github.com/ecoball/go-ecoball/common"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
)

type PoolActor struct {
	txPool *TxPool
}

func NewTxPoolActor(pool *TxPool) (pid *actor.PID, err error) {
	props := actor.FromProducer(func() actor.Actor {
		return &PoolActor{txPool: pool}
	})

	if pid, err = actor.SpawnNamed(props, "TxPoolActor"); nil != err {
		return nil, err
	}
	event.RegisterActor(event.ActorTxPool, pid)

	return
}

func (p *PoolActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Restarting:
	case *types.Transaction:
		log.Info("receive tx:", msg.Hash.HexString())
		go p.handleTransaction(msg)
	case *types.Block:
		log.Debug("new block delete transactions")
		go p.handleNewBlock(msg)
	case common.Hash:
		log.Info("Add New TxList:", msg.HexString())
		p.txPool.AddTxsList(msg)
	default:
		log.Warn("unknown type message:", msg, "type", reflect.TypeOf(msg))
	}
}

//Determine whether a transaction already exists
func (p *PoolActor) isSameTransaction(hash common.Hash) bool {
	//if tr := p.txPool.PendingTx.Same(hash); tr {
	//	return true
	//}
	if p.txPool.txsCache.Contains(hash) {
		return true
	}
	return false
}

func (p *PoolActor) handleTransaction(tx *types.Transaction) error {
	//if exist := p.isSameTransaction(tx.Hash); exist {
	if p.txPool.txsCache.Contains(tx.Hash) {
		log.Warn("transaction already in the txn pool" + tx.Hash.HexString())
		return nil
	}
	p.txPool.txsCache.Add(tx.Hash, nil)

	ret, cpu, net, err := p.txPool.ledger.PreHandleTransaction(tx.ChainID, tx, tx.TimeStamp)
	if err != nil {
		return err
	}
	log.Debug(ret, cpu, net, err)
	//data := tx.Hash.Bytes()
	//for _, v := range tx.Signatures {
	//	if hasSign, err := secp256k1.Verify(data, v.SigData, v.PubKey); nil != err || !hasSign {
	//		log.Warn("check transaction signatures failed:" + tx.Hash.HexString())
	//		return errors.New(log, "check transaction signatures fail:"+tx.Hash.HexString())
	//	}
	//}

	//Send the transaction to ledger to Check legitimacy
	//res, err := event.SendSync(event.ActorLedger, tx, time.Second*2)
	//if nil != err {
	//	log.Warn("send message to ledger actor error: ", tx.Hash.HexString())
	//	return errors.New(log, "send message to ledger actor error: "+tx.Hash.HexString())
	//} else if nil != res {
	//	if err, ok := res.(error); ok {
	//		log.Warn(tx.Hash.HexString(), " Check legitimacy failed: ", err)
	//		return errors.New(log, tx.Hash.HexString()+" Check legitimacy failed: "+fmt.Sprintf("%v", err))
	//	} else {
	//		return errors.New(log, "unidentified message")
	//	}
	//}

	//Verify by adding to the transaction pool
	p.txPool.Push(tx.ChainID, tx)

	//Broadcast transactions on p2p
	if err := event.Send(event.ActorNil, event.ActorP2P, tx); nil != err {
		log.Warn("broadcast transaction failed:" + tx.Hash.HexString())
		//return errors.New(log, "broadcast transaction failed:"+tx.Hash.HexString())
	}

	return nil
}

func (p *PoolActor) handleNewBlock(block *types.Block) {
	for _, v := range block.Transactions {
		p.txPool.Delete(block.ChainID, v.Hash)
	}
}
