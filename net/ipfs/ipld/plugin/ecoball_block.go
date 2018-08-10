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

package main

import (
	"fmt"
	"github.com/ipfs/go-ipfs/core/coredag"
	"github.com/ipfs/go-ipfs/plugin"
	"github.com/ecoball/go-ecoball/net/ipfs/ipld"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	"gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
)

type ecoballPlugin struct{}

var _ plugin.PluginIPLD = (*ecoballPlugin)(nil)

func (this *ecoballPlugin) Name() string {
	return "ipld-ecoball"
}

func (this *ecoballPlugin) Version() string {
	return "0.0.1"
}

func (this *ecoballPlugin) Init() error {
	fmt.Println("ecoball ipld plugin init.")
	return nil
}

func (this *ecoballPlugin) RegisterBlockDecoders(dec format.BlockDecoder) error {
	fmt.Println("ecoball ipld plugin register decoders.")
	dec.Register(cid.EcoballBlock, ipldecoball.DecodeBlock)
	return nil
}

func (this *ecoballPlugin) RegisterInputEncParsers(iec coredag.InputEncParsers) error {
	iec.AddParser("raw", "ecoball-block", ipldecoball.EcoballBlockRawInputParser)
	return nil
}