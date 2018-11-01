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
	//"bytes"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	inner "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/crypto/aes"
)

const (
	unlock byte = 0 //钱包未锁
	locked byte = 1 //钱包已锁
)

type KeyData struct {
	Checksum    [64]byte `json:"Checksum"`
	AccountsMap map[string]string
}

type WalletImpl struct {
	path string
	KeyData
	lockflag   byte
	Cipherkeys []byte //存储加密后的数据
}

/**
方法：内存数据存储到钱包文件中
*/
func (wi *WalletImpl) StoreWallet() error {
	//open file
	file, err := os.OpenFile(wi.path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	//write data
	data := wi.Cipherkeys
	n, err := file.Write(data)
	if n != len(data) || err != nil {
		return err
	}

	return nil
}

/**
方法：将钱包文件的数据导入到内存中
*/
func (wi *WalletImpl) loadWallet() error {
	//open file
	file, err := os.OpenFile(wi.path, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	//read data
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	wi.Cipherkeys = data

	return nil
}

/**
方法：将密钥数据加密
*/
func (wi *WalletImpl) Lock() error {
	//whether the wallet is locked
	/*if wi.lockflag != unlock {
		return errors.New("the wallet has been locked!!")
	}*/

	//marshal keyData
	data, err := json.Marshal(wi.KeyData)
	if nil != err {
		return err
	}

	//encrypt data
	aesKey := wi.Checksum[0:32]
	iv := wi.Checksum[32:48]
	cipherkeyTemp, err := aes.AesEncrypt(data, aesKey, iv)
	if err != nil {
		return err
	}

	//erase data
	/*for i := 0; i < len(wi.Checksum); i++ {
		wi.Checksum[i] = 0
	}*/
	//wi.Accounts = []Account{}

	//wi.lockflag = locked

	wi.Cipherkeys = cipherkeyTemp

	return nil
}

func (wi *WalletImpl) CheckPassword(password []byte) bool {
	return sha512.Sum512(password) == wi.Checksum
}

func (wi *WalletImpl) SetLockedState() {
	wi.lockflag = locked
}

func (wi *WalletImpl) SetUnLockedState() {
	wi.lockflag = unlock
}

/**
方法：将密钥数据解密
*/
func (wi *WalletImpl) Unlock(password []byte) error {
	//Decrypt data
	checksum := sha512.Sum512(password)
	aesKey := checksum[0:32]
	iv := checksum[32:48]
	aeskeys, err := aes.AesDecrypt(wi.Cipherkeys, aesKey, iv)
	if nil != err {
		return err
	}

	//unmarshal data
	wallet := *wi
	str := string(aeskeys)
	result := strings.Index(str, "}}")
	if len(str) > (result + 2) { //代表有脏数据，需要截取
		content := str[0 : result+2]
		aeskeys = []byte(content)
	}
	if err := json.Unmarshal(aeskeys, &wi.KeyData); nil != err {
		*wi = wallet
		return errors.New("Unmarshal faild ! maybe the password is wrong")
	}

	//check password
	if wi.Checksum != checksum {
		*wi = wallet
		return errors.New("password error")
	}
	//wi.lockflag = unlock
	wi.Cipherkeys = nil

	return nil
}

func (wi *WalletImpl) ListKeys() map[string]string {
	return wi.AccountsMap
}

/**
创建公私钥对
*/
func (wi *WalletImpl) CreateKey() ([]byte, []byte, error) {
	//create keys
	_, pri, err := createKey()
	if err != nil {
		return nil, nil, err
	}

	pub, errcode := wi.ImportKey(pri)
	if errcode != nil {
		return nil, nil, errcode
	}
	return pub, pri, nil
}

func (wi *WalletImpl) RemoveKey(password []byte, publickey []byte) error {
	wi.lockflag = locked
	_, ok := wi.AccountsMap[string(publickey)]

	if !ok {
		wi.lockflag = unlock
		return errors.New("publickey is not exist")
	}

	delete(wi.KeyData.AccountsMap, string(publickey))

	errcode := wi.Lock()
	if nil != errcode {
		wi.lockflag = unlock
		return errcode
	}

	//write data
	if err := wi.StoreWallet(); nil != err {
		wi.lockflag = unlock
		return err
	}

	//unlock wallet
	if err := wi.Unlock(password); nil != err {
		wi.lockflag = unlock
		return err
	}

	wi.lockflag = unlock
	return nil
}

/**
导入私钥
**/
func (wi *WalletImpl) ImportKey(privateKey []byte) ([]byte, error) {
	wi.lockflag = locked

	for publickey := range wi.AccountsMap {
		if strings.EqualFold(wi.AccountsMap[publickey], string(privateKey)) {
			wi.lockflag = unlock
			return nil, errors.New("private has exist")
		}
	}

	//export publickey by privatekey
	pub, err := getPublicKey(privateKey)
	if err != nil {
		wi.lockflag = unlock
		return nil, errors.New("get publickey error: " + err.Error())
	}

	//wi.KeyData.Accounts = append(wi.KeyData.Accounts, account)
	wi.KeyData.AccountsMap[inner.ToHex(pub)] = string(privateKey)

	//lock wallet
	errcode := wi.Lock()
	if nil != errcode {
		wi.lockflag = unlock
		return nil, errcode
	}

	//write data
	if err := wi.StoreWallet(); nil != err {
		wi.lockflag = unlock
		return nil, err
	}

	wi.lockflag = unlock
	return pub, nil
}

func (wallet *WalletImpl) ListPublicKey() ([]string, error) {
	if wallet.CheckLocked() {
		return nil, errors.New("Unable to list public keys of a locked wallet")
	}

	keys := []string{}
	for publicKey, _ := range wallet.AccountsMap {
		keys = append(keys, publicKey)
	}

	return keys, nil
}

/**
判断是否为锁定状态
**/
func (wi *WalletImpl) CheckLocked() bool {
	return wi.lockflag == locked
}

func (wallet *WalletImpl) TrySignDigest(digest []byte, publicKey string) (signData []byte, bFind bool) {
	privateKey := []byte{}
	bFound := false
	for public, private := range wallet.AccountsMap {
		if strings.EqualFold(public, publicKey) {
			privateKey = inner.FromHex(private)
			bFound = true
			break
		}
	}

	if !bFound {
		return nil, false
	}

	data, err := signDigest(digest, privateKey)
	if nil != err {
		return nil, false
	}
	return data, true
}
