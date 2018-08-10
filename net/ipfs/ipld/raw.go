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
	"io"
	"fmt"
	"errors"
	"io/ioutil"
	"math"
	"bytes"
	"bufio"
	"encoding/binary"

	"github.com/ecoball/go-ecoball/common"
	"gx/ipfs/QmVzK524a2VWLqyvtBeiHKsUAWYgeAk4DBeZoY7vpNPNRx/go-block-format"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	node "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
	"strconv"
)

const (
	DataShardSize = 256 * 1024
	DataShardEC   = true
)

type EcoballShardInfo struct {
	DataShards     uint32    `json:"DataShardCount"`
	ParityShards   uint32	 `json:"ParityShardCount"`
	DataSize       uint64    `json:"DataSize"`
	ShardHashs     []*cid.Cid `json:"Shards"`

}

type EcoballRawData struct {
	*EcoballShardInfo
	rawdata        []byte
	cid            *cid.Cid
}

// assert that EcoballRawData matches the Node interface for ipld
var _ node.Node = (*EcoballRawData)(nil)

/*
  Block INTERFACE
*/
func (this *EcoballRawData) RawData() []byte {
	if this.rawdata == nil {
		buf := new(bytes.Buffer)
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, this.DataShards)
		buf.Write(data)
		binary.LittleEndian.PutUint32(data, this.ParityShards)
		buf.Write(data)
		size := make([]byte, 8)
		binary.LittleEndian.PutUint64(size, this.DataSize)
		buf.Write(size)

		for _, shardCid := range this.ShardHashs {
			buf.Write(cidToHash(shardCid))
		}

		this.rawdata = buf.Bytes()
	}

	return this.rawdata
}

func (this *EcoballRawData) Cid() *cid.Cid {
	if this.cid == nil {
		this.cid = rawdataToCid(cid.EcoballRawData, this.RawData())
	}

	return this.cid
}

func (this *EcoballRawData) String() string {
	return fmt.Sprintf("<EcoballRawData %s>", this.cid)
}

func (this *EcoballRawData) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"type": "ecoball-rawdata",
	}
}

/*
  Node INTERFACE
*/
func (this *EcoballRawData) Tree(p string, depth int) []string {
	if depth == 0 {
		return nil
	}
	switch p {
	case "Shards":
		return this.shardTreeInputs(nil, depth+1)
	case "":
		out := []string{"DataShardCount","ParityShardCount","DataSize"}
		out = this.shardTreeInputs(out, depth)
		return out
	default:
		return nil
	}
}

func (this *EcoballRawData) shardTreeInputs(out []string, depth int) []string {
	if depth < 2 {
		return out
	}

	for i := range this.ShardHashs {
		inp := "Shards/" + fmt.Sprint(i)
		out = append(out, inp)
	}
	return out
}

func (this *EcoballRawData) Resolve(path []string) (interface{}, []string, error) {
	if len(path) == 0 {
		return nil, nil, fmt.Errorf("zero length path")
	}
	switch path[0] {
	case "DataShardCount":
		return this.DataShards, path[1:], nil
	case "ParityShardCount":
		return this.ParityShards, path[1:], nil
	case "DataSize":
		return this.DataSize, path[1:], nil
	case "Shards":
		if len(path) ==  1{
			var shards []*node.Link
			for i, shardCid := range this.ShardHashs {
				name := fmt.Sprintf("Shards/%d", i)
				shards = append(shards, &node.Link{Cid: shardCid, Name: name})
			}
			return shards, path[1:], nil
		}
		index, err := strconv.Atoi(path[1])
		if err != nil {
			return nil, nil, err
		}

		if index >= len(this.ShardHashs) || index < 0 {
			return nil, nil, fmt.Errorf("index out of range")
		}
		name := fmt.Sprintf("Shards/%d", index)
		return &node.Link{Cid: this.ShardHashs[index], Name: name}, path[2:], nil
	default:
		return nil, nil, fmt.Errorf("no such link")
	}
}

func (this *EcoballRawData) ResolveLink(path []string) (*node.Link, []string, error) {
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

func (this *EcoballRawData) Copy() node.Node {
	panic("don't use this yet")
}

func (this *EcoballRawData) Links() []*node.Link {
	var lnks []*node.Link
	for i, shardCid := range this.ShardHashs {
		lnk := &node.Link{Cid: shardCid}
		lnk.Name = fmt.Sprintf("Shards/%d", i)
		lnks = append(lnks, lnk)
	}
	return lnks
}

func (this *EcoballRawData) Stat() (*node.NodeStat, error) {
	return &node.NodeStat{}, nil
}

func (this *EcoballRawData) Size() (uint64, error) {
	return uint64(len(this.RawData())), nil
}

func DecodeRawData(block blocks.Block) (node.Node, error) {
	prefix := block.Cid().Prefix()

	if prefix.Codec != cid.EcoballRawData {
		return nil, errors.New("invalid CID prefix")
	}

	ecoballRawData := &EcoballRawData{
		EcoballShardInfo:&EcoballShardInfo{},
	}
	r := bufio.NewReader(bytes.NewReader(block.RawData()))

	dataShards, err := readFixedSlice(r, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to read datashards: %s", err)
	}
	ecoballRawData.DataShards = binary.LittleEndian.Uint32(dataShards)

	parityShards, err := readFixedSlice(r, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to read parityshards: %s", err)
	}
	ecoballRawData.ParityShards = binary.LittleEndian.Uint32(parityShards)

	dataSize, err := readFixedSlice(r, 8)
	if err != nil {
		return nil, fmt.Errorf("failed to read datasize: %s", err)
	}
	ecoballRawData.DataSize = binary.LittleEndian.Uint64(dataSize)

	totalShards := ecoballRawData.DataShards + ecoballRawData.ParityShards
	for i:=0; i<int(totalShards); i++ {
		hashData, err := readFixedSlice(r, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to read shard hash: %s", err)
		}
		shardCid := commonHashToCid(cid.EcoballShardData, common.NewHash(hashData))
		ecoballRawData.ShardHashs = append(ecoballRawData.ShardHashs, shardCid)
	}

	return ecoballRawData, nil
}

var _ node.DecodeBlockFunc = DecodeRawData

func encodeBlockShardWithRs(data []byte) ([]node.Node, error) {
	dataShardCount := math.Ceil(float64(len(data)) / DataShardSize)
	parityShardCount := math.Ceil(dataShardCount * 2.0 / 3.0)  //parityshards = 2/3 data shards

	rscode, err := NewRSCode(int(dataShardCount), int(parityShardCount))
	if err != nil {
		return nil, err
	}

	shardRawData, err := rscode.Encode(data)
	if err != nil {
		return nil, err
	}

	var dataShards []node.Node
	var cids []*cid.Cid
	for _, shard := range shardRawData {
		shardCid := rawdataToCid(cid.EcoballShardData, shard)
		dataShard := &EcoballShard{
			ShardInfo: &ShardInfo{shard},
		}
		dataShards = append(dataShards, dataShard)
		cids = append(cids, shardCid)
	}

	shardinfo := &EcoballShardInfo{
		uint32(dataShardCount),
		uint32(parityShardCount),
		uint64(len(data)),
		cids,
	}

	dataNode := &EcoballRawData{
		EcoballShardInfo:shardinfo,
	}

	out := []node.Node {dataNode}
	out = append(out, dataShards...)

	return out, nil
}

func EcoballRawDataInputParser(r io.Reader, mhType uint64, mhLen int) ([]node.Node, error) {
	rawdata, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if DataShardEC {
		return encodeBlockShardWithRs(rawdata)
	}

	shardInfo := &EcoballShardInfo{}

	dataNode := &EcoballRawData{
		shardInfo,
		rawdata,
		rawdataToCid(cid.EcoballRawData, rawdata),
	}
	out := []node.Node {dataNode}

	return out, nil
}