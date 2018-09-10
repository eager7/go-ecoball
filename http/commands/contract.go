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
	"time"

	"github.com/ecoball/go-ecoball/core/types"

	"github.com/ecoball/go-ecoball/account"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
	"github.com/ecoball/go-ecoball/http/common"
	"github.com/ecoball/go-ecoball/common/config"
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
	"github.com/ecoball/go-ecoball/http/common/abi"
	"github.com/ecoball/go-ecoball/common/errors"
	"strconv"
	//"encoding/hex"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"strings"
)

var log = elog.NewLogger("commands", elog.DebugLog)

func SetContract(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch {
	case len(params) == 1:
		if errCode, result := handleSetContract(params); errCode != common.SUCCESS {
			log.Error(errCode.Info())
			return common.NewResponse(errCode, nil)
		} else {
			return common.NewResponse(common.SUCCESS, result)
		}

	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
}

func handleSetContract(params []interface{}) (common.Errcode, string) {

	//Get account address
	/*var (
		code         []byte
		contractName string
		description  string
		invalid      bool = false
		abicode      []byte
	)

	if v, ok := params[0].(string); ok {
		code = innerCommon.FromHex(v)
	} else {
		invalid = true
	}

	if v, ok := params[1].(string); ok {
		contractName = v
	} else {
		invalid = true
	}

	if v, ok := params[2].(string); ok {
		description = v
	} else {
		invalid = true
	}

	if v, ok := params[3].(string); ok {
		abitmp, err := hex.DecodeString(v)
		if err == nil {
			abicode = abitmp
		} else {
			invalid = true
		}
	} else {
		invalid = true
	}

	if invalid {
		return common.INVALID_PARAMS, ""
	}*/

	//time
	//time := time.Now().Unix()

	//generate key pair
	keyData, err := secp256k1.NewECDSAPrivateKey()
	if err != nil {
		return common.GENERATE_KEY_PAIR_FAILED, ""
	}

	public, err := secp256k1.FromECDSAPub(&keyData.PublicKey)
	if err != nil {
		return common.GENERATE_KEY_PAIR_FAILED, ""
	}

	//generate address
	address := account.AddressFromPubKey(public)

	//from address
	//from := account.AddressFromPubKey(common.Account.PublicKey)

	transaction := new(types.Transaction)//{
	//	Payload: &types.InvokeInfo{}}
	
	var invalid bool
	var name string 
	//invoke.Show()
	//account name
	if v, ok := params[0].(string); ok {
		name = v
	} else {
		invalid = true
	}
	
	if invalid {
		return common.INVALID_PARAMS, ""
	}
	
	if err := transaction.Deserialize(innerCommon.FromHex(name)); err != nil {
		fmt.Println(err)
		return common.INVALID_PARAMS, ""
	}
	//send to txpool
	err = event.Send(event.ActorNil, event.ActorTxPool, transaction)
	if nil != err {
		return common.INTERNAL_ERROR, ""
	}

	return common.SUCCESS, address.HexString()
}

func InvokeContract(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch {
	case len(params) == 3:
		if errCode := handleInvokeContract(params); errCode != common.SUCCESS {
			log.Error(errCode.Info())
			return common.NewResponse(errCode, nil)
		}

	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
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
		return nil, errors.New(log, "can not find method " + method)
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
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int8 range"))
					}
					if a >= int64(INT8_MIN) && a <= int64(INT8_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int8 range"))
					}
				case "int16":
					const INT16_MAX = int16(^uint16(0) >> 1)
					const INT16_MIN = ^INT16_MAX
					a, err := strconv.ParseInt(vv, 10, 16)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int16 range"))
					}
					if a >= int64(INT16_MIN) && a <= int64(INT16_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int16 range"))
					}
				case "int32":
					const INT32_MAX = int32(^uint32(0) >> 1)
					const INT32_MIN = ^INT32_MAX
					a, err := strconv.ParseInt(vv, 10, 32)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int32 range"))
					}
					if a >= int64(INT32_MIN) && a <= int64(INT32_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int32 range"))
					}
				case "int64":
					const INT64_MAX = int64(^uint64(0) >> 1)
					const INT64_MIN = ^INT64_MAX
					a, err := strconv.ParseInt(vv, 10, 64)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int64 range"))
					}
					if a >= INT64_MIN && a <= INT64_MAX {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int64 range"))
					}

				case "uint8":
					const UINT8_MIN uint8 = 0
					const UINT8_MAX = ^uint8(0)
					a, err := strconv.ParseUint(vv, 10, 8)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint8 range"))
					}
					if a >= uint64(UINT8_MIN) && a <= uint64(UINT8_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint8 range"))
					}
				case "uint16":
					const UINT16_MIN uint16 = 0
					const UINT16_MAX = ^uint16(0)
					a, err := strconv.ParseUint(vv, 10, 16)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint16 range"))
					}
					if a >= uint64(UINT16_MIN) && a <= uint64(UINT16_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint16 range"))
					}
				case "uint32":
					const UINT32_MIN uint32 = 0
					const UINT32_MAX = ^uint32(0)
					a, err := strconv.ParseUint(vv, 10, 32)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint32 range"))
					}
					if a >= uint64(UINT32_MIN) && a <= uint64(UINT32_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint32 range"))
					}
				case "uint64":
					const UINT64_MIN uint64 = 0
					const UINT64_MAX = ^uint64(0)
					a, err := strconv.ParseUint(vv, 10, 64)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint64 range"))
					}
					if a >= UINT64_MIN && a <= UINT64_MAX {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint64 range"))
					}

				default:
					return nil, errors.New(log, fmt.Sprintln("can't match abi struct field type ", field.Type))
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
				return nil, errors.New(log, fmt.Sprintln("can't match abi struct field type: ", v))
			}
		} else {
			return nil, errors.New(log, "can't match abi struct field name:  " + field.Name)
		}

	}

	bs, err := json.Marshal(args)
	if err != nil {
		return nil, errors.New(log, "json.Marshal failed")
	}
	return bs, nil
}

//func checkParam(abiDef abi.ABI, method string, arg []byte) ([]byte, error){
//	var f interface{}
//
//	if err := json.Unmarshal(arg, &f); err != nil {
//		return nil, err
//	}
//
//	m := f.(map[string]interface{})
//
//	var fields []abi.FieldDef
//	for _, action := range abiDef.Actions {
//		// first: find method
//		if string(action.Name) == method {
//			//fmt.Println("find ", method)
//			for _, struction := range abiDef.Structs {
//				// second: find struct
//				if struction.Name == action.Type {
//					fields = struction.Fields
//				}
//			}
//			break
//		}
//	}
//
//	if fields == nil {
//		return nil, errors.New(log, "can not find method " + method)
//	}
//
//	args := make([]wasmservice.ParamTV, len(fields))
//	for i, field := range fields {
//		v := m[field.Name]
//		if v != nil {
//			args[i].Ptype = field.Type
//
//			switch vv := v.(type) {
//			case string:
//				if field.Type == "string" || field.Type == "account_name" || field.Type == "asset" {
//					args[i].Pval = vv
//				} else {
//					return nil, errors.New(log, fmt.Sprintln("can't match abi struct field type ", field.Type))
//				}
//				fmt.Println(field.Name, "is ", field.Type, "", vv)
//			case float64:
//				switch field.Type {
//				case "int8":
//					const INT8_MAX = int8(^uint8(0) >> 1)
//					const INT8_MIN = ^INT8_MAX
//					if int64(vv) >= int64(INT8_MIN) && int64(vv) <= int64(INT8_MAX) {
//						args[i].Pval = strconv.FormatInt(int64(vv), 10)
//					} else {
//						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int8 range"))
//					}
//				case "int16":
//					const INT16_MAX = int16(^uint16(0) >> 1)
//					const INT16_MIN = ^INT16_MAX
//					if int64(vv) >= int64(INT16_MIN) && int64(vv) <= int64(INT16_MAX) {
//						args[i].Pval = strconv.FormatInt(int64(vv), 10)
//					} else {
//						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int16 range"))
//					}
//				case "int32":
//					const INT32_MAX = int32(^uint32(0) >> 1)
//					const INT32_MIN = ^INT32_MAX
//					if int64(vv) >= int64(INT32_MIN) && int64(vv) <= int64(INT32_MAX) {
//						args[i].Pval = strconv.FormatInt(int64(vv), 10)
//					} else {
//						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int32 range"))
//					}
//				case "int64":
//					args[i].Pval = strconv.FormatInt(int64(vv), 10)
//
//				case "uint8":
//					const UINT8_MIN uint8 = 0
//					const UINT8_MAX = ^uint8(0)
//					if uint64(vv) >= uint64(UINT8_MIN) && uint64(vv) <= uint64(UINT8_MAX) {
//						args[i].Pval = strconv.FormatUint(uint64(vv), 10)
//					} else {
//						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint8 range"))
//					}
//				case "uint16":
//					const UINT16_MIN uint16 = 0
//					const UINT16_MAX = ^uint16(0)
//					if uint64(vv) >= uint64(UINT16_MIN) && uint64(vv) <= uint64(UINT16_MAX) {
//						args[i].Pval = strconv.FormatUint(uint64(vv), 10)
//					} else {
//						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint16 range"))
//					}
//				case "uint32":
//					const UINT32_MIN uint32 = 0
//					const UINT32_MAX = ^uint32(0)
//					if uint64(vv) >= uint64(UINT32_MIN) && uint64(vv) <= uint64(UINT32_MAX) {
//						args[i].Pval = strconv.FormatUint(uint64(vv), 10)
//					} else {
//						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint32 range"))
//					}
//				case "uint64":
//					args[i].Pval = strconv.FormatUint(uint64(vv), 10)
//
//				default:
//					return nil, errors.New(log, fmt.Sprintln("can't match abi struct field type ", field.Type))
//				}
//				//
//				//if field.Type == "int8" || field.Type == "int16" || field.Type == "int32" {
//				//	args[i].Pval = strconv.FormatInt(int64(vv), 10)
//				//} else if field.Type == "uint8" || field.Type == "uint16" || field.Type == "uint32" {
//				//	args[i].Pval = strconv.FormatUint(uint64(vv), 10)
//				//} else {
//				//	return nil, errors.New(log, fmt.Sprintln("can't match abi struct field type ", field.Type))
//				//}
//				fmt.Println(field.Name, "is ", field.Type, "", vv)
//				//case []interface{}:
//				//	fmt.Println(field.Name, "is an array:")
//				//	for i, u := range vv {
//				//		fmt.Println(i, u)
//				//	}
//			default:
//				return nil, errors.New(log, fmt.Sprintln("can't match abi struct field type: %T", v))
//			}
//		} else {
//			return nil, errors.New(log, "can't match abi struct field name:  " + field.Name)
//		}
//
//	}
//
//	bs, err := json.Marshal(args)
//	if err != nil {
//		return nil, errors.New(log, "json.Marshal failed")
//	}
//	return bs, nil
//}

func handleInvokeContract(params []interface{}) common.Errcode {
	var (
		contractName   string
		contractMethod string
		contractParam  string
		parameters     []string
		invalid        bool = false
	)

	if v, ok := params[0].(string); ok {
		contractName = v
	} else {
		invalid = true
	}

	if v, ok := params[1].(string); ok {
		contractMethod = v
	} else {
		invalid = true
	}

	if v, ok := params[2].(string); ok {
		contractParam = v
	} else {
		invalid = true
	}

	if invalid {
		log.Debug("Param error")
		return common.INVALID_PARAMS
	}

	//if "" != contractParam {
	//	parameters = strings.Split(contractParam, " ")
	//}

	//args, err := ParseParams(contractParam)
	//if err != nil {
	//	return common.INVALID_PARAMS
	//}
	//
	//data, err := json.Marshal(args)
	//if err != nil {
	//	return common.INVALID_PARAMS
	//}
	//log.Debug("ParseParams: ", string(data))
	//
	//argbyte, err := BuildWasmContractParam(args)
	//if err != nil {
	//	//t.Errorf("build wasm contract param failed:%s", err)
	//	//return
	//	return common.INVALID_PARAMS
	//}
	//log.Debug("BuildWasmContractParam: ", string(argbyte))

	if "new_account" == contractMethod {
		parameter := strings.Split(contractParam, ",")
		for _, v := range parameter {
			if strings.Contains(v, "0x") {
				parameters = append(parameters, innerCommon.AddressFromPubKey(innerCommon.FromHex(v)).HexString())
			}else {
				parameters = append(parameters, v)
			}
		}
	}else if "pledge" == contractMethod {
		parameters = strings.Split(contractParam, ",")
	}else if "set_account" == contractMethod {
		parameters = append(parameters, contractName)
		parameters = append(parameters, contractParam)
	}else {
		contract, err := ledger.L.GetContract(config.ChainHash, innerCommon.NameToIndex(contractName))

		var abiDef abi.ABI
		err = abi.UnmarshalBinary(contract.Abi, &abiDef)
		if err != nil {
			fmt.Errorf("can not find UnmarshalBinary abi file")
			return common.INVALID_CONTRACT_ABI
		}
	
		log.Debug("contractParam: ", contractParam)
		argbyte, err := checkParam(abiDef, contractMethod, []byte(contractParam))
		if err != nil {
			log.Debug("checkParam error")
			return common.INVALID_PARAMS
		}
	
		parameters = append(parameters, string(argbyte[:]))
	}

	//from address
	//from := account.AddressFromPubKey(common.Account.PublicKey)

	//contract address
	//address := innerCommon.NewAddress(innerCommon.CopyBytes(innerCommon.FromHex(contractAddress)))

	//time
	time := time.Now().UnixNano()

	transaction, err := types.NewInvokeContract(innerCommon.NameToIndex("root"), innerCommon.NameToIndex("root"), config.ChainHash, "owner", contractMethod, parameters, 0, time)
	if nil != err {
		return common.TRX_FAIL
	}

	// Just for test, must delete later
	err = transaction.SetSignature(&config.Root)
	if err != nil {
		return common.INVALID_ACCOUNT
	}

	//send to txpool
	err = event.Send(event.ActorNil, event.ActorTxPool, transaction)
	if nil != err {
		return common.INTERNAL_ERROR
	}

	return common.SUCCESS
}
