// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package ledgerimpl_test

import (
	"encoding/json"
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
)

var root = common.NameToIndex("root")
var worker1 = common.NameToIndex("worker1")
var worker2 = common.NameToIndex("worker2")
var worker3 = common.NameToIndex("worker3")

func TestLedgerImpl_ResetStateDB(t *testing.T) {
	elog.Log.Info("genesis block")
	l := example.Ledger("/tmp/ledger_impl")
	elog.Log.Info("new account block")
	createBlock := CreateAccountBlock(l)

	elog.Log.Info("transfer block:", createBlock.StateHash.HexString())
	transferBlock := TokenTransferBlock(l)

	elog.Log.Info("current block:", transferBlock.StateHash.HexString())
	currentBlock, err := l.GetTxBlock(config.ChainHash, l.GetCurrentHeader(config.ChainHash).Hash)
	errors.CheckErrorPanic(err)
	fmt.Println(transferBlock.JsonString(false))
	fmt.Println(currentBlock.JsonString(false))
	errors.CheckEqualPanic(transferBlock.JsonString(false) == currentBlock.JsonString(false))
	elog.Log.Info("prev block")
	prevBlock, err := l.GetTxBlock(config.ChainHash, currentBlock.PrevHash)
	errors.CheckErrorPanic(err)
	errors.CheckEqualPanic(createBlock.JsonString(false) == prevBlock.JsonString(false))
	elog.Log.Info("reset block to create block")
	errors.CheckErrorPanic(l.ResetStateDB(config.ChainHash, prevBlock.Header))
	elog.Log.Info("reset block:")
	newBlock, err := l.NewTxBlock(config.ChainHash, currentBlock.Transactions, currentBlock.ConsensusData, currentBlock.TimeStamp)
	errors.CheckErrorPanic(err)
	newBlock.SetSignature(&config.Root)
	errors.CheckEqualPanic(currentBlock.JsonString(false) == newBlock.JsonString(false))
}

func CreateAccountBlock(ledger ledger.Ledger) *types.Block {
	elog.Log.Info("CreateAccountBlock------------------------------------------------------\n\n")
	var txs []*types.Transaction
	tokenContract, err := types.NewDeployContract(root, root, config.ChainHash, state.Active, types.VmNative, "system control", nil, nil, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(tokenContract.SetSignature(&config.Root))
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "new_account", []string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 1, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "new_account", []string{"worker3", common.AddressFromPubKey(config.Worker3.PublicKey).HexString()}, 2, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	perm := state.NewPermission(state.Active, state.Owner, 2, []state.KeyFactor{}, []state.AccFactor{{Actor: common.NameToIndex("worker1"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker2"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker3"), Weight: 1, Permission: "active"}})
	param, err := json.Marshal(perm)
	errors.CheckErrorPanic(err)
	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Active, "set_account", []string{"root", string(param)}, 0, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	return example.SaveBlock(ledger, txs, config.ChainHash)
}
func TokenTransferBlock(ledger ledger.Ledger) *types.Block {
	elog.Log.Info("TokenTransferBlock------------------------------------------------------\n\n")
	var txs []*types.Transaction
	transfer, err := types.NewTransfer(root, worker1, config.ChainHash, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)

	transfer, err = types.NewTransfer(root, worker2, config.ChainHash, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)

	transfer, err = types.NewTransfer(root, worker3, config.ChainHash, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	txs = append(txs, transfer)

	return example.SaveBlock(ledger, txs, config.ChainHash)
}
