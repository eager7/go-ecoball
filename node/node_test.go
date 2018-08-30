package main_test

import (
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/consensus/ababft"
	"github.com/ecoball/go-ecoball/consensus/solo"
	"github.com/ecoball/go-ecoball/net"
	"github.com/ecoball/go-ecoball/spectator"
	"github.com/ecoball/go-ecoball/test/example"
	"github.com/ecoball/go-ecoball/txpool"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
)

func TestRunMain(t *testing.T) {
	net.InitNetWork()
	ledger.L = example.Ledger("/tmp/run_test")
	elog.Log.Info("consensus", config.ConsensusAlgorithm)

	//start transaction pool
	txPool, err := txpool.Start(ledger.L)
	errors.CheckErrorPanic(err)
	net.StartNetWork()

	//start consensus
	switch config.ConsensusAlgorithm {
	case "SOLO":
		solo.NewSoloConsensusServer(ledger.L, txPool)
		event.Send(event.ActorNil, event.ActorConsensusSolo, config.ChainHash)
	case "DPOS":
		elog.Log.Info("Start DPOS consensus")
	case "ABABFT":
		elog.Log.Info("enter the branch of ababft consensus", config.ConsensusAlgorithm)
		s, _ := ababft.ServiceABABFTGen(ledger.L, txPool, &config.Worker2)
		s.Start()
		elog.Log.Info("send the start message to ababft")
		event.Send(event.ActorNil, event.ActorConsensus, message.ABABFTStart{})
	default:
		elog.Log.Fatal("unsupported consensus algorithm:", config.ConsensusAlgorithm)
	}

	//start explorer
	go spectator.Bystander(ledger.L)
	if config.StartNode {
		//go example.AutoGenerateTransaction(ledger.L)
		//go example.VotingProducer(ledger.L)
		go example.InvokeContract()
	}

	wait()
}
func wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	elog.Log.Info("ecoball received signal:", sig)
}
