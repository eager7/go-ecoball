package wasmservice

import(
	"fmt"
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"

	"bytes"
)

//C API :void prints(char *p)
func (ws *WasmService) prints (proc *exec.Process, p uint32) uint32{
	msg := proc.LoadAt(p)
	index := bytes.IndexByte(msg,0)
	msg = msg[:index]
	fmt.Printf("%s\n",msg)
	return 0
}

//C API :void prints_l(char *p, uint32 len)
func (ws *WasmService) prints_l(proc *exec.Process, p uint32, len uint32) uint32{
	msg := make([]byte, len)
	proc.ReadAt(msg, p, len)
	fmt.Printf("%s\n",msg)
	return 0
}

//C API :void printi(int64 v)
func (ws *WasmService) printi(proc *exec.Process, v int64) uint32{

	fmt.Printf("%d\n",v)
	return 0
}

//C API :void printui(uint64 v)
func (ws *WasmService) printui(proc *exec.Process, v uint64) uint32{

	fmt.Printf("%d\n",v)
	return 0
}

//C API :void printsf(float v)
func (ws *WasmService) printsf(proc *exec.Process, v float32) uint32{

	fmt.Printf("%f\n",v)
	return 0
}

//C API :void printdf(float v)
func (ws *WasmService) printdf(proc *exec.Process, v float64) uint32{

	fmt.Printf("%f\n",v)
	return 0
}