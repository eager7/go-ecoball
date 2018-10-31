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
	"net/url"
	"os"
	"time"

	"encoding/json"

	"strings"

	clientCommon "github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/http/common/abi"
	"github.com/urfave/cli"
	//innerCommon "github.com/ecoball/go-ecoball/http/common"
	"github.com/ecoball/go-ecoball/common/config"
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
						Usage: "contract name",
					},
					cli.StringFlag{
						Name:  "description, d",
						Usage: "contract description",
					},
					cli.StringFlag{
						Name:  "abipath, ap",
						Usage: "abi file path",
					},
					cli.StringFlag{
						Name:  "permission, per",
						Usage: "active permission",
						Value: "active",
					},
					cli.StringFlag{
						Name:  "chainId, c",
						Usage: "chainId hash",
						Value: "config.hash",
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
						Name:  "sender, s",
						Usage: "sender name",
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
		fmt.Println("Invalid file path: ", fileName)
		return errors.New("Invalid contrace file path")
	}

	//abi file path
	abi_fileName := c.String("abipath")
	if abi_fileName == "" {
		fmt.Println("Invalid abifile path: ", fileName)
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
		fmt.Println("Invalid contract name: ", contractName)
		return errors.New("Invalid contract name")
	}

	//contract description
	description := c.String("description")
	if description == "" {
		fmt.Println("Invalid contract description: ", description)
		return errors.New("Invalid contract description")
	}

	permission := c.String("permission")
	if "" == permission {
		permission = "active"
	}

	info, err := getInfo()
	if err != nil {
		fmt.Println(err)
		return err
	}

	chainHash, err := getMainChainHash()
	if nil != err {
		fmt.Println(err)
		return err
	}

	allPublickeys, err := getPublicKeys()
	if err != nil {
		fmt.Println(err)
		return err
	}

	time := time.Now().UnixNano()
	transaction, err := types.NewDeployContract(common.NameToIndex(contractName), common.NameToIndex(contractName), chainHash, "owner", types.VmWasm, description, data, abibyte, 0, time)
	if nil != err {
		fmt.Println(err)
		return err
	}

	requiredKeys, err := getRequiredKeys(chainHash, permission, contractName)
	if err != nil {
		fmt.Println(err)
		return err
	}

	publickeys := ""
	keyDatas := strings.Split(allPublickeys, ",")
	for _, v := range keyDatas {
		addr := common.AddressFromPubKey(common.FromHex(v))
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

	datas, errcode := signTransaction(chainHash, publickeys, transaction)
	if nil != errcode {
		fmt.Println(errcode)
	}

	//rpc call
	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("transaction", datas)
	err = rpc.NodePost("/invokeContract", values.Encode(), &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
}

func GetContract(chainID common.Hash, index common.AccountName) (*types.DeployInfo, error) {
	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("contractName", index.String())
	values.Set("chainId", chainID.HexString())
	err := rpc.NodePost("/getContract", values.Encode(), &result)
	if nil == err {
		deploy := new(types.DeployInfo)
		if err := deploy.Deserialize(common.FromHex(result.Result)); err != nil {
			return nil, err
		}
		return deploy, nil
	}
	return nil, err
}

func StoreGet(chainID common.Hash, index common.AccountName, key []byte) (value []byte, err error) {
	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("contractName", index.String())
	values.Set("chainId", chainID.HexString())
	values.Set("key", common.ToHex(key))
	err = rpc.NodePost("/storeGet", values.Encode(), &result)
	if nil == err {
		return common.FromHex(result.Result), nil
	}
	return nil, err
}

func GetContractTable(contractName string, accountName string, abiDef abi.ABI, tableName string) ([]byte, error) {
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
		return nil, errors.New("can not find struct of table  " + tableName)
	}

	table := make(map[string]string, len(fields))

	for i, _ := range fields {
		key := []byte(fields[i].Name)
		if fields[i].Name == "balance" { // only for token contract, because KV struct can't support
			key = []byte(accountName)
		} else {
			key = append(key, 0) // C lang string end with 0
		}

		storage, err := StoreGet(config.ChainHash, common.NameToIndex(contractName), key)
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

func invokeContract(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//contract address
	contractName := c.String("name")
	if contractName == "" {
		fmt.Println("Invalid contract name: ", contractName)
		return errors.New("Invalid contract name")
	}

	//contract name
	contractMethod := c.String("method")
	if contractMethod == "" {
		fmt.Println("Invalid contract method: ", contractMethod)
		return errors.New("Invalid contract method")
	}

	//contract parameter
	contractParam := c.String("param")

	var parameters []string

	chainHash, err := getMainChainHash()
	if nil != err {
		fmt.Println(err)
		return err
	}

	allPublickeys, err := getPublicKeys()
	if err != nil {
		fmt.Println(err)
		return err
	}

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
		contract, err := GetContract(chainHash, common.NameToIndex(contractName))
		if err != nil {
			return errors.New("GetContract failed")
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
		GetContractTable(contractName, "root", abiDef, "Account")
	}

	//contract address
	sender := c.String("sender")
	if sender == "" {
		sender = contractName
	}

	//time
	time := time.Now().UnixNano()

	transaction, err := types.NewInvokeContract(common.NameToIndex(sender), common.NameToIndex(contractName), chainHash, "owner", contractMethod, parameters, 0, time)
	if nil != err {
		fmt.Println(err)
		return err
	}

	requiredKeys, err := getRequiredKeys(chainHash, "active", sender)
	if err != nil {
		fmt.Println(err)
		return err
	}

	publickeys := ""
	keyDatas := strings.Split(allPublickeys, ",")
	for _, v := range keyDatas {
		addr := common.AddressFromPubKey(common.FromHex(v))
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

	data, errcode := signTransaction(chainHash, publickeys, transaction)
	if nil != errcode {
		fmt.Println(errcode)
	}

	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("transaction", data)
	err = rpc.NodePost("/invokeContract", values.Encode(), &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
}
