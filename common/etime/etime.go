package etime

import (
	"time"
	"github.com/ecoball/go-ecoball/common/config"
)

func Now() int64 {
	t := time.Now().UnixNano()
	n := t / 1000000 / int64(config.TimeSlot)
	return int64(uint64(n) * uint64(config.TimeSlot))
}
