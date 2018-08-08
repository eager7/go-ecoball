package geneses_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/test/example"
	"math/big"
	"testing"
	"time"
	"encoding/json"
)

var root = common.NameToIndex("root")
var worker1 = common.NameToIndex("worker1")
var worker2 = common.NameToIndex("worker2")
var worker3 = common.NameToIndex("worker3")
var delegate = common.NameToIndex("delegate")

func TestGenesesBlockInit(t *testing.T) {
	elog.Log.Info("genesis block")
	l := example.Ledger("/tmp/genesis")
	elog.Log.Info("new account block")
	createBlock := CreateAccountBlock(l)

	elog.Log.Info("transfer block:", createBlock.StateHash.HexString())
	blockTransfer := TokenTransferBlock(l)

	elog.Log.Info("pledge block:", blockTransfer.StateHash.HexString())
	pledgeBlock := PledgeContract(l)

	elog.Log.Info("voting block:", pledgeBlock.StateHash.HexString())
	votingBlock := VotingContract(l)
	l.StateDB().RequireVotingInfo()

	//CancelPledgeContract(l, *con)
	//showAccountInfo(l)
	//l.StateDB().RequireVotingInfo()

	elog.Log.Info("current block:", blockTransfer.StateHash.HexString())
	currentBlock, err := l.GetTxBlock(l.GetCurrentHeader().Hash)
	errors.CheckErrorPanic(err)
	errors.CheckEqualPanic(votingBlock.JsonString(false) == currentBlock.JsonString(false))
	//showAccountInfo(l)

	elog.Log.Info("prev block")
	prevBlock, err := l.GetTxBlock(currentBlock.PrevHash)
	errors.CheckErrorPanic(err)
	errors.CheckEqualPanic(pledgeBlock.JsonString(false) == prevBlock.JsonString(false))

	elog.Log.Info("reset block to create block")
	errors.CheckErrorPanic(l.ResetStateDB(prevBlock))
	//elog.Log.Debug("reset hash:", l.StateDB().GetHashRoot().HexString())

	elog.Log.Info("reset block:", )
	newBlock, s, err := l.NewTxBlock(currentBlock.Transactions, currentBlock.ConsensusData, currentBlock.TimeStamp)
	errors.CheckErrorPanic(err)
	newBlock.SetSignature(&config.Root)
	//currentBlock.Show(false)
	//newBlock.Show(false)
	example.ShowAccountInfo(s, root)
	errors.CheckEqualPanic(currentBlock.JsonString(false) == newBlock.JsonString(false))

	//elog.Log.Info("new transfer block")
	//elog.Log.Debug(newBlock.JsonString())
	//elog.Log.Warn("22222222222222222222222222222")
	//l.GetCurrentHeader().Show()
	//curBlock.Header.Show()
	//newBlock.Header.Show()


	for i := 0; i < 0; i++ {
		time.Sleep(10 * time.Second)
		fmt.Println(l.RequireResources(root, time.Now().UnixNano()))
	}

	//errors.CheckErrorPanic(l.StateDB().Close())
}
func CreateAccountBlock(ledger ledger.Ledger) *types.Block {
	elog.Log.Info("CreateAccountBlock------------------------------------------------------\n\n")
	var txs []*types.Transaction
	tokenContract, err := types.NewDeployContract(root, root, state.Active, types.VmNative, "system control", nil, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(tokenContract.SetSignature(&config.Root))
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(root, root, state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, state.Owner, "new_account", []string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 1, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, state.Owner, "new_account", []string{"worker3", common.AddressFromPubKey(config.Worker3.PublicKey).HexString()}, 2, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	perm := state.NewPermission(state.Active, state.Owner, 2, []state.KeyFactor{}, []state.AccFactor{{Actor: common.NameToIndex("worker1"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker2"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker3"), Weight: 1, Permission: "active"}})
	param, err := json.Marshal(perm)
	errors.CheckErrorPanic(err)
	invoke, err = types.NewInvokeContract(root, root, state.Active, "set_account", []string{"root", string(param)}, 0, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	return example.SaveBlock(ledger, txs)
}
func TokenTransferBlock(ledger ledger.Ledger) *types.Block {
	elog.Log.Info("TokenTransferBlock------------------------------------------------------\n\n")
	var txs []*types.Transaction
	transfer, err := types.NewTransfer(root, worker1, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)

	transfer, err = types.NewTransfer(root, worker2, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)

	transfer, err = types.NewTransfer(root, worker3, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)

	return example.SaveBlock(ledger, txs)
}
func PledgeContract(ledger ledger.Ledger) *types.Block{
	elog.Log.Info("PledgeContract------------------------------------------------------")
	var txs []*types.Transaction
	tokenContract, err := types.NewDeployContract(delegate, delegate, "active", types.VmNative, "system control", nil, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	tokenContract.SetSignature(&config.Delegate)
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(root, delegate, "owner", "pledge", []string{"root", "worker1", "100", "100"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker1, delegate, "owner", "pledge", []string{"worker1", "worker1", "100", "100"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker2, delegate, "owner", "pledge", []string{"worker2", "worker2", "100", "100"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	txs = append(txs, invoke)

	return example.SaveBlock(ledger, txs)
}
func VotingContract(ledger ledger.Ledger) *types.Block {
	elog.Log.Info("VotingContract------------------------------------------------------\n\n")
	var txs []*types.Transaction
	invoke, err := types.NewInvokeContract(worker1, root, "active", "reg_prod", []string{"worker1"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker2, root, "active", "reg_prod", []string{"worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker1, root, "active", "vote", []string{"worker1", "worker1", "worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)
	elog.Log.Debug("33333333333333333333333333333")
	ledger.GetCurrentHeader().Show()
	return example.SaveBlock(ledger, txs)
}
func CancelPledgeContract(ledger ledger.Ledger, con types.ConsensusData) *types.Block {
	elog.Log.Info("CancelPledgeContract------------------------------------------------------\n\n")
	var txs []*types.Transaction
	invoke, err := types.NewInvokeContract(worker1, delegate, "owner", "cancel_pledge", []string{"worker1", "worker1", "50", "50"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)

	return example.SaveBlock(ledger, txs)
}

func showAccountInfo(l ledger.Ledger) {
	acc, err := l.AccountGet(root)
	errors.CheckErrorPanic(err)
	acc.Show(false)
/*
	acc, err = l.AccountGet(worker1)
	errors.CheckErrorPanic(err)
	acc.Show()

	acc, err = l.AccountGet(worker2)
	errors.CheckErrorPanic(err)
	acc.Show()

	acc, err = l.AccountGet(worker3)
	errors.CheckErrorPanic(err)
	acc.Show()

	acc, err = l.AccountGet(delegate)
	errors.CheckErrorPanic(err)
	acc.Show()*/
}

