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
	. "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/store"
	"sync"
)

var log = elog.NewLogger("state", config.LogLevel)
var AbaToken = "ABA"

type State struct {
	Type   TypeState
	path   string
	trie   Trie
	db     Database
	diskDb *store.LevelDBStore

	Tokens    TokensMap
	Accounts  AccountCache
	Params    ParamsMap
	Producers ProducersMap
	Chains    ChainsMap

	mutex sync.RWMutex
}

/**
 *  @brief create a new mpt trie and a levelDB
 *  @param path - the levelDB store path
 *  @param root - the root of mpt trie, this value decide the state of trie
 */
func NewState(path string, root Hash) (st *State, err error) {
	st = &State{
		Type:      0,
		path:      path,
		trie:      nil,
		db:        nil,
		diskDb:    nil,
		Tokens:    new(TokensMap).Initialize(),
		Accounts:  AccountCache{},
		Params:    new(ParamsMap).Initialize(),
		Producers: new(ProducersMap).Initialize(),
		Chains:    new(ChainsMap).Initialize(),
		mutex:     sync.RWMutex{},
	}
	if err := st.Accounts.Initialize(); err != nil {
		return nil, err
	}
	st.diskDb, err = store.NewLevelDBStore(path, 0, 0)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	st.db = NewDatabase(st.diskDb)
	log.Notice("Open Trie Hash:", root.HexString())
	st.trie, err = st.db.OpenTrie(root)
	if err != nil {
		st.trie, _ = st.db.OpenTrie(Hash{})
	}
	return st, nil
}

func (s *State) StateType() TypeState {
	return s.Type
}

/**
 *  @brief copy a new trie into memory
 */
func (s *State) StateCopy() (*State, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stateCp := &State{
		path:      s.path,
		trie:      s.db.CopyTrie(s.trie),
		Tokens:    s.Tokens.Clone(),
		Accounts:  AccountCache{},
		Params:    s.Params.Clone(),
		Producers: s.Producers.Clone(),
		Chains:    s.Chains.Clone(),
	}
	return stateCp, stateCp.Accounts.Initialize()
}

/**
 *  @brief create a new account and store into mpt trie, meanwhile store the mapping of addr and index
 *  @param index - account's index
 *  @param addr - account's address convert from public key
 */
func (s *State) AddAccount(index AccountName, addr Address, timeStamp int64) (*Account, error) {
	s.mutex.RLock()
	data, err := s.trie.TryGet(index.Bytes())
	s.mutex.RUnlock()
	if err != nil {
		return nil, err
	}
	if data != nil {
		return nil, errors.New("reduplicate name")
	}
	acc, err := NewAccount(s.path, index, addr, timeStamp)
	if err != nil {
		return nil, err
	}
	if err := s.commitAccount(acc); err != nil {
		return nil, err
	}
	//save the mapping of addr and index
	s.mutex.Lock()
	err = s.trie.TryUpdate(addr.Bytes(), acc.Index.Bytes())
	s.mutex.Unlock()
	if err != nil {
		return nil, err
	}
	return acc, s.commitParam(addr.HexString(), uint64(index))
}

/**
 *  @brief 通过用户名返回账户结构体,返回的是对象的拷贝,这样可以避免资源竞争
 *  @param index - the account index
 */
func (s *State) GetAccountByName(index AccountName) (*Account, error) {
	acc := s.Accounts.Get(index)
	if acc != nil {
		return acc.Clone()
	}
	s.mutex.Lock()
	fData, err := s.trie.TryGet(index.Bytes())
	s.mutex.Unlock()
	if err != nil {
		return nil, err
	}
	if fData == nil {
		log.Warn(fmt.Sprintf("no this account named:%s", index.String()))
		return nil, errors.New(fmt.Sprintf("no this account named:%s", index.String()))
	}
	acc = &Account{}
	if err = acc.Deserialize(fData); err != nil {
		return nil, err
	}
	return acc.Clone()
}

/**
 *  @brief search the account by address
 *  @param addr - the account address
 */
func (s *State) GetAccountByAddr(addr Address) (*Account, error) {
	if value, err := s.getParam(addr.HexString()); err != nil {
		return nil, err
	} else {
		if value == 0 {
			return nil, errors.New(fmt.Sprintf("the address:%s is not register be an account", addr.HexString()))
		}
		return s.GetAccountByName(AccountName(value))
	}
}

/**
 *  @brief get the trie root hash
 */
func (s *State) GetHashRoot() Hash {
	return NewHash(s.trie.Hash().Bytes())
}

/**
 *  @brief save the information of mpt trie into levelDB
 */
func (s *State) CommitToDB() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.commitToMemory(); err != nil {
		return err
	}
	return s.db.TrieDB().Commit(s.trie.Hash(), false)
}

/**
 *  @brief reset the mpt state by root hash
 *  @param hash - the hash of mpt witch state will be reset
 */
func (s *State) Reset(hash Hash) error {
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
	s.Accounts.Purge()
	s.Producers.Purge()
	s.Chains.Purge()
	s.Tokens.Purge()
	s.Params.Purge()
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
 *  @brief commit the trie into past trie list
 */
func (s *State) commitToMemory() error {
	root, err := s.trie.Commit(nil)
	if err != nil {
		return err
	}
	log.Debug("commit state db to memory:", root.HexString())
	return nil
}

/**
 *  @brief update the account's information into trie
 *  @param acc - account object
 */
func (s *State) commitAccount(acc *Account) error {
	if acc == nil {
		return errors.New("param acc is nil")
	}
	d, err := acc.Serialize()
	if err != nil {
		return err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.trie.TryUpdate(acc.Index.Bytes(), d); err != nil {
		return err
	}
	s.Accounts.Add(acc)
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
	if err := s.trie.TryUpdate([]byte(key), Uint64ToBytes(value)); err != nil {
		return err
	}
	s.Params.Add(key, value)
	return nil
}

/**
 *  @brief get the param's information from trie
 *  @param key - param name
 */
func (s *State) getParam(key string) (uint64, error) {
	param := s.Params.Get(key)
	if param != nil {
		return param.Value, nil
	}
	s.mutex.Lock()
	data, err := s.trie.TryGet([]byte(key))
	s.mutex.Unlock()
	if err != nil {
		s.Params.Add(key, 0)
		return 0, errors.New(fmt.Sprintf("mpt tree get error:%s", err.Error()))
	}
	if len(data) == 0 {
		return 0, nil
	}
	value := Uint64SetBytes(data)
	s.Params.Add(key, value)
	return value, nil
}