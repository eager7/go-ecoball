package wasmservice
import(

	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"

)

func (ws *WasmService) block_GetTime(proc *exec.Process) int64 {

	return ws.timeStamp
}