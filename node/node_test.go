package main_test

import (
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/utils"
	"github.com/ecoball/go-ecoball/consensus/solo"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/lib-p2p"
	"github.com/ecoball/go-ecoball/test/example"
	"github.com/ecoball/go-ecoball/txpool"
	"golang.org/x/net/context"
	"testing"
)

func TestRunMain(t *testing.T) {
	event.InitMsgDispatcher()
	p2p.InitNetWork(context.Background())

	ledger.L = example.Ledger("/tmp/node_test")
	elog.Log.Info("consensus", config.ConsensusAlgorithm)
	txPool, err := txpool.Start(ledger.L)
	errors.CheckErrorPanic(err)

	switch config.ConsensusAlgorithm {
	case "SOLO":
		solo.NewSoloConsensusServer(ledger.L, txPool, config.User)
	default:
		elog.Log.Fatal("unsupported consensus algorithm:", config.ConsensusAlgorithm)
	}
	if config.StartNode {
		go example.VotingProducer(ledger.L)
	}
	utils.Pause()
}

func TestRunNode(t *testing.T) {
	event.InitMsgDispatcher()
	p2p.InitNetWork(context.Background())

	ledger.L = example.Ledger("/tmp/node_test")
	elog.Log.Info("consensus", config.ConsensusAlgorithm)
	txPool, err := txpool.Start(ledger.L)
	errors.CheckErrorPanic(err)

	switch config.ConsensusAlgorithm {
	case "SOLO":
		solo.NewSoloConsensusServer(ledger.L, txPool, config.User)
	default:
		elog.Log.Fatal("unsupported consensus algorithm:", config.ConsensusAlgorithm)
	}
	if config.StartNode {
		go example.InvokeSingleContract(ledger.L)
	}
	utils.Pause()
}
