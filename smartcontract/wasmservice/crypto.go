package wasmservice

import(
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"crypto/sha256"
	// "golang.org/x/crypto/ripemd160"
	"crypto/sha512"
)

//C API :void sha256(char *data, uint32 len, char *hash )
func (ws *WasmService) sha256(proc *exec.Process, data uint32, len uint32, hash uint32) uint32{
	msg := make([]byte, len)
	proc.ReadAt(msg, data, len)
	temp := sha256.Sum256(msg)
	proc.WriteAt(temp[:], hash, 32)
	return 0
}


//C API :void sha512(char *data, uint32 len, char *hash )
func (ws *WasmService) sha512(proc *exec.Process, data uint32, len uint32, hash uint32) uint32{
	msg := make([]byte, len)
	proc.ReadAt(msg, data, len)
	temp := sha512.Sum512(msg)
	proc.WriteAt(temp[:], hash, 64)
	return 0
}


//C API :void ripemd160(char *data, uint32 len, char *hash )
func (ws *WasmService) ripemd160(proc *exec.Process, p uint32) uint32 {

	return 0
}