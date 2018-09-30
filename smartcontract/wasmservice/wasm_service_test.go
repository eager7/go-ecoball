package wasmservice_test

import (
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
	"testing"
	"io/ioutil"
)

func TestApi(t *testing.T) {
	data, _ := ioutil.ReadFile("../../test/api.wasm")

	paras := make([]wasmservice.ParamTV,2)
	paras[0] = wasmservice.ParamTV{Ptype:"int32",Pval:"007"}
	paras[1] = wasmservice.ParamTV{Ptype:"string",Pval:"IronMan"}

	arg := wasmservice.Param{
		Arg:  paras,
		Addrs: make([]int64,2),
	}

	ws := wasmservice.WasmService{
		Code:      data,
		Args:      arg,
		Method:    "init",
	}
	ws.RegisterApi()
	ws.Execute()
}
