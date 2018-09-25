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

	if ws.context.Tc.Trx.From == common.NameToIndex(string(account_msg)) {
		return 0
	}



	return 0
}
