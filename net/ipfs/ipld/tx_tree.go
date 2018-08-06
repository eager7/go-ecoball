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
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	node "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
	"fmt"
)

type EcoballTxTree struct {
	left  *node.Link
	right *node.Link

	rawdata []byte
	cid     *cid.Cid
}

// assert that Block matches the Node interface for ipld
var _ node.Node = (*EcoballTxTree)(nil)

func cidToHash(c *cid.Cid) []byte {
	h := []byte(c.Hash())
	return h[len(h)-32:]
}

func buildTxTreeRawdata(leftCid *cid.Cid, rightCid *cid.Cid) []byte {
	out := make([]byte, 64)
	lbytes := cidToHash(leftCid)
	copy(out[:32], lbytes)

	rbytes := cidToHash(rightCid)
	copy(out[32:], rbytes)

	return out
}

func mkTxMerkleTree(txs []node.Node) ([]*EcoballTxTree, error) {
	var out []*EcoballTxTree
	var next []node.Node
	layer := txs
	for len(layer) > 1 {
		if len(layer)%2 != 0 {
			layer = append(layer, layer[len(layer)-1])
		}
		for i := 0; i < len(layer)/2; i++ {
			var left, right node.Node
			left = layer[i*2]
			right = layer[(i*2)+1]

			rawdata := buildTxTreeRawdata(left.Cid(), right.Cid())
			t := &EcoballTxTree{
				left:    &node.Link{Cid: left.Cid()},
				right:   &node.Link{Cid: right.Cid()},
				rawdata: rawdata,
				cid:     rawdataToCid(cid.EcoballTxTree, rawdata),
			}

			out = append(out, t)
			next = append(next, t)
		}

		layer = next
		next = nil
	}

	return out, nil
}

/*
  Block INTERFACE
*/
func (this *EcoballTxTree) RawData() []byte {
	return this.rawdata
}

func (this *EcoballTxTree) Cid() *cid.Cid {
	return this.cid
}

func (this *EcoballTxTree) String() string {
	return fmt.Sprintf("<EcoballTxTree %s>", this.cid)
}

func (this *EcoballTxTree) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"type": "ecoball-txtree",
	}
}

/*
  Node INTERFACE
*/
func (this *EcoballTxTree) Tree(p string, depth int) []string {
	return []string{"0", "1"}
}

func (this *EcoballTxTree) Resolve(path []string) (interface{}, []string, error) {
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

func (this *EcoballTxTree) ResolveLink(path []string) (*node.Link, []string, error) {
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

func (this *EcoballTxTree) Copy() node.Node {
	panic("dont use this yet")
}

func (this *EcoballTxTree) Links() []*node.Link {
	return []*node.Link{this.left, this.right}
}

func (this *EcoballTxTree) Stat() (*node.NodeStat, error) {
	return &node.NodeStat{}, nil
}

func (this *EcoballTxTree) Size() (uint64, error) {
	return uint64(len(this.rawdata)), nil
}