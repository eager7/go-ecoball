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
	"github.com/ecoball/go-ecoball/net/address"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"sync"
)

type ShardInfo struct {
	localID    uint32
	role       int
	shardMap   map[uint32]address.PeerMap
	lock       sync.RWMutex
}

func (s *ShardInfo) Initialize() *ShardInfo {
	s.shardMap = make(map[uint32]address.PeerMap)
	return s
}

func (s *ShardInfo) GetShardNodes(shardId uint32) *address.PeerMap {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if peerMap, ok := s.shardMap[shardId]; ok {
		return peerMap.Clone()
	}
	return nil
}

func (s *ShardInfo) AddShardNode(shardId uint32, peerId peer.ID, addr multiaddr.Multiaddr) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if peerMap, ok := s.shardMap[shardId]; ok {
		peerMap.Add(peerId, nil, []multiaddr.Multiaddr{addr}, "")
	} else {
		peerMap := new(address.PeerMap).Initialize()
		peerMap.Add(peerId, nil, []multiaddr.Multiaddr{addr}, "")
		s.shardMap[shardId] = peerMap
	}
}

func (s *ShardInfo) IsValidRemotePeer(p peer.ID) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, shard := range s.shardMap {
		if shard.Contains(p) {
			return true
		}
	}
	return false
}

func (s *ShardInfo) GetLocalId() uint32 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.localID
}

func (s *ShardInfo) SetLocalId(id uint32) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.localID = id
}

func (s *ShardInfo) SetNodeRole(role int) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.role = role
}

func (s *ShardInfo) GetNodeRole() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.role
}

func (s *ShardInfo) JsonString() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var info string
	for id, peerMap := range s.shardMap {
		info += fmt.Sprintf("\nshard id[%d], nodes:", id)
		for node := range peerMap.Iterator() {
			info += fmt.Sprintf("[%s-%s]", node.PeerInfo.ID.Pretty(), node.PeerInfo.Addrs)
		}
	}
	return fmt.Sprintf("local id:%d, the role is :%d, the info map is:%s", s.localID, s.role, info)
}
