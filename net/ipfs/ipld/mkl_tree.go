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
	"fmt"

	mh "github.com/multiformats/go-multihash"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	node "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
)

type EcoballTree struct {
	left  *node.Link
	right *node.Link

	rawdata []byte
	cid     *cid.Cid
}

// assert that Block matches the Node interface for ipld
var _ node.Node = (*EcoballTree)(nil)

func cidToHash(c *cid.Cid) []byte {
	h := []byte(c.Hash())
	return h[len(h)-32:]
}

func hashToLink(b []byte) *node.Link {
	mhb, _ := mh.Encode(b, mh.KECCAK_256)
	c := cid.NewCidV1(cid.EcoballTree, mhb)
	return &node.Link{Cid: c}
}

func mkMerkleTree(n []node.Node) ([]*EcoballTree, error) {
	var out []*EcoballTree
	var next []node.Node
	layer := n
	for len(layer) > 1 {
		if len(layer)%2 != 0 {
			layer = append(layer, layer[len(layer)-1])
		}
		for i := 0; i < len(layer)/2; i++ {
			var left, right node.Node
			left = layer[i*2]
			right = layer[(i*2)+1]
			t := &EcoballTree{
				left:    &node.Link{Cid: left.Cid()},
				right:   &node.Link{Cid: right.Cid()},
			}

			out = append(out, t)
			next = append(next, t)
		}

		layer = next
		next = nil
	}

	return out, nil
}

func DecodeHardTree(data []byte) (*EcoballTree, error) {
	if len(data) != 64 {
		return nil, fmt.Errorf("invalid shardtree data")
	}

	linkL := hashToLink(data[:32])
	linkR := hashToLink(data[32:])

	return &EcoballTree{
		left:  linkL,
		right: linkR,
	}, nil
}

/*
  Block INTERFACE
*/
func (this *EcoballTree) RawData() []byte {
	if this.rawdata == nil {
		data := make([]byte, 64)
		lbytes := cidToHash(this.left.Cid)
		copy(data[:32], lbytes)

		rbytes := cidToHash(this.right.Cid)
		copy(data[32:], rbytes)

		this.rawdata = data
	}
	return this.rawdata
}

func (this *EcoballTree) Cid() *cid.Cid {
	if this.cid == nil {
		this.cid = rawdataToCid(cid.EcoballTree, this.RawData())
	}
	return this.cid
}

func (this *EcoballTree) String() string {
	return fmt.Sprintf("<EcoballTree %s>", this.Cid())
}

func (this *EcoballTree) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"type": "ecoball-tree",
	}
}

/*
  Node INTERFACE
*/
func (this *EcoballTree) Tree(p string, depth int) []string {
	return []string{"0", "1"}
}

func (this *EcoballTree) Resolve(path []string) (interface{}, []string, error) {
	if len(path) == 0 {
		return nil, nil, fmt.Errorf("zero length path")
	}

	switch path[0] {
	case "0":
		return this.left, path[1:], nil
	case "1":
		return this.right, path[1:], nil
	default:
		return nil, nil, fmt.Errorf("no such link")
	}
}

func (this *EcoballTree) ResolveLink(path []string) (*node.Link, []string, error) {
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

func (this *EcoballTree) Copy() node.Node {
	panic("dont use this yet")
}

func (this *EcoballTree) Links() []*node.Link {
	return []*node.Link{this.left, this.right}
}

func (this *EcoballTree) Stat() (*node.NodeStat, error) {
	return &node.NodeStat{}, nil
}

func (this *EcoballTree) Size() (uint64, error) {
	return uint64(len(this.rawdata)), nil
}