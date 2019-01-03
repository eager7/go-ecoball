package main_test

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/utils"
	"github.com/ecoball/go-ecoball/consensus/solo"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/lib-p2p"
	"github.com/ecoball/go-ecoball/spectator"
	"github.com/ecoball/go-ecoball/test/example"
	"github.com/ecoball/go-ecoball/txpool"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"os"
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
	_, ctx := errgroup.WithContext(context.Background())
	p2p.InitNetWork(ctx)
	os.RemoveAll("/tmp/node_test")
	L, err := ledgerimpl.NewLedger("/tmp/node_test", config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey))
	errors.CheckErrorPanic(err)
	elog.Log.Info("consensus", config.ConsensusAlgorithm)
	ledger.L = L

	//start transaction pool
	txPool, err := txpool.Start(ledger.L)
	errors.CheckErrorPanic(err)

	//start consensus
	switch config.ConsensusAlgorithm {
	case "SOLO":
		solo.NewSoloConsensusServer(ledger.L, txPool, config.User)
		//event.Send(event.ActorNil, event.ActorConsensusSolo, &message.RegChain{ChainID: config.ChainHash, Address: common.AddressFromPubKey(config.Root.PublicKey)})
	case "DPOS":
		elog.Log.Info("Start DPOS consensus")
	default:
		elog.Log.Fatal("unsupported consensus algorithm:", config.ConsensusAlgorithm)
	}
	//go rpc.StartRPCServer()
	//start explorer
	go spectator.Bystander(ledger.L)
	if config.StartNode {
		//go example.CreateAccountBlock(config.ChainHash)
		//go example.TokenContract(ledger.L)
		go example.InvokeTicContract(ledger.L)
		//example.RecepitTest(ledger.L)
	}

	utils.Pause()
}
