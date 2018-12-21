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
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message/pb"
)

type HandlerFunc func(data []byte) (err error)

func HdTransactionMsg(data []byte) error {
	tx := new(types.Transaction)
	err := tx.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch tx msg")
	event.Send(event.ActorNil, event.ActorTxPool, tx)
	return nil
}

func HdBlkMsg(data []byte) error {
	blk := new(types.Block)
	err := blk.Deserialize(data)
	if err != nil {
		return err
	}
	log.Debug("dispatch blk msg")
	event.Send(event.ActorNil, event.ActorLedger, blk)
	return nil
}

// MakeHandlers generates a map of MsgTypes to their corresponding handler functions
func MakeHandlers() map[pb.MsgType]HandlerFunc {
	return map[pb.MsgType]HandlerFunc{
		pb.MsgType_APP_MSG_TRN: HdTransactionMsg,
		pb.MsgType_APP_MSG_BLK: HdBlkMsg,
		//TODO add new msg handler at here
	}
}
