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
	"github.com/ecoball/go-ecoball/core/types"
	"sync"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"fmt"
)

type Worker struct {
	ledger ledger.Ledger
	mutex    sync.RWMutex
	workerID uint8
	txList   *types.TxsList
	recCh    chan *types.Transaction
	stopCh   chan bool
}


func NewWorker(workID uint8, ledger ledger.Ledger) *Worker {
	w := &Worker{workerID: workID, recCh: make(chan *types.Transaction, 1000), stopCh: make(chan bool, 1), txList: types.NewTxsList(), ledger:ledger}
	return w
}

func (w *Worker)Run(trx *types.Transaction) {
	w.recCh <- trx
}

func (w *Worker) Start() {
	for {
		select {
		case tx, ok := <-w.recCh:
			if ok {
				fmt.Println("Start PreHandle Transaction")
				ret, cpu, net, err := w.ledger.PreHandleTransaction(tx.ChainID, tx, tx.TimeStamp)
				if err != nil {
					log.Warn(tx.JsonString())
					log.Error("handle tx error:", err)
					continue
				}
				log.Debug(ret, cpu, net, err)
			}
		}
	}
}

