package benchmark

import (
	"testing"
	"runtime"
	"sync"
)

func TestCmd(t *testing.T) {
	runtime.GOMAXPROCS(4)
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		SendTransaction("root", "root", Shard3)
	}()
	go func() {
		defer wg.Done()
		SendTransaction("testeru", "testeru", Shard1)
	}()
	go func() {
		defer wg.Done()
		SendTransaction("testerh", "testerh", Shard2)
	}()
	wg.Wait()
}

func BenchmarkSendTransaction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SendTransaction("root", "root", Shard3)
	}
}