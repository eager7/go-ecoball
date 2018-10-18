package solo_test

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/consensus/solo"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net"
	"github.com/ecoball/go-ecoball/test/example"
	"github.com/ecoball/go-ecoball/txpool"
	"math/big"
	"testing"
	"time"
	"golang.org/x/sync/errgroup"
	"context"
)

func xTestSoloModule(t *testing.T) {
	_, ctx := errgroup.WithContext(context.Background())
	net.InitNetWork(ctx)
	ledger := example.Ledger("/tmp/solo")
	txPool, err := txpool.Start(ledger)
	errors.CheckErrorPanic(err)
	net.StartNetWork()

	solo.NewSoloConsensusServer(ledger, txPool, config.User)
	event.Send(event.ActorNil, event.ActorConsensusSolo, config.ChainHash)
	if config.StartNode {
		go autoGenerateTransaction()
	}
	time.Sleep(time.Second * 10)
}

func autoGenerateTransaction() {
	for {
		time.Sleep(time.Second * 1)
		nonce := uint64(1)
		nonce++
		transfer, err := types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("delegate"), config.ChainHash, "active", new(big.Int).SetUint64(1), nonce, time.Now().UnixNano())
		errors.CheckErrorPanic(err)
		transfer.SetSignature(&config.Root)

		errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	}
}

func TestNewSolo(t *testing.T) {
	_, ctx := errgroup.WithContext(context.Background())
	net.InitNetWork(ctx)
	ledger := example.Ledger("/tmp/solo")
	txPool, err := txpool.Start(ledger)
	errors.CheckErrorPanic(err)
	net.StartNetWork()

	solo.NewSoloConsensusServer(ledger, txPool, config.User)
	event.Send(event.ActorNil, event.ActorConsensusSolo, &message.RegChain{
		ChainID: config.ChainHash,
		Address: common.AddressFromPubKey(config.Root.PublicKey),
		TxHash:  common.Hash{},
	})

	example.CreateAccountBlock(config.ChainHash)
	example.TokenTransferBlock(config.ChainHash)
	example.PledgeContract(config.ChainHash)
	example.CreateNewChain(config.ChainHash)
}
