// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package wasmservice

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"github.com/ecoball/go-ecoball/vm/wasmvm/util"
	"github.com/ecoball/go-ecoball/vm/wasmvm/validate"
	"github.com/ecoball/go-ecoball/vm/wasmvm/wasm"
	"io/ioutil"
	"os"
)

var log = elog.NewLogger("wasm", config.LogLevel)

type WasmService struct {
	state     state.InterfaceState
	tx        *types.Transaction
	Code      []byte
	Args      []uint64
	Method    string
	timeStamp int64
}

func NewWasmService(s state.InterfaceState, tx *types.Transaction, contract *types.DeployInfo, invoke *types.InvokeInfo, timeStamp int64) (*WasmService, error) {
	if contract == nil {
		return nil, errors.New("contract is nil")
	}

	params, err := ParseArguments(invoke.Param)
	if err != nil {
		return nil, err
	}
	ws := &WasmService{
		state:     s,
		tx:        tx,
		Code:      contract.Code,
		Args:      params,
		Method:    string(invoke.Method),
		timeStamp: timeStamp,
	}
	ws.RegisterApi()
	return ws, nil
}

func ParseArguments(param []string) ([]uint64, error) {
	var args []uint64
	for _, v := range param {
		arg, err := common.StringToPointer(v)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return args, nil
}

func ReadWasm(file string) ([]byte, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return raw, nil
}

func (ws *WasmService) Execute() ([]byte, error) {
	bf := bytes.NewBuffer(ws.Code)
	m, err := wasm.ReadModule(bf, importer)
	if err != nil {
		log.Error("could not read module:", err)
		return nil, err
	}

	if m.Export == nil {
		log.Warn("module has no export section")
	}

	vm, err := exec.NewVM(m)
	if err != nil {
		fmt.Printf("could not create VM: %v", err)
	}
	entry, ok := m.Export.Entries[ws.Method]

	if ok == false {
		fmt.Printf("method does not exist!")
	}
	index := int64(entry.Index)
	fIdx := m.Function.Types[int(index)]
	fType := m.Types.Entries[int(fIdx)]

	res, err := vm.ExecCode(index, ws.Args...)
	if err != nil {
		fmt.Printf("err=%v", err)
	}
	fmt.Printf("res:%[1]v (%[1]T)\n", res)
	switch fType.ReturnTypes[0] {
	case wasm.ValueTypeI32:
		return util.Int32ToBytes(res.(uint32)), nil
	case wasm.ValueTypeI64:
		return util.Int64ToBytes(res.(uint64)), nil
	case wasm.ValueTypeF32:
		return util.Float32ToBytes(res.(float32)), nil
	case wasm.ValueTypeF64:
		return util.Float64ToBytes(res.(float64)), nil
	default:
		return nil, errors.New("unknown return type")
	}
}

func importer(name string) (*wasm.Module, error) {
	f, err := os.Open(name + ".wasm")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, err := wasm.ReadModule(f, nil)
	if err != nil {
		return nil, err
	}
	err = validate.VerifyModule(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}


func (ws *WasmService) RegisterApi() {
	functions := wasm.InitNativeFuns()
	//console
	functions.Register("prints", ws.prints)
	functions.Register("prints_l", ws.prints_l)
	functions.Register("printi", ws.printi)
	functions.Register("printui", ws.printui)
	functions.Register("printsf", ws.printsf)
	functions.Register("printdf", ws.printdf)
	//memory
	functions.Register("strlen", ws.strlen)
	functions.Register("strcmp", ws.strcmp)
	functions.Register("memcpy", ws.memcpy)
	functions.Register("memset", ws.memset)
	//crypto
	functions.Register("sha256", ws.sha256)
	functions.Register("sha2512", ws.sha512)
}
