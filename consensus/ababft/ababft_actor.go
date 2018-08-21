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
type Actor_ababft struct {
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
	pid *actor.PID // actor pid
	service_ababft *Service_ababft
}

const(
	pubkey_tag = "ababft"
	signdata_tag = "ababft"
)

var log = elog.NewLogger("ABABFT", elog.NoticeLog)

// to run the go test, please set TestTag to True
const TestTag = false

const threshold_round = 60

var Num_peers int
var Peers_list []Peer_info // Peer information for consensus
var Peers_addr_list []Peer_addr_info // Peer address information for consensus
var Peers_list_account []Peer_info_account // Peer information for consensus
var Self_index int // the index of this peer in the peers list
var current_round_num int // current round number
var current_height_num int // current height, according to the blocks saved in the local ledger
var current_ledger ledger.Ledger

var primary_tag int // 0: verification peer; 1: is the primary peer, who generate the block at current round;
// var signature_preblock_list [][]byte // list for saving the signatures for the previous block
var signature_preblock_list []common.Signature // list for saving the signatures for the previous block
// var signature_BlkF_list [][]byte // list for saving the signatures for the first round block
var signature_BlkF_list []common.Signature // list for saving the signatures for the first round block
var block_firstround Block_FirstRound // temporary parameters for the first round block
var block_secondround Block_SecondRound // temporary parameters for the second round block
var currentheader *types.Header // temporary parameters for the current block header, according to the blocks saved in the local ledger
var current_payload types.AbaBftData // temporary parameters for current payload
var received_signpre_num int // the number of received signatures for the previous block
var cache_signature_preblk []pb.SignaturePreblock // cache the received signatures for the previous block
var block_first_cal *types.Block // cache the first-round block
var received_signblkf_num int // temporary parameters for received signatures for first round block
var TimeoutMsgs = make(map[string]int, 1000) // cache the timeout message
var verified_height uint64

var delta_roundnum int

var syn_status int
// for test 2018.07.31
var Accounts_test []account.Account
// test end



func Actor_ababft_gen(actor_ababft *Actor_ababft) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor {
		return actor_ababft
	})
	pid, err := actor.SpawnNamed(props, "Actor_ababft")
	if err != nil {
		return nil, err
	}
	event.RegisterActor(event.ActorConsensus, pid)
	syn_status = 0
	return pid, err
}

func (actor_c *Actor_ababft) Receive(ctx actor.Context) {
	var err error
	// log.Debug("ababft service receives the message")

	// deal with the message
	switch msg := ctx.Message().(type) {
	case message.ABABFTStart:
		actor_c.status = 2
		log.Debug("start ababft: receive the ababftstart message:", current_height_num,verified_height,current_ledger.GetCurrentHeader())

		// check the status of the main net
		if ok:=current_ledger.StateDB().RequireVotingInfo(); ok!=true {
			// main net has not started yet
			currentheader = current_ledger.GetCurrentHeader()
			current_height_num = int(currentheader.Height)
			current_round_num = 0
			verified_height = uint64(current_height_num) - 1

			log.Debug("ababft is in solo mode!")
			// if soloaccount.PrivateKey != nil {
			if config.StartNode == true {
				// is the solo prime
				actor_c.status = 101
				// generate the solo block
				// consensus data
				var signpre_send []common.Signature
				signpre_send = append(signpre_send, currentheader.Signatures[0])
				conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(current_round_num),signpre_send}}
				// tx list
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
				// generate the block in the form of second round block
				var block_solo *types.Block
				t_time := time.Now().UnixNano()
				block_solo,err = actor_c.service_ababft.ledger.NewTxBlock(txs, conData, t_time)
				block_solo.SetSignature(&soloaccount)
				block_secondround.Blocksecond = *block_solo
				// save (the ledger will broadcast the block after writing the block into the DB)
				if err = actor_c.service_ababft.ledger.SaveTxBlock(block_solo); err != nil {
					// log.Error("save block error:", err)
					println("save solo block error:", err)
					return
				}
				fmt.Println("ababft solo height:",block_solo.Height,block_solo)
				time.Sleep(time.Second * WAIT_RESPONSE_TIME)
				// call itself again
				event.Send(event.ActorNil,event.ActorConsensus,message.ABABFTStart{})
			} else {
				// is the solo peer
				actor_c.status = 102
				// todo
				// no need every time to send a request for solo block

				// send solo syn request
				var requestsyn REQSynSolo
				requestsyn.Reqsyn.PubKey = actor_c.service_ababft.account.PublicKey
				hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(current_height_num+1)))
				requestsyn.Reqsyn.SigData,_ = actor_c.service_ababft.account.Sign(hash_t.Bytes())
				requestsyn.Reqsyn.RequestHeight = uint64(current_height_num)+1
				event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
				log.Info("send solo block requirements:", requestsyn.Reqsyn.RequestHeight, current_height_num)
			}
			return
		}

		// initialization
		// clear and initialize the signature preblock array

		// update the peers list by accountname
		newPeers,err := current_ledger.GetProducerList()
		if err != nil {
			log.Debug("fail to get peer list.")
		}
		log.Debug("ababft now enter into the ababft mode:",newPeers[0],newPeers[1])

		Num_peers = len(newPeers)
		var Peers_list_account_t = make([]string, Num_peers)
		for i := 0; i < Num_peers; i++ {
			// Peers_list_account_t = append(Peers_list_account_t,common.IndexToName(newPeers[i]))
			Peers_list_account_t[i] = newPeers[i].String()
		}
		log.Debug("ababft now enter into the ababft mode:Peers_list_account_t",Peers_list_account_t)
		// sort newPeers
		sort.Strings(Peers_list_account_t)

		Peers_list_account = make([]Peer_info_account, Num_peers)
		Peers_addr_list = make([]Peer_addr_info, Num_peers)
		for i := 0; i < Num_peers; i++ {
			Peers_list_account[i].Accountname = common.NameToIndex(Peers_list_account_t[i])
			Peers_list_account[i].Index = i + 1

			account_info,err := current_ledger.AccountGet(Peers_list_account[i].Accountname)
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
					Peers_addr_list[i].AccAdress = addr_key
					break;
				}
			}
			Peers_addr_list[i].Index = i + 1

			if uint64(selfaccountname) == uint64(Peers_list_account[i].Accountname) {
				// update Self_index, i.e. the corresponding index in the peer address list
				Self_index = i + 1
			}
		}

		fmt.Println("Peers_addr_list:",Peers_addr_list)

		signature_preblock_list = make([]common.Signature, len(Peers_addr_list))
		signature_BlkF_list = make([]common.Signature, len(Peers_addr_list))
		block_firstround = Block_FirstRound{}
		block_secondround = Block_SecondRound{}
		// log.Debug("current_round_num:",current_round_num,Num_peers,Self_index)
		// get the current round number of the block
		currentheader = current_ledger.GetCurrentHeader()
		current_height_num = int(currentheader.Height)

		// todo
		// check following patch:
		// add threshold_round to solve the liveness problem
		lastest_roundnum := int(currentheader.ConsensusData.Payload.(*types.AbaBftData).NumberRound)
		delta_roundnum = current_round_num - lastest_roundnum
		if delta_roundnum > threshold_round && current_height_num > int(verified_height) {
			// as there is a long time since last block, maybe the chain is blocked somewhere
			// to generate the block after the previous block (i.e. the latest verified block)
			var currentblock *types.Block
			currentblock,err = current_ledger.GetTxBlock(currentheader.PrevHash)
			if err != nil {
				fmt.Println("get previous block error.")
			}
			currentheader = currentblock.Header
			current_height_num = current_height_num - 1

			// todo
			// 1. the ledger needs one backward step
			// 2. the peer list also needs one backward step
			// 3. the txpool also needs one backward step or maybe not
			// 4. the blockchain in database needs one backward step
		}

		if currentheader.ConsensusData.Type != types.ConABFT {
			//log.Warn("wrong ConsensusData Type")
			return
		}
		if v,ok:= currentheader.ConsensusData.Payload.(* types.AbaBftData); ok {
			current_payload = *v
		}

		// todo
		// the update of current_round_num
		// current_round_num = int(current_payload.NumberRound)
		// the timeout/changeview message
		// need to check whether the update of current_round_num is necessary


		// signature the current highest block and broadcast
		var signature_preblock common.Signature
		signature_preblock.PubKey = actor_c.service_ababft.account.PublicKey
		signature_preblock.SigData, err = actor_c.service_ababft.account.Sign(currentheader.Hash.Bytes())
		if err != nil {
			return
		}

		// check whether self is the prime or peer
		if current_round_num % Num_peers == (Self_index-1) {
			// if is prime
			primary_tag = 1
			actor_c.status = 3
			received_signpre_num = 0
			// increase the round index
			current_round_num ++
			fmt.Println("ABABFTStart:current_round_num:",current_round_num,Self_index)
			// log.Debug("primary")
			// set up a timer to wait for the signature_preblock from other peera
			t0 := time.NewTimer(time.Second * WAIT_RESPONSE_TIME * 2)
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
			primary_tag = 0
			actor_c.status = 5
			// broadcast the signature_preblock and set up a timer for receiving the data
			var signaturepre_send Signature_Preblock
			signaturepre_send.Signature_preblock.PubKey = signature_preblock.PubKey
			signaturepre_send.Signature_preblock.SigData = signature_preblock.SigData
			// todo
			// for the signature of previous block, maybe the round number is not needed
			signaturepre_send.Signature_preblock.Round = uint32(current_round_num)
			signaturepre_send.Signature_preblock.Height = uint32(currentheader.Height)
			// broadcast
			event.Send(event.ActorConsensus, event.ActorP2P, signaturepre_send)
			// increase the round index
			current_round_num ++
			fmt.Println("ABABFTStart:current_round_num(non primary):",current_round_num,Self_index)
			// log.Debug("non primary")
			// log.Debug("signaturepre_send:",current_round_num,currentheader.Height,signaturepre_send)
			// set up a timer for receiving the data
			t1 := time.NewTimer(time.Second * WAIT_RESPONSE_TIME * 2)
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

	case Signature_Preblock:
		log.Info("receive the preblock signature:",actor_c.status,msg.Signature_preblock)
		if actor_c.status == 102 {
			event.Send(event.ActorNil,event.ActorConsensus,message.ABABFTStart{})
		}
		// the prime will verify the signature for the previous block
		round_in := int(msg.Signature_preblock.Round)
		height_in := int(msg.Signature_preblock.Height)
		// log.Debug("current_round_num:",current_round_num,round_in)
		if round_in >= current_round_num && actor_c.status!=101 && actor_c.status!= 102 {
			// cache the Signature_Preblock
			cache_signature_preblk = append(cache_signature_preblk,msg.Signature_preblock)
			// in case that the signature for the previous block arrived bofore the corresponding block generator was born
		}
		if primary_tag == 1 && (actor_c.status == 2 || actor_c.status == 3){
			// verify the signature
			// first check the round number and height

			// todo
			// maybe round number is not needed for preblock signature
			if round_in >= (current_round_num-1) && height_in >= current_height_num {
				if round_in > (current_round_num - 1) && height_in > current_height_num {
					// todo
					// need to check
					// only require when height difference between the peers is >= 2

					if delta_roundnum > threshold_round {
						if verified_height == uint64(current_height_num) && height_in == (current_height_num+1) {
							return
						}
					}

					// require synchronization, the longest chain is ok
					// send synchronization message
					var requestsyn REQSyn
					requestsyn.Reqsyn.PubKey = actor_c.service_ababft.account.PublicKey
					hash_t,_ := common.DoubleHash(Uint64ToBytes(verified_height+1))
					requestsyn.Reqsyn.SigData,_ = actor_c.service_ababft.account.Sign(hash_t.Bytes())
					requestsyn.Reqsyn.RequestHeight = verified_height+1
					event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
					syn_status = 1
					// todo
					// attention
					// to against the height cheat, do not change the actor_c.status
				} else {
					// check the signature
					pubkey_in := msg.Signature_preblock.PubKey// signaturepre_send.signature_preblock.PubKey = signature_preblock.PubKey
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
					for index,peer_addr := range Peers_addr_list {
						peer_addr_in := common.AddressFromPubKey(pubkey_in)
						if ok := bytes.Equal(peer_addr.AccAdress.Bytes(), peer_addr_in.Bytes()); ok == true {
							found_peer = true
							peer_index = index
							break
						}
					}

					if found_peer == false {
						// the signature is not from the peer in the list
						return
					}
					// 1. check that signature in or not in list of
					if signature_preblock_list[peer_index].SigData != nil {
						// already receive the signature
						return
					}
					// 2. verify the correctness of the signature
					sigdata_in := msg.Signature_preblock.SigData
					header_hash := currentheader.Hash.Bytes()
					var result_verify bool
					result_verify, err = secp256k1.Verify(header_hash, sigdata_in, pubkey_in)
					if result_verify == true {
						// add the incoming signature to signature preblock list
						signature_preblock_list[peer_index].SigData = sigdata_in
						signature_preblock_list[peer_index].PubKey = pubkey_in
						received_signpre_num ++
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
		if primary_tag == 1 && (actor_c.status == 2 || actor_c.status == 3){
			// 1. check the cache cache_signature_preblk
			header_hash := currentheader.Hash.Bytes()
			for _,signpreblk := range cache_signature_preblk {
				round_in := signpreblk.Round
				if int(round_in) != current_round_num {
					continue
				}
				// check the signature
				pubkey_in := signpreblk.PubKey// signaturepre_send.signature_preblock.PubKey = signature_preblock.PubKey
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
				for index,peer_addr := range Peers_addr_list {
					peer_addr_in := common.AddressFromPubKey(pubkey_in)
					if ok := bytes.Equal(peer_addr.AccAdress.Bytes(), peer_addr_in.Bytes()); ok == true {
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
				if signature_preblock_list[peer_index].SigData != nil {
					// already receive the signature
					continue
				}
				// second, verify the correctness of the signature
				sigdata_in := signpreblk.SigData
				var result_verify bool
				result_verify, err = secp256k1.Verify(header_hash, sigdata_in, pubkey_in)
				if result_verify == true {
					// add the incoming signature to signature preblock list
					signature_preblock_list[peer_index].SigData = sigdata_in
					signature_preblock_list[peer_index].PubKey = pubkey_in
					received_signpre_num ++
				} else {
					continue
				}

			}
			// clean the cache_signature_preblk
			// cache_signature_preblk = make([]pb.SignaturePreblock,len(Peers_list)*2)
			cache_signature_preblk = make([]pb.SignaturePreblock,len(Peers_addr_list)*2)
			// fmt.Println("valid sign_pre:",received_signpre_num)
			// fmt.Println("current status root hash:",currentheader.StateHash)

			// 2. check the number of the preblock signature
			if received_signpre_num >= int(len(Peers_addr_list)/3+1) {
				// enough preblock signature, so generate the first-round block, only including the preblock signatures and
				// prepare the ConsensusData
				var signpre_send []common.Signature
				for _,signpre := range signature_preblock_list {
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
				conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(current_round_num),signpre_send}}
				// fmt.Println("conData for blk firstround",conData)
				// prepare the tx list
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
				}
				// log.Debug("obtained tx list", txs[0])
				// generate the first-round block
				var block_first *types.Block
				t_time := time.Now().UnixNano()
				block_first,err = actor_c.service_ababft.ledger.NewTxBlock(txs, conData, t_time)
				block_first.SetSignature(actor_c.service_ababft.account)
				// broadcast the first-round block to peers for them to verify the transactions and wait for the corresponding signatures back
				block_firstround.Blockfirst = *block_first
				event.Send(event.ActorConsensus, event.ActorP2P, block_firstround)
				// log.Debug("first round block:",block_firstround.Blockfirst)
				// fmt.Println("first round block status root hash:",block_first.StateHash)
				log.Info("generate the first round block and send",block_firstround.Blockfirst.Height)

				// for test 2018.07.27
				if TestTag == true {
					event.Send(event.ActorNil,event.ActorConsensus,block_firstround)
					// log.Debug("first round block:",block_firstround.Blockfirst.Header)
				}
				// test end


				// change the statue
				actor_c.status = 4
				// initial the received_signblkf_num to count the signatures for txs (i.e. the first round block)
				received_signblkf_num = 0
				// set the timer for collecting the signature for txs (i.e. the first round block)
				t2 := time.NewTimer(time.Second * WAIT_RESPONSE_TIME)
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
				actor_c.status = 7
				primary_tag = 0 // reset to zero, and the next primary will take the turn
				// send out the timeout message
				var timeoutmsg TimeoutMsg
				timeoutmsg.Toutmsg.RoundNumber = uint64(current_round_num)
				timeoutmsg.Toutmsg.PubKey = actor_c.service_ababft.account.PublicKey
				hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(current_round_num)))
				timeoutmsg.Toutmsg.SigData,_ = actor_c.service_ababft.account.Sign(hash_t.Bytes())
				event.Send(event.ActorConsensus,event.ActorP2P,timeoutmsg)
				// start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
			}
		} else {
			return
		}

	case Block_FirstRound:
		// for test 2018.07.27
		if TestTag == true {
			primary_tag = 0
			actor_c.status = 5
			// log.Debug("debug for first round block")
		}
		// end of test

		log.Info("current height and receive the first round block:",current_height_num, msg.Blockfirst.Header)

		if primary_tag == 0 && (actor_c.status == 2 || actor_c.status == 5) {
			// to verify the first round block
			blockfirst_received := msg.Blockfirst
			// the protocal type is ababft
			if blockfirst_received.ConsensusData.Type == types.ConABFT {
				data_preblk_received := blockfirst_received.ConsensusData.Payload.(*types.AbaBftData)
				// 1. check the round number
				// 1a. current round number
				if data_preblk_received.NumberRound < uint32(current_round_num) {
					return
				} else if data_preblk_received.NumberRound > uint32(current_round_num) {
					// require synchronization, the longest chain is ok
					// in case that somebody may skip the current generator, only the different height can call the synchronization
					if (verified_height+2) < blockfirst_received.Header.Height {
						// send synchronization message
						var requestsyn REQSyn
						requestsyn.Reqsyn.PubKey = actor_c.service_ababft.account.PublicKey
						hash_t,_ := common.DoubleHash(Uint64ToBytes(verified_height+1))
						requestsyn.Reqsyn.SigData,_ = actor_c.service_ababft.account.Sign(hash_t.Bytes())
						requestsyn.Reqsyn.RequestHeight = verified_height+1
						event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
						syn_status = 1
						// todo
						// attention:
						// to against the height cheat, do not change the actor_c.status
					}
				} else {
					// 1b. the round number corresponding to the block generator
					index_g := (current_round_num-1) % Num_peers + 1
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
					for _, peer_addr := range Peers_addr_list {
						peer_addr_in := common.AddressFromPubKey(pukey_g_in)
						if ok := bytes.Equal(peer_addr.AccAdress.Bytes(), peer_addr_in.Bytes()); ok == true {
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
					valid_blk,err = actor_c.verify_header(&blockfirst_received, current_round_num, *currentheader)
					if valid_blk==false {
						println("header check fail")
						return
					}
					// 2. check the preblock signature
					sign_preblk_list := data_preblk_received.PerBlockSignatures
					header_hash := currentheader.Hash.Bytes()
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
						for _, peer_addr := range Peers_addr_list {
							peer_addr_in := common.AddressFromPubKey(sign_preblk.PubKey)
							if ok := bytes.Equal(peer_addr.AccAdress.Bytes(), peer_addr_in.Bytes()); ok == true {
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
					if num_verified < int(len(Peers_addr_list)/3+1){
						// not enough signature
						fmt.Println("not enough signature for second round block")
						return
					}
					// 3. check the txs
					txs_in := blockfirst_received.Transactions
					for index1,tx_in := range txs_in {
						err = actor_c.service_ababft.ledger.CheckTransaction(tx_in)
						if err != nil {
							println("wrong tx, index:", index1)
							return
						}
					}
					// 4. sign the received first-round block
					var sign_blkf_send Signature_BlkF
					sign_blkf_send.Signature_blkf.PubKey = actor_c.service_ababft.account.PublicKey
					sign_blkf_send.Signature_blkf.SigData,err = actor_c.service_ababft.account.Sign(blockfirst_received.Header.Hash.Bytes())
					// 5. broadcast the signature of the first round block
					event.Send(event.ActorConsensus, event.ActorP2P, sign_blkf_send)
					// 6. change the status
					actor_c.status = 6
					// fmt.Println("sign_blkf_send:",sign_blkf_send)
					// clean the cache_signature_preblk
					cache_signature_preblk = make([]pb.SignaturePreblock,len(Peers_addr_list)*2)
					// send the received first-round block to other peers in case that network is not good
					block_firstround.Blockfirst = blockfirst_received
					event.Send(event.ActorConsensus,event.ActorP2P,block_firstround)
					log.Info("generate the signature for first round block",block_firstround.Blockfirst.Height)

					// for test 2018.07.31
					if TestTag == true {
						primary_tag = 1
						actor_c.status = 4
						// create the signature for first-round block for test
						for i:=0;i<Num_peers;i++ {
							var sign_blkf_send1 Signature_BlkF
							sign_blkf_send1.Signature_blkf.PubKey = Accounts_test[i].PublicKey
							sign_blkf_send1.Signature_blkf.SigData,err = Accounts_test[i].Sign(blockfirst_received.Header.Hash.Bytes())
							// fmt.Println("Accounts_test:",i,Accounts_test[i].PublicKey,sign_blkf_send1.Signature_blkf)
							event.Send(event.ActorNil, event.ActorConsensus, sign_blkf_send1)
						}
						fmt.Println("blockfirst_received.Header.Hash:",data_preblk_received.NumberRound,blockfirst_received.Header.Hash, blockfirst_received.Header.MerkleHash,blockfirst_received.Header.StateHash)
					}
					// test end


					// 7. set the timer for waiting the second-round(final) block
					t3 := time.NewTimer(time.Second * WAIT_RESPONSE_TIME)
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
			// actor_c.status = 5
			fmt.Println("timeout test needs to be specified")
		}
		// end test

		if primary_tag == 0 && (actor_c.status == 2 || actor_c.status == 5) {
			// not receive the first round block
			// change the status
			actor_c.status = 8
			primary_tag = 0
			// send out the timeout message
			var timeoutmsg TimeoutMsg
			timeoutmsg.Toutmsg.RoundNumber = uint64(current_round_num)
			timeoutmsg.Toutmsg.PubKey = actor_c.service_ababft.account.PublicKey
			hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(current_round_num)))
			timeoutmsg.Toutmsg.SigData,_ = actor_c.service_ababft.account.Sign(hash_t.Bytes())
			event.Send(event.ActorConsensus,event.ActorP2P,timeoutmsg)

			// for test 2018.08.01
			if TestTag == true {
				primary_tag = 0
				actor_c.status = 5
				for i:=0;i<Num_peers;i++ {
					var timeoutmsg1 TimeoutMsg
					timeoutmsg1.Toutmsg.RoundNumber = uint64(current_round_num+1)
					timeoutmsg1.Toutmsg.PubKey = Accounts_test[i].PublicKey
					hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(current_round_num+1)))
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

	case Signature_BlkF:
		// fmt.Println("Signature_BlkF:",received_signblkf_num,msg.Signature_blkf)
		// the prime will verify the signatures of first-round block from peers
		if primary_tag == 1 && actor_c.status == 4 {
			// verify the signature
			// 1. check the peer in the peers list
			pubkey_in := msg.Signature_blkf.PubKey
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
			for index,peer_addr := range Peers_addr_list {
				peer_addr_in := common.AddressFromPubKey(pubkey_in)
				if ok := bytes.Equal(peer_addr.AccAdress.Bytes(), peer_addr_in.Bytes()); ok == true {
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
			if signature_BlkF_list[peer_index].SigData != nil {
				// already receive the signature
				return
			}
			sigdata_in := msg.Signature_blkf.SigData
			header_hash := block_firstround.Blockfirst.Header.Hash.Bytes()
			var result_verify bool
			result_verify, err = secp256k1.Verify(header_hash, sigdata_in, pubkey_in)
			if result_verify == true {
				// add the incoming signature to signature preblock list
				signature_BlkF_list[peer_index].SigData = sigdata_in
				signature_BlkF_list[peer_index].PubKey = pubkey_in
				received_signblkf_num ++
				return
			} else {
				return
			}

		}

	case SignTxTimeout:
		// fmt.Println("received_signblkf_num:",received_signblkf_num)
		log.Info("start to generate second round block",primary_tag,actor_c.status,received_signblkf_num,int(2*len(Peers_addr_list)/3),signature_BlkF_list)
		if primary_tag == 1 && actor_c.status == 4 {
			// check the number of the signatures of first-round block from peers
			if received_signblkf_num >= int(2*len(Peers_addr_list)/3) {
				// enough first-round block signatures, so generate the second-round(final) block
				// 1. add the first-round block signatures into ConsensusData
				pubkey_tag_b := []byte(pubkey_tag)
				signdata_tag_b := []byte(signdata_tag)
				var sign_tag common.Signature
				sign_tag.PubKey = pubkey_tag_b
				sign_tag.SigData = signdata_tag_b

				ababftdata := block_firstround.Blockfirst.ConsensusData.Payload.(*types.AbaBftData)
				// prepare the ConsensusData
				// add the tag to distinguish preblock signature and second round signature
				ababftdata.PerBlockSignatures = append(ababftdata.PerBlockSignatures, sign_tag)
				for _,signblkf := range signature_BlkF_list {
					/*
					if signblkf != nil {
						var sign_tmp common.Signature
						sign_tmp.SigData = signblkf
						sign_tmp.PubKey = Peers_list[index].PublicKey
						ababftdata.PerBlockSignatures = append(ababftdata.PerBlockSignatures, sign_tmp)
					}
					*/
					// change public key to account address
					if signblkf.SigData != nil {
						var sign_tmp common.Signature
						sign_tmp.SigData = signblkf.SigData
						sign_tmp.PubKey = signblkf.PubKey
						ababftdata.PerBlockSignatures = append(ababftdata.PerBlockSignatures, sign_tmp)
					}
				}

				conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(current_round_num),ababftdata.PerBlockSignatures}}
				// 2. generate the second-round(final) block
				var block_second types.Block
				block_second,err =  actor_c.update_block(block_firstround.Blockfirst, conData)
				block_second.SetSignature(actor_c.service_ababft.account)
				// fmt.Println("block_second:",block_second.Header)

				// 3. broadcast the second-round(final) block
				block_secondround.Blocksecond = block_second
				// the ledger will multicast the block_secondround after the block is saved in the DB
				// event.Send(event.ActorConsensus, event.ActorP2P, block_secondround)

				// for test 2018.07.31
				if TestTag == true {
					for i:=0;i<Num_peers;i++ {
						primary_tag = 0
						actor_c.status = 6
						event.Send(event.ActorNil, event.ActorConsensus, block_secondround)
					}
					time.Sleep(time.Second * 10)
					return
				}
				//

				// 4. save the second-round(final) block to ledger
				if err = actor_c.service_ababft.ledger.SaveTxBlock(&block_second); err != nil {
					// log.Error("save block error:", err)
					println("save block error:", err)
					return
				}
				verified_height = block_second.Height - 1
				// 5. change the status
				actor_c.status = 7
				primary_tag = 0

				fmt.Println("save the generated block", block_second.Height,verified_height)
				// start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
				return
			} else {
				// 1. did not receive enough signatures of first-round block from peers in the assigned time interval
				actor_c.status = 7
				primary_tag = 0 // reset to zero, and the next primary will take the turn
				// 2. reset the stateDB
				//err = actor_c.service_ababft.ledger.ResetStateDB(currentheader.Hash)
				//if err != nil {
				//	log.Debug("ResetStateDB fail")
				//	return
				//}
				// send out the timeout message
				var timeoutmsg TimeoutMsg
				timeoutmsg.Toutmsg.RoundNumber = uint64(current_round_num)
				timeoutmsg.Toutmsg.PubKey = actor_c.service_ababft.account.PublicKey
				hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(current_round_num)))
				timeoutmsg.Toutmsg.SigData,_ = actor_c.service_ababft.account.Sign(hash_t.Bytes())
				event.Send(event.ActorConsensus,event.ActorP2P,timeoutmsg)
				// 3. start/enter the next turn
				event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
			}
		}

	case Block_SecondRound:
		// for test 2018.08.09
		if TestTag == true {
			fmt.Println("get second round block")
			primary_tag = 0
			actor_c.status = 6
		}
		// test end
		log.Info("ababbt peer status:", primary_tag,actor_c.status)
		// check whether it is solo mode
		if actor_c.status == 102 || actor_c.status == 101 {
			if actor_c.status == 102 {
				// solo peer
				blocksecond_received := msg.Blocksecond
				log.Info("ababbt solo block height vs current_height_num:",blocksecond_received.Header.Height,current_height_num)
				if int(blocksecond_received.Header.Height) <= current_height_num {
					return
				} else if int(blocksecond_received.Header.Height) == (current_height_num+1) {
					// check and save
					blocksecond_received := msg.Blocksecond
					if blocksecond_received.ConsensusData.Type == types.ConABFT {
						data_blks_received := blocksecond_received.ConsensusData.Payload.(*types.AbaBftData)
						// check the signature comes from the root
						if ok := bytes.Equal(blocksecond_received.Signatures[0].PubKey,config.Root.PublicKey); ok != true {
							println("the solo block should be signed by the root")
							return
						}

						// check the block header(the consensus data is null)
						var valid_blk bool
						valid_blk,err = actor_c.verify_header(&blocksecond_received, int(data_blks_received.NumberRound), *currentheader)
						if valid_blk==false {
							println("header check fail")
							return
						}
						// save the solo block ( in the form of second-round block)
						if err = actor_c.service_ababft.ledger.SaveTxBlock(&blocksecond_received); err != nil {
							println("save solo block error:", err)
							return
						}
						verified_height = blocksecond_received.Height - 1
						log.Info("verified height of the solo mode:",verified_height)
						event.Send(event.ActorNil, event.ActorConsensus, message.ABABFTStart{})
					}
				} else {
					// send solo syn request
					var requestsyn REQSynSolo
					requestsyn.Reqsyn.PubKey = actor_c.service_ababft.account.PublicKey
					hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(current_height_num+1)))
					requestsyn.Reqsyn.SigData,_ = actor_c.service_ababft.account.Sign(hash_t.Bytes())
					requestsyn.Reqsyn.RequestHeight = uint64(current_height_num)+1
					event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
					log.Info("send requirements:", requestsyn.Reqsyn.RequestHeight, current_height_num)
				}
			}

			return
		}

		if primary_tag == 0 && (actor_c.status == 6 || actor_c.status == 2 || actor_c.status == 5) {
			// to verify the first round block
			blocksecond_received := msg.Blocksecond
			// check the protocal type is ababft
			if blocksecond_received.ConsensusData.Type == types.ConABFT {
				data_blks_received := blocksecond_received.ConsensusData.Payload.(*types.AbaBftData)

				// for test 2018.08.09
				if TestTag == true {
					verified_height = uint64(current_height_num) - 1
					fmt.Println("blocksecond_received.Header.Height:",blocksecond_received.Header.Height,verified_height,current_height_num)
				}
				//

				log.Info("received secondround block:",blocksecond_received.Header.Height,verified_height,current_height_num,data_blks_received.NumberRound,blocksecond_received.Header)
				// 1. check the round number and height
				// 1a. current round number
				if data_blks_received.NumberRound < uint32(current_round_num) || blocksecond_received.Header.Height <= uint64(current_height_num) {
					return
				} else if (blocksecond_received.Header.Height-2) > verified_height {
					// send synchronization message
					var requestsyn REQSyn
					requestsyn.Reqsyn.PubKey = actor_c.service_ababft.account.PublicKey
					hash_t,_ := common.DoubleHash(Uint64ToBytes(verified_height+1))
					requestsyn.Reqsyn.SigData,_ = actor_c.service_ababft.account.Sign(hash_t.Bytes())
					requestsyn.Reqsyn.RequestHeight = verified_height+1
					event.Send(event.ActorConsensus,event.ActorP2P,requestsyn)
					syn_status = 1

					// todo
					// attention:
					// to against the height cheat, do not change the actor_c.status
				} else {
					// here, the add new block into the ledger, data_blks_received.NumberRound >= current_round_num is ok instead of data_blks_received.NumberRound == current_round_num
					// 1b. the round number corresponding to the block generator
					index_g := (int(data_blks_received.NumberRound)-1) % Num_peers + 1
					pukey_g_in := blocksecond_received.Signatures[0].PubKey
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
					for _, peer_addr := range Peers_addr_list {
						peer_addr_in := common.AddressFromPubKey(pukey_g_in)
						if ok := bytes.Equal(peer_addr.AccAdress.Bytes(), peer_addr_in.Bytes()); ok == true {
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
					valid_blk,err = actor_c.verify_header(&blocksecond_received, int(data_blks_received.NumberRound), *currentheader)
					// todo
					// can check the hash and statdb and merker root instead of the total head to speed up
					if valid_blk==false {
						println("header check fail")
						return
					}
					// 2. check the signatures ( for both previous and current blocks) in ConsensusData
					preblkhash := currentheader.Hash
					valid_blk, err = actor_c.verify_signatures(data_blks_received, preblkhash, blocksecond_received.Header)
					if valid_blk==false {
						println("previous and first-round blocks signatures check fail")
						return
					}

					// for test 2018.08.01
					if TestTag == true {
						fmt.Println("received and verified second round:", blocksecond_received.Height, blocksecond_received.Header.Hash, blocksecond_received.MerkleHash, blocksecond_received.StateHash)
						// return
					}
					// test end

					// 3.save the second-round block into the ledger
					if err = actor_c.service_ababft.ledger.SaveTxBlock(&blocksecond_received); err != nil {
						// log.Error("save block error:", err)
						println("save block error:", err)
						return
					}

					// 4. change status
					verified_height = blocksecond_received.Height - 1
					actor_c.status = 8
					primary_tag = 0
					// update the current_round_num
					if int(data_blks_received.NumberRound) > current_round_num {
						current_round_num = int(data_blks_received.NumberRound)
					}

					fmt.Println("Block_SecondRound,current_round_num:",current_round_num)
					// start/enter the next turn
					event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
					// 5. broadcast the received second-round block, which has been checked valid
					// to let other peer know this block
					block_secondround.Blocksecond = blocksecond_received
					// as the ledger will multicast the block after the block is saved in DB, so following code is not need any more
					// event.Send(event.ActorConsensus, event.ActorP2P, block_secondround)
					return
				}
			}
		}
	case BlockSTimeout:
		if primary_tag == 0 && actor_c.status == 5 {
			actor_c.status = 8
			primary_tag = 0
			// reset the state of merkle tree, statehash and so on
			// err = actor_c.service_ababft.ledger.ResetStateDB(currentheader.Hash)
			//if err != nil {
			//	log.Debug("ResetStateDB fail")
			//	return
			//}
			// send out the timeout message
			var timeoutmsg TimeoutMsg
			timeoutmsg.Toutmsg.RoundNumber = uint64(current_round_num)
			timeoutmsg.Toutmsg.PubKey = actor_c.service_ababft.account.PublicKey
			hash_t,_ := common.DoubleHash(Uint64ToBytes(uint64(current_round_num)))
			timeoutmsg.Toutmsg.SigData,_ = actor_c.service_ababft.account.Sign(hash_t.Bytes())
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

		if height_req > uint64(current_height_num - 1) {
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
		blk_syn_v,err1 := actor_c.service_ababft.ledger.GetTxBlockByHeight(height_req)
		if err1 != nil || blk_syn_v == nil {
			log.Debug("not find the block of the corresponding height in the ledger")
			return
		}

		// fmt.Println("blk_syn_v:",blk_syn_v.Header)
		blk_syn_f,err2 := actor_c.service_ababft.ledger.GetTxBlockByHeight(height_req+1)
		if err2 != nil || blk_syn_f == nil {
			log.Debug("not find the block of the corresponding height in the ledger")
			return
		}
		// 3. send the found /blocks
		var blksyn_send Block_Syn
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
			current_pre_blk,_ := current_ledger.GetTxBlock(currentheader.PrevHash)
			// current_blk := blk_syn_f
			//err1 := actor_c.service_ababft.ledger.ResetStateDB(current_pre_blk.Header.StateHash)
			err1 := actor_c.service_ababft.ledger.ResetStateDB(current_pre_blk.Header)
			if err1 != nil {
				fmt.Println("reset status error:", err1)
			}
			// block_first_cal,err = actor_c.service_ababft.ledger.NewTxBlock(current_blk.Transactions,current_blk.Header.ConsensusData, current_blk.Header.TimeStamp)
			// fmt.Println("current_blk.hash verfigy:",current_blk.Header.Hash, currentheader.Hash)
			// fmt.Println("compare merkle hash:", current_blk.Header.MerkleHash, block_first_cal.MerkleHash)
			// fmt.Println("compare state hash:", current_blk.Header.StateHash, block_first_cal.StateHash)

			// currentheader = current_ledger.GetCurrentHeader()
			old_block,_ := current_ledger.GetTxBlock(currentheader.PrevHash)
			currentheader = old_block.Header
			// fmt.Println("after reset: currentheader.Hash:",currentheader.Hash)
			current_height_num = current_height_num - 1
			verified_height = uint64(current_height_num) - 1
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
		if height_req > uint64(current_height_num) {
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

		for i := int(height_req); i <= current_height_num; i++ {
			// get the response blocks from the ledger
			blk_syn_solo,err1 := actor_c.service_ababft.ledger.GetTxBlockByHeight(uint64(i))
			if err1 != nil || blk_syn_solo == nil {
				log.Debug("not find the solo block of the corresponding height in the ledger")
				return
			}
			// send the solo block
			event.Send(event.ActorNil,event.ActorP2P,blk_syn_solo)
			log.Info("send the required solo block:", blk_syn_solo.Height)
		}
		return
	case Block_Syn:
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
			fmt.Println("height_syn_v:",blks_v.Header.Height,current_height_num,verified_height)
			// fmt.Println("blks_v.Header:",blks_v.Header)
		}
		// test end


		height_syn_v := blks_v.Header.Height
		if height_syn_v == (verified_height+1) {
			// the current_height_num has been verified
			// 1. verify the verified block blks_v

			// todo
			// maybe only check the hash is enough

			var result_v bool
			var blk_v_local *types.Block
			blk_v_local,err = actor_c.service_ababft.ledger.GetTxBlockByHeight(verified_height)
			if err != nil {
				log.Debug("get previous block error")
				return
			}

			if ok := bytes.Equal(blks_v.Hash.Bytes(),currentheader.Hash.Bytes()); ok == true {
				// the blks_v is the same as current block, just to verify and save blks_f
				result_v = true
				blks_v.Header = currentheader
				fmt.Println("already have")

			} else {
				// verify blks_v
				result_v,err = actor_c.Blk_syn_verify(blks_v, *blk_v_local)
				fmt.Println("have not yet")
			}


			// for test 2018.08.06
			if TestTag == true {
				// fmt.Println("blks_v.Hash:",blks_v.Hash)
				// fmt.Println("currentheader.Hash:",currentheader.Hash)
				if ok := bytes.Equal(blks_v.Hash.Bytes(),currentheader.Hash.Bytes()); ok == true {
					result_v = true
				}
			}
			// test end

			if result_v == false {
				log.Debug("verification of blks_v fails")
				return
			}
			// 2. verify the verified block blks_f
			var result_f bool
			result_f,err = actor_c.Blk_syn_verify(blks_f, blks_v)
			if result_f == false {
				log.Debug("verification of blks_f fails")
				return
			}
			// 3. save the blocks
			// 3.1 save blks_v
			if ok := bytes.Equal(blks_v.Hash.Bytes(), currentheader.Hash.Bytes()); ok != true {
				// the blks_v is not in the ledger,then save blks_v
				// here need one reset DB
				//err = actor_c.service_ababft.ledger.ResetStateDB(blk_pre.Header.Hash)
				if verified_height < uint64(current_height_num) {
					err = actor_c.service_ababft.ledger.ResetStateDB(blk_v_local.Header)
					if err != nil {
						log.Debug("reset state db error:", err)
						return
					}
				}
				if err = actor_c.service_ababft.ledger.SaveTxBlock(&blks_v); err != nil {
					log.Debug("save block error:", err)
					return
				}
			}  else {
				// the blks_v has been in the ledger
			}
			// 3.2 save blks_f
			if err = actor_c.service_ababft.ledger.SaveTxBlock(&blks_f); err != nil {
				log.Debug("save block error:", err)
				return
			}
			// 4. only the block is sucessfully saved, then change the status
			verified_height = blks_v.Height
			actor_c.status = 8
			primary_tag = 0

			// update the current_round_num
			blk_roundnum := int(blks_v.ConsensusData.Payload.(*types.AbaBftData).NumberRound)
			if current_round_num < blk_roundnum {
				current_round_num = blk_roundnum
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

		} else if height_syn_v >uint64(current_height_num) {
			// the verified block has bigger height
			// send synchronization message
			var requestsyn REQSyn
			requestsyn.Reqsyn.PubKey = actor_c.service_ababft.account.PublicKey
			hash_t,_ := common.DoubleHash(Uint64ToBytes(verified_height+1))
			requestsyn.Reqsyn.SigData,_ = actor_c.service_ababft.account.Sign(hash_t.Bytes())
			requestsyn.Reqsyn.RequestHeight = verified_height+1
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
		fmt.Println("receive the TimeoutMsg:",pubkey_in,round_in,current_round_num)
		// check the peer in the peers list
		if round_in < current_round_num {
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
						actor_c.status = 8
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
		for _, peer_addr := range Peers_addr_list {
			peer_addr_in := common.AddressFromPubKey(pubkey_in)
			if ok := bytes.Equal(peer_addr.AccAdress.Bytes(), peer_addr_in.Bytes()); ok == true {
				// legal peer
				// fmt.Println("TimeoutMsgs:",TimeoutMsgs)
				if _, ok1 := TimeoutMsgs[peer_addr_in.HexString()]; ok1 != true {
					TimeoutMsgs[peer_addr_in.HexString()] = round_in
					//fmt.Println("TimeoutMsgs, add:",TimeoutMsgs[string(pubkey_in)])
				} else if TimeoutMsgs[peer_addr_in.HexString()] >= round_in {
					return
				}

				TimeoutMsgs[peer_addr_in.HexString()] = round_in
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
					if total_count >= int(2*len(Peers_addr_list)/3) {
						// reset the round number
						current_round_num += i
						// start/enter the next turn
						actor_c.status = 8
						primary_tag = 0
						event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{})
						// fmt.Println("reset according to the timeout msg:",i,max_r,current_round_num,count_r[i])
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

func (actor_c *Actor_ababft) verify_header(block_in *types.Block, current_round_num_in int, cur_header types.Header) (bool,error){
	var err error
	header_in := block_in.Header
	txs := block_in.Transactions
	data_preblk_received := block_in.ConsensusData.Payload.(*types.AbaBftData)
	signpre_send := data_preblk_received.PerBlockSignatures
	condata_c := types.ConsensusData{Type:types.ConABFT, Payload:&types.AbaBftData{uint32(current_round_num_in),signpre_send}}
	// fmt.Println("data_preblk_received:",data_preblk_received)
	// fmt.Println("condata_c:",condata_c)
	// fmt.Println("before reset")
	// reset the stateDB
	// fmt.Println("cur_header state hash:",cur_header.Height,cur_header.StateHash)
	// err = actor_c.service_ababft.ledger.ResetStateDB(cur_header.Hash)
	// fmt.Println("after reset",err)

	// generate the block_first_cal for comparison
	block_first_cal,err = actor_c.service_ababft.ledger.NewTxBlock(txs,condata_c, header_in.TimeStamp)
	// fmt.Println("height:",block_in.Height,block_first_cal.Height)
	// fmt.Println("merkle:",block_in.Header.MerkleHash,block_first_cal.Header.MerkleHash)
	// fmt.Println("timestamp:",block_in.Header.TimeStamp,block_first_cal.Header.TimeStamp)
	// fmt.Println("block_first_cal:",block_first_cal.Header, block_first_cal.Header.StateHash)
	// fmt.Println("block_in:",block_in.Header, block_in.Header.StateHash)
	var num_txs int
	num_txs = int(block_in.CountTxs)
	if num_txs != len(txs) {
		println("tx number is wrong")
		return false,nil
	}
	// check Height        uint64
	if current_height_num >= int(header_in.Height) {
		println("the height is not higher than current height")
		return false,nil
	}
	// ConsensusData is checked in the Receive function

	// check PrevHash      common.Hash
	if ok :=bytes.Equal(block_in.PrevHash.Bytes(),cur_header.Hash.Bytes()); ok != true {
		println("prehash is wrong")
		return false,nil
	}
	// check MerkleHash    common.Hash
	if ok := bytes.Equal(block_first_cal.MerkleHash.Bytes(),block_in.MerkleHash.Bytes()); ok != true {
		println("MercleHash is wrong")
		return false,nil
	}
	// fmt.Println("mercle:",block_first_cal.MerkleHash.Bytes(),block_in.MerkleHash.Bytes())

	// check StateHash     common.Hash
	if ok := bytes.Equal(block_first_cal.StateHash.Bytes(),block_in.StateHash.Bytes()); ok != true {
		println("StateHash is wrong")
		return false,nil
	}
	// fmt.Println("statehash:",block_first_cal.StateHash.Bytes(),block_in.StateHash.Bytes())

	// check Bloom         bloom.Bloom
	if ok := bytes.Equal(block_first_cal.Bloom.Bytes(), block_in.Bloom.Bytes()); ok != true {
		println("bloom is wrong")
		return false,nil
	}
	// check Hash common.Hash
	header_cal,err1 := types.NewHeader(header_in.Version, common.NameToIndex("root").Number(), header_in.Height, header_in.PrevHash,
		header_in.MerkleHash, header_in.StateHash, header_in.ConsensusData, header_in.Bloom, header_in.Receipt.BlockCpu, header_in.Receipt.BlockNet,header_in.TimeStamp)
	if ok := bytes.Equal(header_cal.Hash.Bytes(),header_in.Hash.Bytes()); ok != true {
		println("Hash is wrong")
		return false,err1
	}
	// check Signatures    []common.Signature
	signpre_in := block_in.Signatures[0]
	pubkey_g_in := signpre_in.PubKey
	signdata_in := signpre_in.SigData
	var sign_verify bool
	sign_verify, err = secp256k1.Verify(header_in.Hash.Bytes(), signdata_in, pubkey_g_in)
	if sign_verify != true {
		println("signature is wrong")
		return false,err
	}
	return true,err
}

func (actor_c *Actor_ababft) update_block(block_first types.Block, condata types.ConsensusData) (types.Block,error){
	var block_second types.Block
	var err error
	header_in := block_first.Header
	header, _ := types.NewHeader(header_in.Version, common.NameToIndex("root").Number(), header_in.Height, header_in.PrevHash, header_in.MerkleHash,
		header_in.StateHash, condata, header_in.Bloom, header_in.Receipt.BlockCpu, header_in.Receipt.BlockNet, header_in.TimeStamp)
	block_second = types.Block{header, uint32(len(block_first.Transactions)), block_first.Transactions}
	return block_second,err
}

func (actor_c *Actor_ababft) verify_signatures(data_blks_received *types.AbaBftData, preblkhash common.Hash, curheader *types.Header) (bool,error){
	var err error
	// 1. devide the signatures into two part
	var sign_blks_preblk []common.Signature
	var sign_blks_curblk []common.Signature
	pubkey_tag_byte := []byte(pubkey_tag)
	sigdata_tag_byte := []byte(signdata_tag)
	var tag_sign int
	tag_sign = 0
	for _,sign := range data_blks_received.PerBlockSignatures {
		ok1 := bytes.Equal(sign.PubKey, pubkey_tag_byte);
		ok2 := bytes.Equal(sign.SigData,sigdata_tag_byte);
		if ok1 == true && ok2 == true {
			tag_sign = 1
			continue
		}
		if tag_sign == 0 {
			sign_blks_preblk = append(sign_blks_preblk,sign)
		} else if tag_sign == 1 {
			sign_blks_curblk = append(sign_blks_curblk,sign)
		}
	}
	// fmt.Println("sign_blks_preblk:",sign_blks_preblk)
	fmt.Println("sign_blks_curblk:",sign_blks_curblk)

	// 2. check the preblock signature
	var num_verified int
	num_verified = 0
	for index,sign_preblk := range sign_blks_preblk {
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
		for _, peer_addr := range Peers_addr_list {
			peer_addr_in := common.AddressFromPubKey(sign_preblk.PubKey)
			if ok := bytes.Equal(peer_addr.AccAdress.Bytes(), peer_addr_in.Bytes()); ok == true {
				peerin_tag = true
				break
			}
		}
		if peerin_tag == false {
			// there exists signature not from the peer list
			fmt.Println("the signature is not from the peer list, its index is:", index)
			return false,nil
		}
		// 2b. verify the correctness of the signature
		pubkey_in := sign_preblk.PubKey
		sigdata_in := sign_preblk.SigData
		var result_verify bool
		result_verify, err = secp256k1.Verify(preblkhash.Bytes(), sigdata_in, pubkey_in)
		if result_verify == true {
			num_verified++
		}
	}
	// 2c. check the valid signature number
	if num_verified < int(len(Peers_addr_list)/3+1){
		fmt.Println(" not enough signature for the previous block:", num_verified)
		return false,nil
	}


	// 3. check the current block signature
	num_verified = 0
	// calculate firstround block header hash for the check of the first-round block signatures
	conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(data_blks_received.NumberRound),sign_blks_preblk}}
	header_recal, _ := types.NewHeader(curheader.Version, common.NameToIndex("root").Number(), curheader.Height, curheader.PrevHash, curheader.MerkleHash,
		curheader.StateHash, conData, curheader.Bloom, curheader.Receipt.BlockCpu, curheader.Receipt.BlockNet,curheader.TimeStamp)
	blkFhash := header_recal.Hash
	// fmt.Println("blkFhash:",blkFhash)
	// fmt.Println("header_recal for first round signature:", current_round_num,header_recal.StateHash,header_recal.MerkleHash,header_recal.Hash)
	for index,sign_curblk := range sign_blks_curblk {
		// 3a. check the peers in the peer list
		var peerin_tag bool
		peerin_tag = false
		/*
		for _, peer := range Peers_list {
			if ok := bytes.Equal(peer.PublicKey, sign_curblk.PubKey); ok == true {
				peerin_tag = true
				break
			}
		}
		*/
		// change public key to account address
		for _, peer_addr := range Peers_addr_list {
			peer_addr_in := common.AddressFromPubKey(sign_curblk.PubKey)
			if ok := bytes.Equal(peer_addr.AccAdress.Bytes(), peer_addr_in.Bytes()); ok == true {
				peerin_tag = true
				break
			}
		}
		if peerin_tag == false {
			// there exists signature not from the peer list
			fmt.Println("the signature is not from the peer list, its index is:", index)
			return false,nil
		}
		// 3b. verify the correctness of the signature
		pubkey_in := sign_curblk.PubKey
		sigdata_in := sign_curblk.SigData
		var result_verify bool
		result_verify, err = secp256k1.Verify(blkFhash.Bytes(), sigdata_in, pubkey_in)
		if result_verify == true {
			num_verified++
		}
	}
	// 3c. check the valid signature number
	if num_verified < int(2*len(Peers_addr_list)/3){
		fmt.Println(" not enough signature for first round block:", num_verified)
		return false,nil
	}
	return  true,err

	// todo
	// use checkPermission(index common.AccountName, name string, sig []common.Signature) instead
	/*
	// 4. check the current block signature by using function checkPermission
	// 4a. check the peers permission
	err = actor_c.service_ababft.ledger.checkPermission(0, "active",sign_blks_curblk)
	if err != nil {
		log.Debug("signature permission check fail")
		return false,err
	}
	num_verified = 0
	// calculate firstround block header hash for the check of the first-round block signatures
	conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{uint32(current_round_num),sign_blks_preblk}}
	header_recal, _ := types.NewHeader(curheader.Version, curheader.Height, curheader.PrevHash, curheader.MerkleHash,
		curheader.StateHash, conData, curheader.Bloom, curheader.TimeStamp)
	blkFhash := header_recal.Hash
	for _,sign_curblk := range sign_blks_curblk {
		// 4b. verify the correctness of the signature
		pubkey_in := sign_curblk.PubKey
		sigdata_in := sign_curblk.SigData
		var result_verify bool
		result_verify, err = secp256k1.Verify(blkFhash.Bytes(), sigdata_in, pubkey_in)
		if result_verify == true {
			num_verified++
		}
	}
	// 4c. check the valid signature number
	if num_verified < int(2*len(Peers_list)/3+1){
		fmt.Println(" not enough signature for first round block:", num_verified)
		return false,nil
	}
	return  true,err
	*/
}

func (actor_c *Actor_ababft) Blk_syn_verify(block_in types.Block, blk_pre types.Block) (bool,error) {
	var err error
	// 1. check the protocal type is ababft
	if block_in.ConsensusData.Type != types.ConABFT {
		log.Debug("protocal error")
		return false,nil
	}
	// 2. check the block generator
	data_blks_received := block_in.ConsensusData.Payload.(*types.AbaBftData)
	round_num_in := int(data_blks_received.NumberRound)
	index_g := (int(data_blks_received.NumberRound)-1) % Num_peers + 1
	pukey_g_in := block_in.Signatures[0].PubKey

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
	for _, peer_addr := range Peers_addr_list {
		peer_addr_in := common.AddressFromPubKey(pukey_g_in)
		if ok := bytes.Equal(peer_addr.AccAdress.Bytes(), peer_addr_in.Bytes()); ok == true {
			index_g_in = int(peer_addr.Index)
			break
		}
	}
	// for test 2018.08.10
	if TestTag == true {
		fmt.Println("syn round_num_in:",round_num_in,index_g_in,pukey_g_in)
		fmt.Println("peer address list:",Peers_addr_list)
		fmt.Println("block_in.header:",block_in.Height,block_in.Hash,block_in.MerkleHash,block_in.StateHash)
	}
	// test end


	if index_g != index_g_in {
		log.Debug("illegal block generator")
		return false,nil
	}
	// 3. check the block header, except the consensus data
	var valid_blk bool
	valid_blk,err = actor_c.verify_header(&block_in, round_num_in, *blk_pre.Header)
	if valid_blk==false {
		println("header check fail")
		return valid_blk,err
	}

	// 4. check the signatures ( for both previous and current blocks) in ConsensusData
	valid_blk, err = actor_c.verify_signatures(data_blks_received, blk_pre.Header.Hash, block_in.Header)
	if valid_blk==false {
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
