package wasmservice

import(
	"strconv"
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"github.com/ecoball/go-ecoball/vm/wasmvm/util"
)


//for c api: char *itoa(int d)
func(ws *WasmService)itoa(proc *exec.Process, data int64)int32{
	str := strconv.FormatInt(data,10)
	addr,err := proc.VMSetBlock(str)
	if err != nil{
		return -1
	}
	return int32(addr)
}

//for c api: char *i64toa(int d)
func(ws *WasmService)i64toa(proc *exec.Process, data int64)int32{
	str := strconv.FormatInt(data,10)
	addr,err := proc.VMSetBlock(str)
	if err != nil{
		return -1
	}
	return int32(addr)
}

//for c api: int64 atoi(char *s)
func(ws *WasmService)atoi(proc *exec.Process, p int32)int32{
	data, err := proc.VMGetData(int(p))
	if err != nil{
		return -1
	}
	str := util.TrimBuffToString(data)
	out, err := strconv.ParseInt(str, 10, 32)
	if err != nil{
		return -1
	}
	return int32(out)
}

//for c api: int64 atoi64(char *s)
func(ws *WasmService)atoi64(proc *exec.Process, p int32)int64{
	data, err := proc.VMGetData(int(p))
	if err != nil{
		return -1
	}
	str := util.TrimBuffToString(data)
	out, err := strconv.ParseInt(str, 10, 64)
	if err != nil{
		return -1
	}
	return out
}
