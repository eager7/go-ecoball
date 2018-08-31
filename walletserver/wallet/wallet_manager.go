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
	"crypto/sha512"
	//"os"
	//"path/filepath"

	"github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/core/types"
	inner "github.com/ecoball/go-ecoball/common"
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
	RemoveKey(password []byte, publickey string) error
	ImportKey(privateKey string) ([]byte, error)
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
var (
	Wallets = make(map[string]WalletApi) // 后台存储所有钱包
)


func Create(path string, password []byte) error {
	//whether the wallet file exists
	if common.FileExisted(path) {
		return errors.New("The file already exists")
	}

	newWallet := &WalletImpl{
		path:     path,
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
	Wallets[path] = newWallet

	return nil
}

/**
打开钱包
*/
func Open(path string, password []byte) error {
	newWallet := &WalletImpl{
		path:     path,
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

	_, ok := Wallets [ path ]
	
	if ok {
		fmt.Println("exist:", path)
        delete(Wallets, path)
	}
	
	Wallets[path] = newWallet

	return nil
}

func ImportKey(name string, privateKey string)([]byte, error) {
	wallet, ok := Wallets [ name ]
	
	if !ok {
		return nil, errors.New("wallet is not exist")
	}

	if wallet.CheckLocked(){
		return nil, errors.New("wallet is locked")
	}

	return wallet.ImportKey(privateKey)
}

func RemoveKey(name string, password []byte, publickey string) error {
	wallet, ok := Wallets [ name ]
	
	if !ok {
		return errors.New("wallet is not exist")
	}

	if wallet.CheckLocked(){
		return errors.New("wallet is locked")
	}

	if !wallet.CheckPassword(password) {
		return errors.New("wrong passwords!!")
	}

	return wallet.RemoveKey(password, publickey)
}

func CreateKey(name string)([]byte, []byte, error) {
	wallet, ok := Wallets [ name ]
	
	if !ok {
		return nil, nil, errors.New("wallet is not exist")
	}

	if wallet.CheckLocked(){
		return nil, nil, errors.New("wallet is locked")
	}

	return wallet.CreateKey()
}

func Lock(name string) (error) {
	wallet, ok := Wallets [ name ]
	
	if !ok {
		return errors.New("wallet is not exist")
	}

	if wallet.CheckLocked(){
		return errors.New("wallet is locked")
	}

	wallet.SetLockedState()
	if err := wallet.Lock() ; err != nil{
		wallet.SetUnLockedState()
		return err
	}
	return nil
}

func Unlock(name string, password []byte) error {
	wallet, ok := Wallets [ name ]
	
	if !ok {
		return errors.New("wallet is not exist")
	}

	if !wallet.CheckLocked(){
		return errors.New("wallet is unlocked")
	}

	if !wallet.CheckPassword(password) {
		return errors.New("wrong passwords!!")
	}

	wallet.SetUnLockedState()
	if err := wallet.Unlock(password); err != nil{
		wallet.SetLockedState()
		return err
	}
	return nil
}

func ListKeys(name string, password []byte) (map[string]string, error) {
	wallet, ok := Wallets [ name ]
	
	if !ok {
		return nil, errors.New("wallet is not exist")
	}

	if wallet.CheckLocked(){
		return nil, errors.New("wallet is unlocked")
	}

	if !wallet.CheckPassword(password) {
		return nil, errors.New("wrong passwords!!")
	}

	return wallet.ListKeys(), nil
}

func GetPublicKeys() ([]string, error) {
	if len(Wallets) == 0 {
		return nil, errors.New("You don't have any wallet!")
	}

	keys := []string{}
	allLocked := true
	for _, wallet := range Wallets {
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

func List_wallets()([]string, error) {
	if len(Wallets) == 0 {
		return nil, errors.New("You don't have any wallet!")
	}

	keys := []string{}
	for name, wallet := range Wallets {
		if !wallet.CheckLocked() {
			name += "*"
		}
		keys = append(keys, name)
	}

	return keys, nil
}


func SignTransaction(transaction []byte, publicKeys []string) ([]byte, error) {
	Transaction := new(types.Transaction)
	if err := Transaction.Deserialize(transaction); err != nil{
		return nil, err
	}

	for _, publicKey := range publicKeys {
		bFound := false
		for _, wallet := range Wallets {
			if !wallet.CheckLocked() {
				if signData, bHave := wallet.TrySignDigest(transaction, publicKey); bHave {
					sig := new(inner.Signature)
					sig.PubKey = []byte(publicKey)
					sig.SigData = signData

					Transaction.Signatures = append(Transaction.Signatures, *sig)
					if !bFound {
						bFound = true
					}
					break;
				}
			}
		}

		if !bFound {
			return nil, errors.New("Public key not found in unlocked wallets: " + string(publicKey))
		}
	}

	Transaction.Show()
	data, err := Transaction.Serialize()
	if nil != err {
		return nil, err
	}
	return data, nil

}

func SignDigest(data []byte, publicKey string) ([]byte, error) {
	bFound := false
	result := []byte{}
	for _, wallet := range Wallets {
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
