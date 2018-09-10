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

	clientCommon "github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/urfave/cli"
)

var (
	QueryCommands = cli.Command{
		Name:     "query",
		Usage:    "operations for query state",
		Category: "Query",
		Action:   clientCommon.DefaultAction,
		Subcommands: []cli.Command{
			{
				Name:   "balance",
				Usage:  "query account's balance",
				Action: queryBalance,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "address, a",
						Usage: "account address",
					},
				},
			},
		},
	}
)

func get_account(name string) (*state.Account, error) {
	info, err := getInfo()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	//rpc call
	resp, err := rpc.NodeCall("get_account", []interface{}{info.ChainID.HexString(), name})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	if nil != resp["result"] {
		switch resp["result"].(type) {
		case string:
			data := resp["result"].(string)
			accountInfo := new(state.Account)
			accountInfo.Deserialize(common.FromHex(data))
			return accountInfo, nil
		default:
		}
	}
	return nil, nil
}

func queryBalance(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//account address
	address := c.String("address")
	if address == "" {
		fmt.Println("Invalid account address: ", address)
	}

	//rpc call
	resp, err := rpc.NodeCall("query", []interface{}{string("balance"), address})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	//result
	return rpc.EchoResult(resp)
}
