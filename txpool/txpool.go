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
	"context"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/hashicorp/golang-lru"
)

var log = elog.NewLogger("TxPool", elog.NoticeLog)

type TxPool struct {
	ctx        context.Context
	netMsg     <-chan interface{}
	ledger     ledger.Ledger
	txsCache   *lru.Cache
	PendingTxs map[common.Hash]*types.TxsList
	StateDB    map[common.Hash]*state.State
	stop       chan struct{}
}

//start transaction pool
func Start(ctx context.Context, ledger ledger.Ledger) (pool *TxPool, err error) {
	csc, err := lru.New(10000)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("New Lru error:%s", err.Error()))
	}
	//transaction pool
	pool = &TxPool{
		ctx:        ctx,
		netMsg:     nil,
		ledger:     ledger,
		PendingTxs: make(map[common.Hash]*types.TxsList, 1),
		txsCache:   csc,
		StateDB:    make(map[common.Hash]*state.State, 1),
		stop:       make(chan struct{}),
	}
	pool.AddTxsList(config.ChainHash)
	if pool.netMsg, err = event.Subscribe([]mpb.Identify{mpb.Identify_APP_MSG_TRANSACTION}...); err != nil {
		return nil, errors.New(err.Error())
	}
	if pool.StateDB[config.ChainHash], err = ledger.StateDB(config.ChainHash).StateCopy(); err != nil {
		return nil, err
	}
	if _, err = NewTxPoolActor(pool, 3); err != nil {
		return nil, err
	}
	go pool.Subscribe()
	return
}

func (t *TxPool) Stop() {
	t.stop <- struct{}{}
}

func (t *TxPool) GetTxsList(chainID common.Hash) ([]*types.Transaction, error) {
	list, ok := t.PendingTxs[chainID]
	if !ok {
		return nil, errors.New(fmt.Sprintf("can't find this chain:%s", chainID.HexString()))
	}
	txs := list.GetTransactions()
	return txs, nil
}

func (t *TxPool) AddTxsList(hash common.Hash) {
	if _, ok := t.PendingTxs[hash]; ok {
		return
	}
	t.PendingTxs[hash] = types.NewTxsList()
}

func (t *TxPool) Push(chainID common.Hash, tx *types.Transaction) error {
	list, ok := t.PendingTxs[chainID]
	if !ok {
		return errors.New(fmt.Sprintf("can't find this chain:%s", chainID.HexString()))
	}
	list.Push(tx)
	return nil
}

func (t *TxPool) Delete(chainID, txHash common.Hash) {
	if list, ok := t.PendingTxs[chainID]; ok {
		list.Delete(txHash)
	}
}

func (t *TxPool) Subscribe() {
	for {
		select {
		case <-t.stop:
			log.Info("stop tx pool")
			return
		case <-t.ctx.Done():
			log.Info("receive ctx done, stop tx pool")
			return
		case msg := <-t.netMsg:
			in, ok := msg.(*mpb.Message)
			if !ok {
				log.Error("tx pool can't parse msg")
				continue
			}
			log.Info("tx pool receive msg:", in.Identify.String())
			tx := new(types.Transaction)
			if err := tx.Deserialize(in.Payload); err != nil {
				log.Error(err)
				continue
			}
			if err := event.Send(event.ActorNil, event.ActorTxPool, tx); err != nil {
				log.Fatal(err)
			}
		}
	}
}
