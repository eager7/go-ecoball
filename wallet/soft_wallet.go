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
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/ecoball/go-ecoball/crypto/aes"
)

type SoftWallet struct {
	FileName   string
	Cipherkeys []byte
	Keys       map[string][]byte
	CheckSum   [64]byte
}

func (wallet *SoftWallet) SetPassword(password []byte) error {
	wallet.CheckSum = sha512.Sum512(password)
	return wallet.Lock()
}

func (wallet *SoftWallet) SetWalletFileName(fileName string) {
	wallet.FileName = fileName
}

func (wallet *SoftWallet) IsLocked() bool {
	return wallet.CheckSum == [64]byte{}
}

func (wallet *SoftWallet) Lock() error {
	if wallet.IsLocked() {
		return errors.New("wallet has been locked")
	}

	if err := wallet.encryptKeys(); nil != err {
		return err
	}

	wallet.CheckSum = [64]byte{}
	wallet.Keys = make(map[string][]byte)
	return nil
}

func (wallet *SoftWallet) encryptKeys() error {
	if !wallet.IsLocked() {
		plain := plainKeys{wallet.Keys, wallet.CheckSum}
		data, err := plain.Serialize()
		if nil != err {
			return err
		}

		//encrypt data
		aesKey := wallet.CheckSum[0:32]
		iv := wallet.CheckSum[32:48]
		cipherkeyTemp, err := aes.AesEncrypt(data, aesKey, iv)
		if err != nil {
			return err
		}
		wallet.Cipherkeys = cipherkeyTemp
	}

	return nil
}

func (wallet *SoftWallet) Unlock(password []byte) error {
	checkSum := sha512.Sum512(password)
	aesKey := checkSum[0:32]
	iv := checkSum[32:48]
	aeskeys, err := aes.AesDecrypt(wallet.Cipherkeys, aesKey, iv)
	if nil != err {
		return err
	}
	var plain plainKeys
	if err := plain.Deserialize(aeskeys); nil != err {
		return err
	}

	wallet.CheckSum = plain.CheckSum
	wallet.Keys = plain.Keys
	return nil
}

func (wallet *SoftWallet) SaveWalletFile() error {
	if err := wallet.encryptKeys(); nil != err {
		return err
	}

	file, err := os.OpenFile(wallet.FileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	if n, err := file.Write(wallet.Cipherkeys); n != len(wallet.Cipherkeys) || err != nil {
		return errors.New("write wallet file error")
	}

	return nil
}

func (wallet *SoftWallet) LoadWalletFile() error {
	//open file
	file, err := os.OpenFile(wallet.FileName, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	//read data
	wallet.Cipherkeys, err = ioutil.ReadAll(file)
	return err
}

func (wallet *SoftWallet) CheckPassword(password []byte) bool {
	checkSum := sha512.Sum512(password)
	aesKey := checkSum[0:32]
	iv := checkSum[32:48]
	aeskeys, err := aes.AesDecrypt(wallet.Cipherkeys, aesKey, iv)
	if nil != err {
		return false
	}

	var plain plainKeys
	if err := plain.Deserialize(aeskeys); nil != err {
		return false
	}

	return checkSum == plain.CheckSum
}

func (wallet *SoftWallet) ListKeys() (map[string][]byte, error) {
	if wallet.IsLocked() {
		return nil, errors.New("Unable to list keys of a locked wallet")
	}

	return wallet.Keys, nil
}

func (wallet *SoftWallet) ListPublicKey() ([][]byte, error) {
	if wallet.IsLocked() {
		return nil, errors.New("Unable to list public keys of a locked wallet")
	}

	keys := [][]byte{}
	for _, publicKey := range wallet.Keys {
		keys = append(keys, publicKey)
	}

	return keys, nil
}

func (wallet *SoftWallet) ImportKey(privateKey []byte) error {
	if wallet.IsLocked() {
		return errors.New("Unable to import key on a locked wallet")
	}

	publickey, err := GetPublicFromPrivate(privateKey)
	if nil != err {
		return err
	}

	_, ok := wallet.Keys[string(privateKey)]
	if ok {
		return errors.New("Key already in wallet")
	}

	wallet.Keys[string(privateKey)] = publickey
	return wallet.SaveWalletFile()
}

func (wallet *SoftWallet) RemoveKey(privateKey []byte) error {
	if wallet.IsLocked() {
		return errors.New("Unable to remove key from a locked wallet")
	}

	_, ok := wallet.Keys[string(privateKey)]
	if !ok {
		return errors.New("Key not in wallet")
	}

	delete(wallet.Keys, string(privateKey))
	return wallet.SaveWalletFile()
}

func (wallet *SoftWallet) CreateKey() (publicKey []byte, privateKey []byte, err error) {
	if wallet.IsLocked() {
		return nil, nil, errors.New("Unable to create key on a locked wallet")
	}

	if privateKey, publicKey, err = createKey(); nil != err {
		return nil, nil, err
	}

	err = wallet.ImportKey(privateKey)
	if nil != err {
		wallet.SaveWalletFile()
	}
	return
}

func (wallet *SoftWallet) TrySignDigest(digest []byte, publicKey []byte) (signData []byte, bFind bool) {
	privateKey := []byte{}
	bFound := false
	for private, public := range wallet.Keys {
		if bytes.Equal(public, publicKey) {
			privateKey = []byte(private)
			bFound = true
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

type plainKeys struct {
	Keys     map[string][]byte
	CheckSum [64]byte
}

func (plain *plainKeys) Serialize() ([]byte, error) {
	return json.Marshal(*plain)
}

func (plain *plainKeys) Deserialize(data []byte) error {
	return json.Unmarshal(data, plain)
}
