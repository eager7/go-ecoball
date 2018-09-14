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

package sharding

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/sharding/cell"
	"github.com/ecoball/go-ecoball/sharding/committee"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/shard"
)

var (
	log = elog.NewLogger("sharding", elog.DebugLog)
)

type ShardingInstance interface {
	Start()
	MsgDispatch(msg interface{})
}

type Sharding struct {
	ns       *cell.Cell
	instance sc.NodeInstance
}

func MakeSharding() ShardingInstance {
	return &Sharding{ns: cell.MakeCell()}
}

func (s *Sharding) MsgDispatch(msg interface{}) {
	s.instance.MsgDispatch(msg)
}

func (s *Sharding) Start() {
	//read config
	s.ns.LoadConfig()

	if s.ns.NodeType == sc.NodeCommittee {
		s.instance = committee.MakeCommittee(s.ns)
	} else if s.ns.NodeType == sc.NodeShard {
		s.instance = shard.MakeShard(s.ns)
	} else {
		log.Error("unsupport node type")
		return
	}

	s.instance.Start()
}
