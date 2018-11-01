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
	"crypto/sha512"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ecoball/go-ecoball/client/common"
	inner "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
)

/*var (
	Wallet = WalletManeger{Wallets: make(map[string]*SoftWallet)}
)

func init() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	Wallet.Dir = strings.Replace(dir, "\\", "/", -1) + "/wallet"
	Wallet.FileExten = ".data"

}*/

type WalletApi interface {
	StoreWallet() error
	loadWallet() error
	Lock() error
	Unlock(password []byte) error
	CreateKey() ([]byte, []byte, error)
	RemoveKey(password []byte, publickey []byte) error
	ImportKey(privateKey []byte) ([]byte, error)
	ListPublicKey() ([]string, error)
	CheckLocked() bool
	CheckPassword(password []byte) bool
	SetLockedState()
	SetUnLockedState()
	ListKeys() map[string]string
	TrySignDigest(digest []byte, publicKey string) (signData []byte, bFind bool)
}

/*type WalletManeger struct {
	Wallets   map[string]*SoftWallet
	Dir       string
	FileExten string
}*/

const INVALID_TIME int64 = -1

var (
	wallets  = make(map[string]WalletApi) // 后台存储所有钱包
	dir      string
	timeout  int64 = INVALID_TIME
	interval int64 = 0
)

func init() {
	rootDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	rootDir = strings.Replace(rootDir, "\\", "/", -1)
	dir = path.Join(rootDir, "wallet/")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			fmt.Println("could not create directory:", dir, err)
		}
	}
}

func Create(name string, password []byte) error {
	checkTimeout()
	//whether the wallet file exists
	filename := path.Join(dir, name)
	if common.FileExisted(filename) {
		return errors.New("The file already exists")
	}

	newWallet := &WalletImpl{
		path:     filename,
		lockflag: unlock,
		KeyData: KeyData{
			Checksum: sha512.Sum512(password),
			//Accounts: []Account{},
			AccountsMap: make(map[string]string),
		},
	}

	//lock wallet
	err := newWallet.Lock()
	newWallet.lockflag = locked
	if nil != err {
		return err
	}

	//write data
	if err := newWallet.StoreWallet(); nil != err {
		return err
	}

	//unlock wallet
	if err := newWallet.Unlock(password); nil != err {
		return err
	}

	newWallet.lockflag = unlock
	wallets[name] = newWallet

	return nil
}

/**
打开钱包
*/
func Open(name string, password []byte) error {
	checkTimeout()
	filename := path.Join(dir, name)
	newWallet := &WalletImpl{
		path:     filename,
		lockflag: unlock,
		KeyData: KeyData{
			//Accounts: []Account{},
			AccountsMap: make(map[string]string),
		},
	}

	//load data
	err := newWallet.loadWallet()
	if nil != err {
		return err
	}

	//unlock wallet
	if err := newWallet.Unlock(password); nil != err {
		return err
	}
	newWallet.lockflag = unlock

	_, ok := wallets[name]

	if ok {
		fmt.Println("exist:", filename)
		delete(wallets, name)
	}

	wallets[name] = newWallet

	return nil
}

func ImportKey(name string, privateKey string) ([]byte, error) {
	checkTimeout()
	wallet, ok := wallets[name]

	if !ok {
		return nil, errors.New("wallet is not exist")
	}

	if wallet.CheckLocked() {
		return nil, errors.New("wallet is locked")
	}

	return wallet.ImportKey([]byte(privateKey))
}

func RemoveKey(name string, password []byte, publickey string) error {
	checkTimeout()
	wallet, ok := wallets[name]

	if !ok {
		return errors.New("wallet is not exist")
	}

	if wallet.CheckLocked() {
		return errors.New("wallet is locked")
	}

	if !wallet.CheckPassword(password) {
		return errors.New("wrong passwords!!")
	}

	return wallet.RemoveKey(password, []byte(publickey))
}

func CreateKey(name string) ([]byte, []byte, error) {
	checkTimeout()
	wallet, ok := wallets[name]

	if !ok {
		return nil, nil, errors.New("wallet is not exist")
	}

	if wallet.CheckLocked() {
		return nil, nil, errors.New("wallet is locked")
	}

	return wallet.CreateKey()
}

func Lock(name string, check bool) error {
	if !check {
		checkTimeout()
	}
	wallet, ok := wallets[name]

	if !ok {
		return errors.New("wallet is not exist")
	}

	if wallet.CheckLocked() {
		return errors.New("wallet is locked")
	}

	wallet.SetLockedState()
	if err := wallet.Lock(); err != nil {
		wallet.SetUnLockedState()
		return err
	}
	return nil
}

func Unlock(name string, password []byte) error {
	checkTimeout()
	wallet, ok := wallets[name]

	if !ok {
		return errors.New("wallet is not exist")
	}

	if !wallet.CheckLocked() {
		return errors.New("wallet is unlocked")
	}

	if !wallet.CheckPassword(password) {
		return errors.New("wrong passwords!!")
	}

	wallet.SetUnLockedState()
	if err := wallet.Unlock(password); err != nil {
		wallet.SetLockedState()
		return err
	}
	return nil
}

func ListKeys(name string, password []byte) (map[string]string, error) {
	checkTimeout()
	wallet, ok := wallets[name]

	if !ok {
		return nil, errors.New("wallet is not exist")
	}

	if wallet.CheckLocked() {
		return nil, errors.New("wallet is unlocked")
	}

	if !wallet.CheckPassword(password) {
		return nil, errors.New("wrong passwords!!")
	}

	return wallet.ListKeys(), nil
}

func GetPublicKeys() ([]string, error) {
	checkTimeout()
	if len(wallets) == 0 {
		return nil, errors.New("You don't have any wallet or no wallet was opened")
	}

	keys := []string{}
	allLocked := true
	for _, wallet := range wallets {
		if wallet.CheckLocked() {
			continue
		}
		allLocked = false
		publicKeys, err := wallet.ListPublicKey()
		if nil != err {
			continue
		}
		keys = append(keys, publicKeys...)
	}

	if allLocked {
		return nil, errors.New("You don't have any unlocked wallet!")
	}

	return keys, nil
}

func ListWallets() ([]string, error) {
	checkTimeout()
	if len(wallets) == 0 {
		return nil, errors.New("You don't have any wallet!")
	}

	keys := []string{}
	for name, wallet := range wallets {
		if !wallet.CheckLocked() {
			name += "*"
		}
		keys = append(keys, name)
	}

	return keys, nil
}

func SignTransaction(transaction []byte, publicKeys []string) ([]byte, error) {
	checkTimeout()
	Transaction := new(types.Transaction)
	if err := Transaction.Deserialize(transaction); err != nil {
		return nil, err
	}

	for _, publicKey := range publicKeys {
		bFound := false
		for _, wallet := range wallets {
			if !wallet.CheckLocked() {
				if signData, bHave := wallet.TrySignDigest(Transaction.Hash.Bytes(), publicKey); bHave {
					/*flag, err := Verify(Transaction.Hash.Bytes(), inner.FromHex(publicKey), signData); if !flag || err != nil {
						fmt.Println(err)
					}*/
					sig := new(inner.Signature)
					sig.PubKey = inner.CopyBytes(inner.FromHex(publicKey))
					sig.SigData = inner.CopyBytes(signData)

					Transaction.Signatures = append(Transaction.Signatures, *sig)
					if !bFound {
						bFound = true
					}
					break
				}
			}
		}

		if !bFound {
			return nil, errors.New("Public key not found in unlocked wallets: " + string(publicKey))
		}
	}

	data, err := Transaction.Serialize()
	if nil != err {
		return nil, err
	}
	return data, nil

}

func SignDigest(data []byte, publicKey string) ([]byte, error) {
	checkTimeout()
	bFound := false
	result := []byte{}
	for _, wallet := range wallets {
		if !wallet.CheckLocked() {
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

func SetTimeout(seconds int64) error {
	if seconds < 0 {
		return fmt.Errorf("Invalid arguments timeout seconds: %d", seconds)
	}
	interval = seconds
	timeout = time.Now().Unix() + seconds
	return nil
}

func checkTimeout() {
	if INVALID_TIME != timeout && time.Now().Unix() > timeout {
		for name, wallet := range wallets {
			if wallet.CheckLocked() {
				continue
			}
			Lock(name, true)
			timeout = time.Now().Unix() + interval
		}
	}
}
