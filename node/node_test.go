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
)

func TestRunMain(t *testing.T) {
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
		if ledger.StateDB().RequireVotingInfo() {
			event.Send(event.ActorNil, event.ActorConsensus, &message.ABABFTStart{})
		} else {
			c, _ := solo.NewSoloConsensusServer(ledger)
			c.Start()
		}
	default:
		elog.Log.Fatal("unsupported consensus algorithm:", config.ConsensusAlgorithm)
	}
	//start transaction pool
	_, err := txpool.Start()
	errors.CheckErrorPanic(err)
	net.StartNetWork(ledger)

	//start explorer
	go spectator.Bystander(ledger)

	go example.AutoGenerateTransaction(ledger)
	go example.VotingProducer(ledger)
	wait()
}

func TestRunNode(t *testing.T) {
	ledger := example.Ledger("/tmp/run_test")
	elog.Log.Info("consensus", config.ConsensusAlgorithm)
	//start consensus
	switch config.ConsensusAlgorithm {
	case "SOLO":
		//c, _ := solo.NewSoloConsensusServer(ledger)
		//c.Start()
	case "DPOS":
		elog.Log.Info("Start DPOS consensus")
	case "ABABFT":
		s, _ := ababft.Service_ababft_gen(ledger, &config.Root)
		s.Start()
		if ledger.StateDB().RequireVotingInfo() {
			event.Send(event.ActorNil, event.ActorConsensus, &message.ABABFTStart{})
		} else {
			c, _ := solo.NewSoloConsensusServer(ledger)
			c.Start()
		}
	default:
		elog.Log.Fatal("unsupported consensus algorithm:", config.ConsensusAlgorithm)
	}
	//start transaction pool
	_, err := txpool.Start()
	errors.CheckErrorPanic(err)
	net.StartNetWork(ledger)
	//start explorer
	go spectator.Bystander(ledger)
	wait()
}

func wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	elog.Log.Info("ecoball received signal:", sig)
}
