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
	"github.com/ecoball/go-ecoball/common/errors"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"sync"
)

type ShardInfo struct {
	ShardSubCh <-chan interface{}
	LocalID    uint32
	Role       int
	ShardMap   map[uint32]map[peer.ID]multiaddr.Multiaddr // to accelerate the finding speed
	lock       sync.RWMutex
}

func (s *ShardInfo) Initialize() {
	s.ShardSubCh = make(<-chan interface{}, 1)
	s.ShardMap = make(map[uint32]map[peer.ID]multiaddr.Multiaddr)
}

func (s *ShardInfo) GetShardNodes(shardId uint32) (map[peer.ID]multiaddr.Multiaddr, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if works, ok := s.ShardMap[shardId]; ok {
		return works, nil
	} else {
		return nil, errors.New(fmt.Sprintf("cat't find this shard:%d", shardId))
	}
}

func (s *ShardInfo) GetNodeAddress(shardId uint32, peerId peer.ID) multiaddr.Multiaddr {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if works, ok := s.ShardMap[shardId]; ok {
		for pid, addr := range works {
			if pid == peerId {
				return addr
			}
		}
	}
	return nil
}

func (s *ShardInfo) AddShardNode(shardId uint32, peerId peer.ID, addr multiaddr.Multiaddr) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if nodeMap, ok := s.ShardMap[shardId]; ok {
		nodeMap[peerId] = addr
	} else {
		idAddr := make(map[peer.ID]multiaddr.Multiaddr)
		idAddr[peerId] = addr
		s.ShardMap[shardId] = idAddr
	}
}

func (s *ShardInfo) IsValidRemotePeer(p peer.ID) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, shard := range s.ShardMap {
		if _, ok := shard[p]; ok {
			return true
		}
	}
	return false
}

func (s *ShardInfo) GetLocalId() uint32 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.LocalID
}

func (s *ShardInfo) SetLocalId(id uint32) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.LocalID = id
}

func (s *ShardInfo) SetNodeRole(role int) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Role = role
}

func (s *ShardInfo) GetNodeRole() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.Role
}

func (s *ShardInfo) JsonString() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var info string
	for id, works := range s.ShardMap {
		info += fmt.Sprintf("\nshard id[%d], nodes:", id)
		for id, addr := range works {
			info += fmt.Sprintf("[%s-%s]", id.Pretty(), addr.String())
		}
	}
	return fmt.Sprintf("local id:%d, the role is :%d, the info map is:%s", s.LocalID, s.Role, info)
}
