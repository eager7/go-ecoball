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
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/dispatcher"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/txpool"
	"time"
)

var log = elog.NewLogger("Solo", elog.NoticeLog)

type Solo struct {
	stop   chan struct{}
	msg    <-chan interface{}
	ledger ledger.Ledger
	txPool *txpool.TxPool
	Chains map[common.Hash]common.Hash
}

func NewSoloConsensusServer(l ledger.Ledger, txPool *txpool.TxPool) (solo *Solo, err error) {
	solo = &Solo{ledger: l, stop: make(chan struct{}, 1), txPool: txPool, Chains: make(map[common.Hash]common.Hash, 1)}
	actor := &soloActor{solo: solo}
	NewSoloActor(actor)

	messages := []uint32{
		message.APP_MSG_BLK,
	}

	solo.msg, err = dispatcher.Subscribe(messages...)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return solo, nil
}

func (s *Solo) Start(chainID common.Hash) error {
	t := time.NewTimer(time.Second * 1)
	conData := types.ConsensusData{Type: types.ConSolo, Payload: &types.SoloData{}}

	go func() {
		for {
			t.Reset(time.Second * 300)
			select {
			case <-t.C:
				log.Debug("Request transactions from tx pool")
				txs, _ := s.txPool.GetTxsList(chainID)
				block, err := s.ledger.NewTxBlock(chainID, txs, conData, time.Now().UnixNano())
				if err != nil {
					log.Fatal(err)
				}
				if err := block.SetSignature(&config.Root); err != nil {
					log.Fatal(err)
				}
				if err := event.Send(event.ActorConsensusSolo, event.ActorLedger, block); err != nil {
					log.Fatal(err)
				}
			case <-s.stop:
				{
					log.Info("Stop Solo Mode")
					return
				}
			case msg := <-s.msg:
				fmt.Println("receive msg:", msg)
			}
		}
	}()
	return nil
}

func ConsensusWorkerThread(chainID common.Hash, solo *Solo) {
	t := time.NewTimer(time.Second * 1)
	conData := types.ConsensusData{Type: types.ConSolo, Payload: &types.SoloData{}}
	for {
		t.Reset(time.Second * 3)
		select {
		case <-t.C:
			log.Debug("Request transactions from tx pool")
			txs, _ := solo.txPool.GetTxsList(chainID)
			if len(txs) == 0 {
				log.Info("no transaction in this time")
				continue
			}
			block, err := solo.ledger.NewTxBlock(chainID, txs, conData, time.Now().UnixNano())
			if err != nil {
				log.Fatal(err)
			}
			if err := block.SetSignature(&config.Root); err != nil {
				log.Fatal(err)
			}
			if err := event.Send(event.ActorConsensusSolo, event.ActorLedger, block); err != nil {
				log.Fatal(err)
			}
		case <-solo.stop:
			{
				log.Info("Stop Solo Mode")
				return
			}
		case msg := <-solo.msg:
			fmt.Println("receive msg:", msg)
		}
	}
}
