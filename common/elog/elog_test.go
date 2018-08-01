package elog_test

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"testing"
)

func TestLogger_P(t *testing.T) {
	l := elog.NewLogger("Module", elog.NoticeLog)
	l.Notice("Test")
	l.Debug("Test")
	l.Info("Test")
	l.Warn("Test")
	l.Error("Test")
}
