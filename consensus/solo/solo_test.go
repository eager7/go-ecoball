package solo_test

import (
	"testing"
	"github.com/ecoball/go-ecoball/test/example"
	"github.com/ecoball/go-ecoball/txpool"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/consensus/solo"
	"math/big"
	"time"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/event"
)

func TestSoloModule(t *testing.T) {
	ledger := example.Ledger("/tmp/solo")
	_, err := txpool.Start()
	errors.CheckErrorPanic(err)

	c, _ := solo.NewSoloConsensusServer(ledger)
	c.Start()
	autoGenerateTransaction()
	for i := 0; i < 10; i++ {
		autoGenerateTransaction()
		time.Sleep(time.Second * 1)
	}
}

func autoGenerateTransaction() {
		nonce := uint64(1)
		nonce ++
		transfer, err := types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("delegate"), "active", new(big.Int).SetUint64(1), nonce, time.Now().UnixNano())
		errors.CheckErrorPanic(err)
		transfer.SetSignature(&config.Root)

		errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
}