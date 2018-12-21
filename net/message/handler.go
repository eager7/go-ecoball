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
	eactor "github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
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
	return nil
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
	signPreReceive := pb.SignaturePreBlockA{}
	err := signPreReceive.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch signPre msg")
	eactor.Send(0, eactor.ActorConsensus, signPreReceive)
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
	reqSyn := pb.REQSynA{}
	err := reqSyn.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch synchronization request msg")
	eactor.Send(0, eactor.ActorConsensus, reqSyn)
	return nil
}

func HdReqSynSoloMsg(data []byte) error {
	reqSyn := pb.REQSynSolo{}
	err := reqSyn.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch synchronization request msg")
	eactor.Send(0, eactor.ActorConsensus, reqSyn)
	return nil
}

func HdToutMsg(data []byte) error {
	tOutMsg := pb.TimeoutMsg{}
	err := tOutMsg.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch synchronization request msg")
	eactor.Send(0, eactor.ActorConsensus, tOutMsg)
	return nil
}

func HdSignBlkFMsg(data []byte) error {
	signBlkfReceive := pb.SignatureBlkFA{}
	err := signBlkfReceive.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch the signature of first-round block msg")
	eactor.Send(0, eactor.ActorConsensus, signBlkfReceive)
	return nil
}

func HdBlkSMsg(data []byte) error {
	blockSecondRound := pb.BlockSecondRound{}
	err := blockSecondRound.BlockSecond.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch second-round(final) block msg")
	eactor.Send(0, eactor.ActorConsensus, blockSecondRound)
	return nil
}

func HdBlkSynMsg(data []byte) error {
	blkSyn := pb.BlockSynA{}
	err := blkSyn.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch the block according to the synchronization request")
	eactor.Send(0, eactor.ActorConsensus, blkSyn)
	return nil
}

// MakeHandlers generates a map of MsgTypes to their corresponding handler functions
func MakeHandlers() map[pb.MsgType]HandlerFunc {
	return map[pb.MsgType]HandlerFunc{
		pb.MsgType_APP_MSG_TRN:        HdTransactionMsg,
		pb.MsgType_APP_MSG_BLK:        HdBlkMsg,
		pb.MsgType_APP_MSG_SIGNPRE:    HdSignPreMsg,
		pb.MsgType_APP_MSG_BLKF:       HdBlkFMsg,
		pb.MsgType_APP_MSG_REQSYN:     HdReqSynMsg,
		pb.MsgType_APP_MSG_REQSYNSOLO: HdReqSynSoloMsg,
		pb.MsgType_APP_MSG_SIGNBLKF:   HdSignBlkFMsg,
		pb.MsgType_APP_MSG_BLKS:       HdBlkSMsg,
		pb.MsgType_APP_MSG_BLKSYN:     HdBlkSynMsg,
		pb.MsgType_APP_MSG_TIMEOUT:    HdToutMsg,
		//TODO add new msg handler at here
	}
}
