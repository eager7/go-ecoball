package wasmservice

import (
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/http/common/abi"
)

func (ws *WasmService)inline_action(proc *exec.Process, account, action int32) int32{
	//fmt.Println("wasm inline action")
	contract_msg, err := proc.VMGetData(int(account))
	if err != nil{
		return -1
	}
	action_msg, err := proc.VMGetData(int(action))
	if err != nil{
		return -1
	}
	if(len(contract_msg) == 0 || len(action_msg) == 0){
		return -1
	}

	fmt.Println("wasm inline action ", contract_msg, " ", action_msg)

	contractGet, err := ws.state.GetContract(common.NameToIndex("worker"))
	if err != nil {
		fmt.Errorf("can not find contract abi file")
		return -2
	}

	var abiDef abi.ABI
	err = abi.UnmarshalBinary(contractGet.Abi, &abiDef)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return -3
	}

	trans := []byte(`{"from": "worker1", "to": "worker2", "amount": "25", "token_id": "xyx"}`)

	argbyte, err := abi.CheckParam(abiDef, "transfer", trans)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return -4
	}

	var issueParameters []string

	issueParameters = append(issueParameters, string(argbyte[:]))

	invoke := &types.InvokeInfo{Method: []byte("transfer"), Param: issueParameters}

	actionNew, _ := types.NewSimpleAction("worker", invoke)
	//smartcontract.ApplyExec(ws.context, actionNew)

	ws.context.InlineAction = append(ws.context.InlineAction, *actionNew)

	return 0
}
