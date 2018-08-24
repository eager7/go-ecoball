package wasmservice

import(
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"crypto/sha256"
	// "golang.org/x/crypto/ripemd160"
	"crypto/sha512"
)

//C API :int sha256(char *data, uint32 len, char *hash )
func (ws *WasmService) sha256(proc *exec.Process, data int32, length int32, hash int32) int32{
	msg, err := proc.VMGetData(int(data))
	if err != nil{
		return -1
	}
	msglen := len(msg)
	if length > int32(msglen){
		length = int32(msglen)
	}
	temp := sha256.Sum256(msg[:length])
	_, err = proc.VMSetData(int(hash), temp[:])
	if err != nil{
		return -1
	}
	return 0
}

//C API :int sha512(char *data, uint32 len, char *hash )
func (ws *WasmService) sha512(proc *exec.Process, data int32, length int32, hash int32) int32{
	msg, err := proc.VMGetData(int(data))
	if err != nil{
		return -1
	}
	msglen := len(msg)
	if length > int32(msglen){
		length = int32(msglen)
	}
	temp := sha512.Sum512(msg[:length])
	_, err = proc.VMSetData(int(hash), temp[:])
	if err != nil{
		return -1
	}
	return 0
}

//C API :void ripemd160(char *data, uint32 len, char *hash )
func (ws *WasmService) ripemd160(proc *exec.Process, p int32) int32 {

	return 0
}