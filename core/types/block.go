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

package types

import (
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/core/trie"
)

type Block struct {
	*Header
	CountTxs     uint32
	Transactions []*Transaction
}

func NewBlock(chainID common.Hash, prev *Header, stateHash common.Hash, consensusData ConsData, txs []*Transaction, cpu, net float64, timeStamp int64) (*Block, error) {
	if nil == prev {
		return nil, errors.New("invalid parameter preHeader")
	}
	var Bloom bloom.Bloom
	var hashes []common.Hash
	for _, t := range txs {
		hashes = append(hashes, t.Hash)
		Bloom.Add(t.Hash.Bytes())
		Bloom.Add(common.IndexToBytes(t.From))
		Bloom.Add(common.IndexToBytes(t.Addr))
	}
	merkleHash, err := trie.GetMerkleRoot(hashes)
	if err != nil {
		return nil, err
	}

	var cpuLimit, netLimit float64
	if cpu < (config.BlockCpuLimit / 10) {
		cpuLimit = prev.Receipt.BlockCpu * 1.01
		if cpuLimit > config.VirtualBlockCpuLimit {
			cpuLimit = config.VirtualBlockCpuLimit
		}
	} else {
		cpuLimit = prev.Receipt.BlockCpu * 0.99
		if cpuLimit < config.BlockCpuLimit {
			cpuLimit = config.BlockCpuLimit
		}
	}
	if net < (config.BlockNetLimit / 10) {
		netLimit = prev.Receipt.BlockNet * 1.01
		if netLimit > config.VirtualBlockNetLimit {
			netLimit = config.VirtualBlockNetLimit
		}
	} else {
		netLimit = prev.Receipt.BlockNet * 0.99
		if netLimit < config.BlockNetLimit {
			netLimit = config.BlockNetLimit
		}
	}

	header := &Header{
		Version:    VersionHeader,
		ChainID:    chainID,
		TimeStamp:  timeStamp,
		Height:     prev.Height + 1,
		ConsData:   consensusData,
		PrevHash:   prev.Hash,
		MerkleHash: merkleHash,
		StateHash:  stateHash,
		Receipt: BlockReceipt{
			BlockCpu: cpuLimit,
			BlockNet: netLimit,
		},
		Signatures: nil,
		Hash:       common.Hash{},
	}
	if err := header.ComputeHash(); err != nil {
		return nil, err
	}

	block := Block{
		Header:       header,
		CountTxs:     uint32(len(txs)),
		Transactions: txs,
	}
	return &block, nil
}

func (b *Block) SetSignature(account *account.Account) error {
	return b.Header.SetSignature(account)
}

func (b *Block) GetTransaction(hash common.Hash) *Transaction {
	for _, tx := range b.Transactions {
		if hash.Equals(&tx.Hash) {
			return tx
		}
	}
	return nil
}

/**
 *  @brief converts a structure into a sequence of characters
 *  @return []byte - a sequence of characters
 */
func (b *Block) Serialize() (data []byte, err error) {
	p, err := b.Proto()
	if err != nil {
		return nil, err
	}
	data, err = p.Marshal()
	if err != nil {
		return nil, err
	}
	return data, nil
}

/**
 *  @brief converts a sequence of characters into a structure
 *  @param data - a sequence of characters
 */
func (b *Block) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var pbBlock pb.Block
	if err := pbBlock.Unmarshal(data); err != nil {
		return err
	}
	dataHeader, err := pbBlock.Header.Marshal()
	if err != nil {
		return err
	}

	b.Header = new(Header)
	err = b.Header.Deserialize(dataHeader)
	if err != nil {
		return err
	}

	var txs []*Transaction
	for _, tx := range pbBlock.Transactions {
		b, err := tx.Marshal()
		if err != nil {
			return err
		}
		t := new(Transaction)
		if err := t.Deserialize(b); err != nil {
			return err
		}
		txs = append(txs, t)
	}

	b.CountTxs = uint32(len(txs))
	b.Transactions = txs

	return nil
}

func (b *Block) String() string {
	data := b.Header.String()
	data += fmt.Sprintf("{CountTxs:%d}", b.CountTxs)
	for _, v := range b.Transactions {
		data += v.String()
	}
	return string(data)

}

func (b *Block) Identify() mpb.Identify {
	return mpb.Identify_APP_MSG_BLOCK
}

func (b *Block) GetInstance() interface{} {
	return b
}

func (b *Block) Proto() (*pb.Block, error) {
	var block pb.Block
	var err error
	block.Header, err = b.Header.proto()
	if err != nil {
		return nil, err
	}
	var pbTxs []*pb.Transaction
	for _, tx := range b.Transactions {
		pbTx, err := tx.ProtoBuf()
		if err != nil {
			return nil, err
		}
		pbTxs = append(pbTxs, pbTx)
	}
	block.Transactions = append(block.Transactions, pbTxs...)
	return &block, nil
}

func (b *Block) Type() uint32 {
	return 0
}
