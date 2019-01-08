package geneses_test

import (
	"context"
	"encoding/json"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/test/example"
	"github.com/ecoball/go-ecoball/txpool"
	"math/big"
	"testing"
	"time"
)

var root = common.NameToIndex("root")
var worker1 = common.NameToIndex("worker1")
var worker2 = common.NameToIndex("worker2")
var worker3 = common.NameToIndex("worker3")
var delegate = common.NameToIndex("delegate")

func TestGenesesBlockInit(t *testing.T) {
	event.InitMsgDispatcher()
	elog.Log.Info("genesis block")
	ledger.L = example.Ledger("/tmp/genesis")
	_, err := txpool.Start(context.Background(), ledger.L)
	errors.CheckErrorPanic(err)

	elog.Log.Info("new account block")
	createBlock := createAccountBlock(ledger.L, config.ChainHash)

	elog.Log.Info("transfer block:", createBlock.StateHash.HexString())
	blockTransfer := tokenTransferBlock(ledger.L, config.ChainHash)

	elog.Log.Info("pledge block:", blockTransfer.StateHash.HexString())
	pledgeBlock := pledgeContract(ledger.L, config.ChainHash)

	elog.Log.Info("voting block:", pledgeBlock.StateHash.HexString())
	votingBlock := votingContract(ledger.L, config.ChainHash)
	ledger.L.StateDB(config.ChainHash).RequireVotingInfo()

	elog.Log.Info("cancel pledge block:", votingBlock.StateHash.HexString())
	cancelPledgeContract(ledger.L, config.ChainHash)
	elog.Log.Debug(ledger.L.StateDB(config.ChainHash).RequireVotingInfo())
}
func createAccountBlock(ledger ledger.Ledger, chainID common.Hash) *types.Block {
	elog.Log.Info("createAccountBlock--------------------------2----------------------------\n\n")
	var txs []*types.Transaction
	tokenContract, err := types.NewDeployContract(root, root, chainID, state.Active, types.VmNative, "system control", nil, nil, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(tokenContract.SetSignature(&config.Root))
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"delegate", common.AddressFromPubKey(config.Delegate.PublicKey).HexString()}, 0, time.Now().Unix())
	_ = invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().Unix())
	_ = invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 1, time.Now().Unix())
	_ = invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"worker3", common.AddressFromPubKey(config.Worker3.PublicKey).HexString()}, 2, time.Now().Unix())
	_ = invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	perm := state.NewPermission(state.Active, state.Owner, 2, []state.KeyFactor{}, []state.AccFactor{{Actor: common.NameToIndex("worker1"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker2"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker3"), Weight: 1, Permission: "active"}})
	param, err := json.Marshal(perm)
	errors.CheckErrorPanic(err)
	invoke, err = types.NewInvokeContract(root, root, chainID, state.Active, "set_account", []string{"root", string(param)}, 0, time.Now().Unix())
	_ = invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	return example.SaveBlock(ledger, txs, chainID)
}
func tokenTransferBlock(ledger ledger.Ledger, chainID common.Hash) *types.Block {
	elog.Log.Info("tokenTransferBlock---------------------------3---------------------------\n\n")
	var txs []*types.Transaction
	transfer, err := types.NewTransfer(root, worker1, chainID, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	_ = transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)

	transfer, err = types.NewTransfer(root, worker2, chainID, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	_ = transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)

	transfer, err = types.NewTransfer(root, delegate, chainID, "active", new(big.Int).SetUint64(10000), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	_ = transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)

	transfer, err = types.NewTransfer(root, worker3, chainID, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	_ = transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)

	return example.SaveBlock(ledger, txs, chainID)
}
func pledgeContract(ledger ledger.Ledger, chainID common.Hash) *types.Block {
	elog.Log.Info("pledgeContract-----------------------4-------------------------------")
	var txs []*types.Transaction

	invoke, err := types.NewInvokeContract(root, root, chainID, "owner", "pledge", []string{"root", "worker1", "100", "100"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	_ = invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker1, root, chainID, "owner", "pledge", []string{"worker1", "worker1", "200", "200"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	_ = invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker2, root, chainID, "owner", "pledge", []string{"worker2", "worker2", "100", "100"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	_ = invoke.SetSignature(&config.Worker2)
	txs = append(txs, invoke)

	return example.SaveBlock(ledger, txs, chainID)
}
func votingContract(ledger ledger.Ledger, chainID common.Hash) *types.Block {
	elog.Log.Info("votingContract-----------------------5-------------------------------\n\n")
	var txs []*types.Transaction
	invoke, err := types.NewInvokeContract(worker1, root, chainID, "active", "reg_prod", []string{"worker1", "public key 1", "192.168.8.35", "9002", "worker1"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	_ = invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker2, root, chainID, "active", "reg_prod", []string{"worker2", "public key 2", "192.168.8.35", "9002", "worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	_ = invoke.SetSignature(&config.Worker2)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(worker1, root, chainID, "active", "vote", []string{"worker1", "worker1", "worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	_ = invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)
	return example.SaveBlock(ledger, txs, chainID)
}
func cancelPledgeContract(ledger ledger.Ledger, chainID common.Hash) *types.Block {
	elog.Log.Info("cancelPledgeContract---------------------6---------------------------------\n\n")
	var txs []*types.Transaction
	invoke, err := types.NewInvokeContract(worker1, root, chainID, "owner", "cancel_pledge", []string{"worker1", "worker1", "50", "50"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	_ = invoke.SetSignature(&config.Worker1)
	txs = append(txs, invoke)

	return example.SaveBlock(ledger, txs, chainID)
}
