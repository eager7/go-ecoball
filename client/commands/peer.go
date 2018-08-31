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

package commands

import (
	"fmt"
	"os"

	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/urfave/cli"
)

var (
	NetworkCommand = cli.Command{
		Name:        "network",
		Usage:       "network manager",
		Category:    "Network",
		Description: "network manager",
		ArgsUsage:   "[args]",
		Subcommands: []cli.Command{
			{
				Name:   "id",
				Usage:  "show my peer id",
				Action: listMyId,
			},
			{
				Name:   "list",
				Usage:  "show my peers",
				Action: listPeers,
			},
		},
	}
)

func listMyId(ctx *cli.Context) error {
	//rpc call
	resp, err := rpc.NodeCall("netlistmyid", []interface{}{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	//result
	return rpc.EchoResult(resp)
}

func listPeers(ctx *cli.Context) error {
	//rpc call
	resp, err := rpc.NodeCall("netlistmypeer", []interface{}{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	//result
	return rpc.EchoResult(resp)
}
