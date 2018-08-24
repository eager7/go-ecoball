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
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"github.com/ecoball/go-ecoball/vm/wasmvm/util"
	"github.com/ecoball/go-ecoball/vm/wasmvm/validate"
	"github.com/ecoball/go-ecoball/vm/wasmvm/wasm"
	"os"
)

var log = elog.NewLogger("wasm", config.LogLevel)


//TLV格式存储
type Param  struct{
	Arg     []byte  //参数数据
	Count   int     //参数个数
	Addrs   []int   //参数地址
}

type WasmService struct {
	state     state.InterfaceState
	tx        *types.Transaction
	Code      []byte
	Args      Param
	Method    string
	timeStamp int64
}

func NewWasmService(s state.InterfaceState, tx *types.Transaction, contract *types.DeployInfo, invoke *types.InvokeInfo, timeStamp int64) (*WasmService, error) {
	if contract == nil {
		return nil, errors.New("contract is nil")
	}

	ws := &WasmService{
		state:     s,
		tx:        tx,
		Code:      contract.Code,
		Args:      Param{},
		Method:    string(invoke.Method),
		timeStamp: timeStamp,
	}
	ws.RegisterApi()
	return ws, nil
}

const (
	ParaInt32      byte = 0XFF
	ParaString     byte = 0XFE
)

//TLV
func (ws *WasmService) ParseParam(vm *exec.VM)([]uint64, error){

	addr, err := vm.Memmanage.SetBlock(ws.Method)
	if err != nil{
		return nil,err
	}
	paras := make([]uint64,1)
	paras[0] = uint64(addr)

    var length int
    var index  int
	pcount := len(ws.Args.Arg)
	ws.Args.Count = 0
	for index = 0;index < pcount; {
		switch ws.Args.Arg[index]{
		case ParaInt32,ParaString:
			length = int(ws.Args.Arg[index+1])
			data := ws.Args.Arg[index+2:index+length+2]
		    addr, err := vm.Memmanage.SetBlock(data)
		    if err != nil{
		    	return nil, errors.New("para error")
			}
		    ws.Args.Addrs[ws.Args.Count] = addr
		    ws.Args.Count += 1
		    index = index + 2 + length
		default:
			return nil, errors.New("unsupport type")

		}

	}

	return paras,nil
}

func (ws *WasmService) Execute() ([]byte, error) {
	bf := bytes.NewBuffer(ws.Code)
	method := "apply"

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
		log.Error("could not create VM: %v", err)
		return nil, err
	}

	entry, ok := m.Export.Entries[method]
	if ok == false {
		log.Error("method does not exist!")
		return nil, err
	}

	args, err:= ws.ParseParam(vm)

	if err != nil{
		log.Error("parse parameter error!")
		return nil, err
	}
	index := int64(entry.Index)
	fIdx := m.Function.Types[int(index)]
	fType := m.Types.Entries[int(fIdx)]

	res, err := vm.ExecCode(index, args...)
	if err != nil {
		log.Error("err=%v", err)
	}
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
	functions.Register("ABA_prints", ws.prints)
	functions.Register("ABA_prints_l", ws.prints_l)
	functions.Register("ABA_printi", ws.printi)
	functions.Register("ABA_printui", ws.printui)
	functions.Register("ABA_printsf", ws.printsf)
	functions.Register("ABA_printdf", ws.printdf)
	//memory
	functions.Register("ABA_malloc", ws.malloc)
	functions.Register("ABA_strlen", ws.strlen)
	functions.Register("ABA_strcmp", ws.strcmp)
	functions.Register("ABA_memcpy", ws.memcpy)
	functions.Register("ABA_memset", ws.memset)
	//crypto
	functions.Register("ABA_sha256", ws.sha256)
	functions.Register("ABA_sha512", ws.sha512)
}
