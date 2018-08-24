package wasmservice
import(

	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"

)

func(ws *WasmService)block_GetTime(proc *exec.Process)int64{
	return ws.timeStamp
}

//for c api: uint32 get_active_producer(char *s)
func(ws *WasmService)get_active_producer(proc *exec.Process)uint32 {

	return 0
}