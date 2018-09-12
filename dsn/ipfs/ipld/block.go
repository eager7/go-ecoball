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

package ipld

import (
	"io"
	"fmt"
	"errors"
	"io/ioutil"

	"github.com/ecoball/go-ecoball/core/types"
	"gx/ipfs/QmVzK524a2VWLqyvtBeiHKsUAWYgeAk4DBeZoY7vpNPNRx/go-block-format"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	node "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
)

type EcoballBlock struct {
	*types.Header
	rawdata []byte
	cid     *cid.Cid
}

// assert that Block matches the Node interface for ipld
var _ node.Node = (*EcoballBlock)(nil)

/*
  Block INTERFACE
*/
func (this *EcoballBlock) RawData() []byte {
	if this.rawdata == nil {
		data, err := this.Header.Serialize()
		if err != nil {
			return nil
		}

		this.rawdata = data
	}

	return this.rawdata
}

func (this *EcoballBlock) Cid() *cid.Cid {
	if this.cid == nil {
		this.cid = rawdataToCid(cid.EcoballBlock, this.RawData())
	}
	return this.cid
}

func (this *EcoballBlock) String() string {
	return fmt.Sprintf("<EcoballBlock %s>", this.cid)
}

func (this *EcoballBlock) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"type": "ecoball-block",
	}
}

/*
  Node INTERFACE
*/
func (this *EcoballBlock) Tree(p string, depth int) []string {
	return []string{
		"version",
		"timeStamp",
		"height",
		"consensusData",
		"tx",
		"parent",
		"stateHash",
		"bloom",
		"signatures",
		"hash",
	}
}

func (this *EcoballBlock) Resolve(path []string) (interface{}, []string, error) {
	if len(path) == 0 {
		return nil, nil, fmt.Errorf("zero length path")
	}
	switch path[0] {
	case "version":
		return this.Version, path[1:], nil
	case "timeStamp":
		return this.TimeStamp, path[1:], nil
	case "height":
		return this.Height, path[1:], nil
	case "consensusData":
		return this.ConsensusData, path[1:], nil
	case "stateHash":
		return this.StateHash, path[1:], nil
	case "bloom":
		return this.Bloom, path[1:], nil
	case "signatures":
		return this.Signatures, path[1:], nil
	case "hash":
		return this.Hash, path[1:], nil
	case "parent":
		return &node.Link{Cid: commonHashToCid(cid.EcoballBlock, this.PrevHash)}, path[1:], nil
	case "tx":
		return &node.Link{Cid: commonHashToCid(cid.EcoballTree, this.MerkleHash)}, path[1:], nil
	default:
		return nil, nil, fmt.Errorf("no such link")
	}
}

func (this *EcoballBlock) ResolveLink(path []string) (*node.Link, []string, error) {
	out, rest, err := this.Resolve(path)
	if err != nil {
		return nil, nil, err
	}

	link, ok := out.(*node.Link)
	if !ok {
		return nil, nil, fmt.Errorf("object at path was not a link")
	}

	return link, rest, nil
}

func (this *EcoballBlock) Copy() node.Node {
	panic("dont use this yet")
}

func (this *EcoballBlock) Links() []*node.Link {
	return []*node.Link{
		{
			Name: "tx",
			Cid:  commonHashToCid(cid.EcoballTree, this.MerkleHash),
		},
		{
			Name: "parent",
			Cid:  commonHashToCid(cid.EcoballBlock, this.PrevHash),
		},
	}
}

func (this *EcoballBlock) Stat() (*node.NodeStat, error) {
	return &node.NodeStat{}, nil
}

func (this *EcoballBlock) Size() (uint64, error) {
	return uint64(len(this.rawdata)), nil
}


func DecodeBlock(block blocks.Block) (node.Node, error) {
	prefix := block.Cid().Prefix()

	if prefix.Codec != cid.EcoballBlock {
		return nil, errors.New("invalid CID prefix")
	}

	return ParseObjectFromBuffer(block.Cid(), block.RawData())
}

var _ node.DecodeBlockFunc = DecodeBlock

func ParseObjectFromBuffer(c *cid.Cid, b []byte) (*EcoballBlock, error) {
	block := new(types.Block)
	if err := block.Deserialize(b); err !=nil {
		return nil, nil
	}

	return &EcoballBlock{block.Header, b,c}, nil
}

func parseTransactions(txs []*types.Transaction) ([]node.Node, []*EcoballTree, error) {
	var txNodes []node.Node

	for _, tx := range txs {
		txNode := NewEcoballTx(tx)
		txNodes = append(txNodes, txNode)
	}

	txTrees, err := mkMerkleTree(txNodes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to mk tx merkle tree: %s", err)
	}

	return txNodes, txTrees, nil
}

func EcoballBlockRawInputParser(r io.Reader, mhType uint64, mhLen int) ([]node.Node, error) {
	rawdata, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	block := new(types.Block)
	if err := block.Deserialize(rawdata); err != nil {
		return nil, err
	}

	ecoBlock := &EcoballBlock{block.Header, rawdata, rawdataToCid(cid.EcoballBlock, rawdata)}
	ecoTxs, txTrees, err := parseTransactions(block.Transactions)
	if err != nil {
		return nil, err
	}

	out := []node.Node{ecoBlock}
	out = append(out, ecoTxs...)

	for _, txTree := range txTrees {
		out = append(out, txTree)
	}

	return out, nil
}