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

	str, err := dsn.AddFile(params[3].(string),0)
	if err != nil {
		return common.NewResponse(common.INVALID_PARAMS, "DsnAddFile faild")
	}

	return common.NewResponse(common.SUCCESS, str)
}
