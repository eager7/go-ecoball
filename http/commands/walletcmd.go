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

import(
	"math/big"
	"strings"
	"time"
	"errors"

	"github.com/ecoball/go-ecoball/http/common"
	inner "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/wallet"
)

func CreateWallet(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}
	switch params[0].(type){
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
	switch params[0].(type){
	case string:
		name := params[0].(string)
		password := params[1].(string)
		pub, pri, err := wallet.CreateKey(name, []byte(password))
		var key_str string
		key_str += "publickey:" + inner.ToHex(pub) + "\n"+ "privatekey:" + inner.ToHex(pri)
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
		password := params[1].(string)
		//lock wallet
		err := wallet.Lock(name, []byte(password))
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
		password := params[1].(string)
		privateKey := params[2].(string)
		//publickey, err := account.Wallet.ImportKey([]byte(password), inner.FromHex(privateKey))
		publickey, err := wallet.ImportKey(name, []byte(password), privateKey)
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

func Sign_transaction(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch {
	case len(params) == 3:
		if err := handleTransaction(params); err != nil {
			return common.NewResponse(common.INTERNAL_ERROR, err)
		}

	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
}

func handleTransaction(params []interface{}) error {
	var (
		from    string
		to      string
		value   *big.Int
		invalid bool = false
	)

	if v, ok := params[0].(string); ok {
		from = v
	} else {
		invalid = true
	}

	if v, ok := params[1].(string); ok {
		to = v
	} else {
		invalid = true
	}

	if v, ok := params[2].(float64); ok {
		value = big.NewInt(int64(v))
	} else {
		invalid = true
	}

	if invalid {
		return errors.New("params is invalid")
	}

	//time
	time := time.Now().Unix()

	transaction, err := types.NewTransfer(inner.NameToIndex(from), inner.NameToIndex(to), "owner", value, 0, time)
	if nil != err {
		return err
	}

	for name := range wallet.Wallets {
		for publickey := range wallet.Wallets[name].AccountsMap{

			data := transaction.Hash.Bytes()
			signed,_ := secp256k1.Sign(data, inner.FromHex(publickey))
			if hasSign, err := secp256k1.Verify(data, signed, inner.FromHex(publickey)); nil != err || !hasSign {
				log.Warn("check transaction signatures failed:" + transaction.Hash.HexString())
				return errors.New("check transaction signatures fail:" + transaction.Hash.HexString())
			}
		}
	}

	return nil
}
