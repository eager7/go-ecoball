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

package ipldecoball

import (
	"bytes"
	"fmt"
	"errors"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	node "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"

	"gx/ipfs/QmVzK524a2VWLqyvtBeiHKsUAWYgeAk4DBeZoY7vpNPNRx/go-block-format"
)

type ShardInfo struct {
	Data         []byte  `json:"Data"`
}
type EcoballShard struct {
	*ShardInfo
	cid          *cid.Cid
}

// assert that EcoballShard matches the Node interface for ipld
var _ node.Node = (*EcoballShard)(nil)

func NewShard(data []byte) (*EcoballShard){
	info := &ShardInfo{data}
	return &EcoballShard{ShardInfo:info}
}

func DecodeShardData(block blocks.Block) (node.Node, error) {
	prefix := block.Cid().Prefix()

	if prefix.Codec != cid.EcoballShardData {
		return nil, errors.New("invalid CID prefix")
	}

	shard := NewShard(nil)

	data := block.RawData()

	buf := new(bytes.Buffer)
	buf.Write(data[:])
	shard.Data = buf.Bytes()

	return shard, nil
}

var _ node.DecodeBlockFunc = DecodeShardData

/*
  Block INTERFACE
*/
func (this *EcoballShard) RawData() []byte {
	return this.Data
}

func (this *EcoballShard) Cid() *cid.Cid {
	if this.cid == nil {
		this.cid = rawdataToCid(cid.EcoballShardData, this.Data)
	}

	return this.cid
}

func (this *EcoballShard) String() string {
	return fmt.Sprintf("<EcoballShardData %s>", this.Cid())
}

func (this *EcoballShard) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"type": "ecoball-sharddata",
	}
}

/*
  Node INTERFACE
*/
func (this *EcoballShard) Tree(p string, depth int) []string {
	return []string{
		"Data",
	}
}

func (this *EcoballShard) Resolve(path []string) (interface{}, []string, error) {
	if len(path) == 0 {
		return nil, nil, fmt.Errorf("zero length path")
	}
	switch path[0] {
	case "Data":
		return this.Data, path[1:], nil
	default:
		return nil, nil, fmt.Errorf("no such link")
	}
}

func (this *EcoballShard) ResolveLink(path []string) (*node.Link, []string, error) {
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

func (this *EcoballShard) Copy() node.Node {
	panic("dont use this yet")
}

func (this *EcoballShard) Links() []*node.Link {
	return nil
}

func (this *EcoballShard) Stat() (*node.NodeStat, error) {
	return &node.NodeStat{}, nil
}

func (this *EcoballShard) Size() (uint64, error) {
	return uint64(len(this.RawData())), nil
}
