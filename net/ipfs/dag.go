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

package ipfs

import (
	"fmt"
	"bytes"
	"gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"os"
	"sync"
	"context"
	"github.com/ecoball/go-ecoball/net/ipfs/ipld"
	"github.com/ipfs/go-ipfs/core/commands/dag"
	cmd "github.com/ipfs/go-ipfs/commands"
)
var lock sync.Mutex

// Put a block dag node to ipfs
func PutSerBlock(format string, blockData []byte) (string, error) {
	if format != "ecoball-rawblock" {
		return "", fmt.Errorf("invalid block format")
	}

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	file := filepath.Join(dir, "block.bin")

	lock.Lock()
	defer lock.Unlock()

	if err := ioutil.WriteFile(file, blockData, 0644); err != nil {
		return "", err
	}

	args := []string{"dag", "put", "--input-enc","raw", "--format", format, file}

	cid, err := put(args)
	if err != nil {
		return "", err
	}

	if err := os.Remove(file); err != nil {
		return "", err
	}

	log.Debug("Put a block DAG node: %s\n", cid)

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
		return IpfsNode, nil
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

// Get a block from IPFS DAG
func GetSerBlock(cid string) ([]byte, error) {
	cctx := context.Background()
	out := make(chan []byte)

	go ipldecoball.ResolveShardLinks(cctx, IpfsNode, cid, out)
	serBlock := <- out
	if serBlock == nil  {
		return nil, fmt.Errorf("error for resolving block link")
	}

	return serBlock, nil
}