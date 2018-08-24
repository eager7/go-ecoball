package wasmservice

import(

	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"github.com/ecoball/go-ecoball/vm/wasmvm/util"
	"bytes"
)

//只分配内存，未做垃圾回收
//C API: char *malloc(int size)
func (ws *WasmService) malloc(proc *exec.Process, p int32) int32 {
	addr, err := proc.VMmalloc(int(p), exec.DString)
	if err != nil{
		return -1
	}
	return int32(addr)
}

//C API: uint32 strlen(char *p)
func (ws *WasmService) strlen(proc *exec.Process, p int32) int32 {
	data := proc.LoadAt(int(p))
	length := bytes.IndexByte(data,0)
	return int32(length)
}

//C API: int32 strcmp(char *s1, char *s2)
func (ws *WasmService) strcmp(proc *exec.Process, p1 , p2 int32) int32 {

	data1, err := proc.VMGetData(int(p1))
	if err != nil{
		return -1
	}

	data2, err := proc.VMGetData(int(p2))
	if err != nil{
		return -1
	}

	if util.TrimBuffToString(data1) == util.TrimBuffToString(data2) {
		return 0
	} else {
		return 1
	}

}

// C API:void memcpy(void *dest, const void *src, uint32 n)
func (ws *WasmService) memcpy(proc *exec.Process, p1, p2 int32, len int32) int32{

	mem1 := proc.LoadAt(int(p1))
	mem2 := proc.LoadAt(int(p2))
    copy(mem1, mem2[:len])
    return 0
}

// C API:void memset(void *s, char ch, uint32 n)
func (ws *WasmService) memset(proc *exec.Process, p int32, c int32, len int32) int32{

	mem := proc.LoadAt(int(p))
	for i:= 0; i < int(len); i++{
		mem[i] = byte(c)
	}
	return 0
}
