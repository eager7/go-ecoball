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

package message

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/types"
)

type ABABFTStart struct {
	ChainID common.Hash
}
type SoloStop struct{}
type GetCurrentHeader struct{}

type RegChain struct {
	ChainID common.Hash
	Address common.Address
	TxHash  common.Hash
}

type BlockMessage struct {
	ShardID uint32
	Block   shard.BlockInterface
}

type ProducerBlock struct {
	ChainID common.Hash
	Height  uint64
	Type    mpb.Identify
	Hashes  []common.Hash
}

type CheckBlock struct {
	Block  shard.BlockInterface
	Result error
}

type DeleteTx struct {
	ChainID common.Hash
	Hash    common.Hash
}

type Transaction struct {
	ShardID uint32
	Tx      *types.Transaction
}

type NetPacket struct {
	Address   string
	Port      string
	PublicKey string
	Message   types.EcoMessage
}
