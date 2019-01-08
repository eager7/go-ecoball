package state

import (
	"fmt"
	. "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/types"
)

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
	if err := a.NewMptTrie(path); err != nil {
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
	if err := a.NewMptTrie(path); err != nil {
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
