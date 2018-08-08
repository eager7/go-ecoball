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
package account

import (
	//"bytes"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ecoball/go-ecoball/client/common"
	inner "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/crypto/aes"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
)

const (
	unlock byte = 0 //钱包未锁
	locked byte = 1 //钱包已锁
)

type KeyData struct {
	Checksum [64]byte  `json:"Checksum"`
	//Accounts []Account `json:"Accounts"`
	AccountsMap map[string]string
}

type WalletImpl struct {
	path string
	KeyData
	lockflag byte
	Cipherkeys []byte  //存储加密后的数据
}

var (
	//Wallet *WalletImpl //存储当前打开的钱包
	Wallets = make(map[string]*WalletImpl) //
)

/**
创建钱包
*/
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
	err := newWallet.Lock(password)
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

	Wallets[path] = newWallet
	newWallet.lockflag = unlock

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

	wallet, ok := Wallets [ path ]
	
	if ok {
		fmt.Println("exist:", wallet.path)
        delete(Wallets, path)
	}
	
	Wallets[path] = newWallet

	return nil
}

func ImportKey2Wallet(name string, password []byte, privateKey string)([]byte, error) {
	wallet, ok := Wallets [ name ]
	
	if !ok {
		return nil, errors.New("wallet is not exist")
	}

	if wallet.CheckLocked(){
		return nil, errors.New("wallet is locked")
	}

	if (sha512.Sum512(password)) != wallet.Checksum {
		return nil, errors.New("wrong passwords!!")
	}

	return wallet.ImportKey(password, privateKey)
}

func RemoveSpringSliceCopy(slice []Account, start,end int) []Account {
    result := make([]Account, len(slice)-(end-start))
    at :=copy(result, slice[:start])
    copy(result[at:], slice[end:])
    return result
}

func RemoveKeyFromWallet(name string, password []byte, publickey string) error {
	wallet, ok := Wallets [ name ]
	
	if !ok {
		return errors.New("wallet is not exist")
	}

	if wallet.CheckLocked(){
		return errors.New("wallet is locked")
	}

	if (sha512.Sum512(password)) != wallet.Checksum {
		return errors.New("wrong passwords!!")
	}

	return wallet.RemoveKey(password, publickey)
}

func CreateKey2Wallet(name string, password []byte)(Account, error) {
	wallet, ok := Wallets [ name ]
	
	if !ok {
		return Account{}, errors.New("wallet is not exist")
	}

	if wallet.CheckLocked(){
		return Account{}, errors.New("wallet is locked")
	}

	if (sha512.Sum512(password)) != wallet.Checksum {
		return  Account{},errors.New("wrong passwords!!")
	}

	return wallet.CreateKey(password)
}

func LockWallet(name string, password []byte) (error) {
	wallet, ok := Wallets [ name ]
	
	if !ok {
		return errors.New("wallet is not exist")
	}

	if wallet.CheckLocked(){
		return errors.New("wallet is locked")
	}

	wallet.lockflag = locked
	if err := wallet.Lock(password) ; err != nil{
		wallet.lockflag = unlock
		return err
	}
	return nil
}

func UnlockWallet(name string, password []byte) error {
	wallet, ok := Wallets [ name ]
	
	if !ok {
		return errors.New("wallet is not exist")
	}

	if !wallet.CheckLocked(){
		return errors.New("wallet is unlocked")
	}

	wallet.lockflag = unlock
	if err := wallet.Unlock(password); err != nil{
		wallet.lockflag = locked
		return err
	}
	return nil
}

func ListAccountFromWallet(name string, password []byte) ([]Account, error) {
	wallet, ok := Wallets [ name ]
	
	if !ok {
		return nil, errors.New("wallet is not exist")
	}

	if wallet.CheckLocked(){
		return nil, errors.New("wallet is unlocked")
	}

	//return wallet.Accounts, nil
	return nil, nil
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
func (wi *WalletImpl) Lock(password []byte) error {
	//whether the wallet is locked
	/*if wi.lockflag != unlock {
		return errors.New("the wallet has been locked!!")
	}*/

	//whether the password is correct
	if (sha512.Sum512(password)) != wi.Checksum {
		return errors.New("wrong password!!")
	}

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
	for i := 0; i < len(wi.Checksum); i++ {
		wi.Checksum[i] = 0
	}
	//wi.Accounts = []Account{}

	//wi.lockflag = locked

	wi.Cipherkeys = cipherkeyTemp

	return nil
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
	result := strings.Index(str,"}}")
	if len(str) > (result+2) {//代表有脏数据，需要截取
		content := str[0 : result+2]
		aeskeys = []byte(content)
	}
	if err := json.Unmarshal(aeskeys, &wi.KeyData); nil != err {
		*wi = wallet
		return err
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

/**
创建公私钥对
*/
func (wi *WalletImpl) CreateKey(password []byte) (Account, error) {
	//create keys
	ac, err := NewAccount(0)
	if err != nil {
		return Account{}, err
	}

	wi.lockflag = locked
	/*for _, v := range wi.Accounts {
		if v.Equal(ac) {
			wi.lockflag = unlock
			return Account{}, errors.New("key has exist")
		}
	}

	wi.Accounts = append(wi.Accounts, ac)*/
	wi.KeyData.AccountsMap[inner.ToHex(ac.PublicKey)] = inner.ToHex(ac.PrivateKey)
	
	//lock wallet
	errcode := wi.Lock(password)
	if nil != errcode {
		wi.lockflag = unlock
		return Account{}, errcode
	}

	//write data
	if err := wi.StoreWallet(); nil != err {
		wi.lockflag = unlock
		return Account{}, err
	}

	//unlock wallet
	if err := wi.Unlock(password); nil != err {
		wi.lockflag = unlock
		return Account{}, err
	}

	wi.lockflag = unlock
	return ac, nil
}

func (wi *WalletImpl) RemoveKey(password []byte, publickey string) error {
	/*var index int
	bFound := false*/
	wi.lockflag = locked
	/*for i,v := range wi.Accounts {
		if strings.EqualFold(inner.	ToHex(v.PublicKey), publickey) {
			index = i
			bFound = true
			break
		}
	}
	
	if !bFound {
		wi.lockflag = unlock
		return errors.New("publickey no found")
	}
	accs := wi.Accounts
	wi.Accounts = []Account{}
	wi.Accounts = RemoveSpringSlice(accs, index, index+1)*/
	//wi.Accounts = []Account{}

	_, ok := wi.AccountsMap [ publickey ]
	
	if !ok {
		wi.lockflag = unlock
		return errors.New("publickey is not exist")
	}
	
	for v := range wi.AccountsMap {
		fmt.Println(v)
		fmt.Println(wi.AccountsMap[v])
	}

	delete(wi.KeyData.AccountsMap, publickey)

	for v := range wi.AccountsMap {
		fmt.Println(v)
		fmt.Println(wi.AccountsMap[v])
	}

	errcode := wi.Lock(password)
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
func (wi *WalletImpl) ImportKey(password []byte, privateKey string) ([]byte, error) {
	wi.lockflag = locked
	//ac := Account{}
	/*for _,v := range wi.Accounts {
		if bytes.Equal(v.PrivateKey[:], privateKey[:]) {
			ac = v
			wi.lockflag = unlock
			return ac.PublicKey, errors.New("privatekey has exist")
		}
	}*/

	for publickey := range wi.AccountsMap {
		if strings.EqualFold(wi.AccountsMap[publickey], privateKey) {
			wi.lockflag = unlock
			return nil, errors.New("private has exist")
		}
	}

	//export publickey by privatekey 
	pri, err := secp256k1.ToECDSA([]byte(privateKey))
	if err != nil {
		wi.lockflag = unlock
		return nil, errors.New("NewECDSAPrivateKey error: " + err.Error())
	}
	pub, err := secp256k1.FromECDSAPub(&pri.PublicKey)
	if err != nil {
		wi.lockflag = unlock
		return nil, errors.New("new account error: " + err.Error())
	}

	account := Account{
		PrivateKey: []byte(privateKey),
		PublicKey:  pub,
		Alg:        0,
	}
	//wi.KeyData.Accounts = append(wi.KeyData.Accounts, account)
	wi.KeyData.AccountsMap[inner.ToHex(account.PublicKey)] = privateKey

	//lock wallet
	errcode := wi.Lock(password)
	if nil != errcode {
		wi.lockflag = unlock
		return nil, errcode
	}
	
	//write data
	if err := wi.StoreWallet(); nil != err {
		wi.lockflag = unlock
		return nil, err
	}
	
	//unlock wallet
	if err := wi.Unlock(password); nil != err {
		wi.lockflag = unlock
		return nil, err
	}
	wi.lockflag = unlock
	return account.PublicKey, nil
}

func RemoveSpringSlice(slice []Account, start,end int) []Account {
	return append(slice[:start], slice[end:]...)
}
/**
列出所有账号
*/
/*func (wi *WalletImpl) ListAccount() {
	for _, v := range wi.Accounts {
		fmt.Println("PrivateKey: ", inner.ToHex(v.PrivateKey[:]))
		fmt.Println("PublicKey: ", inner.ToHex(v.PublicKey[:]))
	}
}*/

/**
判断是否为锁定状态
**/
func (wi *WalletImpl) CheckLocked() bool {
	return wi.lockflag == locked
}
