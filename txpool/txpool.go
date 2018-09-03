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
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/hashicorp/golang-lru"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
)

var log = elog.NewLogger("TxPool", elog.DebugLog)

type TxPool struct {
	ledger    ledger.Ledger
	PendingTxs map[common.Hash]*types.TxsList //UnPackaged list of legitimate transactions
	txsCache  *lru.Cache
}

//start transaction pool
func Start(ledger ledger.Ledger) (pool *TxPool, err error) {
	csc, err := lru.New(10000)
	if err != nil {
		return nil, errors.New(log, fmt.Sprintf("New Lru error:%s", err.Error()))
	}
	//transaction pool
	pool = &TxPool{ledger: ledger, txsCache: csc}
	pool.PendingTxs = make(map[common.Hash]*types.TxsList, 1)
	pool.AddTxsList(config.ChainHash)
	//transaction pool actor
	if _, err = NewTxPoolActor(pool, 3); nil != err {
		pool = nil
	}

	return
}

func (t *TxPool) GetTxsList(chainID common.Hash) ([]*types.Transaction, error) {
	list, ok := t.PendingTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("can't find this chain:%s", chainID.HexString()))
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
		return errors.New(log, fmt.Sprintf("can't find this chain:%s", chainID.HexString()))
	}
	list.Push(tx)
	return nil
}

func (t *TxPool) Delete(chainID, txHash common.Hash) error {
	list, ok := t.PendingTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("can't find this chain:%s", chainID.HexString()))
	}
	list.Delete(txHash)
	return nil
}
