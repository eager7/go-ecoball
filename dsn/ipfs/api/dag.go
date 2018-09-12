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

package api

import (
	"fmt"
	"bytes"
	"os"
	"sync"
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ecoball/go-ecoball/dsn/ipfs/ipld"
	"github.com/ipfs/go-ipfs/core/commands/dag"
	"github.com/ipfs/go-ipfs/core/coreapi/interface"
	cmd "github.com/ipfs/go-ipfs/commands"
	opt "github.com/ipfs/go-ipfs/core/coreapi/interface/options"
	"gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds"
	//"github.com/ipfs/go-ipfs/core/coreapi"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
)

var lock sync.Mutex
var coreApi iface.CoreAPI = DsnIpfsApi

// PutBlock Put a block dag node to ipfs, API for client
func PutBlock(blockData []byte) (string, error) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	file := filepath.Join(dir, "block.bin")

	lock.Lock()
	defer lock.Unlock()

	if err := ioutil.WriteFile(file, blockData, 0644); err != nil {
		return "", err
	}

	args := []string{"dag", "put", "--input-enc","raw", "--format", "ecoball-rawdata", file}

	cid, err := put(args)
	if err != nil {
		return "", err
	}

	if err := os.Remove(file); err != nil {
		return "", err
	}

	log.Debug("Put a block DAG node: ", cid)

	return cid, nil
}

func put(args []string) (string, error) {
	req, err := NewRequest(args)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	req.Options["encoding"] = cmds.JSON
	req.Command.Type = dagcmd.OutputObject{}
	buf := bytes.NewBuffer(nil)
	wc := writecloser{Writer: buf, Closer: nopCloser{}}
	rsp := cmds.NewWriterResponseEmitter(wc, req, cmds.Encoders[cmds.JSON])
	var env cmd.Context
	env.ConstructNode = func() (*core.IpfsNode, error) {
		//return ipfsCtrl.IpfsNode, nil
		return nil, nil
	}

	Root.Call(req, rsp, &env)

	var result dagcmd.OutputObject
	err = json.Unmarshal(buf.Bytes(), &result)

	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	return result.Cid.String(), nil
}

// PutChainBlock Get a block from IPFS DAG, API for daemon
func PutChainBlock(ctx context.Context, blockData []byte) (string, error) {
	r := bytes.NewReader(blockData)

	if coreApi == nil {
		//coreApi = coreapi.NewCoreAPI(ipfsCtrl.IpfsNode)
		//TODO
		return "", nil
	}
	rp, err := coreApi.Dag().Put(ctx, r, opt.Dag.InputEnc("raw"), opt.Dag.Codec(cid.EcoballRawData))
	if err != nil {
		log.Error("error for putting chain block: ", err)
		return "", err
	}

	cid := rp.Root().String()

	log.Debug("Put a block DAG node: ", cid)

	return cid, nil
}

// GetChainBlock Get a block from IPFS DAG, API for daemon
func GetChainBlock(ctx context.Context, cid string) ([]byte, error) {
	out := make(chan []byte)

	go ipld.ResolveShardLinks(ctx, IpfsNode, cid, out)
	serBlock := <- out
	if serBlock == nil  {
		return nil, fmt.Errorf("error for resolving block link")
	}

	return serBlock, nil
}