package wasmservice

import(
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
)

//for c api: int db_put(char* key, uint32 k_len, char *value, uint32 v_len)
func(ws *WasmService)db_put(proc *exec.Process, key, k_len, value, v_len int32)int32{
	k_msg := make([]byte, k_len)
	v_msg := make([]byte, v_len)
	err := proc.ReadAt(k_msg, int(key), int(k_len))
	if err != nil{
		return -1
	}
	err = proc.ReadAt(v_msg, int(value), int(v_len))
	if err != nil{
		return -1
	}
	err = ws.state.StoreSet(ws.action.ContractAccount,k_msg,v_msg)
	if err != nil{
		return -1
	}
	return 0
}

//for c api: int db_get(char* key, uint32 k_len, char *value, uint32 v_len)
func(ws *WasmService)db_get(proc *exec.Process, key, k_len, value, v_len int32)int32{
	k_msg := make([]byte, k_len)
	err := proc.ReadAt(k_msg, int(key), int(k_len))
	if err != nil{
		return -1
	}
	v_msg,err := ws.state.StoreGet(ws.action.ContractAccount,k_msg)
	if err != nil{
		return -1
	}
	if int(v_len) > len(v_msg){
		   v_len = int32(len(v_msg))
	}
	err = proc.WriteAt(v_msg[:], int(value), int(v_len))
	if err != nil{
		return -1
	}
	return 0
}
