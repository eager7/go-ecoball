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

package address

import (
	"fmt"
	"sync"
)

type ShardInfo struct {
	localId  uint32
	pubkey   string
	role     int
	shardMap map[uint32]PeerMap
	lock     sync.RWMutex
}

func (s *ShardInfo) Initialize() *ShardInfo {
	s.shardMap = make(map[uint32]PeerMap)
	return s
}

func (s *ShardInfo) GetShardNodes(shardId uint32) *PeerMap {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if peerMap, ok := s.shardMap[shardId]; ok {
		return peerMap.Clone()
	}
	return nil
}

func (s *ShardInfo) AddShardNode(shardId uint32, b64Pub, addr, port string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if peerMap, ok := s.shardMap[shardId]; ok {
		peerMap.Add(b64Pub, addr, port)
	} else {
		peerMap := new(PeerMap).Initialize()
		peerMap.Add(b64Pub, addr, port)
		s.shardMap[shardId] = peerMap
	}
}

func (s *ShardInfo) IsValidRemotePeer(b64Pub string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, shard := range s.shardMap {
		if shard.Contains(b64Pub) {
			return true
		}
	}
	return false
}

func (s *ShardInfo) SetLocalPub(b64Pub string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pubkey = b64Pub
}

func (s *ShardInfo) GetLocalPub() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.pubkey
}

func (s *ShardInfo) GetLocalId() uint32 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.localId
}

func (s *ShardInfo) SetLocalId(id uint32) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.localId = id
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
			info += fmt.Sprintf("[%s-%s]", node.Address, node.Port)
		}
	}
	return fmt.Sprintf("local id:%d, the role is :%d, the info map is:%s", s.localId, s.role, info)
}

func (s *ShardInfo) Purge() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for k := range s.shardMap {
		delete(s.shardMap, k)
	}
}
