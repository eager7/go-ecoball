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

package ledgerimpl

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/transaction"
	"github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"math/big"
	"sync"
)

var log = elog.NewLogger("LedgerImpl", elog.NoticeLog)

type LedgerImpl struct {
	ChainMap ChainsMap
	mutex    sync.RWMutex
	path     string
}

func NewLedger(path string, chainID common.Hash, addr common.Address, option ...bool) (l ledger.Ledger, err error) {
	log.Debug("Create Ledger in ", path)
	ll := &LedgerImpl{path: path, ChainMap: new(ChainsMap).Initialize()}

	actor := &LedActor{ledger: ll}
	actor.pid, err = NewLedgerActor(actor)
	if err != nil {
		return nil, err
	}

	if err := ll.NewTxChain(chainID, addr, option...); err != nil {
		return nil, err
	}

	return ll, nil
}

func (l *LedgerImpl) NewTxChain(chainID common.Hash, addr common.Address, option ...bool) (err error) {
	if l.ChainMap.Get(chainID) != nil {
		return nil
	}
	ChainTx, err := transaction.NewTransactionChain(l.path+"/"+chainID.HexString()+"/Transaction", l, option...)
	if err != nil {
		return err
	}
	if !config.DisableSharding && len(option) == 0 {
		if err := ChainTx.GenesesShardBlockInit(chainID, addr); err != nil {
			return err
		}
	} else {
		if err := ChainTx.GenesesBlockInit(chainID, addr); err != nil {
			return err
		}
	}
	l.ChainMap.Add(chainID, ChainTx)
	log.Info("Chains:", l.ChainMap)
	return nil
}
func (l *LedgerImpl) NewTxBlock(chainID common.Hash, txs []*types.Transaction, consData types.ConsData, timeStamp int64) (*types.Block, []*types.Transaction, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.NewBlock(l, txs, consData, timeStamp)
}
func (l *LedgerImpl) GetTxBlock(chainID common.Hash, hash common.Hash) (*types.Block, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetBlock(hash)
}
func (l *LedgerImpl) SaveTxBlock(chainID common.Hash, block *types.Block) error {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	if err := chain.SaveBlock(block); err != nil {
		return err
	}
	return nil
}
func (l *LedgerImpl) GetTxBlockByHeight(chainID common.Hash, height uint64) (*types.Block, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetBlockByHeight(height)
}
func (l *LedgerImpl) GetCurrentHeader(chainID common.Hash) *types.Header {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
		return nil
	}
	return chain.CurrentHeader
}
func (l *LedgerImpl) GetCurrentHeight(chainID common.Hash) uint64 {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
		return 0
	}
	return chain.CurrentHeader.Height
}
func (l *LedgerImpl) GetChainTx(chainID common.Hash) ledger.ChainInterface {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
		return nil
	}
	return chain
}
func (l *LedgerImpl) VerifyTxBlock(chainID common.Hash, block *types.Block) error {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.VerifyTxBlock(block)
}
func (l *LedgerImpl) CheckTransaction(chainID common.Hash, tx *types.Transaction) error {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	if err := chain.CheckTransaction(tx); err != nil {
		log.Warn(tx.String())
		return err
	}
	return nil
}

func (l *LedgerImpl) GetTransaction(chainID, hash common.Hash) (*types.Transaction, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}

	trx, err := chain.GetTransaction(hash)
	if nil != err {
		return nil, err
	}
	return trx, nil
}
func (l *LedgerImpl) HandleTransaction(chainID common.Hash, s *state.State, tx *types.Transaction, timeStamp int64, cpuLimit, netLimit float64) (ret []byte, cpu, net float64, err error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, 0, 0, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.HandleTransaction(s, tx, timeStamp, cpuLimit, netLimit)
}
func (l *LedgerImpl) PreHandleTransaction(chainID common.Hash, s *state.State, tx *types.Transaction, timeStamp int64) (ret []byte, cpu, net float64, err error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, 0, 0, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	if err := chain.CheckTransactionWithDB(s, tx); err != nil {
		return nil, 0, 0, err
	}
	log.Notice("Handle Transaction:", tx.Type.String(), tx.Hash.HexString(), " in temp DB")
	return chain.HandleTransaction(s, tx, timeStamp, chain.CurrentHeader.Receipt.BlockCpu, chain.CurrentHeader.Receipt.BlockNet)
}

func (l *LedgerImpl) ShardPreHandleTransaction(chainID common.Hash, s *state.State, tx *types.Transaction, timeStamp int64) (ret []byte, cpu, net float64, err error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, 0, 0, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	if err := chain.CheckTransactionWithDB(chain.StateDB, tx); err != nil {
		return nil, 0, 0, err
	}
	log.Notice("ShardPreHandleTransaction:", tx.Type.String(), tx.Hash.HexString())
	return chain.HandleTransaction(s, tx, timeStamp, config.BlockCpuLimit, config.BlockNetLimit)
}

func (l *LedgerImpl) AccountGet(chainID common.Hash, index common.AccountName) (*state.Account, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.GetAccountByName(index)
}
func (l *LedgerImpl) AccountAdd(chainID common.Hash, index common.AccountName, addr common.Address, timeStamp int64) (*state.Account, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.AddAccount(index, addr, timeStamp)
}

//func (l *LedgerImpl) AddResourceLimits(from, to common.AccountName, cpu, net float32) error {
//	return l.ChainTx.StateDB.AddResourceLimits(from, to, cpu, net)
//}
func (l *LedgerImpl) StoreSet(chainID common.Hash, index common.AccountName, key, value []byte) (err error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.StoreSet(index, key, value)
}
func (l *LedgerImpl) StoreGet(chainID common.Hash, index common.AccountName, key []byte) (value []byte, err error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.StoreGet(index, key)
}
func (l *LedgerImpl) SetContract(chainID common.Hash, index common.AccountName, t types.VmType, des, code []byte, abi []byte) error {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.SetContract(index, t, des, code, abi)
}
func (l *LedgerImpl) GetContract(chainID common.Hash, index common.AccountName) (*types.DeployInfo, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.GetContract(index)
}
func (l *LedgerImpl) AddPermission(chainID common.Hash, index common.AccountName, perm state.Permission) error {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.AddPermission(index, perm)
}
func (l *LedgerImpl) FindPermission(chainID common.Hash, index common.AccountName, name string) (string, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return "", errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FindPermission(index, name)
}
func (l *LedgerImpl) GetChainList(chainID common.Hash) ([]state.Chain, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetChainList()
}
func (l *LedgerImpl) CheckPermission(chainID common.Hash, index common.AccountName, name string, hash common.Hash, sig []common.Signature) error {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.CheckPermission(index, name, hash, sig)
}
func (l *LedgerImpl) RequireResources(chainID common.Hash, index common.AccountName, timeStamp int64) (float64, float64, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return 0, 0, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.RequireResources(index, config.BlockCpuLimit, config.BlockNetLimit, timeStamp)
}
func (l *LedgerImpl) GetProducerList(chainID common.Hash) ([]state.Elector, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.GetProducerList()
}
func (l *LedgerImpl) AccountGetBalance(chainID common.Hash, index common.AccountName, token string) (uint64, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return 0, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	value, err := chain.StateDB.AccountGetBalance(index, token)
	if err != nil {
		return 0, err
	}
	return value.Uint64(), nil
}
func (l *LedgerImpl) AccountAddBalance(chainID common.Hash, index common.AccountName, token string, value uint64) error {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.AccountAddBalance(index, token, new(big.Int).SetUint64(value))
}
func (l *LedgerImpl) AccountSubBalance(chainID common.Hash, index common.AccountName, token string, value uint64) error {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.AccountSubBalance(index, token, new(big.Int).SetUint64(value))
}
func (l *LedgerImpl) GetTokenInfo(chainID common.Hash, token string) (*state.TokenInfo, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.GetTokenInfo(token)
}
func (l *LedgerImpl) TokenCreate(chainID common.Hash, index common.AccountName, token string, maximum uint64) error {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.AccountAddBalance(index, token, new(big.Int).SetUint64(maximum))
}
func (l *LedgerImpl) TokenIsExisted(chainID common.Hash, token string) bool {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
		return false
	}
	return chain.StateDB.TokenExisted(token)
}
func (l *LedgerImpl) StateDB(chainID common.Hash) *state.State {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
		return nil
	}
	return chain.StateDB
}
func (l *LedgerImpl) ResetStateDB(chainID common.Hash, header *types.Header) error {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.ResetStateDB(header)
}

/*func (l *LedgerImpl) GetGenesesTime(chainID common.Hash) int64 {
	chain := l.ChainMap.Get(chainID)
	if !ok {
		errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
		return 0
	}
	return chain.Geneses.TimeStamp
}*/

func (l *LedgerImpl) SaveShardBlock(chainID common.Hash, block shard.BlockInterface) (err error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.SaveShardBlock(block)
}

func (l *LedgerImpl) GetShardBlockByHash(chainID common.Hash, typ mpb.Identify, hash common.Hash, finalizer bool) (shard.BlockInterface, bool, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, false, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetShardBlockByHash(typ, hash, finalizer)
}

func (l *LedgerImpl) GetShardBlockByHeight(chainID common.Hash, typ mpb.Identify, height uint64, shardID uint32) (shard.BlockInterface, bool, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, false, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetShardBlockByHeight(typ, height, shardID)
}

func (l *LedgerImpl) GetLastShardBlock(chainID common.Hash, typ mpb.Identify) (shard.BlockInterface, bool, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, false, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetLastShardBlock(typ)
}

func (l *LedgerImpl) NewCmBlock(chainID common.Hash, timeStamp int64, shards []shard.Shard) (*shard.CMBlock, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.NewCmBlock(timeStamp, shards)
}

func (l *LedgerImpl) NewMinorBlock(chainID common.Hash, txs []*types.Transaction, timeStamp int64) (*shard.MinorBlock, []*types.Transaction, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.NewMinorBlock(txs, timeStamp)
}

func (l *LedgerImpl) NewFinalBlock(chainID common.Hash, timeStamp int64, hashes []common.Hash) (*shard.FinalBlock, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.NewFinalBlock(timeStamp, hashes)
}

func (l *LedgerImpl) NewViewChangeBlock(chainID common.Hash, timeStamp int64, round uint16) (*shard.ViewChangeBlock, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return nil, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.NewViewChangeBlock(timeStamp, round)
}

func (l *LedgerImpl) GetShardId(chainID common.Hash) (uint32, error) {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return 0, errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetShardId()
}

func (l *LedgerImpl) CheckShardBlock(chainID common.Hash, block shard.BlockInterface) error {
	chain := l.ChainMap.Get(chainID)
	if chain == nil {
		return errors.New(fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.CheckShardBlock(block)
}
