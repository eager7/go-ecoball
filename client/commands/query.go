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
	"strings"

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
				Name:   "listchain",
				Usage:  "query all chain id",
				Action: GetChainList,
				Flags: []cli.Flag{},
			},
			{
				Name:   "account",
				Usage:  "query account's info",
				Action: queryAccount,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "account_name, n",
						Usage: "account name",
					},
				},
			},
		},
	}
)

func GetChainList(c *cli.Context) error {
	resp, err := rpc.NodeCall("Get_ChainList", []interface{}{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if nil != resp["result"] {
		switch resp["result"].(type) {
		case string:
			data := resp["result"].(string)
			//chainList := []state.Chain{}
			chainInfo_str := strings.Split(data, "\n")
			for _, v := range chainInfo_str {
				/*chain := new(state.Chain)
				chain_str := strings.Split(v, ":")
				chain.Index = common.NameToIndex(chain_str[0])
				chain.Hash = common.HexToHash(chain_str[1])
				chainList = append(chainList, *chain)*/
				fmt.Println(v)
			}
			return nil
		default:
		}
	}
	return nil
}

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

func queryAccount(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//account address
	address := c.String("account_name")
	if address == "" {
		fmt.Println("Invalid account address: ", address)
	}

	accountInfo, err := get_account(address)
	accountInfo.Show()
	if nil != err {
		return err
	}
	return nil
}
