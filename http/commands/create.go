package commands

import (
	"fmt"
	"strings"

	innercommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/http/common"
	//"github.com/ecoball/go-ecoball/core/store"
	//"github.com/ecoball/go-ecoball/core/ledgerimpl/transaction"
	//"github.com/ecoball/go-ecoball/core/ledgerimpl/Ledger"
	"encoding/json"
	//"github.com/ecoball/go-ecoball/spectator/notify"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
)

func CreateAccount(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	//switch {
	//case len(params) == 4:
		if errCode := handleCreateAccount(params); errCode != common.SUCCESS {
			log.Error(errCode.Info())
			return common.NewResponse(errCode, nil)
		} else {
			return common.NewResponse(common.SUCCESS, nil)
		}

	//default:
		//return common.NewResponse(common.INVALID_PARAMS, nil)
	//}

	return common.NewResponse(common.SUCCESS, "")
}

func Getinfo(params []interface{}) *common.Response {
	var height uint64 = 1

	blockInfo, errcode := ledger.L.GetTxBlockByHeight(config.ChainHash, height)
	if errcode != nil {
		return common.NewResponse(common.INVALID_PARAMS, "get block faild")
	}
	
	data, errs := blockInfo.Serialize()
	if errs != nil{
		return common.NewResponse(common.INVALID_PARAMS, "Serialize failed")
	}
	return common.NewResponse(common.SUCCESS, innercommon.ToHex(data))
}

func Get_required_keys(params []interface{}) *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}

	switch params[0].(type){
	case string:
		//list account
		chainId := params[0].(string)
		required_keys := params[1].(string)
		permission := params[2].(string)
		transaction_data := params[3].(string)

		key_datas := strings.Split(required_keys, "\n")
		Transaction := new(types.Transaction)
		if err := Transaction.Deserialize(innercommon.FromHex(transaction_data)); err != nil {
			return common.NewResponse(common.INTERNAL_ERROR, err)
		}
		//signTransaction, err := wallet.SignTransaction(inner.FromHex(transaction_data), datas)
		hash := new(innercommon.Hash)
		chainids := hash.FormHexString(chainId)
		data, err := ledger.L.FindPermission(chainids, Transaction.From, permission)
		if err != nil{
			return common.NewResponse(common.INTERNAL_ERROR, err)
		}

		permission_datas := []state.Permission{}
		if err := json.Unmarshal([]byte(data), &permission_datas); err != nil {
			return common.NewResponse(common.INTERNAL_ERROR, err)
		}

		public_address := []innercommon.Address{}
		for _, v := range permission_datas {
			for _, value:= range v.Keys{
				public_address = append(public_address, value.Actor)
			}
		}

		publickeys := ""
		for _, v := range key_datas {
			addr := innercommon.AddressFromPubKey(innercommon.FromHex(v))
			for _, vv := range public_address {
				if addr == vv {
					publickeys += v
					publickeys += "\n"
					break
				}
			}
		}
		if "" != publickeys {
			publickeys = strings.TrimSuffix(publickeys, "\n")
			return common.NewResponse(common.SUCCESS, publickeys)
		}
		return common.NewResponse(common.SUCCESS, nil)
	default:
		return common.NewResponse(common.INVALID_PARAMS, nil)
	}
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

func handleCreateAccount(params []interface{}) common.Errcode {
	invoke := new(types.Transaction)//{
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
		return common.INVALID_PARAMS
	}

	if err := invoke.Deserialize(innercommon.FromHex(name)); err != nil {
		fmt.Println(err)
		return common.INVALID_PARAMS
	}

	//send to txpool
	err := event.Send(event.ActorNil, event.ActorTxPool, invoke)
	if nil != err {
		return common.INTERNAL_ERROR
	}

	return common.SUCCESS
}
