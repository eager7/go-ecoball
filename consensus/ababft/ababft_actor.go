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
	"github.com/ecoball/go-ecoball/core/pb"
	"bytes"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
	"fmt"
	netMsg "github.com/ecoball/go-ecoball/common/message"
	"encoding/binary"
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

	// multiple chain
	chainID common.Hash
	msgChan chan actor.Context // use channel, combined with actor
	msgStop chan struct{}
}

const(
	pubKeyTag   = "ababft"
	signDataTag = "ababft"
)

var log = elog.NewLogger("ABABFT", elog.NoticeLog)

const ThresholdRound = 60


func ActorABABFTGen(chainId common.Hash, actorABABFT *ActorABABFT) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor {
		return actorABABFT
	})
	/*
	chainStr := string("ABABFT")
	chainStr += chainId.HexString()
	pid, err := actor.SpawnNamed(props, chainStr)
	*/
	pid, err := actor.SpawnNamed(props, "ABABFT")
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
	// deal with the message
	switch msg := ctx.Message().(type) {
	case netMsg.ABABFTStart:
		// check the chain ID
		chainIn := msg.ChainID
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this ABABFTStart to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok:=bytes.Equal(msg.ChainID.Bytes(),actorC.chainID.Bytes()); ok != true {
			log.Debug("the message is not for this chain")
			return
		}
		ProcessSTART(actorC)
		*/
		return

	case SignaturePreBlock:
		log.Info("receive the preblock signature:", actorC.status,msg.SignPreBlock)
		// check the chain ID
		chainIn := common.BytesToHash(msg.SignPreBlock.ChainID)
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this SignaturePreBlock to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok := bytes.Equal(actorC.chainID.Bytes(),msg.SignPreBlock.ChainID); ok != true {
			log.Debug("wrong chain ID for preblock signature")
			return
		}
		ProcessSignPreBlk(actorC,msg)
		*/
		return

	case PreBlockTimeout:
		// check the chain ID
		chainIn := msg.ChainID
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this PreBlockTimeout to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok := bytes.Equal(msg.ChainID.Bytes(),actorC.chainID.Bytes()); ok != true {
			log.Debug("wrong chain ID for preblock timeout")
			return
		}
		ProcessPreBlkTimeout(actorC)
		*/
		return

	case BlockFirstRound:
		// check the chain ID
		chainIn := msg.BlockFirst.Header.ChainID
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this BlockFirstRound to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok := bytes.Equal(actorC.chainID.Bytes(),msg.BlockFirst.Header.ChainID.Bytes()); ok != true {
			log.Debug("wrong chain ID for first round block")
			return
		}
		ProcessBlkF(actorC,msg)
		*/
		return

	case TxTimeout:
		// check the chain ID
		chainIn := msg.ChainID
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this TxTimeout to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok := bytes.Equal(msg.ChainID.Bytes(),actorC.chainID.Bytes()); ok != true {
			log.Debug("wrong chain ID for transaction timeout")
			return
		}
		ProcessTxTimeout(actorC)
		*/
		return

	case SignatureBlkF:
		// check the chain ID
		chainIn := common.BytesToHash(msg.signatureBlkF.ChainID)
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this SignatureBlkF to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok := bytes.Equal(msg.signatureBlkF.ChainID,actorC.chainID.Bytes()); ok != true {
			log.Debug("wrong chain ID for signature of the first-round block")
			return
		}
		ProcessSignBlkF(actorC,msg)
		*/
		return

	case SignTxTimeout:
		// check the chain ID
		chainIn := msg.ChainID
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this SignTxTimeout to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok := bytes.Equal(actorC.chainID.Bytes(),msg.ChainID.Bytes()); ok != true {
			log.Debug("wrong chain ID for SignTxTimeout")
			return
		}
		ProcessSignTxTimeout(actorC)
		*/
		return

	case BlockSecondRound:
		// check the chain ID
		chainIn := msg.BlockSecond.Header.ChainID
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this BlockSecondRound to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok := bytes.Equal(actorC.chainID.Bytes(),msg.BlockSecond.Header.ChainID.Bytes()); ok != true {
			log.Debug("wrong chain ID for BlockSecondRound")
			return
		}
		ProcessBlkS(actorC,msg)
		*/
		return

	case BlockSTimeout:
		// check the chain ID
		chainIn := msg.ChainID
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this BlockSTimeout to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok := bytes.Equal(actorC.chainID.Bytes(),msg.ChainID.Bytes()); ok != true {
			log.Debug("wrong chain ID for BlockSTimeout")
			return
		}
		ProcessBlkSTimeout(actorC)
		*/
		return

	case REQSyn:
		// check the chain ID
		chainIn := common.BytesToHash(msg.Reqsyn.ChainID)
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this REQSyn to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok := bytes.Equal(msg.Reqsyn.ChainID,actorC.chainID.Bytes()); ok != true {
			log.Debug("wrong chain ID for REQSyn")
			return
		}
		ProcessREQSyn(actorC,msg)
		*/
		return

	case REQSynSolo:
		// check the chain ID
		chainIn := common.BytesToHash(msg.Reqsyn.ChainID)
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this REQSynSolo to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok := bytes.Equal(msg.Reqsyn.ChainID,actorC.chainID.Bytes()); ok != true {
			log.Debug("wrong chain ID for REQSynSolo")
			return
		}
		ProcessREQSynSolo(actorC,msg)
		*/
		return

	case BlockSyn:
		// check the chain ID
		chainIn := common.BytesToHash(msg.Blksyn.ChainID)
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this BlockSyn to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok := bytes.Equal(actorC.chainID.Bytes(),msg.Blksyn.ChainID); ok != true {
			log.Debug("wrong chain ID for BlockSyn")
			return
		}
		ProcessBlkSyn(actorC,msg)
		*/
		return

	case TimeoutMsg:
		// check the chain ID
		chainIn := common.BytesToHash(msg.Toutmsg.ChainID)
		if msgChan, ok := actorC.serviceABABFT.mapMsgChan[chainIn]; ok {
			log.Info("the chain found, send this TimeoutMsg to the corresponding channel")
			msgChan <- ctx
		} else {
			log.Info("Not found the corresponding chain")
		}
		/*
		if ok := bytes.Equal(actorC.chainID.Bytes(),msg.Toutmsg.ChainID); ok != true {
			log.Debug("wrong chain ID for TimeoutMsg")
			return
		}
		ProcessTimeoutMsg(actorC,msg)
		*/
		return

	case *netMsg.RegChain:
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
	headerPayload:=&types.CMBlockHeader{}
	actorC.blockFirstCal,err = actorC.serviceABABFT.ledger.NewTxBlock(actorC.chainID, txs, headerPayload, conDataC, headerIn.TimeStamp)
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
	headerPayload1:=&types.CMBlockHeader{}
	headerCal,err1 := types.NewHeader(headerPayload1, headerIn.Version, actorC.chainID, headerIn.Height, headerIn.PrevHash,
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
	headerPayload:=&types.CMBlockHeader{}
	header, _ := types.NewHeader(headerPayload, headerIn.Version, actorC.chainID, headerIn.Height, headerIn.PrevHash, headerIn.MerkleHash,
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
	headerPayload:=&types.CMBlockHeader{}
	conData := types.ConsensusData{Type: types.ConABFT, Payload: &types.AbaBftData{NumberRound:uint32(dataBlksReceived.NumberRound), PreBlockSignatures:signBlksPreBlk},}
	headerReCal, _ := types.NewHeader(headerPayload, curHeader.Version, actorC.chainID, curHeader.Height, curHeader.PrevHash, curHeader.MerkleHash,
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

	// todo
	// check the chain ID

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
