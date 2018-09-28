package wasmservice
import(
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"

)

func(ws *WasmService)block_GetTime(proc *exec.Process)int64{
	return ws.timeStamp
}

//for c api: int32 get_active_producer(char *s)
func(ws *WasmService)get_active_producer(proc *exec.Process)int32 {

	return 0
}

//for c api: int is_account(acount_name name)
func(ws *WasmService)is_account(proc *exec.Process, account uint64)int32 {
	_, err := ws.state.GetAccountByName(common.AccountName(account))
	if err != nil{
		return 0
	}
	return 1
}

