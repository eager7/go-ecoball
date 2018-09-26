package wasmservice

import (
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"github.com/ecoball/go-ecoball/common"
)

// C API: inline_action(char *account, char *action, int32 actionData)
func (ws *WasmService)require_auth(proc *exec.Process, account int32) int32{
	account_msg, err := proc.VMGetData(int(account))
	if err != nil{
		return -1
	}

	accountLen := len(account_msg)
	var contractSlice []byte = account_msg[:accountLen - 1]
	if account_msg[accountLen - 1] != 0 {
		contractSlice = append(contractSlice, account_msg[accountLen - 1])
	}

	if ws.action.Permission.Actor == common.NameToIndex(string(account_msg)) {
		return 0
	}

	return -1
}
