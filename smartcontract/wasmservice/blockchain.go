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

//for c api: int32 account_contain(char *s)
func(ws *WasmService)account_contain(proc *exec.Process, s int32)int32 {
	k_msg, err := proc.VMGetData(int(s))
	if err != nil{
		return -1
	}
	account := common.NameToIndex(string(k_msg))
	_, err = ws.state.GetAccountByName(account)
	if err != nil{
		return 0
	}
	return 1
}

