package simulate

import (
	"math/rand"
	"sync/atomic"
	"time"
)

var counter int32

var r1 int32
var r2 int32
var r3 int32

const (
	TOTAL = 100
)

func isDrop() bool {
	countern := atomic.AddInt32(&counter, 1)
	log.Debug("couter ", countern)
	if countern%TOTAL == 0 {
		getRn()
	}

	if countern%TOTAL == r1 {
		return true
	}

	return false
}

func getRn() {
	rand.Seed(time.Now().UnixNano())
	r1 = rand.Int31n(TOTAL)
	r2 = rand.Int31n(TOTAL)
	r3 = rand.Int31n(TOTAL)
}
