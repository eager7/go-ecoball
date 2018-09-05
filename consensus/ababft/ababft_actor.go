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
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common"
	"time"
	"github.com/ecoball/go-ecoball/core/pb"
	"bytes"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
	"fmt"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/account"
	"encoding/binary"
	"sort"
	"github.com/ecoball/go-ecoball/common/config"
	"reflect"
)
type ActorABABFT struct {
	status uint // 1: actor generated,
	// 2: running,
	// 3: as prime, start the new round, collect the tx and previous block signature, then broadcast the first round block
	// 4: as prime, start collect the tx signature and generate the new block, then broadcast
	// 5: as peer, start the new round, signature the current newest block and broadcast
	// 6: as peer, wait for the new block generation, and then update the local ledger
	// 7: as prime, the round end and enters to the next round
	// 8: as peer, the round end and enters to the next round
	// 101: solo prime, before main net start
	// 102: solo peer, before main net start
	pid              *actor.PID // actor pid
	serviceABABFT    *ServiceABABFT
	NumPeers         int
	PeersAddrList    []PeerAddrInfo       // Peer address information for consensus
	PeersListAccount []PeerInfoAccount // Peer information for consensus

	selfIndex int                       // the index of this peer in the peers list
	currentRoundNum int                // current round number
	currentHeightNum      int               // current height, according to the blocks saved in the local ledger
	currentLedger         ledger.Ledger
	currentHeader         *types.Header // temporary parameters for the current block header, according to the blocks saved in the local ledger
	currentHeaderData     types.Header
	signaturePreBlockList []common.Signature // list for saving the signatures for the previous block
	signatureBlkFList     []common.Signature // list for saving the signatures for the first round block
	blockFirstRound       BlockFirstRound    // temporary parameters for the first round block
	blockSecondRound      BlockSecondRound   // temporary parameters for the second round block
	cacheSignaturePreBlk []pb.SignaturePreblock // cache the received signatures for the previous block
	blockFirstCal *types.Block                    // cache the first-round block
	TimeoutMSGs map[string]int      // cache the timeout message

	verifiedHeight uint64
	primaryTag int // 0: verification peer; 1: is the primary peer, who generate the block at current round;
	deltaRoundNum int
	receivedSignPreNum int                      // the number of received signatures for the previous block
	receivedSignBlkFNum int                     // temporary parameters for received signatures for first round block
	synStatus int
	chainID common.Hash  // for multi-chain
}

const(
	pubKeyTag   = "ababft"
	signDataTag = "ababft"
)

var log = elog.NewLogger("ABABFT", elog.NoticeLog)

// to run the go test, please set TestTag to True
const TestTag = false

const ThresholdRound = 60

// var Num_peers int
// var Peers_list []PeerInfo                // Peer information for consensus
// var Peers_addr_list []PeerAddrInfo       // Peer address information for consensus
// var Peers_list_account []PeerInfoAccount // Peer information for consensus
// var Self_index int                       // the index of this peer in the peers list
// var current_round_num int                // current round number
// var current_height_num int               // current height, according to the blocks saved in the local ledger
// var current_ledger ledger.Ledger

// var primary_tag int // 0: verification peer; 1: is the primary peer, who generate the block at current round;
// var signature_preblock_list [][]byte // list for saving the signatures for the previous block
// var signature_preblock_list []common.Signature // list for saving the signatures for the previous block
// var signature_BlkF_list [][]byte // list for saving the signatures for the first round block
// var signature_BlkF_list []common.Signature // list for saving the signatures for the first round block
// var block_firstround BlockFirstRound // temporary parameters for the first round block
// var block_secondround BlockSecondRound // temporary parameters for the second round block
// var currentheader *types.Header // temporary parameters for the current block header, according to the blocks saved in the local ledger
// var currentheader_data types.Header
// var current_payload types.AbaBftData              // temporary parameters for current payload
// var received_signpre_num int                      // the number of received signatures for the previous block
// var cache_signature_preblk []pb.SignaturePreblock // cache the received signatures for the previous block
// var blockFirstCal *types.Block                    // cache the first-round block
// var received_signblkf_num int                     // temporary parameters for received signatures for first round block
// var TimeoutMsgs = make(map[string]int, 1000)      // cache the timeout message
// var verified_height uint64

// var delta_roundnum int

// var syn_status int
// for test 2018.07.31
var Accounts_test []account.Account
// test end



func ActorABABFTGen(chainId common.Hash, actorABABFT *ActorABABFT) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor {
		return actorABABFT
	})
	chainStr := string("ABABFT")
	chainStr += chainId.HexString()
	pid, err := actor.SpawnNamed(props, chainStr)
	if err != nil {
		return nil, err
	}
	event.RegisterActor(event.ActorConsensus, pid)
	actorABABFT.synStatus = 0
	actorABABFT.TimeoutMSGs = make(map[string]int, 1000)
	actorABABFT.chainID = chainId
	return pid, err
}

func (actorC *ActorABABFT) Receive(ctx actor.Context) {
	var err error
	// log.Debug("ababft service receives the message")

	// deal with the message
	switch msg := ctx.Message().(type) {
	case message.ABABFTStart:
		actorC.status = 2
		log.Debug("start ababft: receive the ababftstart message:", actorC.currentHeightNum,actorC.verifiedHeight,actorC.currentLedger.GetCurrentHeader(config.ChainHash))

		// check the status of the main net
		if ok:=actorC.currentLedger.StateDB(config.ChainHash).RequireVotingInfo(); ok!=true {
			// main net has not started yet
			// currentheader = current_ledger.GetCurrentHeader()

			actorC.currentHeightNum = int(actorC.currentHeader.Height)
			actorC.currentRoundNum = 0
			actorC.verifiedHeight = uint64(actorC.currentHeightNum)

			log.Debug("ababft is in solo mode!")
			// if soloaccount.PrivateKey != nil {
			if config.StartNode == true {
				// is the solo prime
				actorC.status = 101
				// generate the solo block
				// consensus data
				var signPreSend []common.Signature
				signPreSend = append(signPreSend, actorC.currentHeader.Signatures[0])
				conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{NumberRound:uint32(actorC.currentRoundNum), PreBlockSignatures:signPreSend,}}
				// tx list
				/*
				value, err := event.SendSync(event.ActorTxPool, message.GetTxs{}, time.Second*1)
				if err != nil {
					log.Error("AbaBFT Consensus tx error:", err)
					return
				}
				txList, ok := value.(*types.TxsList)
				if !ok {
					return
				}
				var txs []*types.Transaction
				for _, v := range txList.Txs {
					txs = append(txs, v)
				}
				*/
				txs, _ := actorC.serviceABABFT.txPool.GetTxsList(config.ChainHash)


				// generate the block in the form of second round block
				var blockSolo *types.Block
				tTime := time.Now().UnixNano()
				blockSolo,err = actorC.serviceABABFT.ledger.NewTxBlock(config.ChainHash, txs, conData, tTime)
				blockSolo.SetSignature(&soloaccount)
				actorC.blockSecondRound.BlockSecond = *blockSolo
				// save (the ledger will broadcast the block after writing the block into the DB)
				/*
				if err = actorC.serviceABABFT.ledger.SaveTxBlock(blockSolo); err != nil {
					// log.Error("save block error:", err)
					println("save solo block error:", err)
					return
				}
				*/
				if err := event.Send(event.ActorNil, event.ActorP2P, blockSolo); err != nil {
					log.Fatal(err)
					// return
				}
				// currentheader = blockSolo.Header
				actorC.currentHeaderData = *(blockSolo.Header)
				actorC.currentHeader = &actorC.currentHeaderData
				actorC.verifiedHeight = blockSolo.Height
				if err := event.Send(event.ActorNil, event.ActorLedger, blockSolo); err != nil {
					log.Fatal(err)
					// return
				}

				fmt.Println("ababft solo height:", blockSolo.Height, blockSolo)
				time.Sleep(time.Second * waitResponseTime)
				// call itself again
				event.Send(event.ActorNil,event.ActorConsensus,message.ABABFTStart{actorC.chainID})
			} else {
				// is the solo peer
				actorC.status = 102
				// todo
				// no need every time to send a request for solo block

				// send solo syn request
				var requestsyn REQSynSolo
				requestsyn.Reqsyn.PubKey = actorC.serviceABABFT.account.PublicKey
				hashTS,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentHeightNum+1)))
				requestsyn.Reqsyn.SigData,_ = actorC.serviceABABFT.account.Sign(hashTS.Bytes())
				requestsyn.Reqsyn.RequestHeight = uint64(actorC.currentHeightNum)+1
				event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
				log.Info("send solo block requirements:", requestsyn.Reqsyn.RequestHeight, actorC.currentHeightNum)
			}
			return
		}

		// initialization
		// clear and initialize the signature preblock array

		// update the peers list by accountname
		newPeers,err := actorC.currentLedger.GetProducerList(config.ChainHash)
		if err != nil {
			log.Debug("fail to get peer list.")
		}
		log.Debug("ababft now enter into the ababft mode:",newPeers[0],newPeers[1])

		actorC.NumPeers = len(newPeers)
		var peersListAccountTS = make([]string, actorC.NumPeers)
		for i := 0; i < actorC.NumPeers; i++ {
			// peersListAccountTS = append(peersListAccountTS,common.IndexToName(newPeers[i]))
			peersListAccountTS[i] = newPeers[i].String()
		}
		log.Debug("ababft now enter into the ababft mode:peersListAccountTS", peersListAccountTS)
		// sort newPeers
		sort.Strings(peersListAccountTS)

		actorC.PeersListAccount = make([]PeerInfoAccount, actorC.NumPeers)
		actorC.PeersAddrList = make([]PeerAddrInfo, actorC.NumPeers)
		for i := 0; i < actorC.NumPeers; i++ {
			actorC.PeersListAccount[i].AccountName = common.NameToIndex(peersListAccountTS[i])
			actorC.PeersListAccount[i].Index = i + 1

			accountInfo,err := actorC.currentLedger.AccountGet(config.ChainHash, actorC.PeersListAccount[i].AccountName)
			if err != nil {
				log.Debug("fail to get account info.")
			}
			accAddrInfo := accountInfo.Permissions["owner"]
			if len(accAddrInfo.Keys)!=1 {
				log.Debug("owner address must be 1 for BP node!")
			}
			for _, k := range accAddrInfo.Keys {
				// check the address is correct
				addrKeys := k.Actor
				// todo
				// the next check can be delete
				if ok := bytes.Equal(accAddrInfo.Keys[addrKeys.HexString()].Actor.Bytes(), addrKeys.Bytes()); ok == true {
					// save the address instead of pukey
					actorC.PeersAddrList[i].AccAddress = addrKeys
					break
				}
			}
			actorC.PeersAddrList[i].Index = i + 1

			if uint64(selfaccountname) == uint64(actorC.PeersListAccount[i].AccountName) {
				// update Self_index, i.e. the corresponding index in the peer address list
				actorC.selfIndex = i + 1
			}
		}

		fmt.Println("Peers_addr_list:",actorC.PeersAddrList)

		actorC.signaturePreBlockList = make([]common.Signature, len(actorC.PeersAddrList))
		actorC.signatureBlkFList = make([]common.Signature, len(actorC.PeersAddrList))
		actorC.blockFirstRound = BlockFirstRound{}
		actorC.blockSecondRound = BlockSecondRound{}
		// log.Debug("current_round_num:",current_round_num,Num_peers,Self_index)
		// get the current round number of the block
		// currentheader = current_ledger.GetCurrentHeader()

		actorC.currentHeightNum = int(actorC.currentHeader.Height)

		// todo
		// check following patch:
		// add ThresholdRound to solve the liveness problem
		latestRoundNum := int(actorC.currentHeader.ConsensusData.Payload.(*types.AbaBftData).NumberRound)
		actorC.deltaRoundNum = actorC.currentRoundNum - latestRoundNum
		if actorC.deltaRoundNum > ThresholdRound && actorC.currentHeightNum > int(actorC.verifiedHeight) {
			// as there is a long time since last block, maybe the chain is blocked somewhere
			// to generate the block after the previous block (i.e. the latest verified block)
			var currentBlock *types.Block
			currentBlock,err = actorC.currentLedger.GetTxBlock(config.ChainHash, actorC.currentHeader.PrevHash)
			if err != nil {
				fmt.Println("get previous block error.")
			}
			// currentheader = currentBlock.Header
			actorC.currentHeaderData = *(currentBlock.Header)
			actorC.currentHeader = &actorC.currentHeaderData

			actorC.currentHeightNum = actorC.currentHeightNum - 1

			// todo
			// 1. the ledger needs one backward step
			// 2. the peer list also needs one backward step
			// 3. the txpool also needs one backward step or maybe not
			// 4. the blockchain in database needs one backward step
		}

		if actorC.currentHeader.ConsensusData.Type != types.ConABFT {
			//log.Warn("wrong ConsensusData Type")
			return
		}
		/*
		if v,ok:= actorC.currentHeader.ConsensusData.Payload.(* types.AbaBftData); ok {
			current_payload = *v
		}
		*/
		// todo
		// the update of current_round_num
		// current_round_num = int(current_payload.NumberRound)
		// the timeout/changeview message
		// need to check whether the update of current_round_num is necessary


		// signature the current highest block and broadcast
		var signaturePreblock common.Signature
		signaturePreblock.PubKey = actorC.serviceABABFT.account.PublicKey
		signaturePreblock.SigData, err = actorC.serviceABABFT.account.Sign(actorC.currentHeader.Hash.Bytes())
		if err != nil {
			return
		}

		// check whether self is the prime or peer
		if actorC.currentRoundNum % actorC.NumPeers == (actorC.selfIndex-1) {
			// if is prime
			actorC.primaryTag = 1
			actorC.status = 3
			actorC.receivedSignPreNum = 0
			// increase the round index
			actorC.currentRoundNum ++
			fmt.Println("ABABFTStart:current_round_num:",actorC.currentRoundNum,actorC.selfIndex)
			// log.Debug("primary")
			// set up a timer to wait for the signaturePreblock from other peera
			t0 := time.NewTimer(time.Second * waitResponseTime * 2)
			go func() {
				select {
				case <-t0.C:
					// timeout for the preblock signature
					err = event.Send(event.ActorConsensus, event.ActorConsensus, PreBlockTimeout{})
					t0.Stop()
				}
			}()
		} else {
			// is peer
			actorC.primaryTag = 0
			actorC.status = 5
			// broadcast the signaturePreblock and set up a timer for receiving the data
			var signaturePreSend SignaturePreBlock
			signaturePreSend.SignPreBlock.PubKey = signaturePreblock.PubKey
			signaturePreSend.SignPreBlock.SigData = signaturePreblock.SigData
			// todo
			// for the signature of previous block, maybe the round number is not needed
			signaturePreSend.SignPreBlock.Round = uint32(actorC.currentRoundNum)
			signaturePreSend.SignPreBlock.Height = uint32(actorC.currentHeader.Height)
			// broadcast
			event.Send(event.ActorConsensus, event.ActorP2P, signaturePreSend)
			// increase the round index
			actorC.currentRoundNum ++
			fmt.Println("ABABFTStart:current_round_num(non primary):",actorC.currentRoundNum,actorC.selfIndex)
			// log.Debug("non primary")
			// log.Debug("signaturePreSend:",current_round_num,currentheader.Height,signaturePreSend)
			// set up a timer for receiving the data
			t1 := time.NewTimer(time.Second * waitResponseTime * 2)
			go func() {
				select {
				case <-t1.C:
					// timeout for the preblock signature
					err = event.Send(event.ActorConsensus, event.ActorConsensus, TxTimeout{})
					t1.Stop()
				}
			}()
		}
		return

	case SignaturePreBlock:
		log.Info("receive the preblock signature:", actorC.status,msg.SignPreBlock)
		if actorC.status == 102 {
			event.Send(event.ActorNil,event.ActorConsensus,message.ABABFTStart{actorC.chainID})
		}
		// the prime will verify the signature for the previous block
		roundIn := int(msg.SignPreBlock.Round)
		heightIn := int(msg.SignPreBlock.Height)
		// log.Debug("current_round_num:",current_round_num,roundIn)
		if roundIn >= actorC.currentRoundNum && actorC.status!=101 && actorC.status!= 102 {
			// cache the SignaturePreBlock
			actorC.cacheSignaturePreBlk = append(actorC.cacheSignaturePreBlk,msg.SignPreBlock)
			// in case that the signature for the previous block arrived bofore the corresponding block generator was born
		}
		if actorC.primaryTag == 1 && (actorC.status == 2 || actorC.status == 3){
			// verify the signature
			// first check the round number and height

			// todo
			// maybe round number is not needed for preblock signature
			if roundIn >= (actorC.currentRoundNum-1) && heightIn >= actorC.currentHeightNum {
				if roundIn > (actorC.currentRoundNum - 1) && heightIn > actorC.currentHeightNum {
					// todo
					// need to check
					// only require when height difference between the peers is >= 2

					if actorC.deltaRoundNum > ThresholdRound {
						if actorC.verifiedHeight == uint64(actorC.currentHeightNum) && heightIn == (actorC.currentHeightNum+1) {
							return
						}
					}

					// require synchronization, the longest chain is ok
					// send synchronization message
					var requestSyn REQSyn
					requestSyn.Reqsyn.PubKey = actorC.serviceABABFT.account.PublicKey
					hashTS,_ := common.DoubleHash(Uint64ToBytes(actorC.verifiedHeight+1))
					requestSyn.Reqsyn.SigData,_ = actorC.serviceABABFT.account.Sign(hashTS.Bytes())
					requestSyn.Reqsyn.RequestHeight = actorC.verifiedHeight+1
					event.Send(event.ActorConsensus,event.ActorP2P, requestSyn)
					actorC.synStatus = 1
					// todo
					// attention
					// to against the height cheat, do not change the actorC.status
				} else {
					// check the signature
					pubKeyIn := msg.SignPreBlock.PubKey // signaturepre_send.signaturePreblock.PubKey = signaturePreblock.PubKey
					// check the pubKeyIn is in the peer list
					var foundPeer bool
					foundPeer = false
					var peerIndex int

					/*
					for index,peer := range Peers_list {
						if ok := bytes.Equal(peer.PublicKey, pubKeyIn); ok == true {
							foundPeer = true
							peerIndex = index
							break
						}
					}
					*/
					// change public key to account address
					for index, peerAddr := range actorC.PeersAddrList {
						peerAddrIns := common.AddressFromPubKey(pubKeyIn)
						if ok := bytes.Equal(peerAddr.AccAddress.Bytes(), peerAddrIns.Bytes()); ok == true {
							foundPeer = true
							peerIndex = index
							break
						}
					}

					if foundPeer == false {
						// the signature is not from the peer in the list
						return
					}
					// 1. check that signature in or not in list of
					if actorC.signaturePreBlockList[peerIndex].SigData != nil {
						// already receive the signature
						return
					}
					// 2. verify the correctness of the signature
					sigDataIn := msg.SignPreBlock.SigData
					headerHashes := actorC.currentHeader.Hash.Bytes()
					var resultVerify bool
					resultVerify, err = secp256k1.Verify(headerHashes, sigDataIn, pubKeyIn)
					if resultVerify == true {
						// add the incoming signature to signature preblock list
						actorC.signaturePreBlockList[peerIndex].SigData = sigDataIn
						actorC.signaturePreBlockList[peerIndex].PubKey = pubKeyIn
						actorC.receivedSignPreNum ++
					} else {
						return
					}
					// log.Debug("signature_preblock_list",signature_preblock_list)
				}
			} else {
				// the message is old
				return
			}
		} else {
			return
		}

	case PreBlockTimeout:
		if actorC.primaryTag == 1 && (actorC.status == 2 || actorC.status == 3){
			// 1. check the cache cache_signature_preblk
			headerHashes := actorC.currentHeader.Hash.Bytes()
			for _, signPreBlk := range actorC.cacheSignaturePreBlk {
				roundIn := signPreBlk.Round
				if int(roundIn) != actorC.currentRoundNum {
					continue
				}
				// check the signature
				pubKeyIn := signPreBlk.PubKey // signaturepre_send.signaturePreblock.PubKey = signaturePreblock.PubKey
				// check the pubKeyIn is in the peer list
				var foundPeer bool
				foundPeer = false
				var peerIndex int
				/*
				for index,peer := range Peers_list {
					if ok := bytes.Equal(peer.PublicKey, pubKeyIn); ok == true {
						foundPeer = true
						peerIndex = index
						break
					}
				}
				*/
				// change public key to account address
				for index, peerAddr := range actorC.PeersAddrList {
					peerAddrIns := common.AddressFromPubKey(pubKeyIn)
					if ok := bytes.Equal(peerAddr.AccAddress.Bytes(), peerAddrIns.Bytes()); ok == true {
						foundPeer = true
						peerIndex = index
						break
					}
				}

				if foundPeer == false {
					// the signature is not from the peer in the list
					continue
				}
				// first check that signature in or not in list of
				if actorC.signaturePreBlockList[peerIndex].SigData != nil {
					// already receive the signature
					continue
				}
				// second, verify the correctness of the signature
				signDataIn := signPreBlk.SigData
				var resultVerify bool
				resultVerify, err = secp256k1.Verify(headerHashes, signDataIn, pubKeyIn)
				if resultVerify == true {
					// add the incoming signature to signature preblock list
					actorC.signaturePreBlockList[peerIndex].SigData = signDataIn
					actorC.signaturePreBlockList[peerIndex].PubKey = pubKeyIn
					actorC.receivedSignPreNum ++
				} else {
					continue
				}

			}
			// clean the cache_signature_preblk
			// cache_signature_preblk = make([]pb.SignaturePreblock,len(Peers_list)*2)
			actorC.cacheSignaturePreBlk = make([]pb.SignaturePreblock,len(actorC.PeersAddrList)*2)
			// fmt.Println("valid sign_pre:",received_signpre_num)
			// fmt.Println("current status root hash:",currentheader.StateHash)

			// 2. check the number of the preblock signature
			if actorC.receivedSignPreNum >= int(len(actorC.PeersAddrList)/3+1) {
				// enough preblock signature, so generate the first-round block, only including the preblock signatures and
				// prepare the ConsensusData
				var signPreSend []common.Signature
				for _, signPre := range actorC.signaturePreBlockList {
					/*
					if signPre != nil {
						var sign_tmp common.Signature
						sign_tmp.SigData = signPre
						sign_tmp.PubKey = Peers_list[index].PublicKey
						signPreSend = append(signPreSend, sign_tmp)
					}
					*/
					// change public key to account address
					if signPre.SigData != nil {
						var signTmp common.Signature
						signTmp.SigData = signPre.SigData
						signTmp.PubKey = signPre.PubKey
						signPreSend = append(signPreSend, signTmp)
					}
				}
				conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{NumberRound:uint32(actorC.currentRoundNum), PreBlockSignatures:signPreSend,}}
				// fmt.Println("conData for blk firstround",conData)
				// prepare the tx list
				/*
				value, err := event.SendSync(event.ActorTxPool, message.GetTxs{}, time.Second*1)
				// log.Debug("tx value:",value)
				if err != nil {
					log.Error("AbaBFT Consensus error:", err)
					return
				}
				txList, ok := value.(*types.TxsList)
				if !ok {
					// log.Error("The format of value error [solo]")
					return
				}

				var txs []*types.Transaction
				for _, v := range txList.Txs {
					txs = append(txs, v)
				}*/
				txs, _ := actorC.serviceABABFT.txPool.GetTxsList(config.ChainHash)
				// log.Debug("obtained tx list", txs[0])
				// generate the first-round block
				var blockFirst *types.Block
				tTime := time.Now().UnixNano()
				blockFirst,err = actorC.serviceABABFT.ledger.NewTxBlock(config.ChainHash, txs, conData, tTime)
				blockFirst.SetSignature(actorC.serviceABABFT.account)
				// broadcast the first-round block to peers for them to verify the transactions and wait for the corresponding signatures back
				actorC.blockFirstRound.BlockFirst = *blockFirst
				event.Send(event.ActorConsensus, event.ActorP2P, actorC.blockFirstRound)
				// log.Debug("first round block:",block_firstround.BlockFirst)
				// fmt.Println("first round block status root hash:",blockFirst.StateHash)
				log.Info("generate the first round block and send",actorC.blockFirstRound.BlockFirst.Height)

				// for test 2018.07.27
				if TestTag == true {
					event.Send(event.ActorNil,event.ActorConsensus,actorC.blockFirstRound)
					// log.Debug("first round block:",block_firstround.BlockFirst.Header)
				}
				// test end


				// change the statue
				actorC.status = 4
				// initial the received_signblkf_num to count the signatures for txs (i.e. the first round block)
				actorC.receivedSignBlkFNum = 0
				// set the timer for collecting the signature for txs (i.e. the first round block)
				t2 := time.NewTimer(time.Second * waitResponseTime)
				go func() {
					select {
					case <-t2.C:
						// timeout for the preblock signature
						err = event.Send(event.ActorConsensus, event.ActorConsensus, SignTxTimeout{})
						t2.Stop()
					}
				}()
			} else {
				// did not receive enough preblock signature in the assigned time interval
				actorC.status = 7
				actorC.primaryTag = 0 // reset to zero, and the next primary will take the turn
				// send out the timeout message
				var timeoutMsg TimeoutMsg
				timeoutMsg.Toutmsg.RoundNumber = uint64(actorC.currentRoundNum)
				timeoutMsg.Toutmsg.PubKey = actorC.serviceABABFT.account.PublicKey
				hashTS,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentRoundNum)))
				timeoutMsg.Toutmsg.SigData,_ = actorC.serviceABABFT.account.Sign(hashTS.Bytes())
				event.Send(event.ActorConsensus,event.ActorP2P, timeoutMsg)
				// start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{actorC.chainID})
			}
		} else {
			return
		}

	case BlockFirstRound:
		// for test 2018.07.27
		if TestTag == true {
			actorC.primaryTag = 0
			actorC.status = 5
			// log.Debug("debug for first round block")
		}
		// end of test

		log.Info("current height and receive the first round block:",actorC.currentHeightNum, msg.BlockFirst.Header)

		if actorC.primaryTag == 0 && (actorC.status == 2 || actorC.status == 5) {
			// to verify the first round block
			blockFirstReceived := msg.BlockFirst
			// the protocal type is ababft
			if blockFirstReceived.ConsensusData.Type == types.ConABFT {
				dataPreBlkReceived := blockFirstReceived.ConsensusData.Payload.(*types.AbaBftData)
				// 1. check the round number
				// 1a. current round number
				if dataPreBlkReceived.NumberRound < uint32(actorC.currentRoundNum) {
					return
				} else if dataPreBlkReceived.NumberRound > uint32(actorC.currentRoundNum) {
					// require synchronization, the longest chain is ok
					// in case that somebody may skip the current generator, only the different height can call the synchronization
					if (actorC.verifiedHeight+2) < blockFirstReceived.Header.Height {
						// send synchronization message
						var requestSyn REQSyn
						requestSyn.Reqsyn.PubKey = actorC.serviceABABFT.account.PublicKey
						hashTS,_ := common.DoubleHash(Uint64ToBytes(actorC.verifiedHeight+1))
						requestSyn.Reqsyn.SigData,_ = actorC.serviceABABFT.account.Sign(hashTS.Bytes())
						requestSyn.Reqsyn.RequestHeight = actorC.verifiedHeight+1
						event.Send(event.ActorConsensus,event.ActorP2P, requestSyn)
						actorC.synStatus = 1
						// todo
						// attention:
						// to against the height cheat, do not change the actorC.status
					}
				} else {
					// 1b. the round number corresponding to the block generator
					indexG := (actorC.currentRoundNum-1) % actorC.NumPeers + 1
					pubKeyGIn := blockFirstReceived.Signatures[0].PubKey
					var indexGIn int
					indexGIn = -1
					/*
					for _, peer := range Peers_list {
						if ok := bytes.Equal(peer.PublicKey, pubKeyGIn); ok == true {
							indexGIn = int(peer.Index)
							break
						}
					}
					*/
					// change public key to account address
					for _, peerAddr := range actorC.PeersAddrList {
						peerAddrIns := common.AddressFromPubKey(pubKeyGIn)
						if ok := bytes.Equal(peerAddr.AccAddress.Bytes(), peerAddrIns.Bytes()); ok == true {
							indexGIn = int(peerAddr.Index)
							break
						}
					}
					if indexG != indexGIn {
						// illegal block generator
						return
					}
					// 1c. check the block header, except the consensus data
					var validBlk bool
					validBlk,err = actorC.verifyHeader(&blockFirstReceived, actorC.currentRoundNum, *(actorC.currentHeader))
					if validBlk ==false {
						println("header check fail")
						return
					}
					// 2. check the preblock signature
					signPreBlkList := dataPreBlkReceived.PreBlockSignatures
					headerHashes := actorC.currentHeader.Hash.Bytes()
					var numVerified int
					numVerified = 0
					for index, signPreBlk := range signPreBlkList {
						// 2a. check the peers in the peer list
						var peerInTag bool
						peerInTag = false
						/*
						for _, peer := range Peers_list {
							if ok := bytes.Equal(peer.PublicKey, signPreBlk.PubKey); ok == true {
								peerInTag = true
								break
							}
						}
						*/
						// change public key to account address
						for _, peerAddr := range actorC.PeersAddrList {
							peerAddrIns := common.AddressFromPubKey(signPreBlk.PubKey)
							if ok := bytes.Equal(peerAddr.AccAddress.Bytes(), peerAddrIns.Bytes()); ok == true {
								peerInTag = true
								break
							}
						}
						if peerInTag == false {
							// there exists signature not from the peer list
							fmt.Println("the signature is not from the peer list, its index is:", index)
							return
						}
						// 2b. verify the correctness of the signature
						pubKeyIn := signPreBlk.PubKey
						signDataIn := signPreBlk.SigData
						var resultVerify bool
						resultVerify, err = secp256k1.Verify(headerHashes, signDataIn, pubKeyIn)
						if resultVerify == true {
							numVerified++
						}
					}
					// 2c. check the valid signature number
					if numVerified < int(len(actorC.PeersAddrList)/3+1){
						// not enough signature
						fmt.Println("not enough signature for second round block")
						return
					}
					// 3. check the txs
					txsIn := blockFirstReceived.Transactions
					for index1, txIn := range txsIn {
						err = actorC.serviceABABFT.ledger.CheckTransaction(config.ChainHash, txIn)
						if err != nil {
							println("wrong tx, index:", index1)
							return
						}
					}
					// 4. sign the received first-round block
					var signBlkFSend SignatureBlkF
					signBlkFSend.signatureBlkF.PubKey = actorC.serviceABABFT.account.PublicKey
					signBlkFSend.signatureBlkF.SigData,err = actorC.serviceABABFT.account.Sign(blockFirstReceived.Header.Hash.Bytes())
					// 5. broadcast the signature of the first round block
					event.Send(event.ActorConsensus, event.ActorP2P, signBlkFSend)
					// 6. change the status
					actorC.status = 6
					// fmt.Println("signBlkFSend:",signBlkFSend)
					// clean the cache_signature_preblk
					actorC.cacheSignaturePreBlk = make([]pb.SignaturePreblock,len(actorC.PeersAddrList)*2)
					// send the received first-round block to other peers in case that network is not good
					actorC.blockFirstRound.BlockFirst = blockFirstReceived
					event.Send(event.ActorConsensus,event.ActorP2P,actorC.blockFirstRound)
					log.Info("generate the signature for first round block",actorC.blockFirstRound.BlockFirst.Height)

					// for test 2018.07.31
					if TestTag == true {
						actorC.primaryTag = 1
						actorC.status = 4
						// create the signature for first-round block for test
						for i:=0;i<actorC.NumPeers;i++ {
							var signBlkfSend1 SignatureBlkF
							signBlkfSend1.signatureBlkF.PubKey = Accounts_test[i].PublicKey
							signBlkfSend1.signatureBlkF.SigData,err = Accounts_test[i].Sign(blockFirstReceived.Header.Hash.Bytes())
							// fmt.Println("Accounts_test:",i,Accounts_test[i].PublicKey,signBlkfSend1.signatureBlkF)
							event.Send(event.ActorNil, event.ActorConsensus, signBlkfSend1)
						}
						fmt.Println("blockFirstReceived.Header.Hash:", dataPreBlkReceived.NumberRound, blockFirstReceived.Header.Hash, blockFirstReceived.Header.MerkleHash, blockFirstReceived.Header.StateHash)
					}
					// test end


					// 7. set the timer for waiting the second-round(final) block
					t3 := time.NewTimer(time.Second * waitResponseTime)
					go func() {
						select {
						case <-t3.C:
							// timeout for the second-round(final) block
							err = event.Send(event.ActorConsensus, event.ActorConsensus, BlockSTimeout{})
							t3.Stop()
						}
					}()
				}
			}
		}

	case TxTimeout:
		// for test 2018.08.01
		if TestTag == true {
			// primary_tag = 0
			// actorC.status = 5
			fmt.Println("timeout test needs to be specified")
		}
		// end test

		if actorC.primaryTag == 0 && (actorC.status == 2 || actorC.status == 5) {
			// not receive the first round block
			// change the status
			actorC.status = 8
			actorC.primaryTag = 0
			// send out the timeout message
			var timeoutMsg TimeoutMsg
			timeoutMsg.Toutmsg.RoundNumber = uint64(actorC.currentRoundNum)
			timeoutMsg.Toutmsg.PubKey = actorC.serviceABABFT.account.PublicKey
			hashTS,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentRoundNum)))
			timeoutMsg.Toutmsg.SigData,_ = actorC.serviceABABFT.account.Sign(hashTS.Bytes())
			event.Send(event.ActorConsensus,event.ActorP2P, timeoutMsg)

			// for test 2018.08.01
			if TestTag == true {
				actorC.primaryTag = 0
				actorC.status = 5
				for i:=0;i<actorC.NumPeers;i++ {
					var timeoutMsg1 TimeoutMsg
					timeoutMsg1.Toutmsg.RoundNumber = uint64(actorC.currentRoundNum+1)
					timeoutMsg1.Toutmsg.PubKey = Accounts_test[i].PublicKey
					hashT,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentRoundNum+1)))
					timeoutMsg1.Toutmsg.SigData,_ = Accounts_test[i].Sign(hashT.Bytes())
					event.Send(event.ActorNil, event.ActorConsensus, timeoutMsg1)
				}
				return
			}
			// end test

			// start/enter the next turn
			event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{actorC.chainID})
			// todo
			// the above needed to be checked
			// here, enter the next term and broadcast the preblock signature with the increased round number has the same effect as the changeview/ nextround message
			// handle of the timeout message has been added, please check case TimeoutMsg
			return
		}

	case SignatureBlkF:
		// fmt.Println("SignatureBlkF:",received_signblkf_num,msg.signatureBlkF)
		// the prime will verify the signatures of first-round block from peers
		if actorC.primaryTag == 1 && actorC.status == 4 {
			// verify the signature
			// 1. check the peer in the peers list
			pubKeyIn := msg.signatureBlkF.PubKey
			var foundPeer bool
			foundPeer = false
			var peerIndex int
			/*
			for index,peer := range Peers_list {
				if ok := bytes.Equal(peer.PublicKey, pubKeyIn); ok == true {
					foundPeer = true
					peerIndex = index
					break
				}
			}
			*/
			// change public key to account address
			for index, peerAddr := range actorC.PeersAddrList {
				peerAddrIns := common.AddressFromPubKey(pubKeyIn)
				if ok := bytes.Equal(peerAddr.AccAddress.Bytes(), peerAddrIns.Bytes()); ok == true {
					foundPeer = true
					peerIndex = index
					break
				}
			}

			if foundPeer == false {
				// the signature is not from the peer in the list
				return
			}
			// 2. verify the correctness of the signature
			if actorC.signatureBlkFList[peerIndex].SigData != nil {
				// already receive the signature
				return
			}
			signDataIn := msg.signatureBlkF.SigData
			headerHashes := actorC.blockFirstRound.BlockFirst.Header.Hash.Bytes()
			var resultVerify bool
			resultVerify, err = secp256k1.Verify(headerHashes, signDataIn, pubKeyIn)
			if resultVerify == true {
				// add the incoming signature to signature preblock list
				actorC.signatureBlkFList[peerIndex].SigData = signDataIn
				actorC.signatureBlkFList[peerIndex].PubKey = pubKeyIn
				actorC.receivedSignBlkFNum ++
				return
			} else {
				return
			}

		}

	case SignTxTimeout:
		// fmt.Println("received_signblkf_num:",received_signblkf_num)
		log.Info("start to generate second round block",actorC.primaryTag, actorC.status,actorC.receivedSignBlkFNum,int(2*len(actorC.PeersAddrList)/3),actorC.signatureBlkFList)
		if actorC.primaryTag == 1 && actorC.status == 4 {
			// check the number of the signatures of first-round block from peers
			if actorC.receivedSignBlkFNum >= int(2*len(actorC.PeersAddrList)/3) {
				// enough first-round block signatures, so generate the second-round(final) block
				// 1. add the first-round block signatures into ConsensusData
				pubKeyTagB := []byte(pubKeyTag)
				signDataTagB := []byte(signDataTag)
				var signTag common.Signature
				signTag.PubKey = pubKeyTagB
				signTag.SigData = signDataTagB

				conABABFTData := actorC.blockFirstRound.BlockFirst.ConsensusData.Payload.(*types.AbaBftData)
				// prepare the ConsensusData
				// add the tag to distinguish preblock signature and second round signature
				conABABFTData.PreBlockSignatures = append(conABABFTData.PreBlockSignatures, signTag)
				for _, signBlkF := range actorC.signatureBlkFList {
					/*
					if signBlkF != nil {
						var sign_tmp common.Signature
						sign_tmp.SigData = signBlkF
						sign_tmp.PubKey = Peers_list[index].PublicKey
						conABABFTData.PreBlockSignatures = append(conABABFTData.PreBlockSignatures, sign_tmp)
					}
					*/
					// change public key to account address
					if signBlkF.SigData != nil {
						var signTmp common.Signature
						signTmp.SigData = signBlkF.SigData
						signTmp.PubKey = signBlkF.PubKey
						conABABFTData.PreBlockSignatures = append(conABABFTData.PreBlockSignatures, signTmp)
					}
				}

				conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{NumberRound:uint32(actorC.currentRoundNum), PreBlockSignatures:conABABFTData.PreBlockSignatures,}}
				// 2. generate the second-round(final) block
				var blockSecond types.Block
				blockSecond,err =  actorC.updateBlock(actorC.blockFirstRound.BlockFirst, conData)
				blockSecond.SetSignature(actorC.serviceABABFT.account)
				// fmt.Println("blockSecond:",blockSecond.Header)

				// 3. broadcast the second-round(final) block
				actorC.blockSecondRound.BlockSecond = blockSecond
				// the ledger will multicast the block_secondround after the block is saved in the DB
				// event.Send(event.ActorConsensus, event.ActorP2P, block_secondround)

				// for test 2018.07.31
				if TestTag == true {
					for i:=0;i<actorC.NumPeers;i++ {
						actorC.primaryTag = 0
						actorC.status = 6
						event.Send(event.ActorNil, event.ActorConsensus, actorC.blockSecondRound)
					}
					time.Sleep(time.Second * 10)
					return
				}
				//

				// 4. save the second-round(final) block to ledger
				/*
				if err = actorC.serviceABABFT.ledger.SaveTxBlock(&blockSecond); err != nil {
					// log.Error("save block error:", err)
					println("save block error:", err)
					return
				}
				*/

				// currentheader = blockSecond.Header
				actorC.currentHeaderData = *(blockSecond.Header)
				actorC.currentHeader = &actorC.currentHeaderData
				actorC.verifiedHeight = blockSecond.Height - 1

				if err := event.Send(event.ActorNil, event.ActorLedger, &blockSecond); err != nil {
					log.Fatal(err)
					// return
				}
				if err := event.Send(event.ActorConsensus, event.ActorP2P, &blockSecond); err != nil {
					log.Fatal(err)
					// return
				}

				// 5. change the status
				actorC.status = 7
				actorC.primaryTag = 0

				fmt.Println("save the generated block", blockSecond.Height,actorC.verifiedHeight)
				// start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{actorC.chainID})
				return
			} else {
				// 1. did not receive enough signatures of first-round block from peers in the assigned time interval
				actorC.status = 7
				actorC.primaryTag = 0 // reset to zero, and the next primary will take the turn
				// 2. reset the stateDB
				//err = actorC.serviceABABFT.ledger.ResetStateDB(currentheader.Hash)
				//if err != nil {
				//	log.Debug("ResetStateDB fail")
				//	return
				//}
				// send out the timeout message
				var timeoutMsg TimeoutMsg
				timeoutMsg.Toutmsg.RoundNumber = uint64(actorC.currentRoundNum)
				timeoutMsg.Toutmsg.PubKey = actorC.serviceABABFT.account.PublicKey
				hashTS,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentRoundNum)))
				timeoutMsg.Toutmsg.SigData,_ = actorC.serviceABABFT.account.Sign(hashTS.Bytes())
				event.Send(event.ActorConsensus,event.ActorP2P, timeoutMsg)
				// 3. start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{actorC.chainID})
			}
		}

	case BlockSecondRound:
		// for test 2018.08.09
		if TestTag == true {
			fmt.Println("get second round block")
			actorC.primaryTag = 0
			actorC.status = 6
		}
		// test end
		log.Info("ababbt peer status:", actorC.primaryTag, actorC.status)
		// check whether it is solo mode
		if actorC.status == 102 || actorC.status == 101 {
			if actorC.status == 102 {
				// solo peer
				blockSecondReceived := msg.BlockSecond
				log.Info("ababbt solo block height vs current_height_num:", blockSecondReceived.Header.Height,actorC.currentHeightNum)
				if int(blockSecondReceived.Header.Height) <= actorC.currentHeightNum {
					return
				} else if int(blockSecondReceived.Header.Height) == (actorC.currentHeightNum+1) {
					// check and save
					blockSecondReceived := msg.BlockSecond
					if blockSecondReceived.ConsensusData.Type == types.ConABFT {
						dataBlkReceived := blockSecondReceived.ConsensusData.Payload.(*types.AbaBftData)
						// check the signature comes from the root
						if ok := bytes.Equal(blockSecondReceived.Signatures[0].PubKey,config.Root.PublicKey); ok != true {
							println("the solo block should be signed by the root")
							return
						}

						// check the block header(the consensus data is null)
						var validBlk bool
						validBlk,err = actorC.verifyHeader(&blockSecondReceived, int(dataBlkReceived.NumberRound), *(actorC.currentHeader))
						if validBlk ==false {
							println("header check fail")
							return
						}
						// save the solo block ( in the form of second-round block)
						/*
						if err = actorC.serviceABABFT.ledger.SaveTxBlock(&blockSecondReceived); err != nil {
							println("save solo block error:", err)
							return
						}
						*/
						// currentheader = blockSecondReceived.Header
						actorC.currentHeaderData = *(blockSecondReceived.Header)
						actorC.currentHeader = &actorC.currentHeaderData
						actorC.verifiedHeight = blockSecondReceived.Height
						actorC.currentHeightNum = int(actorC.verifiedHeight)

						if err := event.Send(event.ActorNil, event.ActorLedger, &blockSecondReceived); err != nil {
							log.Fatal(err)
							// return
						}
						if err := event.Send(event.ActorNil, event.ActorP2P, &blockSecondReceived); err != nil {
							log.Fatal(err)
							// return
						}

						log.Info("verified height of the solo mode:",actorC.verifiedHeight,actorC.currentHeightNum)
						// time.Sleep( time.Second * 2 )
						event.Send(event.ActorNil, event.ActorConsensus, message.ABABFTStart{actorC.chainID})
					}
				} else {
					// send solo syn request
					var reqSynSolo REQSynSolo
					reqSynSolo.Reqsyn.PubKey = actorC.serviceABABFT.account.PublicKey
					hashTS,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentHeightNum+1)))
					reqSynSolo.Reqsyn.SigData,_ = actorC.serviceABABFT.account.Sign(hashTS.Bytes())
					reqSynSolo.Reqsyn.RequestHeight = uint64(actorC.currentHeightNum)+1
					event.Send(event.ActorConsensus,event.ActorP2P, reqSynSolo)
					log.Info("send requirements:", reqSynSolo.Reqsyn.RequestHeight, actorC.currentHeightNum)
				}
			}

			return
		}

		if actorC.primaryTag == 0 && (actorC.status == 6 || actorC.status == 2 || actorC.status == 5) {
			// to verify the first round block
			blockSecondReceived := msg.BlockSecond
			// check the protocol type is ababft
			if blockSecondReceived.ConsensusData.Type == types.ConABFT {
				dataBlkReceived := blockSecondReceived.ConsensusData.Payload.(*types.AbaBftData)

				// for test 2018.08.09
				if TestTag == true {
					actorC.verifiedHeight = uint64(actorC.currentHeightNum) - 1
					fmt.Println("blockSecondReceived.Header.Height:", blockSecondReceived.Header.Height,actorC.verifiedHeight,actorC.currentHeightNum)
				}
				//

				log.Info("received secondround block:", blockSecondReceived.Header.Height,actorC.verifiedHeight,actorC.currentHeightNum, dataBlkReceived.NumberRound, blockSecondReceived.Header)
				// 1. check the round number and height
				// 1a. current round number
				if dataBlkReceived.NumberRound < uint32(actorC.currentRoundNum) || blockSecondReceived.Header.Height <= uint64(actorC.currentHeightNum) {
					return
				} else if (blockSecondReceived.Header.Height-2) > actorC.verifiedHeight {
					// send synchronization message
					var requestSyn REQSyn
					requestSyn.Reqsyn.PubKey = actorC.serviceABABFT.account.PublicKey
					hashTS,_ := common.DoubleHash(Uint64ToBytes(actorC.verifiedHeight+1))
					requestSyn.Reqsyn.SigData,_ = actorC.serviceABABFT.account.Sign(hashTS.Bytes())
					requestSyn.Reqsyn.RequestHeight = actorC.verifiedHeight+1
					event.Send(event.ActorConsensus,event.ActorP2P, requestSyn)
					actorC.synStatus = 1

					// todo
					// attention:
					// to against the height cheat, do not change the actorC.status
				} else {
					// here, the add new block into the ledger, dataBlkReceived.NumberRound >= current_round_num is ok instead of dataBlkReceived.NumberRound == current_round_num
					// 1b. the round number corresponding to the block generator
					indexG := (int(dataBlkReceived.NumberRound)-1) % actorC.NumPeers + 1
					pubKeyGIn := blockSecondReceived.Signatures[0].PubKey
					var indexGIn int
					indexGIn = -1
					/*
					for _, peer := range Peers_list {
						if ok := bytes.Equal(peer.PublicKey, pubKeyGIn); ok == true {
							indexGIn = int(peer.Index)
							break
						}
					}
					*/
					// change public key to account address
					for _, peerAddr := range actorC.PeersAddrList {
						peerAddrIns := common.AddressFromPubKey(pubKeyGIn)
						if ok := bytes.Equal(peerAddr.AccAddress.Bytes(), peerAddrIns.Bytes()); ok == true {
							indexGIn = int(peerAddr.Index)
							break
						}
					}
					if indexG != indexGIn {
						// illegal block generator
						return
					}
					// 1c. check the block header, except the consensus data
					var validBlk bool
					validBlk,err = actorC.verifyHeader(&blockSecondReceived, int(dataBlkReceived.NumberRound), *(actorC.currentHeader))
					// todo
					// can check the hash and statdb and merker root instead of the total head to speed up
					if validBlk ==false {
						println("header check fail")
						return
					}
					// 2. check the signatures ( for both previous and current blocks) in ConsensusData
					preBlkHash := actorC.currentHeader.Hash
					validBlk, err = actorC.verifySignatures(dataBlkReceived, preBlkHash, blockSecondReceived.Header)
					if validBlk ==false {
						println("previous and first-round blocks signatures check fail")
						return
					}

					// for test 2018.08.01
					if TestTag == true {
						fmt.Println("received and verified second round:", blockSecondReceived.Height, blockSecondReceived.Header.Hash, blockSecondReceived.MerkleHash, blockSecondReceived.StateHash)
						// return
					}
					// test end

					// 3.save the second-round block into the ledger
					/*
					if err = actorC.serviceABABFT.ledger.SaveTxBlock(&blockSecondReceived); err != nil {
						// log.Error("save block error:", err)
						println("save block error:", err)
						return
					}
					*/
					// currentheader = blockSecondReceived.Header
					actorC.currentHeaderData = *(blockSecondReceived.Header)
					actorC.currentHeader = &actorC.currentHeaderData
					actorC.verifiedHeight = blockSecondReceived.Height - 1
					if err := event.Send(event.ActorNil, event.ActorLedger, &blockSecondReceived); err != nil {
						log.Fatal(err)
						// return
					}
					if err := event.Send(event.ActorNil, event.ActorP2P, &blockSecondReceived); err != nil {
						log.Fatal(err)
						// return
					}
					// 4. change status
					actorC.status = 8
					actorC.primaryTag = 0
					// update the current_round_num
					if int(dataBlkReceived.NumberRound) > actorC.currentRoundNum {
						actorC.currentRoundNum = int(dataBlkReceived.NumberRound)
					}

					fmt.Println("BlockSecondRound,current_round_num:",actorC.currentRoundNum)
					// start/enter the next turn
					event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{actorC.chainID})
					// 5. broadcast the received second-round block, which has been checked valid
					// to let other peer know this block
					actorC.blockSecondRound.BlockSecond = blockSecondReceived
					// as the ledger will multicast the block after the block is saved in DB, so following code is not need any more
					// event.Send(event.ActorConsensus, event.ActorP2P, block_secondround)
					return
				}
			}
		}
	case BlockSTimeout:
		if actorC.primaryTag == 0 && actorC.status == 5 {
			actorC.status = 8
			actorC.primaryTag = 0
			// reset the state of merkle tree, statehash and so on
			// err = actorC.serviceABABFT.ledger.ResetStateDB(currentheader.Hash)
			//if err != nil {
			//	log.Debug("ResetStateDB fail")
			//	return
			//}
			// send out the timeout message
			var timeoutMsg TimeoutMsg
			timeoutMsg.Toutmsg.RoundNumber = uint64(actorC.currentRoundNum)
			timeoutMsg.Toutmsg.PubKey = actorC.serviceABABFT.account.PublicKey
			hashTS,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentRoundNum)))
			timeoutMsg.Toutmsg.SigData,_ = actorC.serviceABABFT.account.Sign(hashTS.Bytes())
			event.Send(event.ActorConsensus,event.ActorP2P, timeoutMsg)
			// start/enter the next turn
			event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{actorC.chainID})
			return
		}

	case REQSyn:
		// receive the shronization request
		heightReq := msg.Reqsyn.RequestHeight // verified_height+1
		pubKeyIn := msg.Reqsyn.PubKey
		signDataIn := msg.Reqsyn.SigData
		// modify the synchronization code
		// only the verified block will be send back
		// 1. check the height of the verified chain

		if heightReq > uint64(actorC.currentHeightNum - 1) {
			// This peer will reply only when the required height is less or equal to the height of verified block in this peer ledger.
			return
		}
		// check the signature of the request message
		hashTS,_ := common.DoubleHash(Uint64ToBytes(uint64(heightReq)))
		var signVerify bool
		signVerify, err = secp256k1.Verify(hashTS.Bytes(), signDataIn, pubKeyIn)
		if signVerify != true {
			println("Syn request message signature is wrong")
			return
		}

		// fmt.Println("reqsyn:",current_height_num,heightReq)
		// 2. get the response blocks from the ledger
		blkSynV,err1 := actorC.serviceABABFT.ledger.GetTxBlockByHeight(config.ChainHash, heightReq)
		if err1 != nil || blkSynV == nil {
			log.Debug("not find the block of the corresponding height in the ledger")
			return
		}

		// fmt.Println("blkSynV:",blkSynV.Header)
		blkSynF,err2 := actorC.serviceABABFT.ledger.GetTxBlockByHeight(config.ChainHash, heightReq+1)
		if err2 != nil || blkSynF == nil {
			log.Debug("not find the block of the corresponding height in the ledger")
			return
		}
		// 3. send the found /blocks
		var blkSynSend BlockSyn
		blkSynSend.Blksyn.BlksynV,err = blkSynV.Blk2BlkTx()
		if err != nil {
			log.Debug("block_v to blockTx transformation fails")
			return
		}
		blkSynSend.Blksyn.BlksynF,err = blkSynF.Blk2BlkTx()
		if err != nil {
			log.Debug("block_f to blockTx transformation fails")
		}
		event.Send(event.ActorConsensus,event.ActorP2P, blkSynSend)

		// for test 2018.08.02
		if TestTag == true {
			// fmt.Println("blkSynV:",blkSynV.Header)
			// fmt.Println("blkSynF:",blkSynF.Header)
			// fmt.Println("blkSynSend v:",blkSynSend.Blksyn.BlksynV.Header)
			// fmt.Println("blkSynSend f:",blkSynSend.Blksyn.BlksynF.Header)
			// fmt.Println("currentheader.PrevHash:",currentheader.PrevHash)
			// fmt.Println("before reset: currentheader.Hash:",currentheader.Hash)
			currentPreBlk,_ := actorC.currentLedger.GetTxBlock(config.ChainHash, actorC.currentHeader.PrevHash)
			// current_blk := blkSynF
			//err1 := actorC.serviceABABFT.ledger.ResetStateDB(currentPreBlk.Header.StateHash)
			err1 := actorC.serviceABABFT.ledger.ResetStateDB(config.ChainHash, currentPreBlk.Header)
			if err1 != nil {
				fmt.Println("reset status error:", err1)
			}
			// blockFirstCal,err = actorC.serviceABABFT.ledger.NewTxBlock(current_blk.Transactions,current_blk.Header.ConsensusData, current_blk.Header.TimeStamp)
			// fmt.Println("current_blk.hash verfigy:",current_blk.Header.Hash, currentheader.Hash)
			// fmt.Println("compare merkle hash:", current_blk.Header.MerkleHash, blockFirstCal.MerkleHash)
			// fmt.Println("compare state hash:", current_blk.Header.StateHash, blockFirstCal.StateHash)

			// currentheader = current_ledger.GetCurrentHeader()
			oldBlock,_ := actorC.currentLedger.GetTxBlock(config.ChainHash, actorC.currentHeader.PrevHash)
			// currentheader = oldBlock.Header
			actorC.currentHeaderData = *(oldBlock.Header)
			actorC.currentHeader = &actorC.currentHeaderData

			// fmt.Println("after reset: currentheader.Hash:",currentheader.Hash)
			actorC.currentHeightNum = actorC.currentHeightNum - 1
			actorC.verifiedHeight = uint64(actorC.currentHeightNum) - 1
			event.Send(event.ActorNil,event.ActorConsensus, blkSynSend)
		}
		// test end
	case REQSynSolo:
		log.Info("receive the solo block requirement:",msg.Reqsyn.RequestHeight)
		// receive the solo synchronization request
		heightReq := msg.Reqsyn.RequestHeight
		pubKeyIn := msg.Reqsyn.PubKey
		signDataIn := msg.Reqsyn.SigData
		// check the required height
		if heightReq > uint64(actorC.currentHeightNum) {
			return
		}
		// check the signature of the request message
		hashTS,_ := common.DoubleHash(Uint64ToBytes(uint64(heightReq)))
		var signVerify bool
		signVerify, err = secp256k1.Verify(hashTS.Bytes(), signDataIn, pubKeyIn)
		if signVerify != true {
			println("Solo Syn request message signature is wrong")
			return
		}

		for i := int(heightReq); i <= actorC.currentHeightNum; i++ {
			// get the response blocks from the ledger
			blkSynSolo,err1 := actorC.serviceABABFT.ledger.GetTxBlockByHeight(config.ChainHash, uint64(i))
			if err1 != nil || blkSynSolo == nil {
				log.Debug("not find the solo block of the corresponding height in the ledger")
				return
			}
			// send the solo block
			event.Send(event.ActorConsensus,event.ActorP2P, blkSynSolo)
			log.Info("send the required solo block:", blkSynSolo.Height)
		}
		return
	case BlockSyn:
		// for test 2018.08.08
		if TestTag == true {
			actorC.synStatus = 1
		}
		// test end

		if actorC.synStatus != 1 {
			return
		}
		var blkV types.Block
		var blkF types.Block
		err = blkV.BlkTx2Blk(*msg.Blksyn.BlksynV)
		if err != nil {
			log.Debug("blockTx to block_v transformation fails")
		}
		err = blkF.BlkTx2Blk(*msg.Blksyn.BlksynF)
		if err != nil {
			log.Debug("blockTx to block_f transformation fails")
		}
		// fmt.Println("blkV:",blkV.Header)
		// fmt.Println("blkF:",blkF.Header)

		// for test 2018.08.06
		if TestTag == true {
			fmt.Println("heightSynV:", blkV.Header.Height,actorC.currentHeightNum,actorC.verifiedHeight)
			// fmt.Println("blkV.Header:",blkV.Header)
		}
		// test end


		heightSynV := blkV.Header.Height
		if heightSynV == (actorC.verifiedHeight+1) {
			// the current_height_num has been verified
			// 1. verify the verified block blkV

			// todo
			// maybe only check the hash is enough

			var resultV bool
			var blkVLocal *types.Block
			blkVLocal,err = actorC.serviceABABFT.ledger.GetTxBlockByHeight(config.ChainHash, actorC.verifiedHeight)
			if err != nil {
				log.Debug("get previous block error")
				return
			}

			if ok := bytes.Equal(blkV.Hash.Bytes(),actorC.currentHeader.Hash.Bytes()); ok == true {
				// the blkV is the same as current block, just to verify and save blkF
				resultV = true
				blkV.Header = actorC.currentHeader
				fmt.Println("already have")

			} else {
				// verify blkV
				resultV,err = actorC.blkSynVerify(blkV, *blkVLocal)
				fmt.Println("have not yet")
			}


			// for test 2018.08.06
			if TestTag == true {
				// fmt.Println("blkV.Hash:",blkV.Hash)
				// fmt.Println("currentheader.Hash:",currentheader.Hash)
				if ok := bytes.Equal(blkV.Hash.Bytes(),actorC.currentHeader.Hash.Bytes()); ok == true {
					resultV = true
				}
			}
			// test end

			if resultV == false {
				log.Debug("verification of blkV fails")
				return
			}
			// 2. verify the verified block blkF
			var resultF bool
			resultF,err = actorC.blkSynVerify(blkF, blkV)
			if resultF == false {
				log.Debug("verification of blkF fails")
				return
			}
			// 3. save the blocks
			// 3.1 save blkV
			if ok := bytes.Equal(blkV.Hash.Bytes(), actorC.currentHeader.Hash.Bytes()); ok != true {
				// the blkV is not in the ledger,then save blkV
				// here need one reset DB
				//err = actorC.serviceABABFT.ledger.ResetStateDB(blk_pre.Header.Hash)
				if actorC.verifiedHeight < uint64(actorC.currentHeightNum) {
					err = actorC.serviceABABFT.ledger.ResetStateDB(config.ChainHash, blkVLocal.Header)
					if err != nil {
						log.Debug("reset state db error:", err)
						return
					}
				}
				/*
				if err = actorC.serviceABABFT.ledger.SaveTxBlock(&blkV); err != nil {
					log.Debug("save block error:", err)
					return
				}
				*/
				if err := event.Send(event.ActorNil, event.ActorLedger, &blkV); err != nil {
					log.Fatal(err)
					// return
				}
				if err := event.Send(event.ActorConsensus, event.ActorP2P, &blkV); err != nil {
					log.Fatal(err)
					// return
				}
			}  else {
				// the blkV has been in the ledger
			}
			// 3.2 save blkF
			/*
			if err = actorC.serviceABABFT.ledger.SaveTxBlock(&blkF); err != nil {
				log.Debug("save block error:", err)
				return
			}
			*/
			actorC.currentHeaderData = *(blkF.Header)
			actorC.currentHeader = &actorC.currentHeaderData
			actorC.verifiedHeight = blkV.Height

			if err := event.Send(event.ActorNil, event.ActorLedger, &blkF); err != nil {
				log.Fatal(err)
				// return
			}
			if err := event.Send(event.ActorConsensus, event.ActorP2P, &blkF); err != nil {
				log.Fatal(err)
				// return
			}

			// 4. the block is successfully saved, then change the status
			// currentheader = blkF.Header
			actorC.status = 8
			actorC.primaryTag = 0

			// update the current_round_num
			blkRoundNum := int(blkV.ConsensusData.Payload.(*types.AbaBftData).NumberRound)
			if actorC.currentRoundNum < blkRoundNum {
				actorC.currentRoundNum = blkRoundNum
			}

			// start/enter the next turn
			event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{actorC.chainID})

			// todo
			// take care of save and reset

		} else if heightSynV >uint64(actorC.currentHeightNum) {
			// the verified block has bigger height
			// send synchronization message
			var requestSyn REQSyn
			requestSyn.Reqsyn.PubKey = actorC.serviceABABFT.account.PublicKey
			hashTS,_ := common.DoubleHash(Uint64ToBytes(actorC.verifiedHeight+1))
			requestSyn.Reqsyn.SigData,_ = actorC.serviceABABFT.account.Sign(hashTS.Bytes())
			requestSyn.Reqsyn.RequestHeight = actorC.verifiedHeight+1
			event.Send(event.ActorConsensus,event.ActorP2P, requestSyn)
			actorC.synStatus = 1
		}
		// todo
		// only need to check the hash and signature is enough?
		// this may help to speed up the ababft
		return

	case TimeoutMsg:
		// todo
		// the waiting time maybe need to be longer after every time out

		pubKeyIn := msg.Toutmsg.PubKey
		roundIn := int(msg.Toutmsg.RoundNumber)
		signDataIn := msg.Toutmsg.SigData
		fmt.Println("receive the TimeoutMsg:", pubKeyIn, roundIn,actorC.currentRoundNum)
		// check the peer in the peers list
		if roundIn < actorC.currentRoundNum {
			return
		}
		// check the signature
		hashTS,_ := common.DoubleHash(Uint64ToBytes(uint64(roundIn)))
		var signVerify bool
		signVerify, err = secp256k1.Verify(hashTS.Bytes(), signDataIn, pubKeyIn)
		if signVerify != true {
			println("time out message signature is wrong")
			return
		}
		/*
		for _, peer := range Peers_list {
			if ok := bytes.Equal(peer.PublicKey, pubKeyIn); ok == true {
				// legal peer
				// fmt.Println("TimeoutMsgs:",TimeoutMsgs)
				if _, ok1 := TimeoutMsgs[string(pubKeyIn)]; ok1 != true {
					TimeoutMsgs[string(pubKeyIn)] = roundIn
					//fmt.Println("TimeoutMsgs, add:",TimeoutMsgs[string(pubKeyIn)])
				} else if TimeoutMsgs[string(pubKeyIn)] >= roundIn {
					return
				}

				TimeoutMsgs[string(pubKeyIn)] = roundIn
				// to count the number is enough
				var count_r [1000]int
				var max_r int
				max_r = 0
				for _,v := range TimeoutMsgs {
					if v > current_round_num {
						count_r[v-current_round_num]++
					}
					if v > max_r {
						max_r = v
					}
				}

				var total_count int
				total_count = 0
				for i := max_r-current_round_num; i > 0; i-- {
					total_count = total_count + count_r[i]
					if total_count >= int(2*len(Peers_list)/3+1) {
						// reset the round number
						current_round_num += i
						// start/enter the next turn
						actorC.status = 8
						primary_tag = 0
						event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
						// fmt.Println("reset according to the timeout msg:",i,max_r,current_round_num,count_r[i])
						break
					}
				}
				break
			}
		}
		*/
		// change public key to account address
		for _, peerAddr := range actorC.PeersAddrList {
			peerAddrIns := common.AddressFromPubKey(pubKeyIn)
			if ok := bytes.Equal(peerAddr.AccAddress.Bytes(), peerAddrIns.Bytes()); ok == true {
				// legal peer
				// fmt.Println("TimeoutMsgs:",TimeoutMsgs)
				if _, ok1 := actorC.TimeoutMSGs[peerAddrIns.HexString()]; ok1 != true {
					actorC.TimeoutMSGs[peerAddrIns.HexString()] = roundIn
					//fmt.Println("TimeoutMsgs, add:",TimeoutMsgs[string(pubKeyIn)])
				} else if actorC.TimeoutMSGs[peerAddrIns.HexString()] >= roundIn {
					return
				}

				actorC.TimeoutMSGs[peerAddrIns.HexString()] = roundIn
				// to count the number is enough
				var countRS [1000]int
				var maxR int
				maxR = 0
				for _,v := range actorC.TimeoutMSGs {
					if v > actorC.currentRoundNum {
						countRS[v-actorC.currentRoundNum]++
					}
					if v > maxR {
						maxR = v
					}
				}

				var totalCount int
				totalCount = 0
				for i := maxR -actorC.currentRoundNum; i > 0; i-- {
					totalCount = totalCount + countRS[i]
					if totalCount >= int(2*len(actorC.PeersAddrList)/3) {
						// reset the round number
						actorC.currentRoundNum += i
						// start/enter the next turn
						actorC.status = 8
						actorC.primaryTag = 0
						event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{actorC.chainID})
						// fmt.Println("reset according to the timeout msg:",i,maxR,current_round_num,countRS[i])
						break
					}
				}
				break
			}
		}

		// change public key to account address

		return

	case *message.RegChain:
		log.Info("Receive ABABFT Create Message")
		go actorC.serviceABABFT.GenNewChain(msg.ChainID)
		return

	default :
		log.Debug(msg)
		log.Warn("unknown message", reflect.TypeOf(ctx.Message()))
		return
	}
}

func (actorC *ActorABABFT) verifyHeader(blockIn *types.Block, currentRoundNumIn int, curHeader types.Header) (bool,error){
	var err error
	headerIn := blockIn.Header
	txs := blockIn.Transactions
	dataPreBlkReceived := blockIn.ConsensusData.Payload.(*types.AbaBftData)
	signPreSend := dataPreBlkReceived.PreBlockSignatures
	conDataC := types.ConsensusData{Type:types.ConABFT, Payload:&types.AbaBftData{ NumberRound:uint32(currentRoundNumIn), PreBlockSignatures:signPreSend},}
	// fmt.Println("dataPreBlkReceived:",dataPreBlkReceived)
	// fmt.Println("conDataC:",conDataC)
	// fmt.Println("before reset")
	// reset the stateDB
	// fmt.Println("curHeader state hash:",curHeader.Height,curHeader.StateHash)
	// err = actorC.serviceABABFT.ledger.ResetStateDB(curHeader.Hash)
	// fmt.Println("after reset",err)

	// generate the blockFirstCal for comparison
	actorC.blockFirstCal,err = actorC.serviceABABFT.ledger.NewTxBlock(config.ChainHash, txs, conDataC, headerIn.TimeStamp)
	// fmt.Println("height:",blockIn.Height,blockFirstCal.Height)
	// fmt.Println("merkle:",blockIn.Header.MerkleHash,blockFirstCal.Header.MerkleHash)
	// fmt.Println("timestamp:",blockIn.Header.TimeStamp,blockFirstCal.Header.TimeStamp)
	// fmt.Println("blockFirstCal:",blockFirstCal.Header, blockFirstCal.Header.StateHash)
	// fmt.Println("blockIn:",blockIn.Header, blockIn.Header.StateHash)
	log.Info("blockFirstCal:", actorC.blockFirstCal.Height, actorC.blockFirstCal.Header)
	log.Info("blockIn:", blockIn.Height, blockIn.Header)

	var numTxs int
	numTxs = int(blockIn.CountTxs)
	if numTxs != len(txs) {
		println("tx number is wrong")
		return false,nil
	}
	// check Height        uint64
	if actorC.currentHeightNum >= int(headerIn.Height) {
		println("the height is not higher than current height")
		return false,nil
	}
	// ConsensusData is checked in the Receive function

	// check PrevHash      common.Hash
	if ok :=bytes.Equal(blockIn.PrevHash.Bytes(), curHeader.Hash.Bytes()); ok != true {
		println("prehash is wrong")
		return false,nil
	}
	// check MerkleHash    common.Hash
	if ok := bytes.Equal(actorC.blockFirstCal.MerkleHash.Bytes(), blockIn.MerkleHash.Bytes()); ok != true {
		println("MercleHash is wrong")
		return false,nil
	}
	// fmt.Println("mercle:",blockFirstCal.MerkleHash.Bytes(),blockIn.MerkleHash.Bytes())

	// check StateHash     common.Hash
	if ok := bytes.Equal(actorC.blockFirstCal.StateHash.Bytes(), blockIn.StateHash.Bytes()); ok != true {
		println("StateHash is wrong")
		return false,nil
	}
	// fmt.Println("statehash:",blockFirstCal.StateHash.Bytes(),blockIn.StateHash.Bytes())

	// check Bloom         bloom.Bloom
	if ok := bytes.Equal(actorC.blockFirstCal.Bloom.Bytes(), blockIn.Bloom.Bytes()); ok != true {
		println("bloom is wrong")
		return false,nil
	}
	// check Hash common.Hash
	headerCal,err1 := types.NewHeader(headerIn.Version, config.ChainHash, headerIn.Height, headerIn.PrevHash,
		headerIn.MerkleHash, headerIn.StateHash, headerIn.ConsensusData, headerIn.Bloom, headerIn.Receipt.BlockCpu, headerIn.Receipt.BlockNet, headerIn.TimeStamp)
	if ok := bytes.Equal(headerCal.Hash.Bytes(), headerIn.Hash.Bytes()); ok != true {
		println("Hash is wrong")
		return false,err1
	}
	// check Signatures    []common.Signature
	signPreIn := blockIn.Signatures[0]
	pubKeyGIn := signPreIn.PubKey
	signDataIn := signPreIn.SigData
	var signVerify bool
	signVerify, err = secp256k1.Verify(headerIn.Hash.Bytes(), signDataIn, pubKeyGIn)
	if signVerify != true {
		println("signature is wrong")
		return false,err
	}
	return true,err
}

func (actorC *ActorABABFT) updateBlock(blockFirst types.Block, conData types.ConsensusData) (types.Block,error){
	var blockSecond types.Block
	var err error
	headerIn := blockFirst.Header
	header, _ := types.NewHeader(headerIn.Version, config.ChainHash, headerIn.Height, headerIn.PrevHash, headerIn.MerkleHash,
		headerIn.StateHash, conData, headerIn.Bloom, headerIn.Receipt.BlockCpu, headerIn.Receipt.BlockNet, headerIn.TimeStamp)
	blockSecond = types.Block{Header:header, CountTxs:uint32(len(blockFirst.Transactions)), Transactions:blockFirst.Transactions,}
	return blockSecond,err
}

func (actorC *ActorABABFT) verifySignatures(dataBlksReceived *types.AbaBftData, preBlkHash common.Hash, curHeader *types.Header) (bool,error){
	var err error
	// 1. devide the signatures into two part
	var signBlksPreBlk []common.Signature
	var signsCurBlk []common.Signature
	pubKeyTagBytes := []byte(pubKeyTag)
	sigDataTagBytes := []byte(signDataTag)
	var tagSign int
	tagSign = 0
	for _,sign := range dataBlksReceived.PreBlockSignatures {
		ok1 := bytes.Equal(sign.PubKey, pubKeyTagBytes)
		ok2 := bytes.Equal(sign.SigData, sigDataTagBytes)
		if ok1 == true && ok2 == true {
			tagSign = 1
			continue
		}
		if tagSign == 0 {
			signBlksPreBlk = append(signBlksPreBlk,sign)
		} else if tagSign == 1 {
			signsCurBlk = append(signsCurBlk,sign)
		}
	}
	// fmt.Println("signBlksPreBlk:",signBlksPreBlk)
	fmt.Println("signsCurBlk:", signsCurBlk)

	// 2. check the preblock signature
	var numVerified int
	numVerified = 0
	for index, signPreBlk := range signBlksPreBlk {
		// 2a. check the peers in the peer list
		var peerInTag bool
		peerInTag = false
		/*
		for _, peer := range Peers_list {
			if ok := bytes.Equal(peer.PublicKey, signPreBlk.PubKey); ok == true {
				peerInTag = true
				break
			}
		}
		*/
		// change public key to account address
		for _, peerAddr := range actorC.PeersAddrList {
			peerAddrIns := common.AddressFromPubKey(signPreBlk.PubKey)
			if ok := bytes.Equal(peerAddr.AccAddress.Bytes(), peerAddrIns.Bytes()); ok == true {
				peerInTag = true
				break
			}
		}
		if peerInTag == false {
			// there exists signature not from the peer list
			fmt.Println("the signature is not from the peer list, its index is:", index)
			return false,nil
		}
		// 2b. verify the correctness of the signature
		pubKeyIn := signPreBlk.PubKey
		sigDataIn := signPreBlk.SigData
		var resultVerify bool
		resultVerify, err = secp256k1.Verify(preBlkHash.Bytes(), sigDataIn, pubKeyIn)
		if resultVerify == true {
			numVerified++
		}
	}
	// 2c. check the valid signature number
	if numVerified < int(len(actorC.PeersAddrList)/3+1){
		fmt.Println(" not enough signature for the previous block:", numVerified)
		return false,nil
	}


	// 3. check the current block signature
	numVerified = 0
	// calculate firstround block header hash for the check of the first-round block signatures
	conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{NumberRound:uint32(dataBlksReceived.NumberRound), PreBlockSignatures:signBlksPreBlk},}
	headerReCal, _ := types.NewHeader(curHeader.Version, config.ChainHash, curHeader.Height, curHeader.PrevHash, curHeader.MerkleHash,
		curHeader.StateHash, conData, curHeader.Bloom, curHeader.Receipt.BlockCpu, curHeader.Receipt.BlockNet, curHeader.TimeStamp)
	blkFHash := headerReCal.Hash
	// fmt.Println("blkFHash:",blkFHash)
	// fmt.Println("headerReCal for first round signature:", current_round_num,headerReCal.StateHash,headerReCal.MerkleHash,headerReCal.Hash)
	for index, signCurBlk := range signsCurBlk {
		// 3a. check the peers in the peer list
		var peerInTag bool
		peerInTag = false
		/*
		for _, peer := range Peers_list {
			if ok := bytes.Equal(peer.PublicKey, signCurBlk.PubKey); ok == true {
				peerInTag = true
				break
			}
		}
		*/
		// change public key to account address
		for _, peerAddr := range actorC.PeersAddrList {
			peerAddrIns := common.AddressFromPubKey(signCurBlk.PubKey)
			if ok := bytes.Equal(peerAddr.AccAddress.Bytes(), peerAddrIns.Bytes()); ok == true {
				peerInTag = true
				break
			}
		}
		if peerInTag == false {
			// there exists signature not from the peer list
			fmt.Println("the signature is not from the peer list, its index is:", index)
			return false,nil
		}
		// 3b. verify the correctness of the signature
		pubkeyIns := signCurBlk.PubKey
		sigdataIns := signCurBlk.SigData
		var resultVerify bool
		resultVerify, err = secp256k1.Verify(blkFHash.Bytes(), sigdataIns, pubkeyIns)
		if resultVerify == true {
			numVerified++
		}
	}
	// 3c. check the valid signature number
	if numVerified < int(2*len(actorC.PeersAddrList)/3){
		fmt.Println(" not enough signature for first round block:", numVerified)
		return false,nil
	}
	return  true,err

	// todo
	// use checkPermission(index common.AccountName, name string, sig []common.Signature) instead
	/*
	// 4. check the current block signature by using function checkPermission
	// 4a. check the peers permission
	err = actorC.serviceABABFT.ledger.checkPermission(0, "active",signsCurBlk)
	if err != nil {
		log.Debug("signature permission check fail")
		return false,err
	}
	numVerified = 0
	// calculate firstround block header hash for the check of the first-round block signatures
	conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(current_round_num),signBlksPreBlk}}
	headerReCal, _ := types.NewHeader(curHeader.Version, curHeader.Height, curHeader.PrevHash, curHeader.MerkleHash,
		curHeader.StateHash, conData, curHeader.Bloom, curHeader.TimeStamp)
	blkFHash := headerReCal.Hash
	for _,signCurBlk := range signsCurBlk {
		// 4b. verify the correctness of the signature
		pubkey_in := signCurBlk.PubKey
		sigdata_in := signCurBlk.SigData
		var result_verify bool
		result_verify, err = secp256k1.Verify(blkFHash.Bytes(), sigdata_in, pubkey_in)
		if result_verify == true {
			numVerified++
		}
	}
	// 4c. check the valid signature number
	if numVerified < int(2*len(Peers_list)/3+1){
		fmt.Println(" not enough signature for first round block:", numVerified)
		return false,nil
	}
	return  true,err
	*/
}

func (actorC *ActorABABFT) blkSynVerify(blockIn types.Block, blkPre types.Block) (bool,error) {
	var err error
	// 1. check the protocal type is ababft
	if blockIn.ConsensusData.Type != types.ConABFT {
		log.Debug("protocal error")
		return false,nil
	}
	// 2. check the block generator
	dataBlockReceived := blockIn.ConsensusData.Payload.(*types.AbaBftData)
	roundNumIn := int(dataBlockReceived.NumberRound)
	indexG := (int(dataBlockReceived.NumberRound)-1) % actorC.NumPeers + 1
	pukeyGIns := blockIn.Signatures[0].PubKey

	var indexGIn int
	indexGIn = -1
	/*
	for _, peer := range Peers_list {
		if ok := bytes.Equal(peer.PublicKey, pukeyGIns); ok == true {
			indexGIn = int(peer.Index)
			break
		}
	}
	*/
	// change public key to account address
	for _, peerAddr := range actorC.PeersAddrList {
		peerAddrIn := common.AddressFromPubKey(pukeyGIns)
		if ok := bytes.Equal(peerAddr.AccAddress.Bytes(), peerAddrIn.Bytes()); ok == true {
			indexGIn = int(peerAddr.Index)
			break
		}
	}
	// for test 2018.08.10
	if TestTag == true {
		fmt.Println("syn roundNumIn:", roundNumIn, indexGIn, pukeyGIns)
		fmt.Println("peer address list:", actorC.PeersAddrList)
		fmt.Println("blockIn.header:", blockIn.Height, blockIn.Hash, blockIn.MerkleHash, blockIn.StateHash)
	}
	// test end


	if indexG != indexGIn {
		log.Debug("illegal block generator")
		return false,nil
	}
	// 3. check the block header, except the consensus data
	var validBlk bool
	validBlk,err = actorC.verifyHeader(&blockIn, roundNumIn, *blkPre.Header)
	if validBlk ==false {
		println("header check fail")
		return validBlk,err
	}

	// 4. check the signatures ( for both previous and current blocks) in ConsensusData
	validBlk, err = actorC.verifySignatures(dataBlockReceived, blkPre.Header.Hash, blockIn.Header)
	if validBlk ==false {
		println("previous and first-round blocks signatures check fail")
		return false,nil
	}

	return true,nil
}

func Uint64ToBytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, i)
	return buf
}
