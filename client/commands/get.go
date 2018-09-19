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
	"errors"

	clientCommon "github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/urfave/cli"
	"github.com/ecoball/go-ecoball/core/types"
	innerCommon "github.com/ecoball/go-ecoball/http/common"
)

var (
	QueryCommands = cli.Command{
		Name:     "get",
		Usage:    "operations for query state",
		Category: "Get",
		Action:   clientCommon.DefaultAction,
		Subcommands: []cli.Command{
			{
				Name:   "listchain",
				Usage:  "get all chain id",
				Action: GetChainList,
				Flags: []cli.Flag{},
			},
			{
				Name:   "account",
				Usage:  "get account's info",
				Action: queryAccount,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "account_name, n",
						Usage: "account name",
					},
					cli.StringFlag{
						Name:  "chainId, c",
						Usage: "chainId hash",
						Value: "config.hash",
					},
				},
			},
			{
				Name:   "block",
				Usage:  "get block's info",
				Action: getBlock,
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "height, he",
						Usage: "block height",
						Value: 1,
					},
				},
			},
			{
				Name:   "transaction",
				Usage:  "get transaction 's info",
				Action: getTransaction,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "id, i",
						Usage: "transaction hash",
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

	rpc.EchoErrInfo(resp)
	if int64(innerCommon.SUCCESS) == int64(resp["errorCode"].(float64)){
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
	}
	return nil
}

func get_account(chainId common.Hash, name string) (*state.Account, error) {
	//rpc call
	resp, err := rpc.NodeCall("get_account", []interface{}{chainId.HexString(), name})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	if int64(innerCommon.SUCCESS) == int64(resp["errorCode"].(float64)){
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
	}

	rpc.EchoErrInfo(resp)
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
		return errors.New("Invalid account address")
	}

	info, err := getInfo()
	if err != nil {
		fmt.Println(err)
		return err
	}

	chainId := info.ChainID
	chainIdStr := c.String("chainId")
	if "config.hash" != chainIdStr && "" != chainIdStr {
		chainId = common.HexToHash(chainIdStr)
	}

	accountInfo, err := get_account(chainId, address)
	if nil != err {
		return err
	}
	if nil != accountInfo {
		accountInfo.Show()
	}
	return nil
}

func getBlockInfoById(height int64) (*types.Block, error) {
	resp, err := rpc.NodeCall("getBlock", []interface{}{height})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	if int64(innerCommon.SUCCESS) == int64(resp["errorCode"].(float64)){
		if nil != resp["result"] {
			switch resp["result"].(type) {
			case string:
				data := resp["result"].(string)
				blockINfo := new(types.Block)
				blockINfo.Deserialize(common.FromHex(data))
				return blockINfo, nil
			default:
			}
		}
	}
	rpc.EchoErrInfo(resp)
	return nil, nil
}

func getBlock(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//account address
	height := c.Int64("height")
	if height <= 0 {
		fmt.Println("Invalid block id: ", height)
		return errors.New("Invalid block id")
	}

	block, err := getBlockInfoById(height)
	if nil != err {
		return err
	}
	if nil != block {
		block.Show(false)
	}
	//result
	return nil
}

func getTransaction(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//account address
	id := c.String("id")
	if id == "" {
		fmt.Println("Invalid block id: ", id)
		return errors.New("Invalid block id")
	}

	resp, err := rpc.NodeCall("getTransaction", []interface{}{id})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	rpc.EchoErrInfo(resp)
	if int64(innerCommon.SUCCESS) == int64(resp["errorCode"].(float64)) {
		if nil != resp["result"] {
			switch resp["result"].(type) {
			case string:
				data := resp["result"].(string)
				trx := new(types.Transaction)
				trx.Deserialize(common.FromHex(data))
				fmt.Println(trx.JsonString())
				return nil
			default:
			}
		}
	}

	return nil
}
