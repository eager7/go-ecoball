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

package http

import (
	"net/http"
	"strings"

	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/walletserver/wallet"
	"github.com/gin-gonic/gin"
)

func StartHttpServer() (err error) {
	//get router instance
	router := gin.Default()

	//register handle
	router.GET("/wallet/attach", attach)
	router.POST("/wallet/create", createWallet)
	router.POST("/wallet/createKey", createKey)
	router.POST("/wallet/openWallet", openWallet)
	router.POST("/wallet/lockWallet", lockWallet)
	router.POST("/wallet/unlockWallet", unlockWallet)
	router.POST("/wallet/importKey", importKey)
	router.POST("/wallet/removeKey", removeKey)
	router.POST("/wallet/listKey", listKey)
	router.GET("/wallet/listWallets", listWallets)
	router.GET("/wallet/getPublicKeys", getPublicKeys)
	router.POST("/wallet/signTransaction", signTransaction)
	router.POST("/wallet/setTimeout", setTimeout)

	http.ListenAndServe(":"+config.WalletHttpPort, router)
	return nil
}

func attach(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func createWallet(c *gin.Context) {
	var oneWallet WalletNamePassword
	if err := c.BindJSON(&oneWallet); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := wallet.Create(oneWallet.Name, []byte(oneWallet.Password)); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func createKey(c *gin.Context) {
	var oneWallet WalletName
	if err := c.BindJSON(&oneWallet); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	pub, pri, err := wallet.CreateKey(oneWallet.Name)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	info := KeyPair{PrivateKey: pri, PublicKey: pub}
	c.JSON(http.StatusOK, info)
}

func openWallet(c *gin.Context) {
	var oneWallet WalletNamePassword
	if err := c.BindJSON(&oneWallet); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := wallet.Open(oneWallet.Name, []byte(oneWallet.Password)); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func lockWallet(c *gin.Context) {
	var oneWallet WalletName
	if err := c.BindJSON(&oneWallet); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := wallet.Lock(oneWallet.Name, false); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func unlockWallet(c *gin.Context) {
	var oneWallet WalletNamePassword
	if err := c.BindJSON(&oneWallet); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := wallet.Unlock(oneWallet.Name, []byte(oneWallet.Password)); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func importKey(c *gin.Context) {
	var oneWallet WalletImportKey
	if err := c.BindJSON(&oneWallet); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	publickey, err := wallet.ImportKey(oneWallet.Name, oneWallet.PriKey.Key)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	publicKey := OneKey{publickey}
	c.JSON(http.StatusOK, publicKey)
}

func removeKey(c *gin.Context) {
	var oneWallet WalletRemoveKey
	if err := c.BindJSON(&oneWallet); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := wallet.RemoveKey(oneWallet.NamePassword.Name, []byte(oneWallet.NamePassword.Password), oneWallet.PubKey.Key)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func listKey(c *gin.Context) {
	var oneWallet WalletNamePassword
	if err := c.BindJSON(&oneWallet); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	accounts, err := wallet.ListKeys(oneWallet.Name, []byte(oneWallet.Password))
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var pairs = KeyPairs{Pairs: []KeyPair{}}
	for k, v := range accounts {
		onePair := KeyPair{PublicKey: []byte(k), PrivateKey: []byte(v)}
		pairs.Pairs = append(pairs.Pairs, onePair)
	}

	c.JSON(http.StatusOK, pairs)
}

func listWallets(c *gin.Context) {
	wallets, err := wallet.ListWallets()
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	wallet := Wallets{wallets}

	c.JSON(http.StatusOK, wallet)
}

func getPublicKeys(c *gin.Context) {
	data, err := wallet.GetPublicKeys()
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if len(data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no publickeys"})
		return
	}

	publicKeys := Keys{KeyList: []OneKey{}}
	for _, k := range data {
		publicKeys.KeyList = append(publicKeys.KeyList, OneKey{[]byte(k)})
	}

	c.JSON(http.StatusOK, publicKeys)
}

func signTransaction(c *gin.Context) {
	keys := c.PostForm("keys")
	data := c.PostForm("transaction")
	key := strings.Split(keys, "\n")
	signData, err := wallet.SignTransaction([]byte(data), key)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	resultData := TransactionData{signData}

	c.JSON(http.StatusOK, resultData)
}

func setTimeout(c *gin.Context) {
	var oneWallet WalletTimeout
	if err := c.BindJSON(&oneWallet); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := wallet.SetTimeout(oneWallet.Interval)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}
