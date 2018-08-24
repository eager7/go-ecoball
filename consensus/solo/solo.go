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

package solo

import (
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"time"
	"github.com/ecoball/go-ecoball/txpool"
)

var log = elog.NewLogger("Solo", elog.NoticeLog)

type Solo struct {
	stop   chan struct{}
	ledger ledger.Ledger
	txPool *txpool.TxPool
}

func NewSoloConsensusServer(l ledger.Ledger, txPool *txpool.TxPool) (*Solo, error) {
	solo := &Solo{ledger: l, stop:make(chan struct{}, 1), txPool:txPool}
	actor := &soloActor{solo: solo}
	NewSoloActor(actor)
	return solo, nil
}

func (s *Solo) Start() error {
	t := time.NewTimer(time.Second * 1)
	conData := types.ConsensusData{Type: types.ConSolo, Payload: &types.SoloData{}}

	go func() {
		for {
			t.Reset(time.Second * 3)
			select {
				case <-t.C:
					log.Debug("Request transactions from tx pool")
					txs, _ := s.txPool.GetTxsList(config.ChainHash)
					block, err := s.ledger.NewTxBlock(config.ChainHash, txs, conData, time.Now().UnixNano())
					if err != nil {
						log.Fatal(err)
					}
					if err := block.SetSignature(&config.Root); err != nil {
						log.Fatal(err)
					}
					if err := event.Send(event.ActorConsensusSolo, event.ActorLedger, block); err != nil {
						log.Fatal(err)
					}
				case <- s.stop: {
					log.Info("Stop Solo Mode")
					return
				}
			}
		}
	}()
	return nil
}
