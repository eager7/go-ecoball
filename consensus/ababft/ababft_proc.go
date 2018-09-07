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

func ProcessSTART(actorC *ActorABABFT) {
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
				if err != nil {
					log.Fatal(err)
				}
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
				if err != nil {
					log.Fatal(err)
				}
				t1.Stop()
			}
		}()
	}
	return
}

func ProcessSignPreBlk(actorC *ActorABABFT, msg SignaturePreBlock) {
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
			if err != nil {
				log.Fatal(err)
			}
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
			if err != nil {
				log.Fatal(err)
			}
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
					if err != nil {
						log.Fatal(err)
					}
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
	return
}

func ProcessBlkF(actorC *ActorABABFT, msg BlockFirstRound) {
	var err error
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
					requestSyn.Reqsyn.ChainID = actorC.chainID.Bytes()
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
				if err != nil {
					log.Fatal(err)
				}
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
					if err != nil {
						log.Fatal(err)
					}
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
					err = actorC.serviceABABFT.ledger.CheckTransaction(actorC.chainID, txIn)
					if err != nil {
						println("wrong tx, index:", index1)
						return
					}
				}
				// 4. sign the received first-round block
				var signBlkFSend SignatureBlkF
				signBlkFSend.signatureBlkF.ChainID = actorC.chainID.Bytes()
				signBlkFSend.signatureBlkF.PubKey = actorC.serviceABABFT.account.PublicKey
				signBlkFSend.signatureBlkF.SigData,err = actorC.serviceABABFT.account.Sign(blockFirstReceived.Header.Hash.Bytes())
				if err != nil {
					log.Fatal(err)
				}

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
				// 7. set the timer for waiting the second-round(final) block
				t3 := time.NewTimer(time.Second * waitResponseTime)
				go func() {
					select {
					case <-t3.C:
						// timeout for the second-round(final) block
						err = event.Send(event.ActorConsensus, event.ActorConsensus, BlockSTimeout{actorC.chainID})
						if err != nil {
							log.Fatal(err)
						}
						t3.Stop()
					}
				}()
			}
		}
	}
	return
}

func ProcessTxTimeout(actorC *ActorABABFT) {
	if actorC.primaryTag == 0 && (actorC.status == 2 || actorC.status == 5) {
		// not receive the first round block
		// change the status
		actorC.status = 8
		actorC.primaryTag = 0
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
		// todo
		// the above needed to be checked
		// here, enter the next term and broadcast the preblock signature with the increased round number has the same effect as the changeview/ nextround message
		// handle of the timeout message has been added, please check case TimeoutMsg
		return
	}
	return
}

func ProcessSignBlkF(actorC *ActorABABFT, msg SignatureBlkF) {
	var err error
	// the prime will verify the signatures of first-round block from peers
	if actorC.primaryTag == 1 && actorC.status == 4 {
		// verify the signature
		// 1. check the peer in the peers list
		pubKeyIn := msg.signatureBlkF.PubKey
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
		// 2. verify the correctness of the signature
		if actorC.signatureBlkFList[peerIndex].SigData != nil {
			// already receive the signature
			return
		}
		signDataIn := msg.signatureBlkF.SigData
		headerHashes := actorC.blockFirstRound.BlockFirst.Header.Hash.Bytes()
		var resultVerify bool
		resultVerify, err = secp256k1.Verify(headerHashes, signDataIn, pubKeyIn)
		if err != nil {
			log.Fatal(err)
		}
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
	return
}

func ProcessSignTxTimeout(actorC *ActorABABFT) {
	var err error
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
			if err != nil {
				log.Fatal(err)
			}
			blockSecond.SetSignature(actorC.serviceABABFT.account)
			// fmt.Println("blockSecond:",blockSecond.Header)

			// 3. broadcast the second-round(final) block
			actorC.blockSecondRound.BlockSecond = blockSecond
			// the ledger will multicast the block_secondround after the block is saved in the DB
			// event.Send(event.ActorConsensus, event.ActorP2P, block_secondround)

			// 4. save the second-round(final) block to ledger
			// currentheader = blockSecond.Header
			actorC.currentHeaderData = *(blockSecond.Header)
			actorC.currentHeader = &actorC.currentHeaderData
			actorC.verifiedHeight = blockSecond.Height - 1

			if err = event.Send(event.ActorNil, event.ActorLedger, &blockSecond); err != nil {
				log.Fatal(err)
				// return
			}
			if err = event.Send(event.ActorConsensus, event.ActorP2P, &blockSecond); err != nil {
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
			// 2. send out the timeout message
			var timeoutMsg TimeoutMsg
			timeoutMsg.Toutmsg.ChainID = actorC.chainID.Bytes()
			timeoutMsg.Toutmsg.RoundNumber = uint64(actorC.currentRoundNum)
			timeoutMsg.Toutmsg.PubKey = actorC.serviceABABFT.account.PublicKey
			hashTS,_ := common.DoubleHash(Uint64ToBytes(uint64(actorC.currentRoundNum)))
			timeoutMsg.Toutmsg.SigData,_ = actorC.serviceABABFT.account.Sign(hashTS.Bytes())
			event.Send(event.ActorConsensus,event.ActorP2P, timeoutMsg)
			// 3. start/enter the next turn
			event.Send(event.ActorConsensus, event.ActorConsensus, message.ABABFTStart{actorC.chainID})
		}
	}
}

func ProcessBlkS(actorC *ActorABABFT, msg BlockSecondRound) {
	var err error
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
					if err != nil {
						log.Fatal(err)
					}
					if validBlk ==false {
						println("header check fail")
						return
					}
					// save the solo block ( in the form of second-round block)
					// currentheader = blockSecondReceived.Header
					actorC.currentHeaderData = *(blockSecondReceived.Header)
					actorC.currentHeader = &actorC.currentHeaderData
					actorC.verifiedHeight = blockSecondReceived.Height
					actorC.currentHeightNum = int(actorC.verifiedHeight)
					if err = event.Send(event.ActorNil, event.ActorLedger, &blockSecondReceived); err != nil {
						log.Fatal(err)
						// return
					}
					if err = event.Send(event.ActorNil, event.ActorP2P, &blockSecondReceived); err != nil {
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
				reqSynSolo.Reqsyn.ChainID = actorC.chainID.Bytes()
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
			log.Info("received secondround block:", blockSecondReceived.Header.Height,actorC.verifiedHeight,actorC.currentHeightNum, dataBlkReceived.NumberRound, blockSecondReceived.Header)
			// 1. check the round number and height
			// 1a. current round number
			if dataBlkReceived.NumberRound < uint32(actorC.currentRoundNum) || blockSecondReceived.Header.Height <= uint64(actorC.currentHeightNum) {
				return
			} else if (blockSecondReceived.Header.Height-2) > actorC.verifiedHeight {
				// send synchronization message
				var requestSyn REQSyn
				requestSyn.Reqsyn.ChainID = actorC.chainID.Bytes()
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
				if err != nil {
					log.Fatal(err)
				}
				if validBlk ==false {
					println("header check fail")
					return
				}
				// 2. check the signatures ( for both previous and current blocks) in ConsensusData
				preBlkHash := actorC.currentHeader.Hash
				validBlk, err = actorC.verifySignatures(dataBlkReceived, preBlkHash, blockSecondReceived.Header)
				if err != nil {
					log.Fatal(err)
				}
				if validBlk ==false {
					println("previous and first-round blocks signatures check fail")
					return
				}

				// 3.save the second-round block into the ledger
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
	return
}

func ProcessBlkSTimeout(actorC *ActorABABFT) {
	if actorC.primaryTag == 0 && actorC.status == 5 {
		actorC.status = 8
		actorC.primaryTag = 0
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
		return
	}
	return
}

func ProcessREQSyn(actorC *ActorABABFT, msg REQSyn) {
	var err error
	// receive the synchronization request
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
	if err != nil {
		log.Fatal(err)
	}
	if signVerify != true {
		println("Syn request message signature is wrong")
		return
	}
	// fmt.Println("reqsyn:",current_height_num,heightReq)

	// 2. get the response blocks from the ledger
	blkSynV,err1 := actorC.serviceABABFT.ledger.GetTxBlockByHeight(actorC.chainID, heightReq)
	if err1 != nil || blkSynV == nil {
		log.Debug("not find the block of the corresponding height in the ledger")
		return
	}
	// fmt.Println("blkSynV:",blkSynV.Header)
	blkSynF,err2 := actorC.serviceABABFT.ledger.GetTxBlockByHeight(actorC.chainID, heightReq+1)
	if err2 != nil || blkSynF == nil {
		log.Debug("not find the block of the corresponding height in the ledger")
		return
	}

	// 3. send the found /blocks
	var blkSynSend BlockSyn
	blkSynSend.Blksyn.ChainID = actorC.chainID.Bytes()
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
	return
}

func ProcessREQSynSolo(actorC *ActorABABFT, msg REQSynSolo) {
	var err error
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
	if err != nil {
		log.Fatal(err)
	}
	if signVerify != true {
		println("Solo Syn request message signature is wrong")
		return
	}
	for i := int(heightReq); i <= actorC.currentHeightNum; i++ {
		// get the response blocks from the ledger
		blkSynSolo,err1 := actorC.serviceABABFT.ledger.GetTxBlockByHeight(actorC.chainID, uint64(i))
		if err1 != nil || blkSynSolo == nil {
			log.Debug("not find the solo block of the corresponding height in the ledger")
			return
		}
		// send the solo block
		event.Send(event.ActorConsensus,event.ActorP2P, blkSynSolo)
		log.Info("send the required solo block:", blkSynSolo.Height)
	}
	return
}

func ProcessBlkSyn(actorC *ActorABABFT, msg BlockSyn) {
	var err error
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

	heightSynV := blkV.Header.Height
	if heightSynV == (actorC.verifiedHeight+1) {
		// the current_height_num has been verified
		// 1. verify the verified block blkV

		// todo
		// maybe only check the hash is enough

		var resultV bool
		var blkVLocal *types.Block
		blkVLocal,err = actorC.serviceABABFT.ledger.GetTxBlockByHeight(actorC.chainID, actorC.verifiedHeight)
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
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("have not yet")
		}

		if resultV == false {
			log.Debug("verification of blkV fails")
			return
		}
		// 2. verify the verified block blkF
		var resultF bool
		resultF,err = actorC.blkSynVerify(blkF, blkV)
		if err != nil {
			log.Fatal(err)
		}
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
				err = actorC.serviceABABFT.ledger.ResetStateDB(actorC.chainID, blkVLocal.Header)
				if err != nil {
					log.Debug("reset state db error:", err)
					return
				}
			}
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
		requestSyn.Reqsyn.ChainID = actorC.chainID.Bytes()
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
}

func ProcessTimeoutMsg(actorC *ActorABABFT, msg TimeoutMsg) {
	var err error
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
	if err != nil {
		log.Fatal(err)
	}
	if signVerify != true {
		println("time out message signature is wrong")
		return
	}
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
					// fmt.Println("reset according to the timeout msgChan:",i,maxR,current_round_num,countRS[i])
					break
				}
			}
			break
		}
	}
	return
}