package wasmservice_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
	"testing"
)

func TestApi(t *testing.T) {
	code, err := wasmservice.ReadWasm("../../test/api.wasm")
	if err != nil {
		t.Fatal(err)
	}
	ws := &wasmservice.WasmService{
		Code:   code,
		Method: "invoke",
	}
	ws.RegisterApi()
	fmt.Println(ws.Execute())
}
