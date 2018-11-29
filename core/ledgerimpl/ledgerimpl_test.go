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
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/test/example"
	"math/big"
	"testing"
	"time"
	"github.com/ecoball/go-ecoball/core/shard"
	"os"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"github.com/ecoball/go-ecoball/txpool"
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
	newBlock, _, err := l.NewTxBlock(config.ChainHash, currentBlock.Transactions, currentBlock.ConsensusData, currentBlock.TimeStamp)
	errors.CheckErrorPanic(err)
	newBlock.SetSignature(&config.Root)
	errors.CheckEqualPanic(currentBlock.JsonString(false) == newBlock.JsonString(false))
	event.EventStop()
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

func TestInterface(t *testing.T) {
	simulate.LoadConfig()
	os.RemoveAll("/tmp/interface")
	l, err := ledgerimpl.NewLedger("/tmp/interface", config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey), true)
	errors.CheckErrorPanic(err)
	header := shard.CMBlockHeader{
		ChainID:   config.ChainHash,
		Version:   0,
		Height:    10,
		Timestamp: 2340,
		PrevHash:  common.Hash{},
		//ConsData:     example.ConsensusData(),
		LeaderPubKey: []byte("12345678909876554432"),
		Nonce:        23450,
		Candidate: shard.NodeInfo{
			PublicKey: config.Root.PublicKey,
			Address:   "1234",
			Port:      "5678",
		},
		ShardsHash: config.ChainHash,
		COSign:     &types.COSign{},
	}
	errors.CheckErrorPanic(header.ComputeHash())
	shards := []shard.Shard{{
		Member:     []shard.NodeInfo{{
			PublicKey: simulate.GetNodePubKey(),
			Address:   simulate.GetNodeInfo().Address,
			Port:      simulate.GetNodeInfo().Port,
		}},
		MemberAddr: []shard.NodeAddr{{
			Address:   simulate.GetNodeInfo().Address,
			Port:      simulate.GetNodeInfo().Port,
		}},
	}}
	block, err := shard.NewCmBlock(header, shards)
	errors.CheckErrorPanic(l.SaveShardBlock(config.ChainHash, block))
	blockGet, _, err := l.GetShardBlockByHash(config.ChainHash, shard.HeCmBlock, block.Hash(), true)
	errors.CheckErrorPanic(err)
	errors.CheckEqualPanic(block.JsonString() == blockGet.JsonString())

	blockLast, _, err := l.GetLastShardBlock(config.ChainHash, shard.HeCmBlock)
	errors.CheckErrorPanic(err)
	errors.CheckEqualPanic(block.JsonString() == blockLast.JsonString())

	list, err := l.GetProducerList(config.ChainHash)
	errors.CheckErrorPanic(err)
	fmt.Println(list)

	blockMinor, _, err := l.NewMinorBlock(config.ChainHash, []*types.Transaction{example.TestTransfer()}, 0)
	errors.CheckErrorPanic(l.SaveShardBlock(config.ChainHash, blockMinor))

	event.EventStop()
}

func TestShard(t *testing.T) {
	os.RemoveAll("/tmp/shard_test")
	simulate.LoadConfig()
	l, err := ledgerimpl.NewLedger("/tmp/shard_test", config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey), true)
	errors.CheckErrorPanic(err)

	//check get last cm block
	blockNew, _, err := l.GetLastShardBlock(config.ChainHash, shard.HeCmBlock)
	errors.CheckErrorPanic(err)
	//test new cm block
	shards := []shard.Shard{{
		Member:     []shard.NodeInfo{{
			PublicKey: simulate.GetNodePubKey(),
			Address:   simulate.GetNodeInfo().Address,
			Port:      simulate.GetNodeInfo().Port,
		}},
		MemberAddr: []shard.NodeAddr{{
			Address:   simulate.GetNodeInfo().Address,
			Port:      simulate.GetNodeInfo().Port,
		}},
	}}
	blockCM, err := l.NewCmBlock(config.ChainHash, time.Now().UnixNano(), shards)
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(l.SaveShardBlock(config.ChainHash, blockCM))
	//check get cm block
	blockNew, _, err = l.GetShardBlockByHash(config.ChainHash, shard.HeCmBlock, blockCM.Hash(), true)
	errors.CheckErrorPanic(err)
	elog.Log.Info("Committee Block:", blockNew.JsonString())
	errors.CheckEqualPanic(blockCM.JsonString() == blockNew.JsonString())

	//MinorBlock
	blockNew, _, err = l.GetLastShardBlock(config.ChainHash, shard.HeMinorBlock)
	errors.CheckErrorPanic(err)
	blockMinor, _, err := l.NewMinorBlock(config.ChainHash, []*types.Transaction{example.TestTransfer()}, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(l.SaveShardBlock(config.ChainHash, blockMinor))
	blockNew, _, err = l.GetShardBlockByHash(config.ChainHash, shard.HeMinorBlock, blockMinor.Hash(), true)
	errors.CheckErrorPanic(err)
	elog.Log.Info("Minor Block:", blockNew.JsonString())
	errors.CheckEqualPanic(blockMinor.JsonString() == blockNew.JsonString())


	//FinalBlock
	blockNew, _, err = l.GetLastShardBlock(config.ChainHash, shard.HeFinalBlock)
	errors.CheckErrorPanic(err)
	blockFinal, err := l.NewFinalBlock(config.ChainHash, time.Now().UnixNano(), []common.Hash{blockMinor.Hash()})
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(l.SaveShardBlock(config.ChainHash, blockFinal))
	blockNew, _, err = l.GetShardBlockByHash(config.ChainHash, shard.HeFinalBlock, blockFinal.Hash(), true)
	errors.CheckErrorPanic(err)
	elog.Log.Info("Final Block:", blockNew.JsonString())
	errors.CheckEqualPanic(blockFinal.JsonString() == blockNew.JsonString())
	event.EventStop()


	//create a new minor block and the height will auto increment
	blockMinor, _, err = l.NewMinorBlock(config.ChainHash, []*types.Transaction{example.TestTransfer()}, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(l.SaveShardBlock(config.ChainHash, blockMinor))


	event.EventStop()

}

func TestExample(t *testing.T) {
	os.RemoveAll("/tmp/shard_example")
	simulate.LoadConfig()
	l, err := ledgerimpl.NewLedger("/tmp/shard_example", config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey), true)
	errors.CheckErrorPanic(err)

	pid := example.Actor()

	msg := &message.ProducerBlock{
		ChainID: config.ChainHash,
		Height:  2,
		Type:    shard.HeFinalBlock,
	}
	pidL, _ := event.GetActor(event.ActorLedger)
	pidL.Request(msg, pid)
	time.Sleep(time.Second * 1)

	m, _, err := l.NewMinorBlock(config.ChainHash, []*types.Transaction{example.TestTransfer()}, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	elog.Log.Debug(m.JsonString())

	block, _, err := l.GetLastShardBlock(config.ChainHash, shard.HeViewChange)
	errors.CheckErrorPanic(err)
	elog.Log.Debug("vc block:", block.JsonString())

	event.EventStop()
}

func TestMinorBlock(t *testing.T) {
	simulate.LoadConfig()
	os.RemoveAll("/tmp/minor_test")
	l, err := ledgerimpl.NewLedger("/tmp/minor_test", config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey), true)
	errors.CheckErrorPanic(err)
	b, _, err := l.NewMinorBlock(config.ChainHash, nil, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	l.SaveShardBlock(config.ChainHash, b)

	f, err := l.NewFinalBlock(config.ChainHash, time.Now().UnixNano(), []common.Hash{b.Hash()})
	errors.CheckErrorPanic(err)
	l.SaveShardBlock(config.ChainHash, f)


	b, _, err = l.NewMinorBlock(config.ChainHash, nil, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	l.SaveShardBlock(config.ChainHash, b)

	event.EventStop()
}

func xTestSaveBlock(t *testing.T) {
	simulate.LoadConfig()
	os.RemoveAll("/tmp/block_save_test")
	l, err := ledgerimpl.NewLedger("/tmp/block_save_test", config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey), true)
	errors.CheckErrorPanic(err)

	_, err = txpool.Start(l)
	errors.CheckErrorPanic(err)

	shards := []shard.Shard{shard.Shard{
		Member:     []shard.NodeInfo{shard.NodeInfo{
			PublicKey: simulate.GetNodePubKey(),
			Address:   simulate.GetNodeInfo().Address,
			Port:      simulate.GetNodeInfo().Port,
		}},
		MemberAddr: []shard.NodeAddr{shard.NodeAddr{
			Address:   simulate.GetNodeInfo().Address,
			Port:      simulate.GetNodeInfo().Port,
		}},
	}}
	blockCM, err := l.NewCmBlock(config.ChainHash, time.Now().UnixNano(), shards)
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(l.SaveShardBlock(config.ChainHash, blockCM))

	var txs []*types.Transaction
	for i := 0; i < 2; i ++ {
		txs = append(txs, example.TestTransfer())
	}

	b, _, err := l.NewMinorBlock(config.ChainHash, txs, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorLedger, b))
	time.Sleep(time.Millisecond * 2000)

	f, err := l.NewFinalBlock(config.ChainHash, time.Now().UnixNano(), []common.Hash{b.Hash()})
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorLedger, f))
	time.Sleep(time.Second * 5)

	event.EventStop()
}
