package commands

import (
	// "fmt"
	// "strings"

	// innercommon "github.com/ecoball/go-ecoball/common"
	// "github.com/ecoball/go-ecoball/common/config"
	// "github.com/ecoball/go-ecoball/common/event"
	// "github.com/ecoball/go-ecoball/core/types"
	// "github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/http/common"
	"github.com/ecoball/go-ecoball/dsn"
	"strconv"
	//"github.com/ecoball/go-ecoball/core/store"
	//"github.com/ecoball/go-ecoball/core/ledgerimpl/transaction"
	//"github.com/ecoball/go-ecoball/core/ledgerimpl/Ledger"
	// "encoding/json"
	//"github.com/ecoball/go-ecoball/spectator/notify"
//	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
)

func DsnAddFile(params []interface{})  *common.Response {

	if len(params) < 1 {
		log.Error("invalid arguments")
	}

	//era := params[4].(string)
	ear,ok := params[4].(string)
	i, _ := strconv.Atoi(ear)
	if ok{
		str, err := dsn.AddFile(params[3].(string),int8(i))
		if err != nil {
			return common.NewResponse(common.INVALID_PARAMS, "DsnAddFile faild")
		}
		return common.NewResponse(common.SUCCESS, str)
	}

	return common.NewResponse(common.INVALID_PARAMS, "type not ok")
}

func DsnCatFile(params []interface{})  *common.Response {

	if len(params) < 1 {
		log.Error("invalid arguments")
	}


	byteStr, err := dsn.CatFile(params[3].(string))
	if err != nil {
		return common.NewResponse(common.INVALID_PARAMS, "DsnGetFile faild")
	}
	return common.NewResponse(common.SUCCESS, byteStr)


}