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
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
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

	go autoGenerateTransaction(ledger)
	go votingProducer()
	wait()
}

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
	go votingProducer()
	wait()
}

func wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	elog.Log.Info("ecoball received signal:", sig)
}

func autoGenerateTransaction(ledger ledger.Ledger) {
	for ; ;  {
		time.Sleep(time.Second * 2)
		if ledger.StateDB().RequireVotingInfo() {
			elog.Log.Info("Start Consensus Module")
			break
		}
	}
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

func votingProducer() {
	//set smart contract for root delegate
	time.Sleep(time.Second * 5)
	contract, err := types.NewDeployContract(common.NameToIndex("root"), common.NameToIndex("root"), state.Owner, types.VmNative, "system control", nil, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Root))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))

	contract, err = types.NewDeployContract(common.NameToIndex("delegate"), common.NameToIndex("delegate"), state.Owner, types.VmNative, "system control", nil, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Delegate))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))

	//create account worker1, worker2
	time.Sleep(time.Second * 5)
	invoke, err := types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))

	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), state.Owner, "new_account", []string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 1, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))

	//transfer worker1, worker2 aba token
	time.Sleep(time.Second * 5)
	transfer, err := types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("worker1"), state.Owner, new(big.Int).SetUint64(10000), 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))

	transfer, err = types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("worker2"), state.Owner, new(big.Int).SetUint64(10000), 1, time.Now().Unix())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))

	//delegate for worker1 and worker2 cpu,net
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("delegate"), state.Active, "pledge", []string{"root", "worker1", "500", "500"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))

	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("delegate"), state.Active, "pledge", []string{"root", "worker2", "500", "500"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))

	//worker1 and worker2 delegate aba to get votes
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("delegate"), state.Active, "pledge", []string{"worker1", "worker1", "4000", "4000"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), common.NameToIndex("delegate"), state.Active, "pledge", []string{"worker2", "worker2", "4000", "4000"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))

	//worker1, worker2 register to producer
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("root"), state.Active, "reg_prod", []string{"worker1"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), common.NameToIndex("root"), state.Active, "reg_prod", []string{"worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))

	//worker1, worker2 voting to be producer
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("root"), state.Active, "vote", []string{"worker1", "worker1", "worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), common.NameToIndex("root"), state.Active, "vote", []string{"worker1", "worker1", "worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
}