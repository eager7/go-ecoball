package wasmservice

import (
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/http/common/abi"
)

// C API: ABA_inline_action(char *account, int32 accountLen, char *action, int32 actionLen, char *actionData, int32 actionDataLen, char *actor, actorLen, char *perm, int32 permLen)
func (ws *WasmService)inline_action(proc *exec.Process, contract, contractLen, action, actionLen, actionData, actionDataLen, actor, actorLen, perm, permLen int32) int32{
	contract_msg := make([]byte, contractLen)
	err := proc.ReadAt(contract_msg, int(contract), int(contractLen))
	if err != nil{
		return -1
	}

	action_msg := make([]byte, actionLen)
	err = proc.ReadAt(action_msg, int(action), int(actionLen))
	if err != nil{
		return -1
	}

	actionData_msg := make([]byte, actionDataLen)
	err = proc.ReadAt(actionData_msg, int(actionData), int(actionDataLen))
	if err != nil{
		return -1
	}

	actor_msg := make([]byte, actorLen)
	err = proc.ReadAt(actor_msg, int(actor), int(actorLen))
	if err != nil{
		return -1
	}

	perm_msg := make([]byte, permLen)
	err = proc.ReadAt(perm_msg, int(perm), int(permLen))
	if err != nil{
		return -1
	}


	if(len(contract_msg) == 0 || len(action_msg) == 0 || len(actionData_msg) == 0 || len(actor_msg) == 0 || len(perm_msg) == 0) {
		fmt.Println("error, can not read param")
		return -1
	}

	fmt.Println("wasm inline action ", string(contract_msg), " ", string(action_msg), " ", string(actionData_msg), " ", string(actor_msg), "", string(perm_msg))

	// C string end with '\0', but Go not. So delete '\0'
	Length := len(contract_msg)
	var contractSlice []byte = contract_msg[:Length - 1]
	if contract_msg[Length - 1] != 0 {
		contractSlice = append(contractSlice, contract_msg[Length - 1])
	}

	Length = len(action_msg)
	var actionSlice []byte = action_msg[:Length - 1]
	if action_msg[Length - 1] != 0 {
		actionSlice = append(actionSlice, action_msg[Length - 1])
	}

	Length = len(actionData_msg)
	var dataSlice []byte = actionData_msg[:Length - 1]
	if actionData_msg[Length - 1] != 0 {
		dataSlice = append(dataSlice, actionData_msg[Length - 1])
	}

	Length = len(actor_msg)
	var actorSlice []byte = actor_msg[:Length - 1]
	if actor_msg[Length - 1] != 0 {
		actorSlice = append(actorSlice, actor_msg[Length - 1])
	}

	Length = len(perm_msg)
	var permSlice []byte = perm_msg[:Length - 1]
	if perm_msg[Length - 1] != 0 {
		permSlice = append(permSlice, perm_msg[Length - 1])
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

