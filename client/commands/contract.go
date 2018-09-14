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

	"encoding/json"

	clientCommon "github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/http/common/abi"
	"github.com/urfave/cli"
	"strings"
	"strconv"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
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

	//file data
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

	chainId := info.ChainID
	chainIdStr := c.String("chainId")
	if "config.hash" != chainIdStr && "" != chainIdStr {
		chainId = common.HexToHash(chainIdStr)
	}

	publickeys, err := GetPublicKeys()
	if err != nil {
		fmt.Println(err)
		return err
	}

	time := time.Now().UnixNano()
	transaction, err := types.NewDeployContract(common.NameToIndex("root"), common.NameToIndex(contractName), chainId, "owner", types.VmWasm, description, data, abibyte, 0, time)
	if nil != err {
		return err
	}

	required_keys, err := get_required_keys(info.ChainID, publickeys, permission, transaction)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if required_keys == "" {
		fmt.Println("no required_keys")
		return err
	}

	errcode := sign_transaction(info.ChainID, required_keys, transaction)
	if nil != errcode {
		fmt.Println(errcode)
	}

	datas, err := transaction.Serialize()
	if err != nil {
		fmt.Println(err)
		return err
	}

	//rpc call
	//resp, err := rpc.NodeCall("setContract", []interface{}{common.ToHex(data), contractName, description, abi_str})
	resp, err := rpc.NodeCall("setContract", []interface{}{common.ToHex(datas)})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	//result
	return rpc.EchoResult(resp)
}

func checkParam(abiDef abi.ABI, method string, arg []byte) ([]byte, error){
	var f interface{}

	if err := json.Unmarshal(arg, &f); err != nil {
		return nil, err
	}

	m := f.(map[string]interface{})

	var fields []abi.FieldDef
	for _, action := range abiDef.Actions {
		// first: find method
		if string(action.Name) == method {
			//fmt.Println("find ", method)
			for _, struction := range abiDef.Structs {
				// second: find struct
				if struction.Name == action.Type {
					fields = struction.Fields
				}
			}
			break
		}
	}

	if fields == nil {
		return nil, errors.New("can not find method " + method)
	}

	args := make([]wasmservice.ParamTV, len(fields))
	for i, field := range fields {
		v := m[field.Name]
		if v != nil {
			args[i].Ptype = field.Type

			switch vv := v.(type) {
			case string:
				//	if field.Type == "string" || field.Type == "account_name" || field.Type == "asset" {
				//		args[i].Pval = vv
				//	} else {
				//		return nil, errors.New(log, fmt.Sprintln("can't match abi struct field type ", field.Type))
				//	}
				//	fmt.Println(field.Name, "is ", field.Type, "", vv)
				//case float64:
				switch field.Type {
				case "string","account_name","asset":
					args[i].Pval = vv
				case "int8":
					const INT8_MAX = int8(^uint8(0) >> 1)
					const INT8_MIN = ^INT8_MAX
					a, err := strconv.ParseInt(vv, 10, 8)
					if err != nil {
						return nil, errors.New(fmt.Sprintln(vv, "is out of int8 range"))
					}
					if a >= int64(INT8_MIN) && a <= int64(INT8_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(fmt.Sprintln(vv, "is out of int8 range"))
					}
				case "int16":
					const INT16_MAX = int16(^uint16(0) >> 1)
					const INT16_MIN = ^INT16_MAX
					a, err := strconv.ParseInt(vv, 10, 16)
					if err != nil {
						return nil, errors.New(fmt.Sprintln(vv, "is out of int16 range"))
					}
					if a >= int64(INT16_MIN) && a <= int64(INT16_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(fmt.Sprintln(vv, "is out of int16 range"))
					}
				case "int32":
					const INT32_MAX = int32(^uint32(0) >> 1)
					const INT32_MIN = ^INT32_MAX
					a, err := strconv.ParseInt(vv, 10, 32)
					if err != nil {
						return nil, errors.New(fmt.Sprintln(vv, "is out of int32 range"))
					}
					if a >= int64(INT32_MIN) && a <= int64(INT32_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(fmt.Sprintln(vv, "is out of int32 range"))
					}
				case "int64":
					const INT64_MAX = int64(^uint64(0) >> 1)
					const INT64_MIN = ^INT64_MAX
					a, err := strconv.ParseInt(vv, 10, 64)
					if err != nil {
						return nil, errors.New(fmt.Sprintln(vv, "is out of int64 range"))
					}
					if a >= INT64_MIN && a <= INT64_MAX {
						args[i].Pval = vv
					} else {
						return nil, errors.New(fmt.Sprintln(vv, "is out of int64 range"))
					}

				case "uint8":
					const UINT8_MIN uint8 = 0
					const UINT8_MAX = ^uint8(0)
					a, err := strconv.ParseUint(vv, 10, 8)
					if err != nil {
						return nil, errors.New(fmt.Sprintln(vv, "is out of uint8 range"))
					}
					if a >= uint64(UINT8_MIN) && a <= uint64(UINT8_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(fmt.Sprintln(vv, "is out of uint8 range"))
					}
				case "uint16":
					const UINT16_MIN uint16 = 0
					const UINT16_MAX = ^uint16(0)
					a, err := strconv.ParseUint(vv, 10, 16)
					if err != nil {
						return nil, errors.New(fmt.Sprintln(vv, "is out of uint16 range"))
					}
					if a >= uint64(UINT16_MIN) && a <= uint64(UINT16_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(fmt.Sprintln(vv, "is out of uint16 range"))
					}
				case "uint32":
					const UINT32_MIN uint32 = 0
					const UINT32_MAX = ^uint32(0)
					a, err := strconv.ParseUint(vv, 10, 32)
					if err != nil {
						return nil, errors.New(fmt.Sprintln(vv, "is out of uint32 range"))
					}
					if a >= uint64(UINT32_MIN) && a <= uint64(UINT32_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(fmt.Sprintln(vv, "is out of uint32 range"))
					}
				case "uint64":
					const UINT64_MIN uint64 = 0
					const UINT64_MAX = ^uint64(0)
					a, err := strconv.ParseUint(vv, 10, 64)
					if err != nil {
						return nil, errors.New(fmt.Sprintln(vv, "is out of uint64 range"))
					}
					if a >= UINT64_MIN && a <= UINT64_MAX {
						args[i].Pval = vv
					} else {
						return nil, errors.New(fmt.Sprintln(vv, "is out of uint64 range"))
					}

				default:
					return nil, errors.New(fmt.Sprintln("can't match abi struct field type ", field.Type))
				}
				//
				//if field.Type == "int8" || field.Type == "int16" || field.Type == "int32" {
				//	args[i].Pval = strconv.FormatInt(int64(vv), 10)
				//} else if field.Type == "uint8" || field.Type == "uint16" || field.Type == "uint32" {
				//	args[i].Pval = strconv.FormatUint(uint64(vv), 10)
				//} else {
				//	return nil, errors.New(log, fmt.Sprintln("can't match abi struct field type ", field.Type))
				//}
				fmt.Println(field.Name, "is ", field.Type, "", vv)
				//case []interface{}:
				//	fmt.Println(field.Name, "is an array:")
				//	for i, u := range vv {
				//		fmt.Println(i, u)
				//	}
			default:
				return nil, errors.New(fmt.Sprintln("can't match abi struct field type: ", v))
			}
		} else {
			return nil, errors.New("can't match abi struct field name:  " + field.Name)
		}

	}

	bs, err := json.Marshal(args)
	if err != nil {
		return nil, errors.New("json.Marshal failed")
	}
	return bs, nil
}

func GetContract(chainID common.Hash, index common.AccountName) (*types.DeployInfo, error){
	resp, errcode := rpc.NodeCall("GetContract", []interface{}{chainID.HexString(), index.String()})
	if errcode != nil {
		fmt.Fprintln(os.Stderr, errcode)
		return nil, errcode
	}

	deploy := new(types.DeployInfo)
	if nil != resp["result"] {
		switch resp["result"].(type) {
		case string:
			data := resp["result"].(string)
			if err := deploy.Deserialize(common.FromHex(data)); err != nil{
				return nil, err
			}
			return deploy, nil
		default:
		}
	}
	return deploy, nil
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

	info, err := getInfo()
	if err != nil {
		fmt.Println(err)
		return err
	}

	publickeys, err := GetPublicKeys()
	if err != nil {
		fmt.Println(err)
		return err
	}

	if "new_account" == contractMethod {
		parameter := strings.Split(contractParam, ",")
		for _, v := range parameter {
			if strings.Contains(v, "0x") {
				parameters = append(parameters, common.AddressFromPubKey(common.FromHex(v)).HexString())
			}else {
				parameters = append(parameters, v)
			}
		}
	}else if "pledge" == contractMethod || "reg_prod" == contractMethod || "vote" == contractMethod {
		parameters = strings.Split(contractParam, ",")
	}else if "set_account" == contractMethod {
		parameters = strings.Split(contractParam, "--")
	}else if "reg_chain" == contractMethod {
		parameter := strings.Split(contractParam, ",")
		if len(parameter) == 3{
			parameters = append(parameters, parameter[0])
			parameters = append(parameters, parameter[1])
			parameters = append(parameters, common.AddressFromPubKey(common.FromHex(parameter[2])).HexString())
		}else {
			return errors.New("Invalid parameters")
		}
	}else {
		contract, err := GetContract(info.ChainID, common.NameToIndex(contractName))

		var abiDef abi.ABI
		err = abi.UnmarshalBinary(contract.Abi, &abiDef)
		if err != nil {
			return errors.New("can not find UnmarshalBinary abi file")
		}
	
		//log.Debug("contractParam: ", contractParam)
		argbyte, err := checkParam(abiDef, contractMethod, []byte(contractParam))
		if err != nil {
			return errors.New("checkParam error")
		}
	
		parameters = append(parameters, string(argbyte[:]))

		//getContractTable(contractName, "root", abiDef, "stat")
	}

	//from address
	//from := account.AddressFromPubKey(common.Account.PublicKey)

	//contract address
	//address := innerCommon.NewAddress(innerCommon.CopyBytes(innerCommon.FromHex(contractAddress)))

	//time
	time := time.Now().UnixNano()

	transaction, err := types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex(contractName), info.ChainID, "owner", contractMethod, parameters, 0, time)
	if nil != err {
		fmt.Println(err)
		return err
	}

	required_keys, err := get_required_keys(info.ChainID, publickeys, "active", transaction)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if required_keys == "" {
		fmt.Println("no required_keys")
		return err
	}

	errcode := sign_transaction(info.ChainID, required_keys, transaction)
	if nil != errcode {
		fmt.Println(errcode)
	}
	
	data, err := transaction.Serialize()
	if err != nil {
		fmt.Println(err)
		return err
	}
	//rpc call
	resp, err := rpc.NodeCall("invokeContract", []interface{}{common.ToHex(data)})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	//result
	return rpc.EchoResult(resp)
}
