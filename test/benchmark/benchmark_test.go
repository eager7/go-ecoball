package benchmark

import (
	"testing"
	"runtime"
)

func TestCmd(t *testing.T) {
	runtime.GOMAXPROCS(4)
	//SendTransaction("root", "root", Shard1)
	BenchMarkTransaction()
}
