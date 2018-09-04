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

package block

import (
	"github.com/ecoball/go-ecoball/common"
	"math/big"
)

type Shard struct {
	Id          uint32
	Member      []NodeInfo
	MemeberAddr []NodeAddr
}

type CMBlockHeader struct {
	PrevHash       common.Hash
	Height         uint64
	LeaderPubKey   []byte
	CadidatePubKey []byte
	Deposite       big.Int
	ShardsHash     common.Hash /*shards hash, not include node address*/
	Timestamp      uint64
	*COSign
}

type CMBlock struct {
	*CMBlockHeader
	Shards []Shard
}
