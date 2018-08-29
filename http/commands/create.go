package commands

import (
	"fmt"
	//"time"

	innercommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/http/common"
	//"github.com/ecoball/go-ecoball/core/store"
	//"github.com/ecoball/go-ecoball/core/ledgerimpl/transaction"
	//"github.com/ecoball/go-ecoball/core/ledgerimpl/Ledger"
	//"encoding/json"
	//"fmt"
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

func handleCreateAccount(params []interface{}) common.Errcode {
	/*var (
		creator string
		name    string
		owner   string
		//	active  string
		invalid bool = false
	)

	//creator name
	if v, ok := params[0].(string); ok {
		creator = v
	} else {
		invalid = true
	}

	//account name
	if v, ok := params[1].(string); ok {
		name = v
	} else {
		invalid = true
	}

	//owner key
	if v, ok := params[2].(string); ok {
		owner = v
	} else {
		invalid = true
	}*/

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
