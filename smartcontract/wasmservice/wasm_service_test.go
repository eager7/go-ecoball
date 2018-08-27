package wasmservice_test

import (
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
	"testing"
	"io/ioutil"
)

func TestApi(t *testing.T) {
	data, _ := ioutil.ReadFile("../../test/api.wasm")
	arg := wasmservice.Param{
//		Arg: 	[]byte{}
//		Count:   2
//		Addrs:
	}

	ws := wasmservice.WasmService{
		Code:      data,
		Args:      arg,
		Method:    "test",
	}
	ws.RegisterApi()
	ws.Execute()
}
