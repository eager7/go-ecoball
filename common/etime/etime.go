package etime

import (
	"time"
	"github.com/ecoball/go-ecoball/common/config"
)

func Microsecond() int64 {
	return time.Now().UnixNano() / 1000
}

func Millisecond() int64 {
	t := Microsecond()
	n := t / 1000 / int64(config.TimeSlot)
	return int64(uint64(n) * uint64(config.TimeSlot))
}

