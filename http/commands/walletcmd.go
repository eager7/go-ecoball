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
	//"math/big"
	"strings"
	//"time"
	//"errors"

	inner "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/http/common"
	//"github.com/ecoball/go-ecoball/crypto/secp256k1"
	//"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/walletserver/wallet"
)

func CreateWallet(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}
	switch params[0].(type) {
	case string:
		name := params[0].(string)
		password := params[1].(string)
		if err := wallet.Create(name, []byte(password)); nil != err {
			return common.NewResponse(common.INTERNAL_ERROR, err.Error())
		}
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}
	return common.NewResponse(common.SUCCESS, "")
}

func CreateKey(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}
	switch params[0].(type) {
	case string:
		name := params[0].(string)
		pub, pri, err := wallet.CreateKey(name)
		var key_str string
		key_str += "publickey:" + inner.ToHex(pub) + "\n" + "privatekey:" + inner.ToHex(pri)
		if err != nil {
			return common.NewResponse(common.INTERNAL_ERROR, err.Error())
		}
		return common.NewResponse(common.SUCCESS, key_str)
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}
}

func OpenWallet(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch params[0].(type){
	case string:
		name := params[0].(string)
		password := params[1].(string)
		err := wallet.Open(name, []byte(password))
		if err != nil {
			return common.NewResponse(common.INTERNAL_ERROR, err.Error())
		}
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
}

func LockWallet(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch params[0].(type){
	case string:
		name := params[0].(string)
		//lock wallet
		err := wallet.Lock(name)
		if nil != err {
			return common.NewResponse(common.INTERNAL_ERROR, err.Error())
		}
		//account.Cipherkeys = cipherkeysTemp
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
}

func UnlockWallet(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch params[0].(type){
	case string:
		name := params[0].(string)
		password := params[1].(string)
		if err := wallet.Unlock(name, []byte(password)); nil != err {
			return common.NewResponse(common.INTERNAL_ERROR, err.Error())
		}

	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
}

func ImportKey(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch params[0].(type){
	case string:
		name := params[0].(string)
		privateKey := params[1].(string)
		//publickey, err := account.Wallet.ImportKey([]byte(password), inner.FromHex(privateKey))
		publickey, err := wallet.ImportKey(name, privateKey)
		var key_str string
		key_str += "publickey:" + inner.ToHex(publickey)
		if err != nil {
			return common.NewResponse(common.INTERNAL_ERROR, err.Error())
		}
		return common.NewResponse(common.SUCCESS, key_str)
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}
}

func RemoveKey(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch params[0].(type){
	case string:
		name := params[0].(string)
		password := params[1].(string)
		publickey := params[2].(string)
		//publickey, err := account.Wallet.ImportKey([]byte(password), inner.FromHex(privateKey))
		err := wallet.RemoveKey(name, []byte(password), publickey)
		if err != nil {
			return common.NewResponse(common.INTERNAL_ERROR, err.Error())
		}
		return common.NewResponse(common.SUCCESS, nil)
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}
}

func ListKeys(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch params[0].(type){
	case string:
		//list account
		name := params[0].(string)
		password := params[1].(string)
		accounts, err := wallet.ListKeys(name, []byte(password))
		var key_str string
		for k, v := range accounts {
			key_str += "publickey:" + k + "\n"+ "privatekey:" + v
			key_str += "\n"
		}
		key_str = strings.TrimSuffix(key_str, "\n")
		if err != nil {
			return common.NewResponse(common.INTERNAL_ERROR, err.Error())
		}
		return common.NewResponse(common.SUCCESS, key_str)
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}
}

func ListWallets(params []interface{}) *common.Response {
	wallets, err := wallet.List_wallets()
	if err != nil {
		return common.NewResponse(common.INTERNAL_ERROR, err.Error())
	}
	var walletsList string
	for _, k := range wallets {
		walletsList += k
		walletsList += "\n"
	}
	walletsList = strings.TrimSuffix(walletsList, "\n")
	return common.NewResponse(common.SUCCESS, walletsList)
}

func GetPublicKeys(params []interface{}) *common.Response {
	data, err := wallet.GetPublicKeys()
	if err != nil {
		return common.NewResponse(common.INTERNAL_ERROR, err.Error())
	}

	var publickeys string
	for _, k := range data {
		publickeys += k
		publickeys += "\n"
	}
	publickeys = strings.TrimSuffix(publickeys, "\n")

	return common.NewResponse(common.SUCCESS, publickeys)
}
