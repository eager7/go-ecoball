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
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/txpool"
	"time"
)

var log = elog.NewLogger("Solo", elog.NoticeLog)

type Solo struct {
	account account.Account
	stop    chan struct{}
	msg     <-chan interface{}
	ledger  ledger.Ledger
	txPool  *txpool.TxPool
	Chains  map[common.Hash]common.Address
}

func NewSoloConsensusServer(l ledger.Ledger, txPool *txpool.TxPool, acc account.Account) (solo *Solo, err error) {
	solo = &Solo{ledger: l, stop: make(chan struct{}, 1), txPool: txPool, Chains: make(map[common.Hash]common.Address, 1), account: acc}
	actor := &soloActor{solo: solo}
	if actor.pid, err = NewSoloActor(actor); err != nil {
		return nil, err
	}

	msg := []mpb.Identify{
		mpb.Identify_APP_MSG_BLOCK,
	}

	solo.msg, err = event.Subscribe(msg...)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	//start main chain
	solo.Chains[config.ChainHash] = common.AddressFromPubKey(config.Root.PublicKey)
	go ConsensusWorkerThread(config.ChainHash, solo, solo.Chains[config.ChainHash])

	chains, err := l.GetChainList(config.ChainHash)
	if err != nil {
		return nil, err
	}
	for _, c := range chains {
		m := message.RegChain{ChainID: c.Hash, Address: c.Address, TxHash: c.TxHash}
		if err := event.Send(event.ActorNil, event.ActorConsensusSolo, &m); err != nil {
			log.Error("send reg chain message failed:", err)
		}
	}
	return solo, nil
}

func ConsensusWorkerThread(chainID common.Hash, solo *Solo, addr common.Address) {
	time.Sleep(time.Second * 1)
	t := time.NewTimer(time.Second * 1)
	conData := types.ConsData{Type: types.ConSolo, Payload: &types.SoloData{}}
	root := common.AddressFromPubKey(solo.account.PublicKey)
	startNode := root.Equals(&addr)
	for {
		t.Reset(time.Millisecond * time.Duration(config.TimeSlot))
		select {
		case <-t.C:
			if !startNode {
				continue
			}
			//log.Debug("Request transactions from tx pool[", chainID.HexString(), "]")
			txs, _ := solo.txPool.GetTxsList(chainID)
			if len(txs) == 0 {
				log.Info("no transaction in tx pool, will create a empty block")
				//continue
			}
		PACKAGE:
			block, txs, err := solo.ledger.NewTxBlock(chainID, txs, conData, time.Now().UnixNano())
			if err != nil {
				log.Error(err)
				if txs != nil {
					goto PACKAGE
				} else {
					panic(err)
				}
			}
			if err := block.SetSignature(&config.Root); err != nil {
				log.Error("signature failed:", err)
				continue
			}
			if err := solo.ledger.VerifyTxBlock(chainID, block); err != nil {
				log.Warn(err)
				continue
			}
			if err := block.SetSignature(&solo.account); err != nil {
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
			in, ok := msg.(*mpb.Message)
			if !ok {
				log.Error("can't parse msg")
				continue
			}
			log.Info("receive msg:", in.Identify.String())
			block := new(types.Block)
			if err := block.Deserialize(in.Payload); err != nil {
				log.Error(err)
				continue
			}

		}
	}
}
