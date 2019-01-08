package state

import (
	"fmt"
	. "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
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
	acc, err := s.GetAccountByName(index)
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
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return nil, err
	}
	acc.lock.RLock()
	defer acc.lock.RUnlock()
	return acc.GetContract()
}
func (s *State) StoreSet(index AccountName, key, value []byte) (err error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.lock.Lock()
	defer acc.lock.Unlock()
	if err := acc.StoreSet(s.path, key, value); err != nil {
		return err
	}
	return s.commitAccount(acc)
}
func (s *State) StoreGet(index AccountName, key []byte) (value []byte, err error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return nil, err
	}
	acc.lock.Lock()
	defer acc.lock.Unlock()
	return acc.StoreGet(s.path, key)
}
/**
*  @brief get the abi of contract
*  @param index - account's index
 */
func (s *State) GetContractAbi(index AccountName) ([]byte, error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return nil, err
	}
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
	if err := a.newTrie(path); err != nil {
		return err
	}
	defer func() {
		if err := a.diskDb.Close(); err != nil {
			log.Error("disk db close err:", err)
		}
	}()
	log.Debug("StoreSet key:", string(key), "value:", value)
	if err := a.trie.TryUpdate(key, value); err != nil {
		return err
	}
	if _, err := a.trie.Commit(nil); err != nil {
		return err
	}
	if err := a.db.TrieDB().Commit(a.trie.Hash(), false); err != nil {
		return err
	}
	a.Hash = a.trie.Hash()
	return nil
}

func (a *Account) StoreGet(path string, key []byte) (value []byte, err error) {
	if err := a.newTrie(path); err != nil {
		return nil, err
	}
	defer func() {
		if err := a.diskDb.Close(); err != nil {
			log.Error("disk db close err:", err)
		}
	}()
	value, err = a.trie.TryGet(key)
	if err != nil {
		return nil, err
	}
	log.Debug("StoreGet key:", string(key), "value:", value)
	return value, nil
}
