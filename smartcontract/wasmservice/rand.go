package wasmservice

import (
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"math/rand"
	"time"
)

/*
param : scale - the max value of rand number
param : seed  - the seed of rand function
*/
func(ws *WasmService)rand(proc *exec.Process,scale ,seed int32)int32{
	rnd:=rand.New(rand.NewSource(time.Now().Unix()+int64(seed)))
	rand_num:=rnd.Intn(int(scale))
	return int32(rand_num)
}
