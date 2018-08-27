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

type ServiceAbabft struct {
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
	AccAdress  common.Address
	Index      int
}

type PeerInfoAccount struct {
	Accountname common.AccountName
	Index       int
}

func ServiceAbabftGen(l ledger.Ledger, txPool *txpool.TxPool, account *account.Account) (serviceAbabft *ServiceAbabft, err error) {
	var pid *actor.PID

	serviceAbabft = new(ServiceAbabft)

	actorAbabft := &ActorAbabft{}
	pid, err = Actor_ababft_gen(actorAbabft)
	if err != nil {
		return nil, err
	}
	actorAbabft.pid = pid
	actorAbabft.status = 1
	actorAbabft.serviceAbabft = serviceAbabft
	serviceAbabft.Actor = actorAbabft
	serviceAbabft.pid = pid
	serviceAbabft.ledger = l
	serviceAbabft.account = account
	serviceAbabft.txPool = txPool

	current_ledger = l
	primary_tag = 0

	selfaccountname = common.NameToIndex("worker1")
	fmt.Println("selfaccountname:",selfaccountname)

	// cache the root account for solo mode
	soloaccount = config.Root

	return serviceAbabft, err
}

func (serviceAbabft *ServiceAbabft) Start() error {
	var err error
	// start the ababft service
	// build the peers list
	// initialization
	current_height_num = int(current_ledger.GetCurrentHeight(config.ChainHash))
	verified_height = uint64(current_height_num) - 1
	currentheader = &currentheader_data
	currentheader_data = *(current_ledger.GetCurrentHeader(config.ChainHash))

	log.Debug("service start")
	return err
}

func (serviceAbabft *ServiceAbabft) Stop() error {
	// stop the ababft
	return nil
}