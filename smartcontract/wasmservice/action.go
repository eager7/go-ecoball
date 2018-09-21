package wasmservice

import (
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/http/common/abi"
)

// C API: inline_action(char *account, char *action, int32 actionData)
func (ws *WasmService)inline_action(proc *exec.Process, account, action, actionData int32) int32{
	//fmt.Println("wasm inline action")
	contract_msg, err := proc.VMGetData(int(account))
	if err != nil{
		return -1
	}
	action_msg, err := proc.VMGetData(int(action))
	if err != nil{
		return -1
	}
	data, err := proc.VMGetData(int(actionData))
	if err != nil{
		return -1
	}
	if(len(contract_msg) == 0 || len(action_msg) == 0 || len(data) == 0){
		return -1
	}

	fmt.Println("wasm inline action ", contract_msg, " ", action_msg, " ", data)

	contractLen := len(contract_msg)
	var contractSlice []byte = contract_msg[:contractLen - 1]
	if contract_msg[contractLen - 1] != 0 {
		contractSlice = append(contractSlice, contract_msg[contractLen - 1])
	}

	actionLen := len(action_msg)
	var actionSlice []byte = action_msg[:actionLen - 1]
	if action_msg[actionLen - 1] != 0 {
		actionSlice = append(actionSlice, action_msg[actionLen - 1])
	}

	dataLen := len(data)
	var dataSlice []byte = data[:dataLen - 1]
	if data[dataLen - 1] != 0 {
		dataSlice = append(dataSlice, data[dataLen - 1])
	}

	contractGet, err := ws.state.GetContract(common.NameToIndex(string(contractSlice)))
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

	argbyte, err := abi.CheckParam(abiDef, string(actionSlice), dataSlice)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return -4
	}

	var issueParameters []string

	issueParameters = append(issueParameters, string(argbyte[:]))

	invoke := &types.InvokeInfo{Method: actionSlice, Param: issueParameters}

	actionNew, _ := types.NewSimpleAction(string(contractSlice), invoke)

	ws.context.InlineAction = append(ws.context.InlineAction, *actionNew)

	return 0
}
