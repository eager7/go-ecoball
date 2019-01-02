package main_test

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/consensus/solo"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/lib-p2p"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"github.com/ecoball/go-ecoball/spectator"
	"github.com/ecoball/go-ecoball/test/example"
	"github.com/ecoball/go-ecoball/txpool"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"github.com/ecoball/go-ecoball/common/event"
)

func TestRunMain(t *testing.T) {
	_, ctx := errgroup.WithContext(context.Background())
	event.InitMsgDispatcher()
	p2p.InitNetWork(ctx)
	simulate.LoadConfig("/tmp/sharding.json")

	L, err := ledgerimpl.NewLedger("/tmp/node_test", config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey), true)
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
	case "SHARD":
		elog.Log.Debug("Start Shard Mode")
		elog.Log.Warn("unsupported now")
	default:
		elog.Log.Fatal("unsupported consensus algorithm:", config.ConsensusAlgorithm)
	}
	//rpc.StartRPCServer()
	//start explorer
	go spectator.Bystander(ledger.L)
	if config.StartNode {
		//go example.VotingProducer(ledger.L)
	}

	wait()
}

func TestRunNode(t *testing.T) {
	_, ctx := errgroup.WithContext(context.Background())
	p2p.InitNetWork(ctx)
	os.RemoveAll("/tmp/node_test")
	L, err := ledgerimpl.NewLedger("/tmp/node_test", config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey), false)
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

	wait()
}

func wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	elog.Log.Info("ecoball received signal:", sig)
}
