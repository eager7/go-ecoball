package wasmservice

import (
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"github.com/ecoball/go-ecoball/common"
	"fmt"
)

// C API: require_auth(char *account, int32 accountLen)
func (ws *WasmService)require_auth(proc *exec.Process, account, accountLen int32) int32{
	account_msg := make([]byte, accountLen)
	err := proc.ReadAt(account_msg, int(account), int(accountLen))
	if err != nil{
		return -1
	}

	Length := len(account_msg)
	var accountSlice []byte = account_msg[:Length - 1]
	if account_msg[Length - 1] != 0 {
		accountSlice = append(accountSlice, account_msg[Length - 1])
	}

	if ws.action.Permission.Actor == common.NameToIndex(string(accountSlice)) {
		return 0
	}

	fmt.Printf("%s has not %s active permission\n", ws.action.Permission.Actor.String(), accountSlice)
	proc.Terminate()

	return -1
}
