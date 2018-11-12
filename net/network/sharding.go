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

// Implement the message API between committee and shard nodes

package network

import (
	"fmt"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
)

func (net *NetImpl)SendMsgDataToShard(shardId uint16, msgId pb.MsgType, data []byte) error {
	p, err := net.receiver.GetShardLeader(shardId)
	if err != nil {
		return err
	}
	msg := message.New(msgId, data)
	net.SendMsgToPeerWithPeerInfo([]*peerstore.PeerInfo{p}, msg)

	return nil
}

func (net *NetImpl)SendMsgToShards(msg message.EcoBallNetMsg) error {
	if !net.receiver.IsLeaderOrBackup() {
		return fmt.Errorf("sender is not a committee leader or backup")
	}

	shardMembers := net.receiver.GetShardMemebersToReceiveCBlock()
	for _, shard := range shardMembers {
		net.SendMsgToPeerWithPeerInfo(shard, msg)
	}

	return nil
}

func (net *NetImpl)SendMsgToCommittee(msg message.EcoBallNetMsg) error {
	if !net.receiver.IsLeaderOrBackup() {
		return fmt.Errorf("sender is not a committee leader or backup")
	}

	cmMembers := net.receiver.GetCMMemebersToReceiveSBlock()
	net.SendMsgToPeerWithPeerInfo(cmMembers, msg)

	return nil
}