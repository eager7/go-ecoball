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

//for c api: int is_account(char *name, uint32 len)
func(ws *WasmService)is_account(proc *exec.Process, p, length uint32)int32 {
	var name []byte
	err := proc.ReadAt(name,int(p), int(length))
	if err != nil{
		return -1
	}
	account := common.NameToIndex(string(name))
	_, err = ws.state.GetAccountByName(account)
	if err != nil{
		return 0
	}
	return 1
}

