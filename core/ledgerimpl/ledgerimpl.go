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
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/transaction"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"math/big"
	"sync"
	"github.com/ecoball/go-ecoball/core/shard"
)

var log = elog.NewLogger("LedgerImpl", elog.NoticeLog)

type LedgerImpl struct {
	ChainTxs map[common.Hash]*transaction.ChainTx
	mutex    sync.RWMutex
	path     string
	//ChainCt *ChainContract
	//ChainAc *account.ChainAccount
}

func NewLedger(path string, chainID common.Hash, addr common.Address, shard bool) (l ledger.Ledger, err error) {
	ll := &LedgerImpl{path: path, ChainTxs: make(map[common.Hash]*transaction.ChainTx, 1)}
	if err := ll.NewTxChain(chainID, addr, shard); err != nil {
		return nil, err
	}

	actor := &LedActor{ledger: ll}
	actor.pid, err = NewLedgerActor(actor)
	if err != nil {
		return nil, err
	}

	return ll, nil
}

func (l *LedgerImpl) NewTxChain(chainID common.Hash, addr common.Address, shard bool) (err error) {
	if _, ok := l.ChainTxs[chainID]; ok {
		return nil
	}
	ChainTx, err := transaction.NewTransactionChain(l.path+"/"+chainID.HexString()+"/Transaction", l, shard)
	if err != nil {
		return err
	}
	if shard {
		if err := ChainTx.GenesesShardBlockInit(chainID, addr); err != nil {
			return err
		}
	} else {
		if err := ChainTx.GenesesBlockInit(chainID, addr); err != nil {
			return err
		}
	}

	ChainTx.StateDB.TempDB, err = ChainTx.StateDB.FinalDB.CopyState()
	ChainTx.StateDB.TempDB.Type = state.TempType
	if err != nil {
		return err
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.ChainTxs[chainID] = ChainTx
	log.Info("Chains:", l.ChainTxs)
	return nil
}
func (l *LedgerImpl) NewTxBlock(chainID common.Hash, txs []*types.Transaction, consensusData types.ConsensusData, timeStamp int64) (*types.Block, error) {
	//return l.ChainTx.NewBlock(l, txs, consensusData, timeStamp)
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.NewBlock(l, txs, consensusData, timeStamp)
}
func (l *LedgerImpl) GetTxBlock(chainID common.Hash, hash common.Hash) (*types.Block, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetBlock(hash)
}
func (l *LedgerImpl) SaveTxBlock(chainID common.Hash, block *types.Block) error {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	if err := chain.SaveBlock(block); err != nil {
		return err
	}
	return nil
}
func (l *LedgerImpl) GetTxBlockByHeight(chainID common.Hash, height uint64) (*types.Block, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetBlockByHeight(height)
}
func (l *LedgerImpl) GetCurrentHeader(chainID common.Hash) *types.Header {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
		return nil
	}
	return chain.CurrentHeader
}
func (l *LedgerImpl) GetCurrentHeight(chainID common.Hash) uint64 {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
		return 0
	}
	return chain.CurrentHeader.Height
}
func (l *LedgerImpl) GetChainTx(chainID common.Hash) ledger.ChainInterface {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
		return nil
	}
	return chain
}
func (l *LedgerImpl) VerifyTxBlock(chainID common.Hash, block *types.Block) error {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.VerifyTxBlock(block)
}
func (l *LedgerImpl) CheckTransaction(chainID common.Hash, tx *types.Transaction) error {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	if err := chain.CheckTransaction(tx); err != nil {
		log.Warn(tx.JsonString())
		return err
	}
	return nil
}

func (l *LedgerImpl) GetTransaction(chainID, transactionId common.Hash)(*types.Transaction, error){
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}

	trx, err := chain.GetTransaction(transactionId.Bytes())
	if nil != err {
		return nil, err
	}
	return trx, nil
}

func (l *LedgerImpl) PreHandleTransaction(chainID common.Hash, tx *types.Transaction, timeStamp int64) (ret []byte, cpu, net float64, err error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, 0, 0, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	if err := chain.CheckTransactionWithDB(chain.StateDB.TempDB, tx); err != nil {
		return nil, 0, 0, err
	}
	log.Notice("Handle Transaction:", tx.Type.String(), tx.Hash.HexString(), " in temp DB")
	return chain.HandleTransaction(chain.StateDB.TempDB, tx, timeStamp, chain.CurrentHeader.Receipt.BlockCpu, chain.CurrentHeader.Receipt.BlockNet)
}

func (l *LedgerImpl) AccountGet(chainID common.Hash, index common.AccountName) (*state.Account, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.GetAccountByName(index)
}
func (l *LedgerImpl) AccountAdd(chainID common.Hash, index common.AccountName, addr common.Address, timeStamp int64) (*state.Account, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.AddAccount(index, addr, timeStamp)
}

//func (l *LedgerImpl) AddResourceLimits(from, to common.AccountName, cpu, net float32) error {
//	return l.ChainTx.StateDB.AddResourceLimits(from, to, cpu, net)
//}
func (l *LedgerImpl) StoreSet(chainID common.Hash, index common.AccountName, key, value []byte) (err error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.StoreSet(index, key, value)
}
func (l *LedgerImpl) StoreGet(chainID common.Hash, index common.AccountName, key []byte) (value []byte, err error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.StoreGet(index, key)
}
func (l *LedgerImpl) SetContract(chainID common.Hash, index common.AccountName, t types.VmType, des, code []byte, abi []byte) error {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.SetContract(index, t, des, code, abi)
}
func (l *LedgerImpl) GetContract(chainID common.Hash, index common.AccountName) (*types.DeployInfo, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.GetContract(index)
}
func (l *LedgerImpl) AddPermission(chainID common.Hash, index common.AccountName, perm state.Permission) error {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.AddPermission(index, perm)
}
func (l *LedgerImpl) FindPermission(chainID common.Hash, index common.AccountName, name string) (string, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return "", errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.FindPermission(index, name)
}
func (l *LedgerImpl) GetChainList(chainID common.Hash) ([]state.Chain, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetChainList()
}
func (l *LedgerImpl) CheckPermission(chainID common.Hash, index common.AccountName, name string, hash common.Hash, sig []common.Signature) error {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.CheckPermission(index, name, hash, sig)
}
func (l *LedgerImpl) RequireResources(chainID common.Hash, index common.AccountName, timeStamp int64) (float64, float64, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return 0, 0, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.RequireResources(index, l.ChainTxs[chainID].CurrentHeader.Receipt.BlockCpu, l.ChainTxs[chainID].CurrentHeader.Receipt.BlockNet, timeStamp)
}
func (l *LedgerImpl) GetProducerList(chainID common.Hash) ([]common.AccountName, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.GetProducerList()
}
func (l *LedgerImpl) AccountGetBalance(chainID common.Hash, index common.AccountName, token string) (uint64, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return 0, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	value, err := chain.StateDB.FinalDB.AccountGetBalance(index, token)
	if err != nil {
		return 0, err
	}
	return value.Uint64(), nil
}
func (l *LedgerImpl) AccountAddBalance(chainID common.Hash, index common.AccountName, token string, value uint64) error {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.AccountAddBalance(index, token, new(big.Int).SetUint64(value))
}
func (l *LedgerImpl) AccountSubBalance(chainID common.Hash, index common.AccountName, token string, value uint64) error {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.AccountSubBalance(index, token, new(big.Int).SetUint64(value))
}
func (l *LedgerImpl) TokenCreate(chainID common.Hash, index common.AccountName, token string, maximum uint64) error {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.StateDB.FinalDB.AccountAddBalance(index, token, new(big.Int).SetUint64(maximum))
}
func (l *LedgerImpl) TokenIsExisted(chainID common.Hash, token string) bool {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
		return false
	}
	return chain.StateDB.FinalDB.TokenExisted(token)
}
func (l *LedgerImpl) StateDB(chainID common.Hash) *state.State {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
		return nil
	}
	return chain.StateDB.FinalDB
}
func (l *LedgerImpl) ResetStateDB(chainID common.Hash, header *types.Header) error {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.ResetStateDB(header)
}

/*func (l *LedgerImpl) GetGenesesTime(chainID common.Hash) int64 {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
		return 0
	}
	return chain.Geneses.TimeStamp
}*/

func (l *LedgerImpl) SaveShardBlock(chainID common.Hash, block shard.BlockInterface) (err error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.SaveShardBlock(block)
}

func (l *LedgerImpl) GetShardBlockByHash(chainID common.Hash, typ shard.HeaderType, hash common.Hash) (shard.BlockInterface, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetShardBlockByHash(typ, hash)
}

func (l *LedgerImpl) GetShardBlockByHeight(chainID common.Hash, typ shard.HeaderType, height uint64) (shard.BlockInterface, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetShardBlockByHeight(typ, height)
}

func (l *LedgerImpl) GetLastShardBlock(chainID common.Hash, typ shard.HeaderType) (shard.BlockInterface, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetLastShardBlock(typ)
}

func (l *LedgerImpl) GetLastShardBlockById(chainID common.Hash, shardId uint32) (shard.BlockInterface, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetLastShardBlockById(shardId)
}

func (l *LedgerImpl) NewCmBlock(chainID common.Hash, timeStamp int64, shards []shard.Shard) (*shard.CMBlock, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.NewCmBlock(timeStamp, shards)
}

func (l *LedgerImpl) NewMinorBlock(chainID common.Hash, txs []*types.Transaction, timeStamp int64) (*shard.MinorBlock, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.NewMinorBlock(txs, timeStamp)
}


func (l *LedgerImpl) NewFinalBlock(chainID common.Hash, timeStamp int64, minorBlocks []*shard.MinorBlockHeader) (*shard.FinalBlock, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.NewFinalBlock(timeStamp, minorBlocks)
}

func (l *LedgerImpl) CreateFinalBlock(chainID common.Hash, timeStamp int64, hashes []common.Hash) (*shard.FinalBlock, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return nil, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.CreateFinalBlock(timeStamp, hashes)
}

func (l *LedgerImpl) GetShardId(chainID common.Hash) (uint32, error) {
	chain, ok := l.ChainTxs[chainID]
	if !ok {
		return 0, errors.New(log, fmt.Sprintf("the chain:%s is not existed", chainID.HexString()))
	}
	return chain.GetShardId()
}