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
	"strings"

	"github.com/ecoball/go-ecoball/http/common"
	innercommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common/config"
)

func GetBlock(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch {
	case len(params) == 1:
		var(
			height uint64
			invalid bool = false
		)
		if v, ok := params[0].(float64); ok {
			height = uint64(v)
		} else {
			invalid = true
		}
		if invalid {
			return common.NewResponse(common.INVALID_PARAMS, "id is invalid")
		}

		blockInfo, errcode := ledger.L.GetTxBlockByHeight(config.ChainHash, height)
		if errcode != nil {
			return common.NewResponse(common.INVALID_PARAMS, "get block faild")
		}
		
		data, errs := blockInfo.Serialize()
		if errs != nil{
			return common.NewResponse(common.INVALID_PARAMS, "Serialize failed")
		}
		return common.NewResponse(common.SUCCESS, innercommon.ToHex(data))
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
}

func GetTransaction(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch {
	case len(params) == 1:
		var(
			id string
			invalid bool = false
		)
		if v, ok := params[0].(string); ok {
			id = v
		} else {
			invalid = true
		}
		if invalid {
			return common.NewResponse(common.INVALID_PARAMS, "id is invalid")
		}

		hash := new(innercommon.Hash)
		TransactionId := hash.FormHexString(id)

		trx, errcode := ledger.L.GetTransaction(config.ChainHash, TransactionId)
		if errcode != nil {
			return common.NewResponse(common.INVALID_PARAMS, "get block faild")
		}
		
		data, errs := trx.Serialize()
		if errs != nil{
			return common.NewResponse(common.INVALID_PARAMS, "Serialize failed")
		}
		return common.NewResponse(common.SUCCESS, innercommon.ToHex(data))
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
}

func Get_account(params []interface{}) *common.Response {
	if len(params) < 1 {
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch params[0].(type) {
	case string:
		chainId := params[0].(string)
		name := params[1].(string)
		hash := new(innercommon.Hash)
		data, err := ledger.L.AccountGet(hash.FormHexString(chainId), innercommon.NameToIndex(name))
		if err != nil {
			return common.NewResponse(common.INTERNAL_ERROR, "AccountGet failed")
		}

		accountInfo, errcode := data.Serialize()
		if errcode != nil {
			return common.NewResponse(common.INTERNAL_ERROR, "Serialize failed")
		}

		return common.NewResponse(common.SUCCESS, innercommon.ToHex(accountInfo))
		default:
			return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	return common.NewResponse(common.SUCCESS, "")
}

func Get_ChainList(params []interface{}) *common.Response {
	chainList, errcode := ledger.L.GetChainList(config.ChainHash)
	if errcode != nil {
		return common.NewResponse(common.INVALID_PARAMS, "get block faild")
	}

	chainList_str := ""
	for _, v := range chainList{
		/*chainList_str += v.Index.String()
		chainList_str += ":"*/
		chainList_str += v.Hash.HexString()
		chainList_str += "\n"
	}
	chainList_str = strings.TrimSuffix(chainList_str, "\n")
	return common.NewResponse(common.SUCCESS, chainList_str)
}
