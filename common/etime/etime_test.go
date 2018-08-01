package etime_test

import (
	"testing"
	"fmt"
	"github.com/ecoball/go-ecoball/common/etime"
	"time"
)

func TestNow(t *testing.T) {
	for ; ;  {
		fmt.Println(etime.Microsecond(), "ms")
		time.Sleep(time.Millisecond * 100)
	}
}
