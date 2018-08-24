package wasmservice

import(
	"fmt"
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
)

//C API :void prints(char *p)
func (ws *WasmService) prints(proc *exec.Process, p int32) int32{
	msg,err := proc.VMGetData(int(p))
	if err != nil{
		return 0
	}
	fmt.Printf("%s\n",msg)
	return 0
}

//C API :void prints_l(char *p, uint32 len)
func (ws *WasmService) prints_l(proc *exec.Process, p int32, length int32) int32{
	msg,err := proc.VMGetData(int(p))
	if err != nil{
		return 0
	}
	msglen := len(msg)
	if length > int32(msglen){
		length = int32(msglen)
	}
	fmt.Printf("%s\n",msg[:length])
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