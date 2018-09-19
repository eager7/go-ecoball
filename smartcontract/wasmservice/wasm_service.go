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
	"strings"
	"encoding/json"
	"strconv"
	"github.com/ecoball/go-ecoball/smartcontract/context"
)

var log = elog.NewLogger("wasm", config.LogLevel)

type Param  struct{
	Arg     []ParamTV     //参数数据
	Addrs   []int64       //参数地址
}

type ParamTV struct {
	Ptype string `json:"type"`
	Pval  string `json:"value"`
}

type WasmService struct {
	state     state.InterfaceState
	action	  *types.Action
	context   *context.ApplyContext
	Code      []byte
	Args      Param
	Method    string
	timeStamp int64
}

func NewWasmService(s state.InterfaceState, tx *types.Transaction, action *types.Action, context *context.ApplyContext, contract *types.DeployInfo, invoke *types.InvokeInfo, timeStamp int64) (*WasmService, error) {
	if contract == nil {
		return nil, errors.New("contract is nil")
	}

	stringByte := strings.Join(invoke.Param, "\x20\x00") // x20 = space and x00 = null

	var args []ParamTV
	err1 := json.Unmarshal([]byte(stringByte), &args)
	if err1 != nil {
		return nil, errors.New("json.Unmarshal failed")
	}
	log.Debug("NewWasmService ", args)

    num := len(args)
    var param = Param{
    	Arg:	args,
    	Addrs:	make([]int64,num),
	}

	ws := &WasmService{
		state:     s,
		action:    action,
		context:	context,
		Code:      contract.Code,
		Args:      param,
		Method:    string(invoke.Method),
		timeStamp: timeStamp,
	}
	ws.RegisterApi()
	return ws, nil
}

func (ws *WasmService) ParseParam(vm *exec.VM)([]uint64, error){

	method, err := vm.Memmanage.SetBlock(ws.Method)
	if err != nil{
		return nil,err
	}
	param := make([]uint64,1)
	param[0] = uint64(method)
	var(
		addr     int
		v_string string
		v_int64  int64
		v_uint64 uint64
	)
	pcount := len(ws.Args.Arg)
	for index := 0;index < pcount; index++{
		switch ws.Args.Arg[index].Ptype{
		case "string":
			v_string = ws.Args.Arg[index].Pval
		    addr, err = vm.Memmanage.SetBlock(v_string)
		    if err != nil{
		    	return nil, errors.New("para error")
			}
		    ws.Args.Addrs[index] = int64(addr)
		case "int8","int16","int32","int64":
			v_string = ws.Args.Arg[index].Pval
			v_int64, _ = strconv.ParseInt(v_string, 10, 64)
			ws.Args.Addrs[index] = v_int64
		case "uint8","uint16","uint32":
			v_string = ws.Args.Arg[index].Pval
			v_uint64, _ = strconv.ParseUint(v_string, 10, 64)
			ws.Args.Addrs[index] = int64(v_uint64)
		default:
			return nil, errors.New("unsupport type")

		}

	}
	return param,nil
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
	//blockchain
	functions.Register("ABA_account_contain",ws.account_contain)
	functions.Register("ABA_block_GetTime",ws.block_GetTime)
	//memory
	functions.Register("ABA_malloc", ws.malloc)
	functions.Register("ABA_len",    ws.len)
	functions.Register("ABA_strlen", ws.strlen)
	functions.Register("ABA_strcmp", ws.strcmp)
	functions.Register("ABA_memcpy", ws.memcpy)
	functions.Register("ABA_memset", ws.memset)
	//crypto
	functions.Register("ABA_sha256", ws.sha256)
	functions.Register("ABA_sha512", ws.sha512)
	//runtime
	functions.Register("ABA_read_param", ws.read_param)
	//db
	functions.Register("ABA_db_put",ws.db_put)
	functions.Register("ABA_db_get",ws.db_get)
	//encode
	functions.Register("ABA_atoi",  ws.atoi)
	functions.Register("ABA_atoi64",ws.atoi64)
	functions.Register("ABA_itoa",  ws.itoa)
	functions.Register("ABA_i64toa",ws.i64toa)
	//rand
	functions.Register("ABA_rand",ws.rand)
	//inline action
	functions.Register("inline_action", ws.inline_action)

}
