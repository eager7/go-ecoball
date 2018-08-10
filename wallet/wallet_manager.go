// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball.
//
// The go-ecoball is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.
package wallet

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ecoball/go-ecoball/client/common"
)

var (
	Wallet = WalletManeger{Wallets: make(map[string]WalletApi)}
)

func init() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	Wallet.Dir = strings.Replace(dir, "\\", "/", -1) + "/wallet"
	Wallet.FileExten = ".data"

}

type WalletApi interface {
	SetPassword(password []byte) error
	CheckPassword(password []byte) bool
	SetWalletFileName(fileName string)
	IsLocked() bool
	Lock() error
	Unlock(password []byte) error
	SaveWalletFile() error
	LoadWalletFile() error
	ListKeys() (map[string][]byte, error)
	ListPublicKey() ([][]byte, error)
	ImportKey(privateKey []byte) error
	RemoveKey(privateKey []byte) error
	CreateKey() (publicKey []byte, privateKey []byte, err error)
	TrySignDigest(digest []byte, publicKey []byte) (signDigest []byte, bFind bool)
	//GetPrivateKey(publicKey []byte) (privateKey []byte, err error)
}

type WalletManeger struct {
	Wallets   map[string]WalletApi
	Dir       string
	FileExten string
}

func (manager *WalletManeger) Create(name string, password []byte) error {
	//whether the wallet file exists
	fileName := manager.Dir + "/" + name + manager.FileExten
	if _, err := os.Stat(manager.Dir + "/"); os.IsNotExist(err) {
		if err := os.MkdirAll(manager.Dir+"/", 0700); err != nil {
			fmt.Println("could not create directory: ", manager.Dir+"/", " error: ", err)
			return err
		}
	}
	if common.FileExisted(fileName) {
		return errors.New("The wallet file already exists")
	}

	wallet := SoftWallet{Cipherkeys: make([]byte, 10), Keys: make(map[string][]byte)}
	if err := wallet.SetPassword(password); nil != err {
		return err
	}

	wallet.SetWalletFileName(fileName)
	if err := wallet.Unlock(password); nil != err {
		return err
	}
	if err := wallet.Lock(); nil != err {
		return err
	}
	if err := wallet.Unlock(password); nil != err {
		return err
	}

	if err := wallet.SaveWalletFile(); nil != err {
		return err
	}

	manager.Wallets[name] = &wallet

	return nil
}

func (manager *WalletManeger) Open(name string) error {
	fileName := manager.Dir + "/" + name + manager.FileExten
	wallet := SoftWallet{Cipherkeys: make([]byte, 10), Keys: make(map[string][]byte)}
	wallet.SetWalletFileName(fileName)
	if err := wallet.LoadWalletFile(); nil != err {
		return err
	}

	manager.Wallets[name] = &wallet

	return nil
}

func (manager *WalletManeger) ListWallets() []string {
	result := []string{}
	for name, wallet := range manager.Wallets {
		if wallet.IsLocked() {
			result = append(result, name+"*")
		} else {
			result = append(result, name)
		}
	}

	return result
}

func (manager *WalletManeger) ListKeys(name string, password []byte) (map[string][]byte, error) {
	wallet, ok := manager.Wallets[name]
	if !ok {
		return nil, errors.New("Wallet not found: " + name)
	}

	if wallet.IsLocked() {
		return nil, errors.New("Wallet is locked: " + name)
	}

	if !wallet.CheckPassword(password) {
		return nil, errors.New("Wallet password is wrong: " + name)
	}

	return wallet.ListKeys()
}

func (manager *WalletManeger) GetPublicKeys() ([][]byte, error) {
	if len(manager.Wallets) == 0 {
		return nil, errors.New("You don't have any wallet!")
	}

	keys := [][]byte{}
	allLocked := true
	for _, wallet := range manager.Wallets {
		if wallet.IsLocked() {
			continue
		}
		if publicKeys, err := wallet.ListPublicKey(); nil != err {
			if allLocked {
				allLocked = false
			}
			keys = append(keys, publicKeys...)
		}
	}

	if allLocked {
		return nil, errors.New("You don't have any unlocked wallet!")
	}

	return keys, nil
}

func (manager *WalletManeger) LockAll() error {
	for _, wallet := range manager.Wallets {
		if !wallet.IsLocked() {
			if err := wallet.Lock(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (manager *WalletManeger) Lock(name string) error {
	if len(manager.Wallets) == 0 {
		return errors.New("You don't have any wallet!")
	}

	for oneName, wallet := range manager.Wallets {
		if name == oneName {
			if wallet.IsLocked() {
				return nil
			} else {
				return wallet.Lock()
			}
		}
	}

	return errors.New("You don't have wallet: " + name)
}

func (manager *WalletManeger) Unlock(name string, password []byte) error {
	_, ok := manager.Wallets[name]
	if !ok {
		if err := manager.Open(name); nil != err {
			return err
		}
	}

	wallet, _ := manager.Wallets[name]
	if wallet.IsLocked() {
		return errors.New("Wallet is already unlocked: " + name)
	}

	return wallet.Unlock(password)
}

func (manager *WalletManeger) ImportKey(name string, privateKey []byte) error {
	wallet, ok := manager.Wallets[name]
	if !ok {
		return errors.New("Wallet not found: " + name)
	}

	if wallet.IsLocked() {
		return errors.New("Wallet is locked: " + name)
	}

	return wallet.ImportKey(privateKey)
}

func (manager *WalletManeger) RemoveKey(name string, password []byte, privateKey []byte) error {
	wallet, ok := manager.Wallets[name]
	if !ok {
		return errors.New("Wallet not found: " + name)
	}

	if wallet.IsLocked() {
		return errors.New("Wallet is locked: " + name)
	}

	if !wallet.CheckPassword(password) {
		return errors.New("Wallet password is wrong: " + name)
	}

	return wallet.RemoveKey(privateKey)
}

func (manager *WalletManeger) CreateKey(name string) (publicKey []byte, privateKey []byte, err error) {
	wallet, ok := manager.Wallets[name]
	if !ok {
		return nil, nil, errors.New("Wallet not found: " + name)
	}

	if wallet.IsLocked() {
		return nil, nil, errors.New("Wallet is locked: " + name)
	}

	return wallet.CreateKey()
}

func (manager *WalletManeger) SignTransaction(transaction []byte, publicKeys [][]byte) (signTransaction []byte, err error) {
	for _, publicKey := range publicKeys {
		bFound := false
		for _, wallet := range manager.Wallets {
			if !wallet.IsLocked() {
				if signData, bHave := wallet.TrySignDigest(transaction, publicKey); bHave {
					transaction = append(transaction, signData...)
					if !bFound {
						bFound = true
					}
				}
			}
		}

		if !bFound {
			return nil, errors.New("Public key not found in unlocked wallets: " + string(publicKey))
		}
	}

	return signTransaction, nil

}

func (manager *WalletManeger) SignDigest(data []byte, publicKey []byte) ([]byte, error) {
	bFound := false
	result := []byte{}
	for _, wallet := range manager.Wallets {
		if !wallet.IsLocked() {
			if signData, bHave := wallet.TrySignDigest(data, publicKey); bHave {
				if !bFound {
					bFound = true
				}
				result = signData
			}
		}
	}

	if !bFound {
		return nil, errors.New("Public key not found in unlocked wallets: " + string(publicKey))
	}

	return result, nil
}
