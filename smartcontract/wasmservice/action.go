package wasmservice

import (
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/http/common/abi"
	"bytes"
)

// C API: inline_action(char *account, char *action, char *actionData, char *actor, char *perm)
func (ws *WasmService)inline_action(proc *exec.Process, account, action, actionData, actor, perm int32) int32{
	data := proc.LoadAt(int(account))
	length := bytes.IndexByte(data,0)
	contract_msg := data[:length]

	data = proc.LoadAt(int(action))
	length = bytes.IndexByte(data,0)
	action_msg := data[:length]

	data = proc.LoadAt(int(actionData))
	length = bytes.IndexByte(data,0)
	actionData_msg := data[:length]

	data = proc.LoadAt(int(actor))
	length = bytes.IndexByte(data,0)
	actor_msg := data[:length]

	data = proc.LoadAt(int(perm))
	length = bytes.IndexByte(data,0)
	permission := data[:length]


	if(len(contract_msg) == 0 || len(action_msg) == 0 || len(actionData_msg) == 0 || len(permission) == 0 || len(actor_msg) == 0) {
		fmt.Println("error, can not read param")
		return -1
	}

	fmt.Println("wasm inline action ", string(contract_msg), " ", string(action_msg), " ", string(actionData_msg), " ", string(actor_msg), "", string(permission))

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

	dataLen := len(actionData_msg)
	var dataSlice []byte = actionData_msg[:dataLen - 1]
	if actionData_msg[dataLen - 1] != 0 {
		dataSlice = append(dataSlice, actionData_msg[dataLen - 1])
	}

	actorLen := len(actor_msg)
	var actorSlice []byte = actor_msg[:actorLen - 1]
	if actor_msg[actorLen - 1] != 0 {
		actorSlice = append(actorSlice, actor_msg[actorLen - 1])
	}

	permLen := len(permission)
	var permSlice []byte = permission[:permLen - 1]
	if permission[permLen - 1] != 0 {
		permSlice = append(permSlice, permission[permLen - 1])
	}

	if ws.action.ContractAccount != common.NameToIndex(string(contractSlice)) {
		if err := ws.state.CheckAccountPermission(common.NameToIndex(string(actorSlice)), ws.action.ContractAccount, string(permSlice)); err != nil {
			return -5
		}
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

	actionNew, _ := types.NewSimpleAction(string(contractSlice), types.PermissionLevel{common.NameToIndex(string(actorSlice)), string(permSlice)}, invoke)

	ws.context.InlineAction = append(ws.context.InlineAction, *actionNew)

	return 0
}
