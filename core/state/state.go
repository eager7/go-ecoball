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

package state

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/store"
	"sync"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common/config"
)

var log = elog.NewLogger("state", config.LogLevel)
var AbaToken = "ABA"

type TypeState uint8

const (
	FinalType TypeState = 1
	TempType  TypeState = 2
	CopyType  TypeState = 3
)

func (t TypeState) String() string {
	switch t {
	case FinalType:
		return "FinalType"
	case TempType:
		return "TempType"
	case CopyType:
		return "CopyType"
	}
	return "unknown type"
}

type State struct {
	Type   TypeState
	path   string
	trie   Trie
	db     Database
	diskDb *store.LevelDBStore

	tokenMutex sync.RWMutex
	Tokens	map[string]*TokenInfo	//map[token name]TokenInfo

	accMutex sync.RWMutex
	Accounts map[string]*Account	// map[account name]*Account

	paraMutex sync.RWMutex
	Params    map[string]uint64

	prodMutex sync.RWMutex
	Producers map[common.AccountName]uint64

	chainMutex sync.RWMutex
	Chains     map[common.Hash]Chain

	mutex sync.RWMutex
}

/**
 *  @brief create a new mpt trie and a levelDB
 *  @param path - the levelDB store path
 *  @param root - the root of mpt trie, this value decide the state of trie
 */
func NewState(path string, root common.Hash) (st *State, err error) {
	st = &State{path: path}
	st.diskDb, err = store.NewLevelDBStore(path, 0, 0)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	st.db = NewDatabase(st.diskDb)
	log.Notice("Open Trie Hash:", root.HexString())
	st.trie, err = st.db.OpenTrie(root)
	if err != nil {
		st.trie, _ = st.db.OpenTrie(common.Hash{})
	}
	st.Tokens = make(map[string]*TokenInfo, 1)
	st.Accounts = make(map[string]*Account, 1)
	st.Params = make(map[string]uint64, 1)
	st.Producers = make(map[common.AccountName]uint64, 1)
	st.Chains = make(map[common.Hash]Chain, 1)
	return st, nil
}

func (s *State) StateType() TypeState {
	return s.Type
}

/**
 *  @brief copy a new trie into memory
 */
func (s *State) CopyState() (*State, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	params := make(map[string]uint64, 1)
	tokens := make(map[string]*TokenInfo, 1)
	accounts := make(map[string]*Account, 1)
	prods := make(map[common.AccountName]uint64, 1)
	chains := make(map[common.Hash]Chain, 1)

	/*
		s.paraMutex.Lock()
		defer s.paraMutex.Unlock()
		if str, err := json.Marshal(s.Params); err != nil {
			return nil, err
		} else {
			if err := json.Unmarshal(str, &params); err != nil {
				return nil, err
			}
		}
		if str, err := json.Marshal(s.Producers); err != nil {
			return nil, err
		} else {
			if err := json.Unmarshal(str, &prods); err != nil {
				return nil, err
			}
		}
	*/
	s.accMutex.RLock()
	defer s.accMutex.RUnlock()
	for _, v := range s.Accounts {
		data, _ := v.Serialize()
		acc := new(Account)
		acc.Deserialize(data)
		accounts[acc.Index.String()] = acc
	}

	s.tokenMutex.RLock()
	defer s.tokenMutex.RUnlock()
	for _, v := range s.Tokens {
		data, _ := v.Serialize()
		token := new(TokenInfo)
		token.Deserialize(data)
		tokens[token.Symbol] = token
	}

	return &State{
		path:      s.path,
		trie:      s.db.CopyTrie(s.trie),
		Tokens: 	tokens,
		Accounts:  accounts,
		Params:    params,
		Producers: prods,
		Chains:    chains,
	}, nil
}

/**
 *  @brief create a new account and store into mpt trie, meanwhile store the mapping of addr and index
 *  @param index - account's index
 *  @param addr - account's address convert from public key
 */
func (s *State) AddAccount(index common.AccountName, addr common.Address, timeStamp int64) (*Account, error) {
	key := common.IndexToBytes(index)
	s.mutex.RLock()
	data, err := s.trie.TryGet(key)
	s.mutex.RUnlock()
	if err != nil {
		return nil, err
	}
	if data != nil {
		return nil, errors.New(log, "reduplicate name")
	}
	acc, err := NewAccount(s.path, index, addr, timeStamp)
	if err != nil {
		return nil, err
	}
	if err := s.CommitAccount(acc); err != nil {
		return nil, err
	}
	//save the mapping of addr and index
	s.mutex.Lock()
	err = s.trie.TryUpdate(addr.Bytes(), common.IndexToBytes(acc.Index))
	s.mutex.Unlock()
	if err != nil {
		return nil, err
	}
	//s.accMutex.Lock()
	//defer s.accMutex.Unlock()
	//s.Accounts[index.String()] = *acc

	s.commitParam(addr.HexString(), uint64(index))
	return acc, nil
}

/**
 *  @brief store the smart contract of account, every account only has one contract
 *  @param index - account's index
 *  @param t - the virtual machine type
 *  @param des - the description of contract
 *  @param code - the code of contract
 *  @param abi  - the abi of contract
 */
func (s *State) SetContract(index common.AccountName, t types.VmType, des, code, abi []byte) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.mutex.Lock()
	defer acc.mutex.Unlock()
	if err := acc.SetContract(t, des, code, abi); err != nil {
		return err
	}
	return s.CommitAccount(acc)
}

/**
 *  @brief get the code of account
 *  @param index - account's index
 */
func (s *State) GetContract(index common.AccountName) (*types.DeployInfo, error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return nil, err
	}
	acc.mutex.RLock()
	defer acc.mutex.RUnlock()
	return acc.GetContract()
}
func (s *State) StoreSet(index common.AccountName, key, value []byte) (err error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.mutex.Lock()
	defer acc.mutex.Unlock()
	if err := acc.StoreSet(s.path, key, value); err != nil {
		return err
	}
	return s.CommitAccount(acc)
}
func (s *State) StoreGet(index common.AccountName, key []byte) (value []byte, err error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return nil, err
	}
	acc.mutex.Lock()
	defer acc.mutex.Unlock()
	return acc.StoreGet(s.path, key)
}

/**
*  @brief get the abi of contract
*  @param index - account's index
 */
func (s *State) GetContractAbi(index common.AccountName) ([]byte, error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return nil, err
	}
	return acc.Contract.Abi, err
}

/**
 *  @brief search the account by name index
 *  @param index - the account index
 */
func (s *State) GetAccountByName(index common.AccountName) (*Account, error) {
	s.accMutex.RLock()
	defer s.accMutex.RUnlock()
	acc, ok := s.Accounts[index.String()]
	if ok {
		return acc, nil
	}
	key := common.IndexToBytes(index)
	s.mutex.Lock()
	fData, err := s.trie.TryGet(key)
	s.mutex.Unlock()
	if err != nil {
		return nil, err
	}
	if fData == nil {
		log.Warn(fmt.Sprintf("no this account named:%s", index.String()))
		return nil, errors.New(log, fmt.Sprintf("no this account named:%s", index.String()))
	}
	acc = &Account{}
	if err = acc.Deserialize(fData); err != nil {
		return nil, err
	}
	return acc, nil
}

/**
 *  @brief search the account by address
 *  @param addr - the account address
 */
func (s *State) GetAccountByAddr(addr common.Address) (*Account, error) {
	if value, err := s.getParam(addr.HexString()); err != nil {
		return nil, err
	} else {
		if value == 0 {
			return nil, errors.New(log, fmt.Sprintf("the address:%s is not register be an account", addr.HexString()))
		}
		return s.GetAccountByName(common.AccountName(value))
	}
	s.mutex.Lock()
	fData, err := s.trie.TryGet(addr.Bytes())
	s.mutex.Unlock()
	if err != nil {
		return nil, err
	} else {
		if fData == nil {
			return nil, errors.New(log, fmt.Sprintf("can't find this account by address:%s", addr.HexString()))
		} else {
			acc, err := s.GetAccountByName(common.IndexSetBytes(fData))
			if err != nil {
				return nil, err
			}
			return acc, nil
		}
	}
}

/**
 *  @brief update the account's information into trie
 *  @param acc - account object
 */
func (s *State) CommitAccount(acc *Account) error {
	if acc == nil {
		return errors.New(log, "param acc is nil")
	}
	d, err := acc.Serialize()
	if err != nil {
		return err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.trie.TryUpdate(common.IndexToBytes(acc.Index), d); err != nil {
		return err
	}
	s.accMutex.Lock()
	defer s.accMutex.Unlock()
	s.Accounts[acc.Index.String()] = acc
	return nil
}

/**
 *  @brief update the param's information into trie
 *  @param key - param name
 *  @param value - param value
 */
func (s *State) commitParam(key string, value uint64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.trie.TryUpdate([]byte(key), common.Uint64ToBytes(value)); err != nil {
		return err
	}
	s.paraMutex.Lock()
	defer s.paraMutex.Unlock()
	s.Params[key] = value
	return nil
}

/**
 *  @brief get the param's information from trie
 *  @param key - param name
 */
func (s *State) getParam(key string) (uint64, error) {
	s.paraMutex.Lock()
	defer s.paraMutex.Unlock()
	value, ok := s.Params[key]
	if ok {
		return value, nil
	}
	s.mutex.Lock()
	data, err := s.trie.TryGet([]byte(key))
	s.mutex.Unlock()
	if err != nil {
		s.Params[key] = 0
		return 0, errors.New(log, fmt.Sprintf("mpt tree get error:%s", err.Error()))
	}
	if len(data) == 0 {
		return 0, nil
	}
	value = common.Uint64SetBytes(data)
	s.Params[key] = value
	return value, nil
}

/**
 *  @brief get the trie root hash
 */
func (s *State) GetHashRoot() common.Hash {
	return common.NewHash(s.trie.Hash().Bytes())
}

/**
 *  @brief commit the trie into past trie list
 */
func (s *State) CommitToMemory() error {
	root, err := s.trie.Commit(nil)
	if err != nil {
		return err
	}
	log.Debug("commit state db to memory:", root.HexString())
	return nil
}

/**
 *  @brief save the information of mpt trie into levelDB
 */
func (s *State) CommitToDB() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.CommitToMemory(); err != nil {
		return err
	}
	return s.db.TrieDB().Commit(s.trie.Hash(), false)
}

/**
 *  @brief reset the mpt state by root hash
 *  @param hash - the hash of mpt witch state will be reset
 */
func (s *State) Reset(hash common.Hash) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.diskDb.Close(); err != nil {
		return err
	}
	diskDb, err := store.NewLevelDBStore(s.path, 0, 0)
	if err != nil {
		return err
	}
	s.db = NewDatabase(diskDb)
	s.trie, err = s.db.OpenTrie(hash)
	if err != nil {
		return err
	}
	s.accMutex.Lock()
	defer s.accMutex.Unlock()
	for k := range s.Accounts {
		delete(s.Accounts, k)
	}
	s.prodMutex.Lock()
	defer s.prodMutex.Unlock()
	for k := range s.Producers {
		delete(s.Producers, k)
	}
	s.tokenMutex.Lock()
	defer s.tokenMutex.Unlock()
	for t := range s.Tokens {
		delete(s.Tokens, t)
	}
		s.paraMutex.Lock()
	defer s.paraMutex.Unlock()
	for k := range s.Params {
		delete(s.Params, k)
	}
	log.Info("Open Trie Hash:", hash.HexString())
	return nil
}

/**
 *  @brief close level db
 */
func (s *State) Close() error {
	return s.diskDb.Close()
}

/**
 *  @brief get the trie
 */
func (s *State) Trie() Trie {
	return s.trie
}
