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
type ActorAbabft struct {
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
	pid           *actor.PID // actor pid
	serviceAbabft *ServiceABABFT
	NumPeers int
	PeersAddrList []PeerAddrInfo       // Peer address information for consensus
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

	verifiedHeight uint64
	primaryTag int // 0: verification peer; 1: is the primary peer, who generate the block at current round;
	deltaRoundNum int
	receivedSignPreNum int                      // the number of received signatures for the previous block
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
var received_signblkf_num int                     // temporary parameters for received signatures for first round block
var TimeoutMsgs = make(map[string]int, 1000)      // cache the timeout message
// var verified_height uint64

// var delta_roundnum int

var syn_status int
// for test 2018.07.31
var Accounts_test []account.Account
// test end



func ActorAbabftGen(actorAbabft *ActorAbabft) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor {
		return actorAbabft
	})
	pid, err := actor.SpawnNamed(props, "ActorAbabft")
	if err != nil {
		return nil, err
	}
	event.RegisterActor(event.ActorConsensus, pid)
	syn_status = 0
	return pid, err
}

func (actorC *ActorAbabft) Receive(ctx actor.Context) {
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
				var signpre_send []common.Signature
				signpre_send = append(signpre_send, actorC.currentHeader.Signatures[0])
				conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(actorC.currentRoundNum),signpre_send}}
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
				txs, _ := actorC.serviceAbabft.txPool.GetTxsList(config.ChainHash)


				// generate the block in the form of second round block
				var blockSolo *types.Block
				t_time := time.Now().UnixNano()
				blockSolo,err = actorC.serviceAbabft.ledger.NewTxBlock(config.ChainHash, txs, conData, t_time)
				blockSolo.SetSignature(&soloaccount)
				actorC.blockSecondRound.BlockSecond = *blockSolo
				// save (the ledger will broadcast the block after writing the block into the DB)
				/*
				if err = actorC.serviceAbabft.ledger.SaveTxBlock(blockSolo); err != nil {
					// log.Error("save block error:", err)
					println("save solo block error:", err)
					return
				}
				*/
				if err := event.Send(event.ActorNil, event.ActorP2P, blockSolo); err != nil {
					log.Fatal(err)
					// return
				}
				if err := event.Send(event.ActorNil, event.ActorLedger, blockSolo); err != nil {
					log.Fatal(err)
					// return
				}

				// currentheader = blockSolo.Header
				actorC.currentHeaderData = *(blockSolo.Header)
				actorC.currentHeader = &actorC.currentHeaderData

				actorC.verifiedHeight = blockSolo.Height
				fmt.Println("ababft solo height:", blockSolo.Height, blockSolo)
				time.Sleep(time.Second * waitResponseTime)
				// call itself again
				event.Send(event.ActorNil,event.ActorConsensus,message.ABABFTStart{})
			} else {
				// is the solo peer
				actorC.status = 102
				// todo
				// no need every time to send a request for solo block

				// send solo syn request
				var requestsyn REQSynSolo
				requestsyn.Reqsyn.PubKey = actorC.serviceAbabft.account.PublicKey
				hashTS,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentHeightNum+1)))
				requestsyn.Reqsyn.SigData,_ = actorC.serviceAbabft.account.Sign(hashTS.Bytes())
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
		var Peers_list_account_t = make([]string, actorC.NumPeers)
		for i := 0; i < actorC.NumPeers; i++ {
			// Peers_list_account_t = append(Peers_list_account_t,common.IndexToName(newPeers[i]))
			Peers_list_account_t[i] = newPeers[i].String()
		}
		log.Debug("ababft now enter into the ababft mode:Peers_list_account_t",Peers_list_account_t)
		// sort newPeers
		sort.Strings(Peers_list_account_t)

		actorC.PeersListAccount = make([]PeerInfoAccount, actorC.NumPeers)
		actorC.PeersAddrList = make([]PeerAddrInfo, actorC.NumPeers)
		for i := 0; i < actorC.NumPeers; i++ {
			actorC.PeersListAccount[i].AccountName = common.NameToIndex(Peers_list_account_t[i])
			actorC.PeersListAccount[i].Index = i + 1

			account_info,err := actorC.currentLedger.AccountGet(config.ChainHash, actorC.PeersListAccount[i].AccountName)
			if err != nil {
				log.Debug("fail to get account info.")
			}
			acc_addr_info := account_info.Permissions["owner"]
			if len(acc_addr_info.Keys)!=1 {
				log.Debug("owner address must be 1 for BP node!")
			}
			for _, k := range acc_addr_info.Keys {
				// check the address is correct
				addr_key := k.Actor
				// todo
				// the next check can be delete
				if ok := bytes.Equal(acc_addr_info.Keys[addr_key.HexString()].Actor.Bytes(), addr_key.Bytes()); ok == true {
					// save the address instead of pukey
					actorC.PeersAddrList[i].AccAddress = addr_key
					break;
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
		lastestRoundNum := int(actorC.currentHeader.ConsensusData.Payload.(*types.AbaBftData).NumberRound)
		actorC.deltaRoundNum = actorC.currentRoundNum - lastestRoundNum
		if actorC.deltaRoundNum > ThresholdRound && actorC.currentHeightNum > int(actorC.verifiedHeight) {
			// as there is a long time since last block, maybe the chain is blocked somewhere
			// to generate the block after the previous block (i.e. the latest verified block)
			var currentblock *types.Block
			currentblock,err = actorC.currentLedger.GetTxBlock(config.ChainHash, actorC.currentHeader.PrevHash)
			if err != nil {
				fmt.Println("get previous block error.")
			}
			// currentheader = currentblock.Header
			actorC.currentHeaderData = *(currentblock.Header)
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
		signaturePreblock.PubKey = actorC.serviceAbabft.account.PublicKey
		signaturePreblock.SigData, err = actorC.serviceAbabft.account.Sign(actorC.currentHeader.Hash.Bytes())
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
			var signaturepre_send SignaturePreBlock
			signaturepre_send.SignPreBlock.PubKey = signaturePreblock.PubKey
			signaturepre_send.SignPreBlock.SigData = signaturePreblock.SigData
			// todo
			// for the signature of previous block, maybe the round number is not needed
			signaturepre_send.SignPreBlock.Round = uint32(actorC.currentRoundNum)
			signaturepre_send.SignPreBlock.Height = uint32(actorC.currentHeader.Height)
			// broadcast
			event.Send(event.ActorConsensus, event.ActorP2P, signaturepre_send)
			// increase the round index
			actorC.currentRoundNum ++
			fmt.Println("ABABFTStart:current_round_num(non primary):",actorC.currentRoundNum,actorC.selfIndex)
			// log.Debug("non primary")
			// log.Debug("signaturepre_send:",current_round_num,currentheader.Height,signaturepre_send)
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
			event.Send(event.ActorNil,event.ActorConsensus,message.ABABFTStart{})
		}
		// the prime will verify the signature for the previous block
		round_in := int(msg.SignPreBlock.Round)
		height_in := int(msg.SignPreBlock.Height)
		// log.Debug("current_round_num:",current_round_num,round_in)
		if round_in >= actorC.currentRoundNum && actorC.status!=101 && actorC.status!= 102 {
			// cache the SignaturePreBlock
			actorC.cacheSignaturePreBlk = append(actorC.cacheSignaturePreBlk,msg.SignPreBlock)
			// in case that the signature for the previous block arrived bofore the corresponding block generator was born
		}
		if actorC.primaryTag == 1 && (actorC.status == 2 || actorC.status == 3){
			// verify the signature
			// first check the round number and height

			// todo
			// maybe round number is not needed for preblock signature
			if round_in >= (actorC.currentRoundNum-1) && height_in >= actorC.currentHeightNum {
				if round_in > (actorC.currentRoundNum - 1) && height_in > actorC.currentHeightNum {
					// todo
					// need to check
					// only require when height difference between the peers is >= 2

					if actorC.deltaRoundNum > ThresholdRound {
						if actorC.verifiedHeight == uint64(actorC.currentHeightNum) && height_in == (actorC.currentHeightNum+1) {
							return
						}
					}

					// require synchronization, the longest chain is ok
					// send synchronization message
					var requestsyn REQSyn
					requestsyn.Reqsyn.PubKey = actorC.serviceAbabft.account.PublicKey
					hash_t,_ := common.DoubleHash(Uint64ToBytes(actorC.verifiedHeight+1))
					requestsyn.Reqsyn.SigData,_ = actorC.serviceAbabft.account.Sign(hash_t.Bytes())
					requestsyn.Reqsyn.RequestHeight = actorC.verifiedHeight+1
					event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
					syn_status = 1
					// todo
					// attention
					// to against the height cheat, do not change the actorC.status
				} else {
					// check the signature
					pubkey_in := msg.SignPreBlock.PubKey // signaturepre_send.signaturePreblock.PubKey = signaturePreblock.PubKey
					// check the pubkey_in is in the peer list
					var foundPeer bool
					foundPeer = false
					var peer_index int

					/*
					for index,peer := range Peers_list {
						if ok := bytes.Equal(peer.PublicKey, pubkey_in); ok == true {
							foundPeer = true
							peer_index = index
							break
						}
					}
					*/
					// change public key to account address
					for index, peerAddr := range actorC.PeersAddrList {
						peerAddrIns := common.AddressFromPubKey(pubkey_in)
						if ok := bytes.Equal(peerAddr.AccAddress.Bytes(), peerAddrIns.Bytes()); ok == true {
							foundPeer = true
							peer_index = index
							break
						}
					}

					if foundPeer == false {
						// the signature is not from the peer in the list
						return
					}
					// 1. check that signature in or not in list of
					if actorC.signaturePreBlockList[peer_index].SigData != nil {
						// already receive the signature
						return
					}
					// 2. verify the correctness of the signature
					sigDataIn := msg.SignPreBlock.SigData
					headerHashes := actorC.currentHeader.Hash.Bytes()
					var resultVerify bool
					resultVerify, err = secp256k1.Verify(headerHashes, sigDataIn, pubkey_in)
					if resultVerify == true {
						// add the incoming signature to signature preblock list
						actorC.signaturePreBlockList[peer_index].SigData = sigDataIn
						actorC.signaturePreBlockList[peer_index].PubKey = pubkey_in
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
			header_hash := actorC.currentHeader.Hash.Bytes()
			for _,signpreblk := range actorC.cacheSignaturePreBlk {
				round_in := signpreblk.Round
				if int(round_in) != actorC.currentRoundNum {
					continue
				}
				// check the signature
				pubkey_in := signpreblk.PubKey// signaturepre_send.signaturePreblock.PubKey = signaturePreblock.PubKey
				// check the pubkey_in is in the peer list
				var found_peer bool
				found_peer = false
				var peer_index int
				/*
				for index,peer := range Peers_list {
					if ok := bytes.Equal(peer.PublicKey, pubkey_in); ok == true {
						found_peer = true
						peer_index = index
						break
					}
				}
				*/
				// change public key to account address
				for index,peer_addr := range actorC.PeersAddrList {
					peer_addr_in := common.AddressFromPubKey(pubkey_in)
					if ok := bytes.Equal(peer_addr.AccAddress.Bytes(), peer_addr_in.Bytes()); ok == true {
						found_peer = true
						peer_index = index
						break
					}
				}

				if found_peer == false {
					// the signature is not from the peer in the list
					continue
				}
				// first check that signature in or not in list of
				if actorC.signaturePreBlockList[peer_index].SigData != nil {
					// already receive the signature
					continue
				}
				// second, verify the correctness of the signature
				sigdata_in := signpreblk.SigData
				var result_verify bool
				result_verify, err = secp256k1.Verify(header_hash, sigdata_in, pubkey_in)
				if result_verify == true {
					// add the incoming signature to signature preblock list
					actorC.signaturePreBlockList[peer_index].SigData = sigdata_in
					actorC.signaturePreBlockList[peer_index].PubKey = pubkey_in
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
				var signpre_send []common.Signature
				for _,signpre := range actorC.signaturePreBlockList {
					/*
					if signpre != nil {
						var sign_tmp common.Signature
						sign_tmp.SigData = signpre
						sign_tmp.PubKey = Peers_list[index].PublicKey
						signpre_send = append(signpre_send, sign_tmp)
					}
					*/
					// change public key to account address
					if signpre.SigData != nil {
						var sign_tmp common.Signature
						sign_tmp.SigData = signpre.SigData
						sign_tmp.PubKey = signpre.PubKey
						signpre_send = append(signpre_send, sign_tmp)
					}
				}
				conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(actorC.currentRoundNum),signpre_send}}
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
				txs, _ := actorC.serviceAbabft.txPool.GetTxsList(config.ChainHash)
				// log.Debug("obtained tx list", txs[0])
				// generate the first-round block
				var block_first *types.Block
				t_time := time.Now().UnixNano()
				block_first,err = actorC.serviceAbabft.ledger.NewTxBlock(config.ChainHash, txs, conData, t_time)
				block_first.SetSignature(actorC.serviceAbabft.account)
				// broadcast the first-round block to peers for them to verify the transactions and wait for the corresponding signatures back
				actorC.blockFirstRound.BlockFirst = *block_first
				event.Send(event.ActorConsensus, event.ActorP2P, actorC.blockFirstRound)
				// log.Debug("first round block:",block_firstround.BlockFirst)
				// fmt.Println("first round block status root hash:",block_first.StateHash)
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
				received_signblkf_num = 0
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
				var timeoutmsg TimeoutMsg
				timeoutmsg.Toutmsg.RoundNumber = uint64(actorC.currentRoundNum)
				timeoutmsg.Toutmsg.PubKey = actorC.serviceAbabft.account.PublicKey
				hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentRoundNum)))
				timeoutmsg.Toutmsg.SigData,_ = actorC.serviceAbabft.account.Sign(hash_t.Bytes())
				event.Send(event.ActorConsensus,event.ActorP2P,timeoutmsg)
				// start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
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
			blockfirst_received := msg.BlockFirst
			// the protocal type is ababft
			if blockfirst_received.ConsensusData.Type == types.ConABFT {
				data_preblk_received := blockfirst_received.ConsensusData.Payload.(*types.AbaBftData)
				// 1. check the round number
				// 1a. current round number
				if data_preblk_received.NumberRound < uint32(actorC.currentRoundNum) {
					return
				} else if data_preblk_received.NumberRound > uint32(actorC.currentRoundNum) {
					// require synchronization, the longest chain is ok
					// in case that somebody may skip the current generator, only the different height can call the synchronization
					if (actorC.verifiedHeight+2) < blockfirst_received.Header.Height {
						// send synchronization message
						var requestsyn REQSyn
						requestsyn.Reqsyn.PubKey = actorC.serviceAbabft.account.PublicKey
						hash_t,_ := common.DoubleHash(Uint64ToBytes(actorC.verifiedHeight+1))
						requestsyn.Reqsyn.SigData,_ = actorC.serviceAbabft.account.Sign(hash_t.Bytes())
						requestsyn.Reqsyn.RequestHeight = actorC.verifiedHeight+1
						event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
						syn_status = 1
						// todo
						// attention:
						// to against the height cheat, do not change the actorC.status
					}
				} else {
					// 1b. the round number corresponding to the block generator
					index_g := (actorC.currentRoundNum-1) % actorC.NumPeers + 1
					pukey_g_in := blockfirst_received.Signatures[0].PubKey
					var index_g_in int
					index_g_in = -1
					/*
					for _, peer := range Peers_list {
						if ok := bytes.Equal(peer.PublicKey, pukey_g_in); ok == true {
							index_g_in = int(peer.Index)
							break
						}
					}
					*/
					// change public key to account address
					for _, peer_addr := range actorC.PeersAddrList {
						peer_addr_in := common.AddressFromPubKey(pukey_g_in)
						if ok := bytes.Equal(peer_addr.AccAddress.Bytes(), peer_addr_in.Bytes()); ok == true {
							index_g_in = int(peer_addr.Index)
							break
						}
					}
					if index_g != index_g_in {
						// illegal block generator
						return
					}
					// 1c. check the block header, except the consensus data
					var valid_blk bool
					valid_blk,err = actorC.verifyHeader(&blockfirst_received, actorC.currentRoundNum, *(actorC.currentHeader))
					if valid_blk==false {
						println("header check fail")
						return
					}
					// 2. check the preblock signature
					sign_preblk_list := data_preblk_received.PreBlockSignatures
					header_hash := actorC.currentHeader.Hash.Bytes()
					var num_verified int
					num_verified = 0
					for index,sign_preblk := range sign_preblk_list {
						// 2a. check the peers in the peer list
						var peerin_tag bool
						peerin_tag = false
						/*
						for _, peer := range Peers_list {
							if ok := bytes.Equal(peer.PublicKey, sign_preblk.PubKey); ok == true {
								peerin_tag = true
								break
							}
						}
						*/
						// change public key to account address
						for _, peer_addr := range actorC.PeersAddrList {
							peer_addr_in := common.AddressFromPubKey(sign_preblk.PubKey)
							if ok := bytes.Equal(peer_addr.AccAddress.Bytes(), peer_addr_in.Bytes()); ok == true {
								peerin_tag = true
								break
							}
						}
						if peerin_tag == false {
							// there exists signature not from the peer list
							fmt.Println("the signature is not from the peer list, its index is:", index)
							return
						}
						// 2b. verify the correctness of the signature
						pubkey_in := sign_preblk.PubKey
						sigdata_in := sign_preblk.SigData
						var result_verify bool
						result_verify, err = secp256k1.Verify(header_hash, sigdata_in, pubkey_in)
						if result_verify == true {
							num_verified++
						}
					}
					// 2c. check the valid signature number
					if num_verified < int(len(actorC.PeersAddrList)/3+1){
						// not enough signature
						fmt.Println("not enough signature for second round block")
						return
					}
					// 3. check the txs
					txs_in := blockfirst_received.Transactions
					for index1,tx_in := range txs_in {
						err = actorC.serviceAbabft.ledger.CheckTransaction(config.ChainHash, tx_in)
						if err != nil {
							println("wrong tx, index:", index1)
							return
						}
					}
					// 4. sign the received first-round block
					var sign_blkf_send SignatureBlkF
					sign_blkf_send.signatureBlkF.PubKey = actorC.serviceAbabft.account.PublicKey
					sign_blkf_send.signatureBlkF.SigData,err = actorC.serviceAbabft.account.Sign(blockfirst_received.Header.Hash.Bytes())
					// 5. broadcast the signature of the first round block
					event.Send(event.ActorConsensus, event.ActorP2P, sign_blkf_send)
					// 6. change the status
					actorC.status = 6
					// fmt.Println("sign_blkf_send:",sign_blkf_send)
					// clean the cache_signature_preblk
					actorC.cacheSignaturePreBlk = make([]pb.SignaturePreblock,len(actorC.PeersAddrList)*2)
					// send the received first-round block to other peers in case that network is not good
					actorC.blockFirstRound.BlockFirst = blockfirst_received
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
							signBlkfSend1.signatureBlkF.SigData,err = Accounts_test[i].Sign(blockfirst_received.Header.Hash.Bytes())
							// fmt.Println("Accounts_test:",i,Accounts_test[i].PublicKey,signBlkfSend1.signatureBlkF)
							event.Send(event.ActorNil, event.ActorConsensus, signBlkfSend1)
						}
						fmt.Println("blockfirst_received.Header.Hash:",data_preblk_received.NumberRound,blockfirst_received.Header.Hash, blockfirst_received.Header.MerkleHash,blockfirst_received.Header.StateHash)
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
			var timeoutmsg TimeoutMsg
			timeoutmsg.Toutmsg.RoundNumber = uint64(actorC.currentRoundNum)
			timeoutmsg.Toutmsg.PubKey = actorC.serviceAbabft.account.PublicKey
			hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentRoundNum)))
			timeoutmsg.Toutmsg.SigData,_ = actorC.serviceAbabft.account.Sign(hash_t.Bytes())
			event.Send(event.ActorConsensus,event.ActorP2P,timeoutmsg)

			// for test 2018.08.01
			if TestTag == true {
				actorC.primaryTag = 0
				actorC.status = 5
				for i:=0;i<actorC.NumPeers;i++ {
					var timeoutmsg1 TimeoutMsg
					timeoutmsg1.Toutmsg.RoundNumber = uint64(actorC.currentRoundNum+1)
					timeoutmsg1.Toutmsg.PubKey = Accounts_test[i].PublicKey
					hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentRoundNum+1)))
					timeoutmsg1.Toutmsg.SigData,_ = Accounts_test[i].Sign(hash_t.Bytes())
					event.Send(event.ActorNil, event.ActorConsensus, timeoutmsg1)
				}
				return
			}
			// end test

			// start/enter the next turn
			event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
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
			pubkey_in := msg.signatureBlkF.PubKey
			var found_peer bool
			found_peer = false
			var peer_index int
			/*
			for index,peer := range Peers_list {
				if ok := bytes.Equal(peer.PublicKey, pubkey_in); ok == true {
					found_peer = true
					peer_index = index
					break
				}
			}
			*/
			// change public key to account address
			for index,peer_addr := range actorC.PeersAddrList {
				peer_addr_in := common.AddressFromPubKey(pubkey_in)
				if ok := bytes.Equal(peer_addr.AccAddress.Bytes(), peer_addr_in.Bytes()); ok == true {
					found_peer = true
					peer_index = index
					break
				}
			}

			if found_peer == false {
				// the signature is not from the peer in the list
				return
			}
			// 2. verify the correctness of the signature
			if actorC.signatureBlkFList[peer_index].SigData != nil {
				// already receive the signature
				return
			}
			sigdata_in := msg.signatureBlkF.SigData
			header_hash := actorC.blockFirstRound.BlockFirst.Header.Hash.Bytes()
			var result_verify bool
			result_verify, err = secp256k1.Verify(header_hash, sigdata_in, pubkey_in)
			if result_verify == true {
				// add the incoming signature to signature preblock list
				actorC.signatureBlkFList[peer_index].SigData = sigdata_in
				actorC.signatureBlkFList[peer_index].PubKey = pubkey_in
				received_signblkf_num ++
				return
			} else {
				return
			}

		}

	case SignTxTimeout:
		// fmt.Println("received_signblkf_num:",received_signblkf_num)
		log.Info("start to generate second round block",actorC.primaryTag, actorC.status,received_signblkf_num,int(2*len(actorC.PeersAddrList)/3),actorC.signatureBlkFList)
		if actorC.primaryTag == 1 && actorC.status == 4 {
			// check the number of the signatures of first-round block from peers
			if received_signblkf_num >= int(2*len(actorC.PeersAddrList)/3) {
				// enough first-round block signatures, so generate the second-round(final) block
				// 1. add the first-round block signatures into ConsensusData
				pubkey_tag_b := []byte(pubKeyTag)
				signdata_tag_b := []byte(signDataTag)
				var sign_tag common.Signature
				sign_tag.PubKey = pubkey_tag_b
				sign_tag.SigData = signdata_tag_b

				ababftdata := actorC.blockFirstRound.BlockFirst.ConsensusData.Payload.(*types.AbaBftData)
				// prepare the ConsensusData
				// add the tag to distinguish preblock signature and second round signature
				ababftdata.PreBlockSignatures = append(ababftdata.PreBlockSignatures, sign_tag)
				for _,signblkf := range actorC.signatureBlkFList {
					/*
					if signblkf != nil {
						var sign_tmp common.Signature
						sign_tmp.SigData = signblkf
						sign_tmp.PubKey = Peers_list[index].PublicKey
						ababftdata.PreBlockSignatures = append(ababftdata.PreBlockSignatures, sign_tmp)
					}
					*/
					// change public key to account address
					if signblkf.SigData != nil {
						var sign_tmp common.Signature
						sign_tmp.SigData = signblkf.SigData
						sign_tmp.PubKey = signblkf.PubKey
						ababftdata.PreBlockSignatures = append(ababftdata.PreBlockSignatures, sign_tmp)
					}
				}

				conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(actorC.currentRoundNum),ababftdata.PreBlockSignatures}}
				// 2. generate the second-round(final) block
				var block_second types.Block
				block_second,err =  actorC.updateBlock(actorC.blockFirstRound.BlockFirst, conData)
				block_second.SetSignature(actorC.serviceAbabft.account)
				// fmt.Println("block_second:",block_second.Header)

				// 3. broadcast the second-round(final) block
				actorC.blockSecondRound.BlockSecond = block_second
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
				if err = actorC.serviceAbabft.ledger.SaveTxBlock(&block_second); err != nil {
					// log.Error("save block error:", err)
					println("save block error:", err)
					return
				}
				*/
				if err := event.Send(event.ActorNil, event.ActorLedger, &block_second); err != nil {
					log.Fatal(err)
					// return
				}
				if err := event.Send(event.ActorConsensus, event.ActorP2P, &block_second); err != nil {
					log.Fatal(err)
					// return
				}

				// currentheader = block_second.Header
				actorC.currentHeaderData = *(block_second.Header)
				actorC.currentHeader = &actorC.currentHeaderData

				actorC.verifiedHeight = block_second.Height - 1
				// 5. change the status
				actorC.status = 7
				actorC.primaryTag = 0

				fmt.Println("save the generated block", block_second.Height,actorC.verifiedHeight)
				// start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
				return
			} else {
				// 1. did not receive enough signatures of first-round block from peers in the assigned time interval
				actorC.status = 7
				actorC.primaryTag = 0 // reset to zero, and the next primary will take the turn
				// 2. reset the stateDB
				//err = actorC.serviceAbabft.ledger.ResetStateDB(currentheader.Hash)
				//if err != nil {
				//	log.Debug("ResetStateDB fail")
				//	return
				//}
				// send out the timeout message
				var timeoutmsg TimeoutMsg
				timeoutmsg.Toutmsg.RoundNumber = uint64(actorC.currentRoundNum)
				timeoutmsg.Toutmsg.PubKey = actorC.serviceAbabft.account.PublicKey
				hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentRoundNum)))
				timeoutmsg.Toutmsg.SigData,_ = actorC.serviceAbabft.account.Sign(hash_t.Bytes())
				event.Send(event.ActorConsensus,event.ActorP2P,timeoutmsg)
				// 3. start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
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
				blocksecond_received := msg.BlockSecond
				log.Info("ababbt solo block height vs current_height_num:",blocksecond_received.Header.Height,actorC.currentHeightNum)
				if int(blocksecond_received.Header.Height) <= actorC.currentHeightNum {
					return
				} else if int(blocksecond_received.Header.Height) == (actorC.currentHeightNum+1) {
					// check and save
					blocksecond_received := msg.BlockSecond
					if blocksecond_received.ConsensusData.Type == types.ConABFT {
						data_blks_received := blocksecond_received.ConsensusData.Payload.(*types.AbaBftData)
						// check the signature comes from the root
						if ok := bytes.Equal(blocksecond_received.Signatures[0].PubKey,config.Root.PublicKey); ok != true {
							println("the solo block should be signed by the root")
							return
						}

						// check the block header(the consensus data is null)
						var valid_blk bool
						valid_blk,err = actorC.verifyHeader(&blocksecond_received, int(data_blks_received.NumberRound), *(actorC.currentHeader))
						if valid_blk==false {
							println("header check fail")
							return
						}
						// save the solo block ( in the form of second-round block)
						/*
						if err = actorC.serviceAbabft.ledger.SaveTxBlock(&blocksecond_received); err != nil {
							println("save solo block error:", err)
							return
						}
						*/
						if err := event.Send(event.ActorNil, event.ActorLedger, &blocksecond_received); err != nil {
							log.Fatal(err)
							// return
						}
						if err := event.Send(event.ActorNil, event.ActorP2P, &blocksecond_received); err != nil {
							log.Fatal(err)
							// return
						}
						// currentheader = blocksecond_received.Header
						actorC.currentHeaderData = *(blocksecond_received.Header)
						actorC.currentHeader = &actorC.currentHeaderData

						actorC.verifiedHeight = blocksecond_received.Height
						actorC.currentHeightNum = int(actorC.verifiedHeight)
						log.Info("verified height of the solo mode:",actorC.verifiedHeight,actorC.currentHeightNum)
						// time.Sleep( time.Second * 2 )
						event.Send(event.ActorNil, event.ActorConsensus, message.ABABFTStart{})
					}
				} else {
					// send solo syn request
					var requestsyn REQSynSolo
					requestsyn.Reqsyn.PubKey = actorC.serviceAbabft.account.PublicKey
					hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentHeightNum+1)))
					requestsyn.Reqsyn.SigData,_ = actorC.serviceAbabft.account.Sign(hash_t.Bytes())
					requestsyn.Reqsyn.RequestHeight = uint64(actorC.currentHeightNum)+1
					event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
					log.Info("send requirements:", requestsyn.Reqsyn.RequestHeight, actorC.currentHeightNum)
				}
			}

			return
		}

		if actorC.primaryTag == 0 && (actorC.status == 6 || actorC.status == 2 || actorC.status == 5) {
			// to verify the first round block
			blockSecondReceived := msg.BlockSecond
			// check the protocal type is ababft
			if blockSecondReceived.ConsensusData.Type == types.ConABFT {
				data_blks_received := blockSecondReceived.ConsensusData.Payload.(*types.AbaBftData)

				// for test 2018.08.09
				if TestTag == true {
					actorC.verifiedHeight = uint64(actorC.currentHeightNum) - 1
					fmt.Println("blockSecondReceived.Header.Height:", blockSecondReceived.Header.Height,actorC.verifiedHeight,actorC.currentHeightNum)
				}
				//

				log.Info("received secondround block:", blockSecondReceived.Header.Height,actorC.verifiedHeight,actorC.currentHeightNum,data_blks_received.NumberRound, blockSecondReceived.Header)
				// 1. check the round number and height
				// 1a. current round number
				if data_blks_received.NumberRound < uint32(actorC.currentRoundNum) || blockSecondReceived.Header.Height <= uint64(actorC.currentHeightNum) {
					return
				} else if (blockSecondReceived.Header.Height-2) > actorC.verifiedHeight {
					// send synchronization message
					var requestsyn REQSyn
					requestsyn.Reqsyn.PubKey = actorC.serviceAbabft.account.PublicKey
					hash_t,_ := common.DoubleHash(Uint64ToBytes(actorC.verifiedHeight+1))
					requestsyn.Reqsyn.SigData,_ = actorC.serviceAbabft.account.Sign(hash_t.Bytes())
					requestsyn.Reqsyn.RequestHeight = actorC.verifiedHeight+1
					event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
					syn_status = 1

					// todo
					// attention:
					// to against the height cheat, do not change the actorC.status
				} else {
					// here, the add new block into the ledger, data_blks_received.NumberRound >= current_round_num is ok instead of data_blks_received.NumberRound == current_round_num
					// 1b. the round number corresponding to the block generator
					index_g := (int(data_blks_received.NumberRound)-1) % actorC.NumPeers + 1
					pukey_g_in := blockSecondReceived.Signatures[0].PubKey
					var index_g_in int
					index_g_in = -1
					/*
					for _, peer := range Peers_list {
						if ok := bytes.Equal(peer.PublicKey, pukey_g_in); ok == true {
							index_g_in = int(peer.Index)
							break
						}
					}
					*/
					// change public key to account address
					for _, peer_addr := range actorC.PeersAddrList {
						peer_addr_in := common.AddressFromPubKey(pukey_g_in)
						if ok := bytes.Equal(peer_addr.AccAddress.Bytes(), peer_addr_in.Bytes()); ok == true {
							index_g_in = int(peer_addr.Index)
							break
						}
					}
					if index_g != index_g_in {
						// illegal block generator
						return
					}
					// 1c. check the block header, except the consensus data
					var valid_blk bool
					valid_blk,err = actorC.verifyHeader(&blockSecondReceived, int(data_blks_received.NumberRound), *(actorC.currentHeader))
					// todo
					// can check the hash and statdb and merker root instead of the total head to speed up
					if valid_blk==false {
						println("header check fail")
						return
					}
					// 2. check the signatures ( for both previous and current blocks) in ConsensusData
					preblkhash := actorC.currentHeader.Hash
					valid_blk, err = actorC.verifySignatures(data_blks_received, preblkhash, blockSecondReceived.Header)
					if valid_blk==false {
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
					if err = actorC.serviceAbabft.ledger.SaveTxBlock(&blockSecondReceived); err != nil {
						// log.Error("save block error:", err)
						println("save block error:", err)
						return
					}
					*/

					if err := event.Send(event.ActorNil, event.ActorLedger, &blockSecondReceived); err != nil {
						log.Fatal(err)
						// return
					}
					if err := event.Send(event.ActorNil, event.ActorP2P, &blockSecondReceived); err != nil {
						log.Fatal(err)
						// return
					}
					// 4. change status
					// currentheader = blockSecondReceived.Header
					actorC.currentHeaderData = *(blockSecondReceived.Header)
					actorC.currentHeader = &actorC.currentHeaderData

					actorC.verifiedHeight = blockSecondReceived.Height - 1
					actorC.status = 8
					actorC.primaryTag = 0
					// update the current_round_num
					if int(data_blks_received.NumberRound) > actorC.currentRoundNum {
						actorC.currentRoundNum = int(data_blks_received.NumberRound)
					}

					fmt.Println("BlockSecondRound,current_round_num:",actorC.currentRoundNum)
					// start/enter the next turn
					event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
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
			// err = actorC.serviceAbabft.ledger.ResetStateDB(currentheader.Hash)
			//if err != nil {
			//	log.Debug("ResetStateDB fail")
			//	return
			//}
			// send out the timeout message
			var timeoutmsg TimeoutMsg
			timeoutmsg.Toutmsg.RoundNumber = uint64(actorC.currentRoundNum)
			timeoutmsg.Toutmsg.PubKey = actorC.serviceAbabft.account.PublicKey
			hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentRoundNum)))
			timeoutmsg.Toutmsg.SigData,_ = actorC.serviceAbabft.account.Sign(hash_t.Bytes())
			event.Send(event.ActorConsensus,event.ActorP2P,timeoutmsg)
			// start/enter the next turn
			event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
			return
		}

	case REQSyn:
		// receive the shronization request
		height_req := msg.Reqsyn.RequestHeight // verified_height+1
		pubkey_in := msg.Reqsyn.PubKey
		signdata_in := msg.Reqsyn.SigData
		// modify the synchronization code
		// only the verified block will be send back
		// 1. check the height of the verified chain

		if height_req > uint64(actorC.currentHeightNum - 1) {
			// This peer will reply only when the required height is less or equal to the height of verified block in this peer ledger.
			return
		}
		// check the signature of the request message
		hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(height_req)))
		var sign_verify bool
		sign_verify, err = secp256k1.Verify(hash_t.Bytes(), signdata_in, pubkey_in)
		if sign_verify != true {
			println("Syn request message signature is wrong")
			return
		}

		// fmt.Println("reqsyn:",current_height_num,height_req)
		// 2. get the response blocks from the ledger
		blk_syn_v,err1 := actorC.serviceAbabft.ledger.GetTxBlockByHeight(config.ChainHash, height_req)
		if err1 != nil || blk_syn_v == nil {
			log.Debug("not find the block of the corresponding height in the ledger")
			return
		}

		// fmt.Println("blk_syn_v:",blk_syn_v.Header)
		blk_syn_f,err2 := actorC.serviceAbabft.ledger.GetTxBlockByHeight(config.ChainHash, height_req+1)
		if err2 != nil || blk_syn_f == nil {
			log.Debug("not find the block of the corresponding height in the ledger")
			return
		}
		// 3. send the found /blocks
		var blksyn_send BlockSyn
		blksyn_send.Blksyn.BlksynV,err = blk_syn_v.Blk2BlkTx()
		if err != nil {
			log.Debug("block_v to blockTx transformation fails")
			return
		}
		blksyn_send.Blksyn.BlksynF,err = blk_syn_f.Blk2BlkTx()
		if err != nil {
			log.Debug("block_f to blockTx transformation fails")
		}
		event.Send(event.ActorConsensus,event.ActorP2P,blksyn_send)

		// for test 2018.08.02
		if TestTag == true {
			// fmt.Println("blk_syn_v:",blk_syn_v.Header)
			// fmt.Println("blk_syn_f:",blk_syn_f.Header)
			// fmt.Println("blksyn_send v:",blksyn_send.Blksyn.BlksynV.Header)
			// fmt.Println("blksyn_send f:",blksyn_send.Blksyn.BlksynF.Header)
			// fmt.Println("currentheader.PrevHash:",currentheader.PrevHash)
			// fmt.Println("before reset: currentheader.Hash:",currentheader.Hash)
			currentPreBlk,_ := actorC.currentLedger.GetTxBlock(config.ChainHash, actorC.currentHeader.PrevHash)
			// current_blk := blk_syn_f
			//err1 := actorC.serviceAbabft.ledger.ResetStateDB(currentPreBlk.Header.StateHash)
			err1 := actorC.serviceAbabft.ledger.ResetStateDB(config.ChainHash, currentPreBlk.Header)
			if err1 != nil {
				fmt.Println("reset status error:", err1)
			}
			// blockFirstCal,err = actorC.serviceAbabft.ledger.NewTxBlock(current_blk.Transactions,current_blk.Header.ConsensusData, current_blk.Header.TimeStamp)
			// fmt.Println("current_blk.hash verfigy:",current_blk.Header.Hash, currentheader.Hash)
			// fmt.Println("compare merkle hash:", current_blk.Header.MerkleHash, blockFirstCal.MerkleHash)
			// fmt.Println("compare state hash:", current_blk.Header.StateHash, blockFirstCal.StateHash)

			// currentheader = current_ledger.GetCurrentHeader()
			old_block,_ := actorC.currentLedger.GetTxBlock(config.ChainHash, actorC.currentHeader.PrevHash)
			// currentheader = old_block.Header
			actorC.currentHeaderData = *(old_block.Header)
			actorC.currentHeader = &actorC.currentHeaderData

			// fmt.Println("after reset: currentheader.Hash:",currentheader.Hash)
			actorC.currentHeightNum = actorC.currentHeightNum - 1
			actorC.verifiedHeight = uint64(actorC.currentHeightNum) - 1
			event.Send(event.ActorNil,event.ActorConsensus,blksyn_send)
		}
		// test end
	case REQSynSolo:
		log.Info("receive the solo block requirement:",msg.Reqsyn.RequestHeight)
		// receive the solo synchronization request
		height_req := msg.Reqsyn.RequestHeight
		pubkey_in := msg.Reqsyn.PubKey
		signdata_in := msg.Reqsyn.SigData
		// check the required height
		if height_req > uint64(actorC.currentHeightNum) {
			return
		}
		// check the signature of the request message
		hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(height_req)))
		var sign_verify bool
		sign_verify, err = secp256k1.Verify(hash_t.Bytes(), signdata_in, pubkey_in)
		if sign_verify != true {
			println("Solo Syn request message signature is wrong")
			return
		}

		for i := int(height_req); i <= actorC.currentHeightNum; i++ {
			// get the response blocks from the ledger
			blkSynSolo,err1 := actorC.serviceAbabft.ledger.GetTxBlockByHeight(config.ChainHash, uint64(i))
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
			syn_status = 1
		}
		// test end

		if syn_status != 1 {
			return
		}
		var blks_v types.Block
		var blks_f types.Block
		err = blks_v.BlkTx2Blk(*msg.Blksyn.BlksynV)
		if err != nil {
			log.Debug("blockTx to block_v transformation fails")
		}
		err = blks_f.BlkTx2Blk(*msg.Blksyn.BlksynF)
		if err != nil {
			log.Debug("blockTx to block_f transformation fails")
		}
		// fmt.Println("blks_v:",blks_v.Header)
		// fmt.Println("blks_f:",blks_f.Header)

		// for test 2018.08.06
		if TestTag == true {
			fmt.Println("height_syn_v:",blks_v.Header.Height,actorC.currentHeightNum,actorC.verifiedHeight)
			// fmt.Println("blks_v.Header:",blks_v.Header)
		}
		// test end


		height_syn_v := blks_v.Header.Height
		if height_syn_v == (actorC.verifiedHeight+1) {
			// the current_height_num has been verified
			// 1. verify the verified block blks_v

			// todo
			// maybe only check the hash is enough

			var resultV bool
			var blk_v_local *types.Block
			blk_v_local,err = actorC.serviceAbabft.ledger.GetTxBlockByHeight(config.ChainHash, actorC.verifiedHeight)
			if err != nil {
				log.Debug("get previous block error")
				return
			}

			if ok := bytes.Equal(blks_v.Hash.Bytes(),actorC.currentHeader.Hash.Bytes()); ok == true {
				// the blks_v is the same as current block, just to verify and save blks_f
				resultV = true
				blks_v.Header = actorC.currentHeader
				fmt.Println("already have")

			} else {
				// verify blks_v
				resultV,err = actorC.blkSynVerify(blks_v, *blk_v_local)
				fmt.Println("have not yet")
			}


			// for test 2018.08.06
			if TestTag == true {
				// fmt.Println("blks_v.Hash:",blks_v.Hash)
				// fmt.Println("currentheader.Hash:",currentheader.Hash)
				if ok := bytes.Equal(blks_v.Hash.Bytes(),actorC.currentHeader.Hash.Bytes()); ok == true {
					resultV = true
				}
			}
			// test end

			if resultV == false {
				log.Debug("verification of blks_v fails")
				return
			}
			// 2. verify the verified block blks_f
			var result_f bool
			result_f,err = actorC.blkSynVerify(blks_f, blks_v)
			if result_f == false {
				log.Debug("verification of blks_f fails")
				return
			}
			// 3. save the blocks
			// 3.1 save blks_v
			if ok := bytes.Equal(blks_v.Hash.Bytes(), actorC.currentHeader.Hash.Bytes()); ok != true {
				// the blks_v is not in the ledger,then save blks_v
				// here need one reset DB
				//err = actorC.serviceAbabft.ledger.ResetStateDB(blk_pre.Header.Hash)
				if actorC.verifiedHeight < uint64(actorC.currentHeightNum) {
					err = actorC.serviceAbabft.ledger.ResetStateDB(config.ChainHash, blk_v_local.Header)
					if err != nil {
						log.Debug("reset state db error:", err)
						return
					}
				}
				/*
				if err = actorC.serviceAbabft.ledger.SaveTxBlock(&blks_v); err != nil {
					log.Debug("save block error:", err)
					return
				}
				*/
				if err := event.Send(event.ActorNil, event.ActorLedger, &blks_v); err != nil {
					log.Fatal(err)
					// return
				}
				if err := event.Send(event.ActorConsensus, event.ActorP2P, &blks_v); err != nil {
					log.Fatal(err)
					// return
				}
			}  else {
				// the blks_v has been in the ledger
			}
			// 3.2 save blks_f
			/*
			if err = actorC.serviceAbabft.ledger.SaveTxBlock(&blks_f); err != nil {
				log.Debug("save block error:", err)
				return
			}
			*/
			if err := event.Send(event.ActorNil, event.ActorLedger, &blks_f); err != nil {
				log.Fatal(err)
				// return
			}
			if err := event.Send(event.ActorConsensus, event.ActorP2P, &blks_f); err != nil {
				log.Fatal(err)
				// return
			}

			// 4. only the block is sucessfully saved, then change the status
			// currentheader = blks_f.Header
			actorC.currentHeaderData = *(blks_f.Header)
			actorC.currentHeader = &actorC.currentHeaderData

			actorC.verifiedHeight = blks_v.Height
			actorC.status = 8
			actorC.primaryTag = 0

			// update the current_round_num
			blk_roundnum := int(blks_v.ConsensusData.Payload.(*types.AbaBftData).NumberRound)
			if actorC.currentRoundNum < blk_roundnum {
				actorC.currentRoundNum = blk_roundnum
			}

			// start/enter the next turn
			event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})

			// for test 2018.08.07
			if TestTag == true {
				fmt.Println("save block by syn:",blks_f.Header.Height,blks_f.Header.Hash,blks_f.Header.StateHash,blks_f.Header.MerkleHash)
			}
			// test end

			// todo
			// take care of save and reset

		} else if height_syn_v >uint64(actorC.currentHeightNum) {
			// the verified block has bigger height
			// send synchronization message
			var requestsyn REQSyn
			requestsyn.Reqsyn.PubKey = actorC.serviceAbabft.account.PublicKey
			hash_t,_ := common.DoubleHash(Uint64ToBytes(actorC.verifiedHeight+1))
			requestsyn.Reqsyn.SigData,_ = actorC.serviceAbabft.account.Sign(hash_t.Bytes())
			requestsyn.Reqsyn.RequestHeight = actorC.verifiedHeight+1
			event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
			syn_status = 1
		}
		// todo
		// only need to check the hash and signature is enough?
		// this may help to speed up the ababft
		return

	case TimeoutMsg:
		// todo
		// the waiting time maybe need to be longer after every time out

		pubkey_in := msg.Toutmsg.PubKey
		round_in := int(msg.Toutmsg.RoundNumber)
		signdata_in := msg.Toutmsg.SigData
		fmt.Println("receive the TimeoutMsg:",pubkey_in,round_in,actorC.currentRoundNum)
		// check the peer in the peers list
		if round_in < actorC.currentRoundNum {
			return
		}
		// check the signature
		hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(round_in)))
		var sign_verify bool
		sign_verify, err = secp256k1.Verify(hash_t.Bytes(), signdata_in, pubkey_in)
		if sign_verify != true {
			println("time out message signature is wrong")
			return
		}
		/*
		for _, peer := range Peers_list {
			if ok := bytes.Equal(peer.PublicKey, pubkey_in); ok == true {
				// legal peer
				// fmt.Println("TimeoutMsgs:",TimeoutMsgs)
				if _, ok1 := TimeoutMsgs[string(pubkey_in)]; ok1 != true {
					TimeoutMsgs[string(pubkey_in)] = round_in
					//fmt.Println("TimeoutMsgs, add:",TimeoutMsgs[string(pubkey_in)])
				} else if TimeoutMsgs[string(pubkey_in)] >= round_in {
					return
				}

				TimeoutMsgs[string(pubkey_in)] = round_in
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
			peerAddrIns := common.AddressFromPubKey(pubkey_in)
			if ok := bytes.Equal(peerAddr.AccAddress.Bytes(), peerAddrIns.Bytes()); ok == true {
				// legal peer
				// fmt.Println("TimeoutMsgs:",TimeoutMsgs)
				if _, ok1 := TimeoutMsgs[peerAddrIns.HexString()]; ok1 != true {
					TimeoutMsgs[peerAddrIns.HexString()] = round_in
					//fmt.Println("TimeoutMsgs, add:",TimeoutMsgs[string(pubkey_in)])
				} else if TimeoutMsgs[peerAddrIns.HexString()] >= round_in {
					return
				}

				TimeoutMsgs[peerAddrIns.HexString()] = round_in
				// to count the number is enough
				var countRS [1000]int
				var maxR int
				maxR = 0
				for _,v := range TimeoutMsgs {
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
						event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
						// fmt.Println("reset according to the timeout msg:",i,maxR,current_round_num,countRS[i])
						break
					}
				}
				break
			}
		}

		// change public key to account address

		return

	default :
		log.Debug(msg)
		log.Warn("unknown message", reflect.TypeOf(ctx.Message()))
		return
	}
}

func (actorC *ActorAbabft) verifyHeader(blockIn *types.Block, currentRoundNumIn int, curHeader types.Header) (bool,error){
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
	// err = actorC.serviceAbabft.ledger.ResetStateDB(curHeader.Hash)
	// fmt.Println("after reset",err)

	// generate the blockFirstCal for comparison
	actorC.blockFirstCal,err = actorC.serviceAbabft.ledger.NewTxBlock(config.ChainHash, txs, conDataC, headerIn.TimeStamp)
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

func (actorC *ActorAbabft) updateBlock(blockFirst types.Block, conData types.ConsensusData) (types.Block,error){
	var blockSecond types.Block
	var err error
	headerIn := blockFirst.Header
	header, _ := types.NewHeader(headerIn.Version, config.ChainHash, headerIn.Height, headerIn.PrevHash, headerIn.MerkleHash,
		headerIn.StateHash, conData, headerIn.Bloom, headerIn.Receipt.BlockCpu, headerIn.Receipt.BlockNet, headerIn.TimeStamp)
	blockSecond = types.Block{Header:header, CountTxs:uint32(len(blockFirst.Transactions)), Transactions:blockFirst.Transactions,}
	return blockSecond,err
}

func (actorC *ActorAbabft) verifySignatures(dataBlksReceived *types.AbaBftData, preBlkHash common.Hash, curHeader *types.Header) (bool,error){
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
	err = actorC.serviceAbabft.ledger.checkPermission(0, "active",signsCurBlk)
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

func (actorC *ActorAbabft) blkSynVerify(blockIn types.Block, blkPre types.Block) (bool,error) {
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
