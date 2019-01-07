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
	contractSlice := make([]byte, contractLen)
	err := proc.ReadAt(contractSlice, int(contract), int(contractLen))
	if err != nil{
		return -1
	}

	actionSlice := make([]byte, actionLen)
	err = proc.ReadAt(actionSlice, int(action), int(actionLen))
	if err != nil{
		return -1
	}

	dataSlice := make([]byte, actionDataLen)
	err = proc.ReadAt(dataSlice, int(actionData), int(actionDataLen))
	if err != nil{
		return -1
	}

	actorSlice := make([]byte, actorLen)
	err = proc.ReadAt(actorSlice, int(actor), int(actorLen))
	if err != nil{
		return -1
	}

	permSlice := make([]byte, permLen)
	err = proc.ReadAt(permSlice, int(perm), int(permLen))
	if err != nil{
		return -1
	}


	if(len(contractSlice) == 0 || len(actionSlice) == 0 || len(dataSlice) == 0 || len(actorSlice) == 0 || len(permSlice) == 0) {
		fmt.Println("error, can not read param")
		return -1
	}

	fmt.Println("wasm inline action ", string(contractSlice), " ", string(actionSlice), " ", string(dataSlice), " ", string(actorSlice), "", string(permSlice))

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

	issueParameters, err := abi.CheckParam(abiDef, string(actionSlice), dataSlice)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return -4
	}

	invoke := &types.InvokeInfo{Method: actionSlice, Param: issueParameters}

	actionNew, _ := types.NewSimpleAction(string(contractSlice), types.PermissionLevel{common.NameToIndex(string(actorSlice)), string(permSlice)}, invoke)

	ws.context.InlineAction = append(ws.context.InlineAction, *actionNew)

	return 0
}

