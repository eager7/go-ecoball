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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	clientCommon "github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/http/commands"
	"github.com/ecoball/go-ecoball/http/common/abi"
	"github.com/urfave/cli"
)

var (
	QueryCommands = cli.Command{
		Name:     "query",
		Usage:    "operations for query info",
		Category: "Query",
		Action:   clientCommon.DefaultAction,
		Subcommands: []cli.Command{
			{
				Name:   "chain",
				Usage:  "get all chain information",
				Action: getAllChainInfo,
			},
			{
				Name:   "account",
				Usage:  "get account's info by name and chain hash(the default is the main chain hash)",
				Action: getAccount,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "account name",
					},
					cli.StringFlag{
						Name:  "chainHash, c",
						Usage: "chain hash",
					},
				},
			},
			{
				Name:   "token",
				Usage:  "get token's info by name and chain hash(the default is the main chain hash)",
				Action: getTokenInfo,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "token name",
					},
					cli.StringFlag{
						Name:  "chainHash, c",
						Usage: "chain hash",
					},
				},
			},
			{
				Name:   "block",
				Usage:  "get block's info by height and chain hash(the default is the main chain hash)",
				Action: getBlockInfo,
				Flags: []cli.Flag{
					cli.Uint64Flag{
						Name:  "height, t",
						Usage: "block height",
						Value: 1,
					},
					cli.StringFlag{
						Name:  "chainHash, c",
						Usage: "chain hash",
					},
				},
			},
			{
				Name:   "transaction",
				Usage:  "get transaction's info by hash and chain hash(the default is the main chain hash)",
				Action: getTransaction,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "hash, a",
						Usage: "transaction hash",
					},
					cli.StringFlag{
						Name:  "chainHash, c",
						Usage: "chain hash",
					},
				},
			},
		},
	}
)

func getAllChainInfo(c *cli.Context) error {
	var result []state.Chain
	err := rpc.NodeGet("/query/allChainInfo", &result)
	if nil == err {
		for _, v := range result {
			fmt.Println(hex.EncodeToString(v.Hash[:]))
		}
	}
	return err
}

func getAccount(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//account name
	name := c.String("name")
	if name == "" {
		fmt.Println("Please input a valid account name")
		return errors.New("Invalid account name")
	}

	//chainHash
	var chainHash innerCommon.Hash
	var err error
	chainHashStr := c.String("chainHash")
	if "" == chainHashStr {
		chainHash, err = getMainChainHash()

	} else {
		var hashTemp []byte
		hashTemp, err = hex.DecodeString(chainHashStr)
		copy(chainHash[:], hashTemp)
	}

	if nil != err {
		fmt.Println(err)
		return err
	}

	//http request
	var result state.Account
	requestData := commands.AccountName{Name: name, ChainHash: chainHash}
	err = rpc.NodePost("/query/getAccountInfo", &requestData, &result)
	if nil == err {
		fmt.Println(result.JsonString(true))
	}
	return err
}

func getTokenInfo(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//token name
	name := c.String("name")
	if name == "" {
		fmt.Println("Please input a valid token name")
		return errors.New("Invalid token name")
	}

	//chainHash
	var chainHash innerCommon.Hash
	var err error
	chainHashStr := c.String("chainHash")
	if "" == chainHashStr {
		chainHash, err = getMainChainHash()

	} else {
		var hashTemp []byte
		hashTemp, err = hex.DecodeString(chainHashStr)
		copy(chainHash[:], hashTemp)
	}

	if nil != err {
		fmt.Println(err)
		return err
	}

	//http request
	var result state.TokenInfo
	requestData := commands.TokenName{Name: name, ChainHash: chainHash}
	err = rpc.NodePost("/query/getTokenInfo", &requestData, &result)
	if nil == err {
		fmt.Println(result.JsonString(true))
	}
	return err
}

func getBlockInfo(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//block hight
	height := c.Uint64("height")
	if 0 == height {
		fmt.Println("Invalid block height")
		return errors.New("Invalid block height")
	}

	//chainHash
	var chainHash innerCommon.Hash
	var err error
	chainHashStr := c.String("chainHash")
	if "" == chainHashStr {
		chainHash, err = getMainChainHash()

	} else {
		var hashTemp []byte
		hashTemp, err = hex.DecodeString(chainHashStr)
		copy(chainHash[:], hashTemp)
	}

	if nil != err {
		fmt.Println(err)
		return err
	}

	//http request
	var result types.Block
	requestData := commands.BlockHeight{Height: height, ChainHash: chainHash}
	err = rpc.NodePost("/query/getBlockInfo", &requestData, &result)
	if nil == err {
		fmt.Println(result.JsonString(true))
	}
	return err
}

func getTransaction(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//transaction address
	hashStr := c.String("hash")
	if hashStr == "" {
		fmt.Println("Please input a valid transaction hash")
		return errors.New("Invalid transaction hash")
	}

	var hash innerCommon.Hash
	err := json.Unmarshal([]byte(hashStr), &hash)
	if nil != err {
		fmt.Println(err)
		return err
	}

	//chainHash
	var chainHash innerCommon.Hash
	chainHashStr := c.String("chainHash")
	if "" == chainHashStr {
		chainHash, err = getMainChainHash()

	} else {
		var hashTemp []byte
		hashTemp, err = hex.DecodeString(chainHashStr)
		copy(chainHash[:], hashTemp)
	}

	if nil != err {
		fmt.Println(err)
		return err
	}

	//http request
	var result types.Transaction
	requestData := commands.TransactionHash{Hash: hash, ChainHash: chainHash}
	err = rpc.NodePost("/query/getTransaction", &requestData, &result)
	if nil == err {
		fmt.Println(result.JsonString())
	}
	return err
}

//other query method
func getMainChainHash() (innerCommon.Hash, error) {
	var result innerCommon.Hash
	err := rpc.NodeGet("/query/mainChainHash", &result)
	if nil != err {
		return innerCommon.Hash{}, err
	}

	return result, nil
}

func getRequiredKeys(chainHash innerCommon.Hash, permission string, account string) ([]innerCommon.Address, error) {
	var result []innerCommon.Address
	requestData := commands.PermissionPublicKeys{Name: account, Permission: permission, ChainHash: chainHash}
	err := rpc.NodePost("/query/getRequiredKeys", &requestData, &result)
	if nil != err {
		return result, nil
	}

	return []innerCommon.Address{}, err
}

func getContract(chainID innerCommon.Hash, index innerCommon.AccountName) (*types.DeployInfo, error) {
	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("contractName", index.String())
	values.Set("chainId", chainID.HexString())
	err := rpc.NodePost("/query/getContract", values.Encode(), &result)
	if nil == err {
		deploy := new(types.DeployInfo)
		if err := deploy.Deserialize(innerCommon.FromHex(result.Result)); err != nil {
			return nil, err
		}
		return deploy, nil
	}
	return nil, err
}

func storeGet(chainID innerCommon.Hash, index innerCommon.AccountName, key []byte) (value []byte, err error) {
	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("contractName", index.String())
	values.Set("chainId", chainID.HexString())
	values.Set("key", innerCommon.ToHex(key))
	err = rpc.NodePost("/query/storeGet", values.Encode(), &result)
	if nil == err {
		return innerCommon.FromHex(result.Result), nil
	}
	return nil, err
}

func getContractTable(contractName string, accountName string, abiDef abi.ABI, tableName string) ([]byte, error) {
	var fields []abi.FieldDef
	for _, table := range abiDef.Tables {
		if string(table.Name) == tableName {
			for _, struction := range abiDef.Structs {
				if struction.Name == table.Type {
					fields = struction.Fields
				}
			}
		}
	}

	if fields == nil {
		return nil, errors.New("can not find struct of table: " + tableName)
	}

	table := make(map[string]string, len(fields))

	for i, _ := range fields {
		key := []byte(fields[i].Name)
		if fields[i].Name == "balance" { // only for token contract, because KV struct can't support
			key = []byte(accountName)
		} else {
			key = append(key, 0) // C lang string end with 0
		}

		storage, err := storeGet(config.ChainHash, common.NameToIndex(contractName), key)
		if err != nil {
			return nil, errors.New("can not get store " + fields[i].Name)
		}
		fmt.Println(fields[i].Name + ": " + string(storage))
		table[fields[i].Name] = string(storage)
	}

	js, _ := json.Marshal(table)
	fmt.Println("json format: ", string(js))

	return nil, nil
}
