package main_test

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/consensus/ababft"
	"github.com/ecoball/go-ecoball/consensus/solo"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net"
	"github.com/ecoball/go-ecoball/spectator"
	"github.com/ecoball/go-ecoball/test/example"
	"github.com/ecoball/go-ecoball/txpool"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestRunNode(t *testing.T) {
	ledger := example.Ledger("/tmp/run_test")
	elog.Log.Info("consensus", config.ConsensusAlgorithm)
	//start consensus
	switch config.ConsensusAlgorithm {
	case "SOLO":
		c, _ := solo.NewSoloConsensusServer(ledger)
		c.Start()
	case "DPOS":
		elog.Log.Info("Start DPOS consensus")
	case "ABABFT":
		s, _ := ababft.Service_ababft_gen(ledger, &config.Root)
		s.Start()
	default:
		elog.Log.Fatal("unsupported consensus algorithm:", config.ConsensusAlgorithm)
	}
	//start transaction pool
	_, err := txpool.Start()
	errors.CheckErrorPanic(err)
	net.StartNetWork(ledger)

	//start explorer
	go spectator.Bystander(ledger)

	go autoGenerateTransaction()

	wait()
}

func wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	elog.Log.Info("ecoball received signal:", sig)
}

func autoGenerateTransaction() {
	for {
		time.Sleep(time.Second * 1)
		nonce := uint64(1)
		nonce++
		transfer, err := types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("delegate"), "active", new(big.Int).SetUint64(1), nonce, time.Now().UnixNano())
		errors.CheckErrorPanic(err)
		transfer.SetSignature(&config.Root)

		errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	}
}
