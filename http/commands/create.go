package commands

import (
	//"time"

	//innercommon "github.com/ecoball/go-ecoball/common"
	//"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/http/common"
	//"encoding/json"
	//"fmt"
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
	if v, ok := params[0].(string); ok {
		name = v
	} else {
		invalid = true
	}

	//owner key
	if v, ok := params[0].(string); ok {
		owner = v
	} else {
		invalid = true
	}*/

	invoke := &types.Transaction{}
	var invalid bool
	var name string 
	//account name
	if v, ok := params[0].(string); ok {
		name = v
	} else {
		invalid = true
	}

	if invalid {
		return common.INVALID_PARAMS
	}

	//json.Unmarshal([]byte(name), invoke);
	invoke.StringJson(name)
	//str := invoke.JsonString()
	//fmt.Println(str)
	//fmt.Println(name)

	//send to txpool
	err := event.Send(event.ActorNil, event.ActorTxPool, invoke)
	if nil != err {
		return common.INTERNAL_ERROR
	}

	return common.SUCCESS
}
