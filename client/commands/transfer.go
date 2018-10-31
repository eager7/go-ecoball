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
	"net/url"
	//"os"

	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/urfave/cli"

	"math/big"
	"time"

	clientCommon "github.com/ecoball/go-ecoball/client/common"
	inner "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	//"github.com/ecoball/go-ecoball/common/config"
)

var (
	TransferCommands = cli.Command{
		Name:        "transfer",
		Usage:       "user ABA transfer",
		Category:    "Transfer",
		Description: "Transfer ABA to other users",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "from, f",
				Usage: "sender name",
			},
			cli.StringFlag{
				Name:  "to, t",
				Usage: "receiver name",
			},
			cli.Int64Flag{
				Name:  "value, v",
				Usage: "ABA amount",
			},
			cli.StringFlag{
				Name:  "chainId, c",
				Usage: "chainId hash",
				Value: "config.hash",
			},
		},
		Action: transferAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			return cli.NewExitError("", 1)
		},
	}
)

func transferAction(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	from := c.String("from")
	if from == "" {
		fmt.Println("Please input a valid from account")
		return errors.New("Invalid sender name")
	}

	to := c.String("to")
	if to == "" {
		fmt.Println("Please input a valid to account", to)
		return errors.New("Invalid revicer name")
	}

	value := c.Int64("value")
	if value == 0 {
		fmt.Println("Please input a valid aba amount")
		return errors.New("Invalid aba amount")
	}

	bigValue := big.NewInt(value)

	info, err := getInfo()
	if err != nil {
		fmt.Println(err)
		return err
	}

	chainId := info.ChainID
	chainIdStr := c.String("chainId")
	if "config.hash" != chainIdStr && "" != chainIdStr {
		chainId = inner.HexToHash(chainIdStr)
	}

	publickeys, err := GetPublicKeys()
	if err != nil {
		fmt.Println(err)
		return err
	}

	//time
	time := time.Now().UnixNano()

	transaction, err := types.NewTransfer(inner.NameToIndex(from), inner.NameToIndex(to), chainId, "owner", bigValue, 0, time)
	if nil != err {
		return err
	}

	permission := "active"
	required_keys, err := get_required_keys(info.ChainID, publickeys, permission, transaction)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if required_keys == "" {
		fmt.Println("no required_keys")
		return err
	}

	data, errcode := sign_transaction(info.ChainID, required_keys, transaction)
	if nil != errcode {
		fmt.Println(errcode)
	}

	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("transfer", data)
	err = rpc.NodePost("/transfer", values.Encode(), &result)
	fmt.Println(result.Result)

	return err
}
