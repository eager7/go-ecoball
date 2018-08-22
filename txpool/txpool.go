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
)

var log = elog.NewLogger("TxPool", elog.DebugLog)

type TxPool struct {
	ledger    ledger.Ledger
	PendingTx *types.TxsList //UnPackaged list of legitimate transactions
	txsCache  *lru.Cache
}

//start transaction pool
func Start(ledger ledger.Ledger) (pool *TxPool, err error) {
	csc, err := lru.New(10000)
	if err != nil {
		return nil, errors.New(log, fmt.Sprintf("New Lru error:%s", err.Error()))
	}
	//transaction pool
	pool = &TxPool{ledger: ledger, PendingTx: types.NewTxsList(), txsCache: csc}

	//transaction pool actor
	if _, err = NewTxPoolActor(pool); nil != err {
		pool = nil
	}

	return
}

func (t *TxPool) GetTxsList(chainID common.Hash) ([]*types.Transaction, error) {
	txs := t.PendingTx.GetTransactions()
	return txs, nil
}