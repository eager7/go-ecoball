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
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/transaction"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"math/big"
	"sync"
)

var log = elog.NewLogger("LedgerImpl", elog.DebugLog)

type LedgerImpl struct {
	ChainTxs map[common.Hash]*transaction.ChainTx
	mutex    sync.RWMutex
	path     string
	//ChainCt *ChainContract
	//ChainAc *account.ChainAccount
}

func NewLedger(path string) (l ledger.Ledger, err error) {
	ll := &LedgerImpl{path: path, ChainTxs: make(map[common.Hash]*transaction.ChainTx, 1)}
	if err := ll.NewTxChain(config.ChainHash); err != nil {
		return nil, err
	}

	actor := &LedActor{ledger: ll}
	actor.pid, err = NewLedgerActor(actor)
	if err != nil {
		return nil, err
	}

	return ll, nil
}

func (l *LedgerImpl) NewTxChain(chainID common.Hash) (err error) {
	ChainTx, err := transaction.NewTransactionChain(l.path+"/"+chainID.HexString()+"/Transaction", l)
	if err != nil {
		return err
	}
	if err := ChainTx.GenesesBlockInit(); err != nil {
		return err
	}
	ChainTx.StateDB.TempDB, err = ChainTx.StateDB.FinalDB.CopyState()
	if err != nil {
		return err
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.ChainTxs[chainID] = ChainTx
	return nil
}
func (l *LedgerImpl) NewTxBlock(chainID common.Hash, txs []*types.Transaction, consensusData types.ConsensusData, timeStamp int64) (*types.Block, error) {
	//return l.ChainTx.NewBlock(l, txs, consensusData, timeStamp)
	return l.ChainTxs[chainID].NewBlock(l, txs, consensusData, timeStamp)
}
func (l *LedgerImpl) GetTxBlock(chainID common.Hash, hash common.Hash) (*types.Block, error) {
	return l.ChainTxs[chainID].GetBlock(hash)
}
func (l *LedgerImpl) SaveTxBlock(chainID common.Hash, block *types.Block) error {
	//if err := l.ChainTx.SaveBlock(block); err != nil {
	if err := l.ChainTxs[chainID].SaveBlock(block); err != nil {
		return err
	}
	return nil
}
func (l *LedgerImpl) GetTxBlockByHeight(chainID common.Hash, height uint64) (*types.Block, error) {
	return l.ChainTxs[chainID].GetBlockByHeight(height)
}
func (l *LedgerImpl) GetCurrentHeader(chainID common.Hash) *types.Header {
	return l.ChainTxs[chainID].CurrentHeader
}
func (l *LedgerImpl) GetCurrentHeight(chainID common.Hash) uint64 {
	return l.ChainTxs[chainID].CurrentHeader.Height
}
func (l *LedgerImpl) GetChainTx(chainID common.Hash) ledger.ChainInterface {
	return l.ChainTxs[chainID]
}
func (l *LedgerImpl) VerifyTxBlock(chainID common.Hash, block *types.Block) error {
	return l.ChainTxs[chainID].VerifyTxBlock(block)
}
func (l *LedgerImpl) CheckTransaction(chainID common.Hash, tx *types.Transaction) error {
	if err := l.ChainTxs[chainID].CheckTransaction(tx); err != nil {
		return err
	}
	return nil
}

func (l *LedgerImpl) PreHandleTransaction(chainID common.Hash, tx *types.Transaction, timeStamp int64) (ret []byte, cpu, net float64, err error) {
	if err := l.ChainTxs[chainID].CheckTransactionWithDB(l.ChainTxs[chainID].StateDB.TempDB, tx); err != nil {
		return nil, 0, 0, err
	}
	return l.ChainTxs[chainID].HandleTransaction(l.ChainTxs[chainID].StateDB.TempDB, tx, timeStamp, l.ChainTxs[chainID].CurrentHeader.Receipt.BlockCpu, l.ChainTxs[chainID].CurrentHeader.Receipt.BlockNet)
}

func (l *LedgerImpl) AccountGet(chainID common.Hash, index common.AccountName) (*state.Account, error) {
	return l.ChainTxs[chainID].StateDB.FinalDB.GetAccountByName(index)
}
func (l *LedgerImpl) AccountAdd(chainID common.Hash, index common.AccountName, addr common.Address, timeStamp int64) (*state.Account, error) {
	return l.ChainTxs[chainID].StateDB.FinalDB.AddAccount(index, addr, timeStamp)
}

//func (l *LedgerImpl) AddResourceLimits(from, to common.AccountName, cpu, net float32) error {
//	return l.ChainTx.StateDB.AddResourceLimits(from, to, cpu, net)
//}
func (l *LedgerImpl) StoreSet(chainID common.Hash, index common.AccountName, key, value []byte) (err error) {
	return l.ChainTxs[chainID].StateDB.FinalDB.StoreSet(index, key, value)
}
func (l *LedgerImpl) StoreGet(chainID common.Hash, index common.AccountName, key []byte) (value []byte, err error) {
	return l.ChainTxs[chainID].StateDB.FinalDB.StoreGet(index, key)
}
func (l *LedgerImpl) SetContract(chainID common.Hash, index common.AccountName, t types.VmType, des, code []byte) error {
	return l.ChainTxs[chainID].StateDB.FinalDB.SetContract(index, t, des, code)
}
func (l *LedgerImpl) GetContract(chainID common.Hash, index common.AccountName) (*types.DeployInfo, error) {
	return l.ChainTxs[chainID].StateDB.FinalDB.GetContract(index)
}
func (l *LedgerImpl) AddPermission(chainID common.Hash, index common.AccountName, perm state.Permission) error {
	return l.ChainTxs[chainID].StateDB.FinalDB.AddPermission(index, perm)
}
func (l *LedgerImpl) FindPermission(chainID common.Hash, index common.AccountName, name string) (string, error) {
	return l.ChainTxs[chainID].StateDB.FinalDB.FindPermission(index, name)
}
func (l *LedgerImpl) CheckPermission(chainID common.Hash, index common.AccountName, name string, hash common.Hash, sig []common.Signature) error {
	return l.ChainTxs[chainID].StateDB.FinalDB.CheckPermission(index, name, hash, sig)
}
func (l *LedgerImpl) RequireResources(chainID common.Hash, index common.AccountName, timeStamp int64) (float64, float64, error) {
	return l.ChainTxs[chainID].StateDB.FinalDB.RequireResources(index, l.ChainTxs[chainID].CurrentHeader.Receipt.BlockCpu, l.ChainTxs[chainID].CurrentHeader.Receipt.BlockNet, timeStamp)
}
func (l *LedgerImpl) GetProducerList(chainID common.Hash) ([]common.AccountName, error) {
	return l.ChainTxs[chainID].StateDB.FinalDB.GetProducerList()
}
func (l *LedgerImpl) AccountGetBalance(chainID common.Hash, index common.AccountName, token string) (uint64, error) {
	value, err := l.ChainTxs[chainID].StateDB.FinalDB.AccountGetBalance(index, token)
	if err != nil {
		return 0, err
	}
	return value.Uint64(), nil
}
func (l *LedgerImpl) AccountAddBalance(chainID common.Hash, index common.AccountName, token string, value uint64) error {
	return l.ChainTxs[chainID].StateDB.FinalDB.AccountAddBalance(index, token, new(big.Int).SetUint64(value))
}
func (l *LedgerImpl) AccountSubBalance(chainID common.Hash, index common.AccountName, token string, value uint64) error {
	return l.ChainTxs[chainID].StateDB.FinalDB.AccountSubBalance(index, token, new(big.Int).SetUint64(value))
}
func (l *LedgerImpl) TokenCreate(chainID common.Hash, index common.AccountName, token string, maximum uint64) error {
	return l.ChainTxs[chainID].StateDB.FinalDB.AccountAddBalance(index, token, new(big.Int).SetUint64(maximum))
}
func (l *LedgerImpl) TokenIsExisted(chainID common.Hash, token string) bool {
	return l.ChainTxs[chainID].StateDB.FinalDB.TokenExisted(token)
}
func (l *LedgerImpl) StateDB(chainID common.Hash) *state.State {
	return l.ChainTxs[chainID].StateDB.FinalDB
}
func (l *LedgerImpl) ResetStateDB(chainID common.Hash, header *types.Header) error {
	return l.ChainTxs[chainID].ResetStateDB(header)
}

func (l *LedgerImpl) GetGenesesTime(chainID common.Hash) int64 {
	return l.ChainTxs[chainID].Geneses.TimeStamp
}
