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
	"time"

	"encoding/hex"
	clientCommon "github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/urfave/cli"
)

var (
	VoteCommands = cli.Command{
		Name:        "vote",
		Usage:       "vote operate",
		Category:    "Vote",
		Description: "you can vote to producer",
		ArgsUsage:   "[args]",
		Action:      clientCommon.DefaultAction,
		Subcommands: []cli.Command{
			{
				Name:   "vote",
				Usage:  "vote",
				Action: vote,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "voter, v",
						Usage: "voter",
					},
					cli.StringFlag{
						Name:  "producer1, f",
						Usage: "first producer voted",
					},
					cli.StringFlag{
						Name:  "producer2, s",
						Usage: "second producer voted",
					},
					cli.StringFlag{
						Name:  "chainHash, c",
						Usage: "chain hash(the default is the main chain hash)",
					},
				},
			},
		},
	}
)

func vote(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//contract address
	voter := c.String("voter")
	if voter == "" {
		fmt.Println("Please input a valid contract account name")
		return errors.New("Invalid contract account name")
	}

	producer1 := c.String("producer1")
	if producer1 == "" {
		fmt.Println("Please input a valid contract account name")
		return errors.New("Invalid contract account name")
	}

	producer2 := c.String("producer2")
	if producer2 == "" {
		fmt.Println("Please input a valid contract account name")
		return errors.New("Invalid contract account name")
	}

	//chainHash
	var chainHash common.Hash
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

	//public keys
	allPublickeys, err := getPublicKeys()
	if err != nil {
		fmt.Println(err)
		return err
	}

	var parameters []string

	parameters = append(parameters, voter)
	parameters = append(parameters, producer1)
	parameters = append(parameters, producer2)

	//time
	time := time.Now().UnixNano()

	transaction, err := types.NewInvokeContract(common.NameToIndex(voter), common.NameToIndex("root"), chainHash, "owner", "pledge", parameters, 0, time)
	if nil != err {
		fmt.Println(err)
		return err
	}

	requiredKeys, err := getRequiredKeys(chainHash, "active", voter)
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
	data, errcode := signTransaction(chainHash, publickeys, transaction.Hash.Bytes())
	if nil != errcode {
		fmt.Println(errcode)
		return errcode
	}

	for _, v := range data.Signature {
		transaction.AddSignature(v.PublicKey.Key, v.SignData)
	}

	datas, err := transaction.Serialize()
	if err != nil {
		fmt.Println(err)
		return err
	}

	var result rpc.SimpleResult
	trx_str := hex.EncodeToString(datas)

	err = rpc.NodePost("/invokeContract", &trx_str, &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
}
