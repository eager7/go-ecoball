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
	"sync"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/shard"
)

const magicNum = 999

type PoolActor struct {
	txPool *TxPool

	wg     sync.WaitGroup
	worker map[string]Worker
}

func NewTxPoolActor(pool *TxPool, n uint8) (pid *actor.PID, err error) {
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
	case *message.RegChain:
		log.Info("Add New TxList:", msg.ChainID.HexString())
		p.txPool.AddTxsList(msg.ChainID)
	case *shard.MinorBlock:
		for _, v := range msg.Transactions {
			log.Info("Delete tx:", v.Hash.HexString())
			p.txPool.Delete(msg.ChainID, v.Hash)
		}
	case *shard.FinalBlock:
	case *shard.CMBlock:
	default:
		log.Warn("unknown type message:", msg, "type", reflect.TypeOf(msg))
	}
}

//Determine whether a transaction already exists
func (p *PoolActor) isSameTransaction(hash common.Hash) bool {
	if p.txPool.txsCache.Contains(hash) {
		return true
	}
	return false
}

func (p *PoolActor) handleTransaction(tx *types.Transaction) error {
	if p.txPool.txsCache.Contains(tx.Hash) {
		log.Warn("transaction already in the txn pool" + tx.Hash.HexString())
		return nil
	}
	p.txPool.txsCache.Add(tx.Hash, nil)

	if !config.DisableSharding {
		lastCMBlock, err := p.txPool.ledger.GetLastShardBlock(tx.ChainID, shard.HeCmBlock)
		if err != nil {
			return err
		}
		numShard := len(lastCMBlock.GetObject().(shard.CMBlock).Shards)
		if numShard == 0 {
			log.Warn("the node network is not work")
			return nil
		}
		log.Notice("the shard number is ", numShard)
		var handle bool
		shardId, err := p.txPool.ledger.GetShardId(tx.ChainID)
		if err != nil {
			return err
		}
		log.Info("the shard id is ", shardId)
		if tx.Type == types.TxTransfer || tx.Addr == common.NameToIndex("root") {
			if uint64(shardId) == uint64(tx.From)%magicNum%uint64(numShard) + 1{
				log.Info("put the tx ", tx.Hash.HexString(), "to txPool")
				handle = true
			}
		} else {
			if uint64(shardId) == uint64(tx.Addr)%magicNum%uint64(numShard) + 1{
				handle = true
			}
		}
		if handle || config.DisableSharding {
			//ret, cpu, net, err := p.txPool.ledger.ShardPreHandleTransaction(tx.ChainID, tx, tx.TimeStamp)
			//if err != nil {
			//	log.Warn(tx.JsonString())
			//	return err
			//}
			//log.Debug(ret, cpu, net, err)
			p.txPool.Push(tx.ChainID, tx)
		}
	} else {
		//ret, cpu, net, err := p.txPool.ledger.PreHandleTransaction(tx.ChainID, tx, tx.TimeStamp)
		//if err != nil {
		//	log.Warn(tx.JsonString())
		//	return err
		//}
		//log.Debug(ret, cpu, net, err)
		p.txPool.Push(tx.ChainID, tx)
	}



	if err := event.Send(event.ActorNil, event.ActorP2P, tx); nil != err {
		log.Warn("broadcast transaction failed:", err.Error(), tx.Hash.HexString())
	}

	//ctx.Sender().Request(cpu, ctx.Self())

	return nil
}

func (p *PoolActor) handleNewBlock(block *types.Block) {
	for _, v := range block.Transactions {
		log.Info("Delete tx:", v.Hash.HexString())
		p.txPool.Delete(block.ChainID, v.Hash)
	}
}
