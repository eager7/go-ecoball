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
	"bytes"
	"context"
	"github.com/ipfs/go-ipfs/core"
	"github.com/AsynkronIT/protoactor-go/log"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	node "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
)

func ResolveShardLinks(cctx context.Context, node *core.IpfsNode, pcid string, out chan []byte) {
	var err error
	c, err := cid.Decode(pcid)
	if err != nil {
		log.Error(err)
		out <- nil
		return
	}
	object, err := node.Resolver.DAG.Get(cctx, c)
	if err != nil {
		log.Error(err)
		out <- nil
		return
	}

	links := object.Links()
	if len(links) == 0 {
		err = fmt.Errorf("no child link under the object")
		log.Error(err)
		out <- nil
		return
	}

	shardInfo, err := ResolveShardInfo(object)
	if err != nil {
		log.Error(err)
		out <- nil
		return
	}
	//fmt.Println("shards info:", shardInfo.DataShards,shardInfo.ParityShards, shardInfo.DataSize)
	var shards [][]byte
	var failShards uint32
	for _, link := range links {
		c := link.Cid
		ln, err := node.Resolver.DAG.Get(cctx, c)
		if err != nil && failShards > shardInfo.ParityShards {
			log.Error(err)
			out <- nil
			return
		} else {
			shards = append(shards, ln.RawData())
		}
		// maybe this is a bug in the RS, need m+n shards to recover
/*
		if len(shards) > int(shardInfo.DataShards) {
			break
		}
*/
	}
	fmt.Println("data shards for recover:", len(shards))
	buf := new(bytes.Buffer)
	err = RecoverShards(shardInfo, shards, buf)
	if err == nil {
		out <- buf.Bytes()
		fmt.Println("successed in recovering data shards")
	} else {
		out <- nil
		fmt.Println("error for recovering data shards:", err)
	}
}

func RecoverShards(info *EcoballShardInfo, shards [][]byte, w io.Writer) error {
	rscode, err := NewRSCode(int(info.DataShards), int(info.ParityShards))
	if err != nil {
		return err
	}

	return rscode.Recover(shards, uint64(info.DataSize), w)
}

func ResolveShardInfo(n node.Node) (*EcoballShardInfo, error) {
	name := []string{"DataShardCount"}
	dataShards, _, err := n.Resolve(name)
	if err != nil {
		return nil, fmt.Errorf("error for resolving datashard")
	}

	name = []string{"ParityShardCount"}
	parityShards, _, err := n.Resolve(name)
	if err != nil {
		return nil, fmt.Errorf("error for resolving parityshard")
	}

	name = []string{"DataSize"}
	size, _, err := n.Resolve(name)
	if err != nil {
		return nil, fmt.Errorf("error for resolving data size")
	}

	return &EcoballShardInfo{
		dataShards.(uint32),
		parityShards.(uint32),
		size.(uint64),
		nil,
	}, nil
}