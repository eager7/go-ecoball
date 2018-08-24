package wasmservice

/*
import "encoding/binary"


import(

	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"encoding/binary"
)

//for c api: void read_params(void)
func(ws *WasmService)read_action_data(proc *exec.Process) uint32{
	legnth := len(ws.args.arg)
	for i := 0; i < legnth; i++{

	}
	addr, err := proc.VMSetBlock(ws.args.arg)  //将参数写进vm的内存，参数以TLV格式存储
	if err != nil{
		return 0
	}
	ws.args.index = addr
	return 0
}


func(ws *WasmService)read_param_i(proc *exec.Process, p int32) uint32{
	addr := ws.args.addrs[1]
	mem, err := proc.VMGetData(addr)
	if err != nil{
		return 0
	}
	return binary.LittleEndian.Uint32(mem)
}

//for c api: int32 *read_param_s(char *s)
func(ws *WasmService)read_param_s(proc *exec.Process, p int32) uint32{
	mem, err := proc.VMGetData(uint64(p))
	if err != nil{
		return 0
	}
	return binary.LittleEndian.Uint32(mem)
}
*/