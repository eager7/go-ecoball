package geneses_test

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
	"math/big"
	"testing"
	"time"
	"os"
	"fmt"
	"github.com/ecoball/go-ecoball/common/event"
)

var log = elog.NewLogger("geneses_test", elog.DebugLog)

var root = common.NameToIndex("root")
var token = common.NameToIndex("token")
var worker1 = common.NameToIndex("worker1")
var worker2 = common.NameToIndex("worker2")
var worker3 = common.NameToIndex("worker3")
var delegate = common.NameToIndex("delegate")
var voting = common.NameToIndex("voting")
func TestGenesesBlockInit(t *testing.T) {
	os.RemoveAll("/tmp/geneses/")
	l, err := ledgerimpl.NewLedger("/tmp/geneses")
	if err != nil {
		t.Fatal(err)
	}
	con, err := types.InitConsensusData(time.Now().Unix())
	CreateAccountBlock(l, con, t)
	TokenTransferBlock(l, con, t)
	ShowAccountInfo(l, t)

	PledgeContract(l, con, t)
	ShowAccountInfo(l, t)

	VotingContract(l, con, t)
	ShowAccountInfo(l, t)
	l.StateDB().RequireVotingInfo()

	CancelPledgeContract(l, con, t)
	ShowAccountInfo(l, t)
	l.StateDB().RequireVotingInfo()

	for i := 0; i < 0; i++ {
		time.Sleep(10 * time.Second)
		fmt.Println(l.RequireResources(root, time.Now().UnixNano()))
	}

	l.StateDB().Close()
}

func CreateAccountBlock(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	log.Info("CreateAccountBlock------------------------------------------------------")
	var txs []*types.Transaction
	tokenContract, err := types.NewDeployContract(root, root, state.Active, types.VmNative, "system control", nil, 0, time.Now().Unix())
	if err != nil {
		log.Error(err)
		t.Fatal(err)
	}
	if err := tokenContract.SetSignature(&config.Root); err != nil {
		t.Fatal(err)
	}
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(root, root, state.Owner, "new_account",
		[]string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, state.Owner, "new_account",
		[]string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 1, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, state.Owner, "new_account",
		[]string{"worker3", common.AddressFromPubKey(config.Worker3.PublicKey).HexString()}, 2, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	perm := state.NewPermission(state.Active, state.Owner, 2, []state.KeyFactor{}, []state.AccFactor{{Actor: common.NameToIndex("worker1"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker2"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker3"), Weight: 1, Permission: "active"}})
	param, err := json.Marshal(perm)
	if err != nil {
		t.Fatal(err)
	}
	invoke, err = types.NewInvokeContract(root, root, state.Active, "set_account", []string{"root", string(param)}, 0, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	block, err := ledger.NewTxBlock(txs, *con, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&config.Root)
	if err := ledger.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}

func SetTokenAccountBlock(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	perm := state.NewPermission("active", "owner", 2, []state.KeyFactor{}, []state.AccFactor{{Actor: worker1, Weight: 1, Permission: "active"}, {Actor: worker2, Weight: 1, Permission: "active"}})
	param, err := json.Marshal(perm)
	if err != nil {
		t.Fatal(err)
	}
	invoke, err := types.NewInvokeContract(worker3, root, "owner", "set_account", []string{"root", string(param)}, 0, time.Now().Unix())
	invoke.SetSignature(&config.Worker3)
	transfer, err := types.NewTransfer(root, worker3, "owner", new(big.Int).SetUint64(1000), 100, time.Now().Unix())
	transfer.SetSignature(&config.Root)

	txs := []*types.Transaction{invoke, transfer}
	block, err := ledger.NewTxBlock(txs, *con, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&config.Root)
	if err := ledger.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}
//transfer worker1-500, worker2-500, worker3-500
func TokenTransferBlock(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	log.Info("TokenTransferBlock------------------------------------------------------")
	var txs []*types.Transaction
	transfer, err := types.NewTransfer(root, worker1, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)
	transfer, err = types.NewTransfer(root, worker2, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)
	transfer, err = types.NewTransfer(root, worker3, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)

	block, err := ledger.NewTxBlock(txs, *con, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&config.Root)
	if err := ledger.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}

func ShowAccountInfo(l ledger.Ledger, t *testing.T) {
	acc, err := l.AccountGet(root)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()

	acc, err = l.AccountGet(worker1)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()

	acc, err = l.AccountGet(worker2)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()

	acc, err = l.AccountGet(worker3)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()

	acc, err = l.AccountGet(delegate)
	if err != nil {
		t.Fatal(err)
	}
	acc.Show()
}

func AddTokenAccount(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	var txs []*types.Transaction
	invoke, err := types.NewInvokeContract(root, root, "owner", "new_account",
		[]string{"token", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	code, err := wasmservice.ReadWasm("../../../test/token/token.wasm")
	if err != nil {
		t.Fatal(err)
	}
	tokenContract, err := types.NewDeployContract(token, token, "active", types.VmWasm, "system control", code, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	tokenContract.SetSignature(&config.Worker1)
	txs = append(txs, tokenContract)

	invoke, err = types.NewInvokeContract(token, token, "owner", "create",
		[]string{"token", "aba", "10000"}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)

	block, err := ledger.NewTxBlock(txs, *con, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&config.Root)
	if err := ledger.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}

func ContractStore(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	var txs []*types.Transaction
	code, err := wasmservice.ReadWasm("../../../test/store/store.wasm")
	if err != nil {
		t.Fatal(err)
	}
	tokenContract, err := types.NewDeployContract(root, worker3, "active", types.VmWasm, "system control", code, 0, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	tokenContract.SetSignature(&config.Root)
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(root, worker3, "owner", "StoreSet",
		[]string{"pct", "panchangtao"}, 0, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, worker3, "owner", "StoreGet",
		[]string{"pct"}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	block, err := ledger.NewTxBlock(txs, *con, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&config.Root)
	if err := ledger.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}
//root delegate 100 to worker1, worker1 delegate 100 to self, worker2 delegate 100 to self
func PledgeContract(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	log.Info("PledgeContract------------------------------------------------------")
	var txs []*types.Transaction
	tokenContract, err := types.NewDeployContract(delegate, delegate, "active", types.VmNative, "system control", nil, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	tokenContract.SetSignature(&config.Delegate)
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(root, delegate, "owner", "pledge", []string{"root", "worker1", "100", "100"}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker1, delegate, "owner", "pledge", []string{"worker1", "worker1", "100", "100"}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker2, delegate, "owner", "pledge", []string{"worker2", "worker2", "100", "100"}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Worker2)
	txs = append(txs, invoke)

	block, err := ledger.NewTxBlock(txs, *con, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&config.Root)
	if err := ledger.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}
func CancelPledgeContract(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	log.Info("CancelPledgeContract------------------------------------------------------\n\n")
	var txs []*types.Transaction
	invoke, err := types.NewInvokeContract(worker1, delegate, "owner", "cancel_pledge",
		[]string{"worker1", "worker1", "50", "50"}, 0, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)
	block, err := ledger.NewTxBlock(txs, *con, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&config.Root)
	if err := ledger.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}

func VotingContract(ledger ledger.Ledger, con *types.ConsensusData, t *testing.T) {
	log.Info("VotingContract------------------------------------------------------\n\n")
	var txs []*types.Transaction
	invoke, err := types.NewInvokeContract(worker1, root, "active", "reg_prod", []string{"worker1"}, 0, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker2, root, "active", "reg_prod", []string{"worker2"}, 0, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Worker2)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker1, root, "active", "vote", []string{"worker1", "worker1", "worker2"}, 0, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)

	block, err := ledger.NewTxBlock(txs, *con, time.Now().UnixNano())
	if err != nil {
		t.Fatal(err)
	}
	block.SetSignature(&config.Root)
	if err := ledger.VerifyTxBlock(block); err != nil {
		t.Fatal(err)
	}
	if err := ledger.SaveTxBlock(block); err != nil {
		t.Fatal(err)
	}
}

func xTestError(t *testing.T) {
	i := 0
	for ; ; {
		i++
		fmt.Println("##############################################################################", i)
		fmt.Println(os.RemoveAll("/tmp/geneses/"))
		time.Sleep(time.Microsecond * 100)
		TestGenesesBlockInit(t)
		event.EventStop()
	}
}