package ababft

import (
	"github.com/ecoball/go-ecoball/common/message"
	"fmt"
	"time"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/types"
	"sort"
	"github.com/ecoball/go-ecoball/common"
	"bytes"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
	"github.com/ecoball/go-ecoball/core/pb"
)

func ProcessSTARTABABFT(actorC *ActorABABFT) {
	var err error
	actorC.status = 2
	log.Debug("start ababft: receive the ababftstart message:", actorC.currentHeightNum,actorC.verifiedHeight,actorC.currentLedger.GetCurrentHeader(actorC.chainID))
	// check the status of the main net
	if ok:=actorC.currentLedger.StateDB(actorC.chainID).RequireVotingInfo(); ok!=true {
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
			txs, err1 := actorC.serviceABABFT.txPool.GetTxsList(actorC.chainID)
			if err1 != nil {
				log.Fatal(err1)
				// return
			}

			// generate the block in the form of second round block
			var blockSolo *types.Block
			tTime := time.Now().UnixNano()
			headerPayload:=&types.CMBlockHeader{}
			// headerPayload.LeaderPubKey = actorC.serviceABABFT.account.PublicKey
			blockSolo,err = actorC.serviceABABFT.ledger.NewTxBlock(actorC.chainID, txs, headerPayload, conData, tTime)
			if err != nil {
				log.Fatal(err)
			}
			blockSolo.SetSignature(&soloaccount)
			actorC.blockSecondRound.BlockSecond = *blockSolo
			// save (the ledger will broadcast the block after writing the block into the DB)
			if err = event.Send(event.ActorNil, event.ActorP2P, blockSolo); err != nil {
				log.Fatal(err)
				// return
			}
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
			requestsyn.Reqsyn.ChainID = actorC.chainID.Bytes()
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
	newPeers,err := actorC.currentLedger.GetProducerList(actorC.chainID)
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

		accountInfo,err := actorC.currentLedger.AccountGet(actorC.chainID, actorC.PeersListAccount[i].AccountName)
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
		currentBlock,err = actorC.currentLedger.GetTxBlock(actorC.chainID, actorC.currentHeader.PrevHash)
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
		// set up a timer to wait for the signaturePreblock from other peers
		t0 := time.NewTimer(time.Second * waitResponseTime * 2)
		go func() {
			select {
			case <-t0.C:
				// timeout for the preblock signature
				err = event.Send(event.ActorConsensus, event.ActorConsensus, PreBlockTimeout{ChainID:actorC.chainID})
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
		signaturePreSend.SignPreBlock.ChainID = actorC.chainID.Bytes()
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
				err = event.Send(event.ActorConsensus, event.ActorConsensus, TxTimeout{actorC.chainID})
				t1.Stop()
			}
		}()
	}
	return
}

func ProcessSignPreBlkABABFT(actorC *ActorABABFT, msg SignaturePreBlock) {
	var err error
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
				requestSyn.Reqsyn.ChainID = actorC.chainID.Bytes()
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
				if err != nil {
					log.Fatal(err)
					return
				}
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
}

func ProcessPreBlkTimeout(actorC *ActorABABFT) {
	var err error
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
			txs, _ := actorC.serviceABABFT.txPool.GetTxsList(actorC.chainID)
			// log.Debug("obtained tx list", txs[0])
			// generate the first-round block
			var blockFirst *types.Block
			tTime := time.Now().UnixNano()
			headerPayload:=&types.CMBlockHeader{}
			// headerPayload.LeaderPubKey = actorC.serviceABABFT.account.PublicKey
			blockFirst,err = actorC.serviceABABFT.ledger.NewTxBlock(actorC.chainID, txs, headerPayload, conData, tTime)
			blockFirst.SetSignature(actorC.serviceABABFT.account)
			// broadcast the first-round block to peers for them to verify the transactions and wait for the corresponding signatures back
			actorC.blockFirstRound.BlockFirst = *blockFirst
			event.Send(event.ActorConsensus, event.ActorP2P, actorC.blockFirstRound)
			// log.Debug("first round block:",block_firstround.BlockFirst)
			// fmt.Println("first round block status root hash:",blockFirst.StateHash)
			log.Info("generate the first round block and send",actorC.blockFirstRound.BlockFirst.Height)

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
					err = event.Send(event.ActorConsensus, event.ActorConsensus, SignTxTimeout{actorC.chainID})
					t2.Stop()
				}
			}()
		} else {
			// did not receive enough preblock signature in the assigned time interval
			actorC.status = 7
			actorC.primaryTag = 0 // reset to zero, and the next primary will take the turn
			// send out the timeout message
			var timeoutMsg TimeoutMsg
			timeoutMsg.Toutmsg.ChainID = actorC.chainID.Bytes()
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
}

func ProcessBlkF() {

}
