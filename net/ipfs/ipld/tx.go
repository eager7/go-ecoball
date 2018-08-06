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
	"errors"
	types "github.com/ecoball/go-ecoball/core/types"
	"gx/ipfs/QmVzK524a2VWLqyvtBeiHKsUAWYgeAk4DBeZoY7vpNPNRx/go-block-format"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	node "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
)

type EcoballTx struct {
	*types.Transaction
	rawdata []byte
	cid     *cid.Cid
}

// assert that tx matches the Node interface for ipld
var _ node.Node = (*EcoballTx)(nil)

func NewEcoballTx(t *types.Transaction) *EcoballTx {
	rawdata, err:= t.Serialize()
	if err != nil {
		return nil
	}
	return &EcoballTx{
		Transaction: t,
		cid:         rawdataToCid(cid.EcoballTx, rawdata),
		rawdata:     rawdata,
	}
}

func DecodeEcoballTx(block blocks.Block) (node.Node, error) {
	prefix := block.Cid().Prefix()

	if prefix.Codec != cid.EcoballTx {
		return nil, errors.New("invalid CID prefix")
	}

	tx := new(types.Transaction)
	if err := tx.Deserialize(block.RawData()); err !=nil {
		return nil, nil
	}

	return &EcoballTx{tx, block.RawData(),block.Cid()}, nil
}

/*
  Block INTERFACE
*/
func (this *EcoballTx) RawData() []byte {
	return this.rawdata
}

func (this *EcoballTx) Cid() *cid.Cid {
	return this.cid
}

func (this *EcoballTx) String() string {
	return fmt.Sprintf("<EcoballTx %s>", this.cid)
}

func (this *EcoballTx) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"type": "ecoball-tx",
	}
}

/*
  Node INTERFACE
*/
func (this *EcoballTx) Tree(p string, depth int) []string {
	return []string{
		"version",
		"type",
		"from",
		"permission",
		"addr",
		"nonce",
		"timestamp",
		"payload",
		"signatures",
		"hash",
	}
}

func (this *EcoballTx) Resolve(path []string) (interface{}, []string, error) {
	if len(path) == 0 {
		return nil, nil, fmt.Errorf("zero length path")
	}
	switch path[0] {
	case "version":
		return this.Version, path[1:], nil
	case "type":
		return this.Type, path[1:], nil
	case "from":
		return this.From, path[1:], nil
	case "permission":
		return this.Permission, path[1:], nil
	case "addr":
		return this.Addr, path[1:], nil
	case "nonce":
		return this.Nonce, path[1:], nil
	case "timestamp":
		return this.TimeStamp, path[1:], nil
	case "payload":
		return this.Payload, path[1:], nil
	case "signatures":
		return this.Signatures, path[1:], nil
	case "hash":
		return this.Hash, path[1:], nil
	default:
		return nil, nil, fmt.Errorf("no such link")
	}
}


func (this *EcoballTx) ResolveLink(path []string) (*node.Link, []string, error) {
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

func (this *EcoballTx) Copy() node.Node {
	panic("dont use this yet")
}

func (this *EcoballTx) Links() []*node.Link {
	return nil
}

func (this *EcoballTx) Stat() (*node.NodeStat, error) {
	return &node.NodeStat{}, nil
}

func (this *EcoballTx) Size() (uint64, error) {
	return uint64(len(this.rawdata)), nil
}