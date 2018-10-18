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
	"github.com/ecoball/go-ecoball/net/message"
)

func (net *NetImpl)SendBlockToShards(blkmsg message.EcoBallNetMsg) {
	if !net.receiver.IsLeaderOrBackup() {
		log.Error("I am not a committee leader or backup")
		return
	}

	shardMembers := net.receiver.GetShardMemebersToReceiveCBlock()
	for _, shard := range shardMembers {
		net.SendMsgToPeersWithId(shard, blkmsg)
	}
}

func (net *NetImpl)SendBlockToCommittee(blkmsg message.EcoBallNetMsg) {
	if !net.receiver.IsLeaderOrBackup() {
		log.Error("I am not a shard leader or backup")
		return
	}

	cmMembers := net.receiver.GetCMMemebersToReceiveSBlock()
	net.SendMsgToPeersWithId(cmMembers, blkmsg)
}