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
	"errors"
	"fmt"

	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/urfave/cli"

	"math/big"
	"time"

	clientCommon "github.com/ecoball/go-ecoball/client/common"
	inner "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
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
				Name:  "chainHash, c",
				Usage: "chain hash(the default is the main chain hash)",
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
	if value <= 0 {
		fmt.Println("Invalid aba amount ", value)
		return errors.New("Invalid aba amount")
	}

	bigValue := big.NewInt(value)

	//chainHash
	var chainHash inner.Hash
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

	//public keys
	allPublickeys, err := getPublicKeys()
	if err != nil {
		fmt.Println(err)
		return err
	}

	//time
	time := time.Now().UnixNano()

	transaction, err := types.NewTransfer(inner.NameToIndex(from), inner.NameToIndex(to), chainHash, "owner", bigValue, 0, time)
	if nil != err {
		fmt.Println(err)
		return err
	}

	permission := "active"
	requiredKeys, err := getRequiredKeys(chainHash, permission, from)
	if err != nil {
		fmt.Println(err)
		return err
	}

	publickeys := clientCommon.IntersectionKeys(allPublickeys, requiredKeys)
	if 0 == len(publickeys.KeyList) {
		fmt.Println("no publickeys")
		return errors.New("no publickeys")
	}

	//sign
	data, errcode := signTransaction(chainHash, publickeys, transaction.Hash[:])
	if nil != errcode {
		fmt.Println(errcode)
		return errcode
	}

	for _, v := range data {
		transaction.AddSignature(v.PublicKey.Key, v.SignData)
	}

	var result rpc.SimpleResult
	err = rpc.NodePost("/invokeContract", transaction, &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
}
