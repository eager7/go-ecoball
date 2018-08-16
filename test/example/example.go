package example

import (
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"math/big"
	"time"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"os"
	"github.com/ecoball/go-ecoball/common/event"
)

var log = elog.NewLogger("example", elog.InfoLog)

func AddAccount(state *state.State) error {
	from := common.NewAddress(common.FromHex("01b1a6569a557eafcccc71e0d02461fd4b601aea"))
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	indexFrom := common.NameToIndex("from")
	indexAddr := common.NameToIndex("addr")
	if _, err := state.AddAccount(indexFrom, from, time.Now().UnixNano()); err != nil {
		return nil
	}
	if _, err := state.AddAccount(indexAddr, addr, time.Now().UnixNano()); err != nil {
		return nil
	}
	return nil
}

func TestInvoke(method string) *types.Transaction {
	indexFrom := common.NameToIndex("from")
	indexAddr := common.NameToIndex("addr")
	invoke, err := types.NewInvokeContract(indexFrom, indexAddr, "", method, []string{"01b1a6569a557eafcccc71e0d02461fd4b601aea", "Token.Test", "20000"}, 0, time.Now().Unix())
	if err != nil {
		panic(err)
		return nil
	}
	acc := account.Account{PrivateKey: config.Root.PrivateKey, PublicKey: config.Root.PublicKey, Alg: 0}
	if err := invoke.SetSignature(&acc); err != nil {
		panic(err)
	}
	return invoke
}

func TestDeploy(code []byte) *types.Transaction {
	indexFrom := common.NameToIndex("from")
	indexAddr := common.NameToIndex("addr")
	deploy, err := types.NewDeployContract(indexFrom, indexAddr, "", types.VmWasm, "test deploy", code, 0, time.Now().Unix())
	if err != nil {
		panic(err)
		return nil
	}
	acc := account.Account{PrivateKey: config.Root.PrivateKey, PublicKey: config.Root.PublicKey, Alg: 0}
	if err := deploy.SetSignature(&acc); err != nil {
		panic(err)
	}
	return deploy
}

func TestTransfer() *types.Transaction {
	indexFrom := common.NameToIndex("from")
	indexAddr := common.NameToIndex("addr")
	value := big.NewInt(100)
	tx, err := types.NewTransfer(indexFrom, indexAddr, "", value, 0, time.Now().Unix())
	if err != nil {
		fmt.Println(err)
		return nil
	}
	acc := account.Account{PrivateKey: config.Root.PrivateKey, PublicKey: config.Root.PublicKey, Alg: 0}
	if err := tx.SetSignature(&acc); err != nil {
		fmt.Println(err)
		return nil
	}
	return tx
}

func Ledger(path string) ledger.Ledger {
	os.RemoveAll(path)
	l, err := ledgerimpl.NewLedger(path)
	errors.CheckErrorPanic(err)
	return l
}

func SaveBlock(ledger ledger.Ledger, txs []*types.Transaction) *types.Block {
	con, err := types.InitConsensusData(TimeStamp())
	errors.CheckErrorPanic(err)
	block, err := ledger.NewTxBlock(txs, *con, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	block.SetSignature(&config.Root)
	errors.CheckErrorPanic(ledger.VerifyTxBlock(block))
	errors.CheckErrorPanic(ledger.SaveTxBlock(block))
	return block
}

func TimeStamp() int64 {
	tm, err := time.Parse("02/01/2006 15:04:05 PM", "21/02/1990 00:00:00 AM")
	errors.CheckErrorPanic(err)
	return tm.UnixNano()
}

func ConsensusData() types.ConsensusData {
	con, _ := types.InitConsensusData(TimeStamp())
	return *con
}

func ShowAccountInfo(s *state.State, index common.AccountName) {
	acc, err := s.GetAccountByName(index)
	errors.CheckErrorPanic(err)
	acc.Show()
}


func AutoGenerateTransaction(ledger ledger.Ledger) {
	for ; ;  {
		time.Sleep(time.Second * 2)
		if ledger.StateDB().RequireVotingInfo() {
			elog.Log.Info("Start Consensus Module")
			break
		}
	}
	for {
		time.Sleep(time.Second * 5)
		nonce := uint64(1)
		nonce++
		transfer, err := types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("delegate"), "active", new(big.Int).SetUint64(1), nonce, time.Now().UnixNano())
		errors.CheckErrorPanic(err)
		transfer.SetSignature(&config.Root)

		errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	}
}

func VotingProducer() {
	//set smart contract for root delegate
	time.Sleep(time.Second * 15)
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