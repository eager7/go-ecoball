package net_test

import (
	"context"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/net"
	"github.com/ecoball/go-ecoball/test/example"
	"testing"
)

func TestNet(t *testing.T) {
	elog.Log.Debug("net test program...")

	ctx, cancel := context.WithCancel(context.Background())
	net.InitNetWork(ctx)
	net.StartNetWork(nil)

	example.Wait()
	cancel()
}
