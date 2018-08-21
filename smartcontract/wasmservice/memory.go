package wasmservice

import(

	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"bytes"
)

func (ws *WasmService) malloc(proc *exec.Process, p uint32) uint32 {

	return 0
}

//C API: uint32 strlen(char *p)
func (ws *WasmService) strlen(proc *exec.Process, p uint32) uint32 {
	mem := proc.LoadAt(p)
	index := bytes.IndexByte(mem,0)
	return uint32(index)

}

//C API: int32 strcmp(char *s1, char *s2)
func (ws *WasmService) strcmp(proc *exec.Process, p1 , p2 uint32) uint32 {
	mem1 := proc.LoadAt(p1)
	index1 := bytes.IndexByte(mem1,0)
	para1 := mem1[:index1]

	mem2 := proc.LoadAt(p2)
	index2 := bytes.IndexByte(mem2,0)
	para2 := mem2[:index2]

	ret := bytes.Equal(para1,para2)
	if ret {
		return 0
	}
	return 1

}

// C API:void memcpy(void *dest, const void *src, uint32 n)
func (ws *WasmService) memcpy(proc *exec.Process, p1, p2 uint32, len uint32) uint32{

	mem1 := proc.LoadAt(p1)
	mem2 := proc.LoadAt(p2)
    copy(mem1, mem2[:len])
    return 0
}

// C API:void memset(void *s, char ch, uint32 n)
func (ws *WasmService) memset(proc *exec.Process, p uint32, c uint32, len uint32) uint32{

	mem := proc.LoadAt(p)
	for i:= 0; i < int(len); i++{
		mem[i] = byte(c)
	}
	return 0
}
