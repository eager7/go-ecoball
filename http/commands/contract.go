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
	"fmt"
	"github.com/ecoball/go-ecoball/http/common/abi"
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
	}else if "pledge" == contractMethod || "reg_prod" == contractMethod || "vote" == contractMethod {
		parameters = strings.Split(contractParam, ",")
	}else if "set_account" == contractMethod {
		parameters = strings.Split(contractParam, "--")
	}else if "reg_chain" == contractMethod {
		parameter := strings.Split(contractParam, ",")
		if len(parameter) == 3{
			parameters = append(parameters, parameter[0])
			parameters = append(parameters, parameter[1])
			parameters = append(parameters, innerCommon.AddressFromPubKey(innerCommon.FromHex(parameter[2])).HexString())
		}else {
			return common.INVALID_PARAMS
		}
	}else {
		contract, err := ledger.L.GetContract(config.ChainHash, innerCommon.NameToIndex(contractName))

		var abiDef abi.ABI
		err = abi.UnmarshalBinary(contract.Abi, &abiDef)
		if err != nil {
			fmt.Errorf("can not find UnmarshalBinary abi file")
			return common.INVALID_CONTRACT_ABI
		}
	
		log.Debug("contractParam: ", contractParam)
		argbyte, err := abi.CheckParam(abiDef, contractMethod, []byte(contractParam))
		if err != nil {
			log.Debug("checkParam error")
			return common.INVALID_PARAMS
		}
	
		parameters = append(parameters, string(argbyte[:]))

		abi.GetContractTable(contractName, "root", abiDef, "accounts")
	}

	//from address
	//from := account.AddressFromPubKey(common.Account.PublicKey)

	//contract address
	//address := innerCommon.NewAddress(innerCommon.CopyBytes(innerCommon.FromHex(contractAddress)))

	//time
	time := time.Now().UnixNano()

	transaction, err := types.NewInvokeContract(innerCommon.NameToIndex("root"), innerCommon.NameToIndex(contractName), config.ChainHash, "owner", contractMethod, parameters, 0, time)
	if nil != err {
		return common.TRX_FAIL
	}

	// Just for test, must delete later
	err = transaction.SetSignature(&config.Root)
	if err != nil {
		return common.INVALID_ACCOUNT
	}
	transaction.Show()

	//send to txpool
	err = event.Send(event.ActorNil, event.ActorTxPool, transaction)
	if nil != err {
		return common.INTERNAL_ERROR
	}

	return common.SUCCESS
}
