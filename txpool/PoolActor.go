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
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/types"
	"reflect"
	"sync"
	"github.com/ecoball/go-ecoball/common/message/mpb"
)

const magicNum = 999

var CurrentTxN = 0

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
	log.Notice("receive type message:", reflect.TypeOf(ctx.Message()))
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Restarting:
	case *types.Transaction:
		log.Debug("receive tx:", CurrentTxN, "type:", msg.Type.String(), "Hash:", msg.Hash.HexString())
		CurrentTxN++
		go p.handleTransaction(msg)
	case *types.Block:
		log.Debug("new block delete transactions")
		go p.handleNewBlock(msg)
	case *message.RegChain:
		log.Debug("Add New TxList:", msg.ChainID.HexString())
		p.txPool.AddTxsList(msg.ChainID)
	case *shard.MinorBlock:
		for _, v := range msg.Transactions {
			//log.Info("Delete tx:", v.Hash.HexString())
			p.txPool.Delete(msg.ChainID, v.Hash)
		}
	case *shard.FinalBlock:
		if s, err := p.txPool.ledger.StateDB(msg.ChainID).CopyState(); err == nil {
			p.txPool.StateDB[msg.ChainID] = s
			log.Debug("update tx pool state:", s.GetHashRoot().HexString())
		} else {
			log.Warn("update tx pool state error:", err)
		}
	case *shard.CMBlock:
	case message.DeleteTx:
		p.txPool.Delete(msg.ChainID, msg.Hash)
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
	txClone, err := tx.Clone()
	if err != nil {
		event.PublishCustom(err.Error(), tx.Hash.String())
		return err
	}
	if ret, err := p.preHandleTransaction(txClone); err != nil {
		event.PublishCustom(err.Error(), tx.Hash.String())
		return err
	} else {
		event.PublishCustom(string(ret), tx.Hash.String())
	}
	p.txPool.txsCache.Add(tx.Hash, nil)

	lastCMBlock, _, err := p.txPool.ledger.GetLastShardBlock(tx.ChainID, mpb.Identify_APP_MSG_CM_BLOCK)
	if err != nil {
		return err
	}
	numShard := len(lastCMBlock.GetInstance().(*shard.CMBlock).Shards)
	if numShard == 0 {
		log.Warn("the node network is not work, last cm block:", lastCMBlock.String())
		return nil
	}
	var handle bool
	shardId, err := p.txPool.ledger.GetShardId(tx.ChainID)
	if err != nil {
		return err
	}
	log.Debug("the shard id is ", shardId)
	var toShard uint64
	if tx.Type == types.TxTransfer || tx.Addr == common.NameToIndex("root") {
		toShard = uint64(tx.From)%magicNum%uint64(numShard) + 1
		log.Debug("the handle shard id is ", toShard)
		if uint64(shardId) == toShard {
			log.Debug("put the transfer tx:", tx.From, tx.Hash.HexString(), "to txPool")
			handle = true
		}
	} else {
		toShard = uint64(tx.Addr)%magicNum%uint64(numShard) + 1
		log.Debug("the handle shard id is ", toShard)
		if uint64(shardId) == toShard {
			log.Debug("put the contract tx:", tx.Addr, tx.Hash.HexString(), "to txPool")
			handle = true
		}
	}
	if handle {
		p.txPool.Push(tx.ChainID, tx)
	}

	if err := event.Send(event.ActorNil, event.ActorP2P, message.Transaction{ShardID: uint32(toShard), Tx: tx}); nil != err {
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

func (p *PoolActor) preHandleTransaction(tx *types.Transaction) ([]byte, error) {
	s, ok := p.txPool.StateDB[tx.ChainID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("can't find the chain:%s", tx.ChainID.HexString()))
	}
	ret, _, _, err := p.txPool.ledger.ShardPreHandleTransaction(tx.ChainID, s, tx, tx.TimeStamp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
