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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	clientCommon "github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	innercommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/urfave/cli"
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
						Usage: "active public key(the default is the owner public key)",
					},
					cli.StringFlag{
						Name:  "permission, p",
						Usage: "active permission",
						Value: "active",
					},
					cli.StringFlag{
						Name:  "chainHash, c",
						Usage: "chain hash(the default is the main chain hash)",
					},
					cli.Float64Flag{
						Name:  "max-cpu-usage-ms",
						Usage: "Maximum CPU consumption",
						Value: 0,
					},
					cli.Float64Flag{
						Name:  "max-net-usage",
						Usage: "Maximum bandwidth",
						Value: 0,
					},
				},
			},
		},
	}
)

func newAccount(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//creator
	creator := c.String("creator")
	if creator == "" {
		fmt.Println("Please input a valid creator name")
		return errors.New("Invalid creator name")
	}

	//name
	name := c.String("name")
	if name == "" {
		fmt.Println("Please input a valid account name")
		return errors.New("Invalid account name")
	}

	if err := innercommon.AccountNameCheck(name); nil != err {
		fmt.Println(err)
		return err
	}

	//owner key
	owner := c.String("owner")
	if "" == owner {
		fmt.Println("Please input a valid owner key")
		return errors.New("Invalid owner key")
	}

	//active key
	active := c.String("active")
	if "" == active {
		active = owner
	}

	permission := c.String("permission")

	max_cpu_usage_ms := c.Float64("max-cpu-usage-ms")
	if max_cpu_usage_ms < 0 {
		fmt.Println("Invalid max-cpu-usage-ms ", max_cpu_usage_ms)
		return errors.New("Invalid max-cpu-usage-ms")
	}

	max_net_usage := c.Float64("max-net-usage")
	if max_net_usage < 0 {
		fmt.Println("Invalid max_net_usage ", max_net_usage)
		return errors.New("Invalid max_net_usage")
	}

	//chainHash
	var chainHash innercommon.Hash
	var err error
	chainHashStr := c.String("chainHash")
	if "" == chainHashStr {
		chainHash, err = getMainChainHash()

	} else {
		json.Unmarshal([]byte(chainHashStr), &chainHash)
	}

	if nil != err {
		fmt.Println(err)
		return err
	}

	allPublickeys, err := getPublicKeys()
	if err != nil {
		fmt.Println(err)
		return err
	}

	creatorAccount := innercommon.NameToIndex(creator)
	timeStamp := time.Now().UnixNano()

	invoke, err := types.NewInvokeContract(creatorAccount, creatorAccount, chainHash, "owner", "new_account",
		[]string{name, innercommon.AddressFromPubKey(innercommon.FromHex(owner)).HexString()}, 0, timeStamp)
	if err != nil {
		fmt.Println(err)
	}

	invoke.Receipt.Cpu = max_cpu_usage_ms
	invoke.Receipt.Net = max_net_usage
	//invoke.SetSignature(&config.Root)

	requiredKeys, err := getRequiredKeys(chainHash, permission, creator)
	if err != nil {
		fmt.Println(err)
		return err
	}

	publickeys := ""
	keyDatas := strings.Split(allPublickeys, ",")
	for _, v := range keyDatas {
		addr := innercommon.AddressFromPubKey(innercommon.FromHex(v))
		for _, vv := range requiredKeys {
			if addr == vv {
				publickeys += v
				publickeys += "\n"
				break
			}
		}
	}

	if "" == publickeys {
		fmt.Println("no publickeys")
		return errors.New("no publickeys")
	}

	data, errcode := signTransaction(chainHash, publickeys, invoke)
	if nil != errcode {
		fmt.Println(errcode)
	}

	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("transaction", data)
	err = rpc.NodePost("/invokeContract", values.Encode(), &result)
	fmt.Println(result.Result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
}
