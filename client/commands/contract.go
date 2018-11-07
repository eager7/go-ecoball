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
	"io/ioutil"
	"os"
	"time"

	"encoding/hex"
	"encoding/json"

	"strings"

	clientCommon "github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/http/common/abi"
	"github.com/urfave/cli"
)

var (
	ContractCommands = cli.Command{
		Name:        "contract",
		Usage:       "contract operate",
		Category:    "Contract",
		Description: "you could deploy or execute contract",
		ArgsUsage:   "[args]",
		Action:      clientCommon.DefaultAction,
		Subcommands: []cli.Command{
			{
				Name:   "deploy",
				Usage:  "deploy contract",
				Action: setContract,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "path, p",
						Usage: "contract file path",
					},
					cli.StringFlag{
						Name:  "name, n",
						Usage: "contract acount name",
					},
					cli.StringFlag{
						Name:  "description, d",
						Usage: "contract description",
					},
					cli.StringFlag{
						Name:  "abipath, i",
						Usage: "abi file path",
					},
					cli.StringFlag{
						Name:  "permission, r",
						Usage: "active permission",
						Value: "active",
					},
					cli.StringFlag{
						Name:  "chainHash, c",
						Usage: "chain hash(the default is the main chain hash)",
					},
				},
			},
			{
				Name:   "invoke",
				Usage:  "invoke contract",
				Action: invokeContract,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "contract name",
					},
					cli.StringFlag{
						Name:  "method, m",
						Usage: "contract method",
					},
					cli.StringFlag{
						Name:  "param, p",
						Usage: "method parameters",
					},
					cli.StringFlag{
						Name:  "invoker, i",
						Usage: "Invoker account name",
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

func setContract(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//contract file path
	fileName := c.String("path")
	if fileName == "" {
		fmt.Println("Please input a valid contrace file path")
		return errors.New("Invalid contrace file path")
	}

	//abi file path
	abi_fileName := c.String("abipath")
	if abi_fileName == "" {
		fmt.Println("Please input a valid abi file path")
		return errors.New("Invalid abi file path")
	}

	//file data
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("open file failed")
		return errors.New("open file failed: " + fileName)
	}

	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("read contract filr err: ", err.Error())
		return err
	}

	//abifile data
	abifile, err := os.OpenFile(abi_fileName, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("open file failed")
		return errors.New("open file failed: " + abi_fileName)
	}

	defer abifile.Close()
	abidata, err := ioutil.ReadAll(abifile)
	if err != nil {
		fmt.Println("read abi filr err: ", err.Error())
		return err
	}

	var contractAbi abi.ABI
	if err = json.Unmarshal(abidata, &contractAbi); err != nil {
		fmt.Errorf("ABI Unmarshal failed")
		return err
	}

	abibyte, err := abi.MarshalBinary(contractAbi)
	if err != nil {
		fmt.Errorf("ABI MarshalBinary failed")
		return err
	}

	//contract name
	contractName := c.String("name")
	if contractName == "" {
		fmt.Println("Please input your account name")
		return errors.New("Invalid account name")
	}

	//contract description
	description := c.String("description")
	if description == "" {
		fmt.Println("Please input a valid contract description")
		return errors.New("Invalid contract description")
	}

	permission := c.String("permission")

	//chainHash
	var chainHash common.Hash
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

	time := time.Now().UnixNano()
	transaction, err := types.NewDeployContract(common.NameToIndex(contractName), common.NameToIndex(contractName), chainHash, "owner", types.VmWasm, description, data, abibyte, 0, time)
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

	requiredKeys, err := getRequiredKeys(chainHash, permission, contractName)
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
	signData, errcode := signTransaction(chainHash, publickeys, transaction.Hash.Bytes())
	if nil != errcode {
		fmt.Println(errcode)
		return errcode
	}

	for _, v := range signData.Signature {
		transaction.AddSignature(v.PublicKey.Key, v.SignData)
	}

	//rpc call
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

func invokeContract(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//contract address
	contractName := c.String("name")
	if contractName == "" {
		fmt.Println("Please input a valid contract account name")
		return errors.New("Invalid contract account name")
	}

	//contract method
	contractMethod := c.String("method")
	if contractMethod == "" {
		fmt.Println("Please input a valid contract method")
		return errors.New("Invalid contract method")
	}

	//contract parameter
	contractParam := c.String("param")

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
	if "new_account" == contractMethod {
		parameter := strings.Split(contractParam, ",")
		for _, v := range parameter {
			if strings.Contains(v, "0x") {
				parameters = append(parameters, common.AddressFromPubKey(common.FromHex(v)).HexString())
			} else {
				parameters = append(parameters, v)
			}
		}
	} else if "pledge" == contractMethod || "cancel_pledge" == contractMethod || "reg_prod" == contractMethod || "vote" == contractMethod {
		parameters = strings.Split(contractParam, ",")
	} else if "set_account" == contractMethod {
		parameters = strings.Split(contractParam, "--")
	} else if "reg_chain" == contractMethod {
		parameter := strings.Split(contractParam, ",")
		if len(parameter) == 3 {
			parameters = append(parameters, parameter[0])
			parameters = append(parameters, parameter[1])
			parameters = append(parameters, common.AddressFromPubKey(common.FromHex(parameter[2])).HexString())
		} else {
			return errors.New("Invalid parameters")
		}
	} else {
		contract, err := getContract(chainHash, common.NameToIndex(contractName))
		if err != nil {
			return errors.New("getContract failed")
		}

		var abiDef abi.ABI
		err = abi.UnmarshalBinary(contract.Abi, &abiDef)
		if err != nil {
			return errors.New("can not find UnmarshalBinary abi file")
		}

		//log.Debug("contractParam: ", contractParam)
		argbyte, err := abi.CheckParam(abiDef, contractMethod, []byte(contractParam))
		if err != nil {
			fmt.Println(err.Error())
			return errors.New("checkParam error")
		}

		parameters = append(parameters, string(argbyte[:]))
	}

	//contract address
	invoker := c.String("invoker")
	if invoker == "" {
		fmt.Println("Please input a valid invoker account name")
		return errors.New("Invalid invoker account name")
	}

	//time
	time := time.Now().UnixNano()

	transaction, err := types.NewInvokeContract(common.NameToIndex(invoker), common.NameToIndex(contractName), chainHash, "owner", contractMethod, parameters, 0, time)
	if nil != err {
		fmt.Println(err)
		return err
	}

	requiredKeys, err := getRequiredKeys(chainHash, "active", invoker)
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
