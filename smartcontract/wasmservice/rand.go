package wasmservice

import (
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"math/rand"
)

/*
param : scale - the max value of rand number
param : seed  - the seed of rand function
*/
func(ws *WasmService)rand(proc *exec.Process,scale ,seed int32)int32{
	var rand_num int
	rand_obj := rand.New(rand.NewSource(int64(seed)))
	if seed == 0 {
		rand_obj.Seed(1)
	}

	if scale == 0 {
		rand_num = rand_obj.Int()
	} else {
		rand_num = rand_obj.Intn(int(scale))
	}

	return int32(rand_num)
}
