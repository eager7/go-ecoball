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
//
// The following is the ababft consensus algorithm.
// Author: Xu Wang, 2018.07.16

package ababft

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"fmt"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/txpool"
)

// in this version, the peers take turns to generate the block
const (
	waitResponseTime = 2
)

type stateAbabft byte
const (
	Initialization stateAbabft = 0x00
	Primary        stateAbabft = 0x01
	Backup         stateAbabft = 0x02
)

var selfaccountname common.AccountName
var soloaccount account.Account

type ServiceABABFT struct {
	Actor *ActorAbabft // save the actor object
	pid   *actor.PID
	ledger ledger.Ledger
	account *account.Account
	txPool *txpool.TxPool
}

type PeerInfo struct {
	PublicKey  []byte
	Index      int
}

type PeerAddrInfo struct {
	AccAddress common.Address
	Index      int
}

type PeerInfoAccount struct {
	AccountName common.AccountName
	Index       int
}

func ServiceABABFTGen(l ledger.Ledger, txPool *txpool.TxPool, account *account.Account) (serviceABABFT *ServiceABABFT, err error) {
	var pid *actor.PID

	serviceABABFT = new(ServiceABABFT)

	actorABABFT := &ActorAbabft{}
	pid, err = ActorAbabftGen(actorABABFT)
	if err != nil {
		return nil, err
	}
	actorABABFT.pid = pid
	actorABABFT.status = 1
	actorABABFT.serviceAbabft = serviceABABFT
	serviceABABFT.Actor = actorABABFT
	serviceABABFT.pid = pid
	serviceABABFT.ledger = l
	serviceABABFT.account = account
	serviceABABFT.txPool = txPool

	serviceABABFT.Actor.currentLedger = l
	serviceABABFT.Actor.primaryTag = 0

	selfaccountname = common.NameToIndex("worker1")
	fmt.Println("selfaccountname:",selfaccountname)

	// cache the root account for solo mode
	soloaccount = config.Root

	return serviceABABFT, err
}

func (serviceABABFT *ServiceABABFT) Start() error {
	var err error
	// start the ababft service
	// build the peers list
	// initialization
	serviceABABFT.Actor.currentHeightNum = int(serviceABABFT.Actor.currentLedger.GetCurrentHeight(config.ChainHash))
	serviceABABFT.Actor.verifiedHeight = uint64(serviceABABFT.Actor.currentHeightNum) - 1
	serviceABABFT.Actor.currentHeader = &(serviceABABFT.Actor.currentHeaderData)
	serviceABABFT.Actor.currentHeaderData = *(serviceABABFT.Actor.currentLedger.GetCurrentHeader(config.ChainHash))

	log.Debug("service start")
	return err
}

func (serviceABABFT *ServiceABABFT) Stop() error {
	// stop the ababft
	return nil
}