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
	"errors"
	"fmt"
	//"os"
	"strconv"
	"time"
	"net/url"

	clientCommon "github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	innercommon "github.com/ecoball/go-ecoball/common"
	//"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/urfave/cli"
	//outerCommon "github.com/ecoball/go-ecoball/http/common"
)

var (
	CreateCommands = cli.Command{
		Name:     "create",
		Usage:    "create operations",
		Category: "Create",
		Action:   clientCommon.DefaultAction,
		Subcommands: []cli.Command{
			{
				Name:   "account",
				Usage:  "create account",
				Action: newAccount,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "creator, c",
						Usage: "creator name",
					},
					cli.StringFlag{
						Name:  "name, n",
						Usage: "account name",
					},
					cli.StringFlag{
						Name:  "owner, o",
						Usage: "owner public key",
					},
					cli.StringFlag{
						Name:  "active, a",
						Usage: "active public key",
						Value: "owner",
					},
					cli.StringFlag{
						Name:  "permission, p",
						Usage: "active permission",
						Value: "active",
					},
					cli.StringFlag{
						Name:  "chainId",
						Usage: "chainId hash",
						Value: "config.hash",
					},
					cli.StringFlag{
						Name:  "max-cpu-usage-ms",
						Usage: "max-cpu-usage-ms",
						Value: "0",
					},
					cli.StringFlag{
						Name:  "max-net-usage",
						Usage: "max-net-usage",
						Value: "0",
					},
				},
			},
		},
	}
)

func getInfo() (*types.Block, error) {
	var result clientCommon.SimpleResult
	err := rpc.NodeGet("/getInfo", &result)
	if nil == err {
		blockINfo := new(types.Block)
		err := blockINfo.Deserialize(innercommon.FromHex(result.Result))
		if nil == err {
			return blockINfo, nil
		}
	}
	return nil, err

	/*resp, err := rpc.NodeCall("getInfo", []interface{}{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	if int64(outerCommon.SUCCESS) == int64(resp["errorCode"].(float64)) {
		if nil != resp["result"] {
			switch resp["result"].(type) {
			case string:
				data := resp["result"].(string)
				blockINfo := new(types.Block)
				blockINfo.Deserialize(innercommon.FromHex(data))
				//blockINfo.Show(true)
				return blockINfo, nil
			default:
			}
		}
	}
	return nil, nil*/
}

func get_required_keys(chainId innercommon.Hash, required_keys, permission string, trx *types.Transaction) (string, error) {
	/*data, err := trx.Serialize()
	if err != nil {
		return "", err
	}*/

	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("permission", permission)
	values.Set("chainId", chainId.HexString())
	values.Set("keys", required_keys)
	values.Set("name", trx.From.String())
	err := rpc.NodePost("/get_required_keys", values.Encode(), &result)
	if nil == err {
		return result.Result, nil
	}
	return "", err
	/*data, err := trx.Serialize()
	if err != nil {
		return "", err
	}

	resp, errcode := rpc.NodeCall("get_required_keys", []interface{}{chainId.HexString(), required_keys, permission, innercommon.ToHex(data)})
	if errcode != nil {
		fmt.Fprintln(os.Stderr, errcode)
		return "", errcode
	}

	if int64(outerCommon.SUCCESS) == int64(resp["errorCode"].(float64)){
		if nil != resp["result"] {
			switch resp["result"].(type) {
			case string:
				data := resp["result"].(string)
				return data, nil
			default:
			}
		}
	}
	return "", err*/
}

func newAccount(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//creator
	creator := c.String("creator")
	if creator == "" {
		fmt.Println("Invalid creator name")
		return errors.New("Invalid creator name")
	}

	//name
	name := c.String("name")
	if name == "" {
		fmt.Println("Invalid account name")
		return errors.New("Invalid account name")
	}

	if err := innercommon.AccountNameCheck(name); nil != err {
		fmt.Println(err)
		return err
	}

	//owner key
	owner := c.String("owner")
	if "" == owner {
		fmt.Println("Invalid owner key")
		return errors.New("Invalid owner key")
	}

	//active key
	active := c.String("active")
	if "" == active {
		active = owner
	}

	permission := c.String("permission")
	if "" == permission {
		permission = "active"
	}

	max_cpu_usage_ms, err := strconv.ParseFloat(c.String("max-cpu-usage-ms"), 64)
	if err != nil {
		fmt.Println(err)
		return err
	}

	max_net_usage, err := strconv.ParseFloat(c.String("max-net-usage"), 64)
	if err != nil {
		fmt.Println(err)
		return err
	}

	info, err := getInfo()
	if err != nil {
		fmt.Println(err)
		return err
	}

	chainId := info.ChainID
	chainIdStr := c.String("chainId")
	if "config.hash" != chainIdStr && "" != chainIdStr {
		chainId = innercommon.HexToHash(chainIdStr)
	}

	publickeys, err := GetPublicKeys()
	if err != nil {
		fmt.Println("get publicKey failed")
		fmt.Println(err)
		return err
	}

	creatorAccount := innercommon.NameToIndex(creator)
	timeStamp := time.Now().UnixNano()

	invoke, err := types.NewInvokeContract(creatorAccount, creatorAccount, chainId, "owner", "new_account",
		[]string{name, innercommon.AddressFromPubKey(innercommon.FromHex(owner)).HexString()}, 0, timeStamp)
	if err != nil {
		fmt.Println(err)
	}

	invoke.Receipt.Cpu = max_cpu_usage_ms
	invoke.Receipt.Net = max_net_usage
	//invoke.SetSignature(&config.Root)

	required_keys, err := get_required_keys(info.ChainID, publickeys, permission, invoke)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(required_keys)
	fmt.Println(info.ChainID.HexString())

	if required_keys == "" {
		fmt.Println("no required_keys")
		return err
	}

	errcode := sign_transaction(info.ChainID, required_keys, invoke)
	if nil != errcode {
		fmt.Println(errcode)
	}

	//rpc call
	//resp, err := rpc.Call("createAccount", []interface{}{creator, name, owner, active})
	data, err := invoke.Serialize()
	if err != nil {
		fmt.Println(err)
		return err
	}

	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("transaction", innercommon.ToHex(data))
	err = rpc.NodePost("/invokeContract", values.Encode(), &result)
	fmt.Println(result.Result)

	return err
}
