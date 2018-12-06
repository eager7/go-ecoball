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

func BenchmarkSendTransaction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SendTransaction("root", "root", Shard3)
	}
}