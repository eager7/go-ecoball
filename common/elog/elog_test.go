package elog_test

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	l := elog.NewLogger("Module", elog.NoticeLog)
	now := time.Now().UnixNano()
	for i := 0; i < 100000; i++ {
		l.Notice("--------------------------------Test-----------------------------------")
		l.Debug("--------------------------------Test-----------------------------------")
		l.Info("--------------------------------Test-----------------------------------")
		l.Warn("--------------------------------Test-----------------------------------")
		l.Error("--------------------------------Test-----------------------------------")
	}
	end := time.Now().UnixNano()
	l.Info("time:", (end - now)/1000000)
}
