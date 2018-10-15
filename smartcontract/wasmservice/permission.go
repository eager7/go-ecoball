package wasmservice

import (
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"github.com/ecoball/go-ecoball/common"
	"bytes"
	"fmt"
)

// C API: inline_action(char *account, char *action, int32 actionData)
func (ws *WasmService)require_auth(proc *exec.Process, account int32) int32{
	data := proc.LoadAt(int(account))
	length := bytes.IndexByte(data,0)
	account_msg := data[:length]

	accountLen := len(account_msg)
	var contractSlice []byte = account_msg[:accountLen - 1]
	if account_msg[accountLen - 1] != 0 {
		contractSlice = append(contractSlice, account_msg[accountLen - 1])
	}

	if ws.action.Permission.Actor == common.NameToIndex(string(contractSlice)) {
		return 0
	}

	fmt.Printf("%s has not %s's active permission\n", ws.action.Permission.Actor.String(), contractSlice)
	proc.Terminate()

	return -1
}
