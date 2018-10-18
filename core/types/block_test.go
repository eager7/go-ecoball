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

package types_test

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/utils"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/test/example"
	"math/big"
	"testing"
	"time"
	"github.com/ecoball/go-ecoball/common/elog"
)

func TestBlockCreate(t *testing.T) {
	ledger := example.Ledger("/tmp/block_create")
	root := common.NameToIndex("root")
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

	block, err := ledger.NewTxBlock(config.ChainHash, txs, example.ConsensusData(), time.Now().UnixNano())
	block.SetSignature(&config.Root)
	data, err := block.Serialize()
	errors.CheckErrorPanic(err)

	blockNew := new(types.Block)
	errors.CheckErrorPanic(blockNew.Deserialize(data))
	elog.Log.Debug(block.JsonString(false))
	elog.Log.Info(blockNew.JsonString(false))
	errors.CheckEqualPanic(block.JsonString(false) == blockNew.JsonString(false))

	errors.CheckErrorPanic(ledger.VerifyTxBlock(config.ChainHash, block))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorLedger, block))
	time.Sleep(time.Millisecond * 500)

	reBlock, err := ledger.GetTxBlock(config.ChainHash, blockNew.Hash)
	errors.CheckErrorPanic(err)
	errors.CheckEqualPanic(blockNew.JsonString(false) == reBlock.JsonString(false))
}

func xTestBlockNew(t *testing.T) {
	ledger := example.Ledger("/tmp/block_new")
	root := common.NameToIndex("root")
	var txs []*types.Transaction
	tokenContract, err := types.NewDeployContract(root, root, config.ChainHash, state.Active, types.VmNative, "system control", nil, nil, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(tokenContract.SetSignature(&config.Root))
	txs = append(txs, tokenContract)

	invoke, err := types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().Unix())
	invoke.SetSignature(&config.Root)
	txs = append(txs, invoke)

	for i := 0; i < 3000; i++ {
		transfer, err := types.NewTransfer(root, common.NameToIndex("worker1"), config.ChainHash, "active", new(big.Int).SetUint64(1), 101, time.Now().UnixNano())
		errors.CheckErrorPanic(err)
		transfer.SetSignature(&config.Root)
		txs = append(txs, transfer)
	}

	con, err := types.InitConsensusData(example.TimeStamp())
	errors.CheckErrorPanic(err)
	block, err := ledger.NewTxBlock(config.ChainHash, txs, *con, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	block.SetSignature(&config.Root)
	errors.CheckErrorPanic(ledger.VerifyTxBlock(config.ChainHash, block))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorLedger, block))
	data, err := block.Serialize()
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(utils.FileWrite("/tmp/block.data", data))
}
