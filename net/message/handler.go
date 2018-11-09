// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball.
//
// The go-ecoball is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.

package message

import (
	"github.com/ecoball/go-ecoball/core/types"
	eactor "github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/net/message/pb"
)

func HdTransactionMsg(data []byte) error {
	tx := new(types.Transaction)
	err := tx.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch tx msg")
	eactor.Send(0, eactor.ActorTxPool, tx)
	return  nil
}

func HdBlkMsg(data []byte) error {
	blk := new(types.Block)
	err := blk.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch blk msg")
	eactor.Send(0, eactor.ActorLedger, blk)
	return nil
}

func HdSignPreMsg(data []byte) error {
	signpre_receive := pb.SignaturePreBlockA{}
	err := signpre_receive.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch signpre msg")
	eactor.Send(0, eactor.ActorConsensus, signpre_receive)
	return nil
}

func HdBlkFMsg(data []byte) error {
	blockFirstRound := pb.BlockFirstRound{}
	err := blockFirstRound.BlockFirst.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch first round block msg")
	eactor.Send(0, eactor.ActorConsensus, blockFirstRound)
	return nil
}

func HdReqSynMsg(data []byte) error {
	reqsyn := pb.REQSynA{}
	err := reqsyn.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch synchronization request msg")
	eactor.Send(0, eactor.ActorConsensus, reqsyn)
	return nil
}

func HdReqSynSoloMsg(data []byte) error {
	reqsyn := pb.REQSynSolo{}
	err := reqsyn.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch synchronization request msg")
	eactor.Send(0, eactor.ActorConsensus, reqsyn)
	return nil
}

func HdToutMsg(data []byte) error {
	toutmsg := pb.TimeoutMsg{}
	err := toutmsg.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch synchronization request msg")
	eactor.Send(0, eactor.ActorConsensus, toutmsg)
	return nil
}

func HdSignBlkFMsg(data []byte) error {
	signblkf_receive := pb.SignatureBlkFA{}
	err := signblkf_receive.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch the signature of first-round block msg")
	eactor.Send(0, eactor.ActorConsensus, signblkf_receive)
	return nil
}

func HdBlkSMsg(data []byte) error {
	block_secondround := pb.BlockSecondRound{}
	err := block_secondround.BlockSecond.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch second-round(final) block msg")
	eactor.Send(0, eactor.ActorConsensus, block_secondround)
	return nil
}

func HdBlkSynMsg(data []byte) error {
	blksyn := pb.BlockSynA{}
	err := blksyn.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch the block according to the synchronization request")
	eactor.Send(0, eactor.ActorConsensus, blksyn)
	return nil
}

// MakeHandlers generates a map of MsgTypes to their corresponding handler functions
func MakeHandlers() map[pb.MsgType]HandlerFunc {
	return map[pb.MsgType]HandlerFunc{
		pb.MsgType_APP_MSG_TRN:       HdTransactionMsg,
		pb.MsgType_APP_MSG_BLK:       HdBlkMsg,
		pb.MsgType_APP_MSG_SIGNPRE:   HdSignPreMsg,
		pb.MsgType_APP_MSG_BLKF:      HdBlkFMsg,
		pb.MsgType_APP_MSG_REQSYN:    HdReqSynMsg,
		pb.MsgType_APP_MSG_REQSYNSOLO:HdReqSynSoloMsg,
		pb.MsgType_APP_MSG_SIGNBLKF:  HdSignBlkFMsg,
		pb.MsgType_APP_MSG_BLKS:      HdBlkSMsg,
		pb.MsgType_APP_MSG_BLKSYN:    HdBlkSynMsg,
		pb.MsgType_APP_MSG_TIMEOUT:   HdToutMsg,
		//TODO add new msg handler at here
	}
}