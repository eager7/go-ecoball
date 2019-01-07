package solo_test

import (
	"context"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/consensus/solo"
	"github.com/ecoball/go-ecoball/lib-p2p"
	"github.com/ecoball/go-ecoball/test/example"
	"github.com/ecoball/go-ecoball/txpool"
	"golang.org/x/sync/errgroup"
	"testing"
)

func TestNewSolo(t *testing.T) {
	_, ctx := errgroup.WithContext(context.Background())
	event.InitMsgDispatcher()
	p2p.InitNetWork(ctx)
	ledger := example.Ledger("/tmp/solo")
	txPool, err := txpool.Start(ctx, ledger)
	errors.CheckErrorPanic(err)
	if _, err := solo.NewSoloConsensusServer(ledger, txPool, config.User); err != nil {
		t.Fatal(err)
	}
	example.CreateAccountBlock(config.ChainHash)
	example.TokenTransferBlock(config.ChainHash)
	example.PledgeContract(config.ChainHash)
	example.CreateNewChain(config.ChainHash)
}
