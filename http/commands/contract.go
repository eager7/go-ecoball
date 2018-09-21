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
	"github.com/ecoball/go-ecoball/core/types"

	"github.com/ecoball/go-ecoball/account"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
	"github.com/ecoball/go-ecoball/http/common"
	"fmt"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"time"
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
	case len(params) == 1:
		if errCode, result := handleInvokeContract(params); errCode != common.SUCCESS {
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

func GetContract(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch params[0].(type){
	case string:
		//list account
		chainId := params[0].(string)
		accountName := params[1].(string)
		hash := new(innerCommon.Hash)
		chainids := hash.FormHexString(chainId)
		contract, err := ledger.L.GetContract(chainids, innerCommon.NameToIndex(accountName))
		if err != nil {
			return common.NewResponse(common.INTERNAL_ERROR, err.Error())
		}

		data, err := contract.Serialize()
		if err != nil {
			return common.NewResponse(common.INTERNAL_ERROR, err.Error())
		}
		return common.NewResponse(common.SUCCESS, innerCommon.ToHex(data))
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}
}

func StoreGet(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch params[0].(type){
	case string:
		//list account
		chainId := params[0].(string)
		accountName := params[1].(string)
		key := params[2].(string)
		hash := new(innerCommon.Hash)
		chainids := hash.FormHexString(chainId)
		storage, err := ledger.L.StoreGet(chainids, innerCommon.NameToIndex(accountName), innerCommon.FromHex(key))
		if err != nil {
			return common.NewResponse(common.INTERNAL_ERROR, err.Error())
		}

		fmt.Println(string(storage))
		return common.NewResponse(common.SUCCESS, innerCommon.ToHex(storage))
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}
}

func handleInvokeContract(params []interface{}) (common.Errcode, string) {
	transaction := new(types.Transaction)

	var invalid bool
	var name string

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
	//err := event.Send(event.ActorNil, event.ActorTxPool, transaction)
	res, err := event.SendSync(event.ActorTxPool, transaction, 5*time.Second)
	if nil != err {
		return common.INTERNAL_ERROR, ""
	}

	result := res.(string)
	fmt.Println("result: ", result)

	return common.SUCCESS, result
}
