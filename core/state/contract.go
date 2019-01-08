package state

import (
	"fmt"
	. "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/trie"
	"github.com/ecoball/go-ecoball/core/types"
)

/**
 *  @brief store the smart contract of account, every account only has one contract
 *  @param index - account's index
 *  @param t - the virtual machine type
 *  @param des - the description of contract
 *  @param code - the code of contract
 *  @param abi  - the abi of contract
 */
func (s *State) SetContract(index AccountName, t types.VmType, des, code, abi []byte) error {
	acc, err := s.getAccountByName(index)
	if err != nil {
		return err
	}
	acc.lock.Lock()
	defer acc.lock.Unlock()
	if err := acc.SetContract(t, des, code, abi); err != nil {
		return err
	}
	return s.commitAccount(acc)
}

/**
 *  @brief get the code of account
 *  @param index - account's index
 */
func (s *State) GetContract(index AccountName) (*types.DeployInfo, error) {
	acc, err := s.getAccountByName(index)
	if err != nil {
		return nil, err
	}
	acc.lock.RLock()
	defer acc.lock.RUnlock()
	return acc.GetContract()
}
func (s *State) StoreSet(index AccountName, key, value []byte) (err error) {
	acc, err := s.getAccountByName(index)
	if err != nil {
		return err
	}
	acc.lock.Lock()
	defer acc.lock.Unlock()
	if err := acc.StoreSet(s.Mpt.Path(), key, value); err != nil {
		return err
	}
	return s.commitAccount(acc)
}
func (s *State) StoreGet(index AccountName, key []byte) (value []byte, err error) {
	acc, err := s.getAccountByName(index)
	if err != nil {
		return nil, err
	}
	acc.lock.Lock()
	defer acc.lock.Unlock()
	return acc.StoreGet(s.Mpt.Path(), key)
}

/**
*  @brief get the abi of contract
*  @param index - account's index
 */
func (s *State) GetContractAbi(index AccountName) ([]byte, error) {
	acc, err := s.getAccountByName(index)
	if err != nil {
		return nil, err
	}
	acc.lock.RLock()
	defer acc.lock.RUnlock()
	return acc.Contract.Abi, err
}

/**
 *  @brief add a smart contract into a account data
 *  @param t - the type of virtual machine
 *  @param des - the description of smart contract
 *  @param code - the code of smart contract
 */
func (a *Account) SetContract(t types.VmType, des, code []byte, abi []byte) error {
	a.Contract.TypeVm = t
	a.Contract.Describe = CopyBytes(des)
	a.Contract.Code = CopyBytes(code)
	a.Contract.Abi = CopyBytes(abi)
	return nil
}

/**
 *  @brief get a smart contract from a account data
 */
func (a *Account) GetContract() (*types.DeployInfo, error) {
	if a.Contract.TypeVm == 0 {
		return nil, errors.New(fmt.Sprintf("this account[%s] is not set contract", a.Index.String()))
	}
	return &a.Contract, nil
}

func (a *Account) StoreSet(path string, key, value []byte) (err error) {
	if err := a.TrieOpen(path); err != nil {
		return err
	}
	defer a.TrieClose()
	log.Debug("StoreSet key:", string(key), "value:", value)
	if err := a.mpt.Put(key, value); err != nil {
		return err
	}
	if err := a.mpt.Commit(); err != nil {
		return err
	}
	a.Hash = a.mpt.Hash()
	return nil
}

func (a *Account) StoreGet(path string, key []byte) (value []byte, err error) {
	if err := a.TrieOpen(path); err != nil {
		return nil, err
	}
	defer a.TrieClose()
	value, err = a.mpt.Get(key)
	if err != nil {
		return nil, err
	}
	log.Debug("StoreGet key:", string(key), "value:", string(value))
	return value, nil
}

func (a *Account) TrieOpen(path string) (err error) {
	if a.mpt != nil {
		return nil
	}
	a.mpt, err = trie.NewMptTrie(path+"/"+a.Index.String(), a.Hash)
	if err != nil {
		return err
	}
	return nil
}

func (a *Account) TrieClose() {
	if err := a.mpt.Close(); err != nil {
		log.Error("disk db close err:", err)
	}
	a.mpt = nil
}
