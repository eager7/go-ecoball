package wasmservice

import(
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
)

//for c api: int db_store(char* key, char *value)
func(ws *WasmService)db_store(proc *exec.Process, key int32, value int32 )int32{
	k_msg, err := proc.VMGetData(int(key))
	if err != nil{
		return -1
	}
	v_msg, err := proc.VMGetData(int(value))
	if err != nil{
		return -1
	}
	ws.state.StoreSet(ws.addr,k_msg,v_msg)
	return 0
}

//for c api:char *db_get(char *key)
func(ws *WasmService)db_get(proc *exec.Process, key int32)int32{
	k_msg, err := proc.VMGetData(int(key))
	if err != nil{
		return -1
	}
	value,err := ws.state.StoreGet(ws.addr,k_msg)
	if err != nil{
		return -1
	}
	addr,err := proc.VMSetBlock(value)
	if err != nil{
		return -1
	}
	return int32(addr)
}

//for c api:
func(ws *WasmService)db_update(proc *exec.Process)int32{
	return 0
}

//for c api:
func(ws *WasmService)db_remove(proc *exec.Process)int32{
	return 0
}
//for c api:
func(ws *WasmService)db_find(proc *exec.Process)int32{
	return 0
}