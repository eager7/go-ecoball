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

var T *TxPool

type TxPool struct {
	netMsg     <-chan interface{}
	ledger     ledger.Ledger
	PendingTxs map[common.Hash]*types.TxsList //UnPackaged list of legitimate transactions
	txsCache   *lru.Cache
	StateDB    map[common.Hash]*state.State
	stop       chan struct{}
}

//start transaction pool
func Start(ledger ledger.Ledger) (pool *TxPool, err error) {
	csc, err := lru.New(10000)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("New Lru error:%s", err.Error()))
	}
	//transaction pool
	pool = &TxPool{
		netMsg:     nil,
		ledger:     ledger,
		PendingTxs: nil,
		txsCache:   csc,
		StateDB:    make(map[common.Hash]*state.State, 0),
		stop:       make(chan struct{}),
	}
	pool.PendingTxs = make(map[common.Hash]*types.TxsList, 1)
	pool.AddTxsList(config.ChainHash)
	topics := []mpb.Identify{
		mpb.Identify_APP_MSG_TRANSACTION,
	}
	pool.netMsg, err = event.Subscribe(topics...)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	s, err := ledger.StateDB(config.ChainHash).CopyState()
	if err != nil {
		return nil, err
	}
	pool.StateDB[config.ChainHash] = s
	if _, err = NewTxPoolActor(pool, 3); nil != err {
		pool = nil
	}
	T = pool
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

func (t *TxPool) Delete(chainID, txHash common.Hash) error {
	list, ok := t.PendingTxs[chainID]
	if !ok {
		return errors.New(fmt.Sprintf("can't find this chain:%s", chainID.HexString()))
	}
	list.Delete(txHash)
	return nil
}

func (t *TxPool) Subscribe() {
	for {
		select {
		case <-t.stop:
			{
				log.Info("stop tx pool")
				return
			}
		case msg := <-t.netMsg:
			in, ok := msg.(*mpb.Message)
			if !ok {
				log.Error("can't parse msg")
				continue
			}
			log.Info("receive msg:", in.Identify.String())
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
