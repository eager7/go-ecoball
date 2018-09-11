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
	"github.com/ecoball/go-ecoball/core/types"
	"bytes"
	"time"
	netMsg "github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/common/event"
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

type ServiceABABFT struct {
	// Actor *ActorABABFT // save the actor object
	pid   *actor.PID
	ledger ledger.Ledger
	account *account.Account
	txPool *txpool.TxPool
	// mapPID map[common.Hash]*actor.PID  // for multi-chain
	mapActor map[common.Hash]*ActorABABFT
	mapNewChainBlk map[common.Hash]types.Header
	// msgChan    <-chan interface{} // only the main chain can generate the subchain
	// mapMsgChan map[common.Hash]chan actor.Context
	mapMsgChan map[common.Hash]chan interface{}
	mapMsgstop map[common.Hash]chan struct{}
	// stop   chan struct{}
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
	chainHash := config.ChainHash
	serviceABABFT = new(ServiceABABFT)
	actorABABFT := &ActorABABFT{}

	pid, err = ActorABABFTGen(chainHash,actorABABFT)
	if err != nil {
		return nil, err
	}
	serviceABABFT.pid = pid

	actorABABFT.pid = pid
	actorABABFT.status = 1
	actorABABFT.serviceABABFT = serviceABABFT
	actorABABFT.addressRoot = common.AddressFromPubKey(config.Root.PublicKey)
	address := common.AddressFromPubKey(account.PublicKey)
	actorABABFT.startNode = actorABABFT.addressRoot.Equals(&address)
	// serviceABABFT.Actor = actorABABFT
	serviceABABFT.mapActor = make(map[common.Hash]*ActorABABFT)
	serviceABABFT.mapActor[chainHash] = actorABABFT
	serviceABABFT.mapNewChainBlk = make(map[common.Hash]types.Header)
	// serviceABABFT.mapMsgChan = make(map[common.Hash]chan actor.Context)
	serviceABABFT.mapMsgChan = make(map[common.Hash]chan interface{})
	serviceABABFT.mapMsgstop = make(map[common.Hash]chan struct{})
	// serviceABABFT.mapPID[chainHash] = pid
	serviceABABFT.ledger = l
	serviceABABFT.account = account
	serviceABABFT.txPool = txPool

	serviceABABFT.mapActor[chainHash].currentLedger = l
	serviceABABFT.mapActor[chainHash].primaryTag = 0
	/*
	// define the channel
	messages := []uint32{
		commonMsg.APP_MSG_BLKS,
	}
	actorABABFT.msgChan, err = dispatcher.Subscribe(messages...)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	*/
	actorABABFT.msgChan = make(chan interface{}, 256)
	serviceABABFT.mapMsgChan[chainHash] = actorABABFT.msgChan
	actorABABFT.msgStop = make(chan struct{})
	serviceABABFT.mapMsgstop[chainHash] = actorABABFT.msgStop

	selfaccountname = common.NameToIndex("root")
	fmt.Println("selfaccountname:",selfaccountname)

	// cache the root account for solo mode
	actorABABFT.soloaccount = *account // here, the soloaccount is the same as ABABFT account

	return serviceABABFT, err
}

func (serviceABABFT *ServiceABABFT) Start() error {
	var err error
	// start the ababft service
	// build the peers list
	// initialization
	// chainHash,_ := common.DoubleHash(config.Root.PublicKey)
	chainHash := config.ChainHash
	serviceABABFT.mapActor[chainHash].currentHeightNum = int(serviceABABFT.mapActor[chainHash].currentLedger.GetCurrentHeight(chainHash))
	serviceABABFT.mapActor[chainHash].verifiedHeight = uint64(serviceABABFT.mapActor[chainHash].currentHeightNum) - 1
	serviceABABFT.mapActor[chainHash].currentHeader = &(serviceABABFT.mapActor[chainHash].currentHeaderData)
	serviceABABFT.mapActor[chainHash].currentHeaderData = *(serviceABABFT.mapActor[chainHash].currentLedger.GetCurrentHeader(chainHash))

	// start the thread for the chain
	go ConsensusABABFTThread(serviceABABFT.mapActor[chainHash])

	log.Debug("service start")
	return err
}

func (serviceABABFT *ServiceABABFT) GenNewChain(chainID common.Hash, root common.Address) {
	// generate the actor
	// add the new actor to the chain map
	// 1. check whether the chain exists
	if _,ok := serviceABABFT.mapActor[chainID]; ok {
		log.Info("the chain is existed:", chainID.HexString())
		return
	}

	// only the original main chain can generate a new chain
	// 2. check the Txblock corresponding to the new chain
	if _,ok := serviceABABFT.mapNewChainBlk[chainID]; ok {
		// 3. check the height
		TxHeight := serviceABABFT.mapNewChainBlk[chainID].Height
		if serviceABABFT.mapActor[config.ChainHash].currentHeader.Height <= TxHeight {
			time.Sleep(time.Second*10)
			return
		}
		// 4. check the header, and check the transaction is in the original main chain
		TxBlock,err := serviceABABFT.mapActor[config.ChainHash].currentLedger.GetTxBlockByHeight(config.ChainHash,TxHeight)
		if err != nil {
			log.Info("Fail to obtain the corresponding block, when generating new chain")
			return
		}
		if ok := bytes.Equal(serviceABABFT.mapNewChainBlk[chainID].Hash.Bytes(),TxBlock.Hash.Bytes()); ok == true {
			// 5. create an Actor for the new chain
			actorABABFT := &ActorABABFT{}
			/*
			var pid *actor.PID

			pid, err = ActorABABFTGen(chainID,actorABABFT)
			if err != nil {
				log.Info("error when create new Actor for new chain:", chainID.HexString())
				return
			}
			actorABABFT.pid = pid
			*/
			actorABABFT.pid = serviceABABFT.pid
			actorABABFT.synStatus = 0
			actorABABFT.TimeoutMSGs = make(map[string]int, 1000)
			actorABABFT.chainID = chainID
			actorABABFT.status = 1
			actorABABFT.serviceABABFT = serviceABABFT
			actorABABFT.addressRoot = root
			address := common.AddressFromPubKey(actorABABFT.serviceABABFT.account.PublicKey)
			actorABABFT.startNode = actorABABFT.addressRoot.Equals(&address)

			// 6. register the new chain
			serviceABABFT.mapActor[chainID] = actorABABFT
			// serviceABABFT.mapPID[chainID] = pid
			/*
			// define the channel
			messages := []uint32{
				commonMsg.APP_MSG_BLKS,
			}

			actorABABFT.msgChan, err = dispatcher.Subscribe(messages...)
			if err != nil {
				log.Error(err)
				return
			}
			*/
			// actorABABFT.msgChan = make(chan actor.Context, 256)
			actorABABFT.msgChan = make(chan interface{}, 256)
			serviceABABFT.mapMsgChan[chainID] = actorABABFT.msgChan
			actorABABFT.msgStop = make(chan struct{})
			serviceABABFT.mapMsgstop[chainID] = actorABABFT.msgStop

			// 7. initialization
			serviceABABFT.mapActor[chainID].currentHeightNum = int(serviceABABFT.mapActor[chainID].currentLedger.GetCurrentHeight(chainID))
			serviceABABFT.mapActor[chainID].verifiedHeight = uint64(serviceABABFT.mapActor[chainID].currentHeightNum) - 1
			serviceABABFT.mapActor[chainID].currentHeader = &(serviceABABFT.mapActor[chainID].currentHeaderData)
			serviceABABFT.mapActor[chainID].currentHeaderData = *(serviceABABFT.mapActor[chainID].currentLedger.GetCurrentHeader(chainID))
			// start the thread for the chain
			go ConsensusABABFTThread(serviceABABFT.mapActor[chainID])

			// 8. start the actor
			event.Send(event.ActorNil, event.ActorConsensus, netMsg.ABABFTStart{chainID})

		} else {
			log.Info("Fail to pass the header check, when generating new chain")
			// delete element from the map
			delete(serviceABABFT.mapNewChainBlk,chainID)
			return
		}
	} else {
		serviceABABFT.mapNewChainBlk[chainID] = *(serviceABABFT.mapActor[config.ChainHash].currentHeader)
		time.Sleep(time.Second * 10)
		// todo
		// following need to be checked in case of infinite recursion
		serviceABABFT.GenNewChain(chainID, root)
		return
	}

	return
}

func (serviceABABFT *ServiceABABFT) Stop() error {
	// stop the ababft
	return nil
}