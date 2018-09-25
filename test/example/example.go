package example

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/http/common/abi"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"
	"encoding/hex"
)

var interval = time.Millisecond * 100
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
	invoke, err := types.NewInvokeContract(indexFrom, indexAddr, config.ChainHash, "", method, []string{"01b1a6569a557eafcccc71e0d02461fd4b601aea", "Token.Test", "20000"}, 0, time.Now().UnixNano())
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
	deploy, err := types.NewDeployContract(indexFrom, indexAddr, config.ChainHash, "", types.VmWasm, "test deploy", code, nil, 0, time.Now().UnixNano())
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
	indexFrom := common.NameToIndex("root")
	indexAddr := common.NameToIndex("delegate")
	value := big.NewInt(100)
	tx, err := types.NewTransfer(indexFrom, indexAddr, config.ChainHash, "", value, 0, time.Now().UnixNano())
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
	l, err := ledgerimpl.NewLedger(path, config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey))
	errors.CheckErrorPanic(err)
	return l
}

func SaveBlock(ledger ledger.Ledger, txs []*types.Transaction, chainID common.Hash) *types.Block {
	con, err := types.InitConsensusData(TimeStamp())
	errors.CheckErrorPanic(err)
	block, err := ledger.NewTxBlock(chainID, txs, *con, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	block.SetSignature(&config.Root)
	errors.CheckErrorPanic(ledger.VerifyTxBlock(block.ChainID, block))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorLedger, block))
	time.Sleep(time.Millisecond * 500)
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

func AutoGenerateTransaction(ledger ledger.Ledger) {
	for {
		time.Sleep(time.Second * 2)
		if ledger.StateDB(config.ChainHash).RequireVotingInfo() {
			elog.Log.Info("Start Consensus Module")
			break
		}
	}
	for {
		time.Sleep(time.Second * 5)
		nonce := uint64(1)
		nonce++
		transfer, err := types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("delegate"), config.ChainHash, "active", new(big.Int).SetUint64(1), nonce, time.Now().UnixNano())
		errors.CheckErrorPanic(err)
		transfer.SetSignature(&config.Root)

		errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	}
}

func VotingProducer(ledger ledger.Ledger) {
	//set smart contract for root delegate
	time.Sleep(time.Second * 5)
	log.Warn("Start Voting Producer")
	contract, err := types.NewDeployContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, types.VmNative, "system control", nil, nil, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Root))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))
	time.Sleep(time.Millisecond * 500)

	/*contract, err = types.NewDeployContract(common.NameToIndex("delegate"), common.NameToIndex("delegate"), config.ChainHash, state.Owner, types.VmNative, "system control", nil, nil, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Delegate))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))
	time.Sleep(time.Millisecond * 500)*/

	invoke, err := types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, "new_account", []string{"delegate", common.AddressFromPubKey(config.Delegate.PublicKey).HexString()}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//create account worker1, worker2
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, "new_account", []string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 1, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//transfer worker1, worker2 aba token
	time.Sleep(time.Second * 5)
	transfer, err := types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("worker1"), config.ChainHash, state.Owner, new(big.Int).SetUint64(10000), 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(time.Millisecond * 500)

	transfer, err = types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("worker2"), config.ChainHash, state.Owner, new(big.Int).SetUint64(10000), 1, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(time.Millisecond * 500)

	//delegate for worker1 and worker2 cpu,net
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Active, "pledge", []string{"root", "worker1", "500", "500"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Active, "pledge", []string{"root", "worker2", "500", "500"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//worker1 and worker2 delegate aba to get votes
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("root"), config.ChainHash, state.Active, "pledge", []string{"worker1", "worker1", "4000", "4000"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), common.NameToIndex("root"), config.ChainHash, state.Active, "pledge", []string{"worker2", "worker2", "4000", "4000"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//worker1, worker2 register to producer
	/*
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("root"), config.ChainHash, state.Active, "reg_prod", []string{"worker1"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)
	*/

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), common.NameToIndex("root"), config.ChainHash, state.Active, "reg_prod", []string{"worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//worker1, worker2 register to producer
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Active, "reg_prod", []string{"root"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//worker1, worker2 voting to be producer
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("root"), config.ChainHash, state.Active, "vote", []string{"worker1", "root", "worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), common.NameToIndex("root"), config.ChainHash, state.Active, "vote", []string{"worker2", "root", "worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	go AutoGenerateTransaction(ledger)
}

func CreateAccountBlock(chainID common.Hash) {
	log.Info("-----------------------------CreateAccountBlock")
	root := common.NameToIndex("root")
	tokenContract, err := types.NewDeployContract(root, root, chainID, state.Active, types.VmNative, "system control", nil, nil, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(tokenContract.SetSignature(&config.Root))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, tokenContract))
	time.Sleep(time.Second * 2)

	invoke, err := types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"delegate", common.AddressFromPubKey(config.Delegate.PublicKey).HexString()}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	invoke, err = types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	invoke, err = types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 1, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	invoke, err = types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"worker3", common.AddressFromPubKey(config.Worker3.PublicKey).HexString()}, 2, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	perm := state.NewPermission(state.Active, state.Owner, 2, []state.KeyFactor{}, []state.AccFactor{{Actor: common.NameToIndex("worker1"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker2"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker3"), Weight: 1, Permission: "active"}})
	param, err := json.Marshal(perm)
	errors.CheckErrorPanic(err)
	invoke, err = types.NewInvokeContract(root, root, chainID, state.Active, "set_account", []string{"root", string(param)}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	time.Sleep(time.Second * 2)
}

func TokenTransferBlock(chainID common.Hash) {
	log.Info("-----------------------------TokenTransferBlock")
	root := common.NameToIndex("root")
	transfer, err := types.NewTransfer(root, common.NameToIndex("worker1"), chainID, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(interval)

	transfer, err = types.NewTransfer(root, common.NameToIndex("delegate"), chainID, "active", new(big.Int).SetUint64(10000), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(interval)

	transfer, err = types.NewTransfer(root, common.NameToIndex("worker2"), chainID, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(interval)

	transfer, err = types.NewTransfer(root, common.NameToIndex("worker3"), chainID, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(interval)

	time.Sleep(time.Second * 2)
}
func PledgeContract(chainID common.Hash) {
	log.Info("-----------------------------PledgeContract")
	root := common.NameToIndex("root")

	/*delegate := common.NameToIndex("delegate")

	tokenContract, err := types.NewDeployContract(delegate, delegate, chainID, "active", types.VmNative, "system control", nil, nil, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	tokenContract.SetSignature(&config.Delegate)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, tokenContract))
	time.Sleep(time.Second * 2)*/

	invoke, err := types.NewInvokeContract(root, root, chainID, "owner", "pledge", []string{"root", "worker1", "100", "100"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), root, chainID, "owner", "pledge", []string{"worker1", "worker1", "200", "200"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), root, chainID, "owner", "pledge", []string{"worker2", "worker2", "100", "100"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)
	time.Sleep(time.Second * 2)
}
func CreateNewChain(chainID common.Hash) {
	log.Info("-----------------------------CreateNewChain")
	invoke, err := types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), chainID, "active", "reg_chain", []string{"root", "solo", common.AddressFromPubKey(config.Root.PublicKey).HexString()}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)
	time.Sleep(time.Second * 5)
}

func InvokeContract(ledger ledger.Ledger) {
	log.Info("-----------------------------CreateAccountBlock")
	root := common.NameToIndex("root")
	tokenContract, err := types.NewDeployContract(root, root, config.ChainHash, state.Active, types.VmNative, "system control", nil, nil, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(tokenContract.SetSignature(&config.Root))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, tokenContract))
	time.Sleep(time.Second * 2)

	invoke, err := types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "new_account", []string{"worker", common.AddressFromPubKey(config.Worker.PublicKey).HexString()}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	time.Sleep(time.Second * 2)

	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	time.Sleep(time.Second * 2)

	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "new_account", []string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	time.Sleep(time.Second * 2)

	log.Info("-----------------------------TokenTransferBlock")
	transfer, err := types.NewTransfer(root, common.NameToIndex("worker"), config.ChainHash, "active", new(big.Int).SetUint64(1000), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(interval)

	transfer, err = types.NewTransfer(root, common.NameToIndex("worker1"), config.ChainHash, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(interval)

	transfer, err = types.NewTransfer(root, common.NameToIndex("worker2"), config.ChainHash, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(interval)

	time.Sleep(time.Second * 5)

	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "pledge", []string{"root", "worker", "100", "100"}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	time.Sleep(time.Second * 2)

	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "pledge", []string{"root", "worker1", "100", "100"}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	time.Sleep(time.Second * 2)

	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "pledge", []string{"root", "worker2", "100", "100"}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	time.Sleep(time.Second * 2)

	//perm := state.NewPermission(state.Active, state.Owner, 1, []state.KeyFactor{}, []state.AccFactor{{Actor: common.NameToIndex("worker"), Weight: 1, Permission: "active"}})
	//param, err := json.Marshal(perm)
	//errors.CheckErrorPanic(err)
	//invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), root, config.ChainHash, state.Owner, "set_account", []string{"worker2", string(param)}, 0, time.Now().UnixNano())
	//invoke.SetSignature(&config.Worker2)
	//errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	//time.Sleep(interval)
	//
	//time.Sleep(time.Second * 2)


	log.Warn("Start Invoke contract")

	path := os.Getenv("GOPATH")

	// contract file1 data
	file, err := os.OpenFile(path + "/src/github.com/ecoball/go-ecoball/test/abaToken/inline_action.wasm", os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("open file inline_action.wasm failed")
		return
	}

	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("read contract filr err: ", err.Error())
		return
	}

	// abi file, common for contract file1 and file2
	abifile, err := os.OpenFile(path + "/src/github.com/ecoball/go-ecoball/test/abaToken/simple_token.abi", os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("open file simple_token.abi failed")
		return
	}

	defer abifile.Close()
	abidata, err := ioutil.ReadAll(abifile)
	if err != nil {
		fmt.Println("read contract filr err: ", err.Error())
		return
	}

	//contract file2 data
	file2, err := os.OpenFile(path + "/src/github.com/ecoball/go-ecoball/test/abaToken/inline_action2.wasm", os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("open file inline_action2.wasm failed")
		return
	}

	defer file2.Close()
	data2, err := ioutil.ReadAll(file2)
	if err != nil {
		fmt.Println("read contract filr err: ", err.Error())
		return
	}


	var contractAbi abi.ABI
	if err = json.Unmarshal(abidata, &contractAbi); err != nil {
		fmt.Errorf("ABI Unmarshal failed")
		return
	}

	abibyte, err := abi.MarshalBinary(contractAbi)
	if err != nil {
		fmt.Errorf("ABI MarshalBinary failed")
		return
	}
	fmt.Println("abibyte: ", hex.EncodeToString(abibyte))

	// deploy first contract
	contract, err := types.NewDeployContract(common.NameToIndex("worker"), common.NameToIndex("worker"), config.ChainHash, state.Owner, types.VmWasm, "test", data, abibyte, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Worker))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))
	time.Sleep(time.Millisecond * 1500)


	// deploy second contract
	contract, err = types.NewDeployContract(common.NameToIndex("worker2"), common.NameToIndex("worker2"), config.ChainHash, state.Owner, types.VmWasm, "test", data2, abibyte, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Worker2))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))
	time.Sleep(time.Millisecond * 1500)


	contractGet, err := ledger.GetContract(config.ChainHash, common.NameToIndex("worker"))
	if err != nil {
		fmt.Errorf("can not find contract abi file")
		return
	}

	var abiDef abi.ABI
	err = abi.UnmarshalBinary(contractGet.Abi, &abiDef)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	//transfer := []byte(`{"from": "gm2tsojvgene", "to": "hellozhongxh", "quantity": "100.0000 EOS", "memo": "-100"}`)
	//create := []byte(`{"creator": "worker1", "max_supply": "800", "token_id": "xyx"}`)
	// first contract create
	create := []byte(`["worker1", "800", "xyx"]`)

	argbyte, err := abi.CheckParam(abiDef, "create", create)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	var parameters []string

	parameters = append(parameters, string(argbyte[:]))

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("worker"), config.ChainHash, state.Owner, "create", parameters, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 2500)


	// second contract create
	create = []byte(`["worker1", "800", "xxx"]`)

	argbyte, err = abi.CheckParam(abiDef, "create", create)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	var parameters2 []string

	parameters2 = append(parameters2, string(argbyte[:]))

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("worker2"), config.ChainHash, state.Owner, "create", parameters2, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 2500)

	// second contract issue
	issue := []byte(`{"to": "worker1", "amount": "100", "token_id": "xxx"}`)

	argbyte, err = abi.CheckParam(abiDef, "issue", issue)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	var issueParameters2 []string

	issueParameters2 = append(issueParameters2, string(argbyte[:]))

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("worker2"), config.ChainHash, state.Owner, "issue", issueParameters2, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 2500)


	// first contract issue, inline call second contract "transfer"
	issue = []byte(`{"to": "worker1", "amount": "100", "token_id": "xyx"}`)

	argbyte, err = abi.CheckParam(abiDef, "issue", issue)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	var issueParameters []string

	issueParameters = append(issueParameters, string(argbyte[:]))

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("worker"), config.ChainHash, state.Owner, "issue", issueParameters, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 2500)


	//trans := []byte(`{"from": "worker1", "to": "worker2", "amount": "20", "token_id": "xyx"}`)
	//
	//argbyte, err = abi.CheckParam(abiDef, "transfer", trans)
	//if err != nil {
	//	fmt.Errorf("can not find UnmarshalBinary abi file")
	//	return
	//}
	//
	//var transferParameters []string
	//
	//transferParameters = append(transferParameters, string(argbyte[:]))
	//
	//invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("worker"), config.ChainHash, state.Owner, "transfer", transferParameters, 0, time.Now().UnixNano())
	//errors.CheckErrorPanic(err)
	//invoke.SetSignature(&config.Worker1)
	//errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	//time.Sleep(time.Millisecond * 2500)

}

func QueryContractData(ledger ledger.Ledger) {
	time.Sleep(time.Second * 30)

	log.Warn("Query Contract Data")

	contractGet, err := ledger.GetContract(config.ChainHash, common.NameToIndex("root"))
	if err != nil {
		fmt.Errorf("can not find contract abi file")
		return
	}

	var abiDef abi.ABI
	err = abi.UnmarshalBinary(contractGet.Abi, &abiDef)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	abi.GetContractTable("root", "root", abiDef, "accounts")

	abi.GetContractTable("root", "worker1", abiDef, "accounts")

	abi.GetContractTable("root", "worker2", abiDef, "accounts")
}

func Wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	log.Info("ecoball received signal:", sig)
}
