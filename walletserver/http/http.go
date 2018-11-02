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
	"strconv"
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
	name := c.PostForm("name")
	password := c.PostForm("password")
	if err := wallet.Create(name, []byte(password)); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func createKey(c *gin.Context) {
	name := c.PostForm("name")
	pub, pri, err := wallet.CreateKey(name)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	info := KeyPair{PrivateKey: pri, PublicKey: pub}
	c.JSON(http.StatusOK, info)
}

func openWallet(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")
	if err := wallet.Open(name, []byte(password)); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func lockWallet(c *gin.Context) {
	name := c.PostForm("name")
	if err := wallet.Lock(name, false); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func unlockWallet(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")
	if err := wallet.Unlock(name, []byte(password)); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func importKey(c *gin.Context) {
	name := c.PostForm("name")
	privateKey := c.PostForm("privateKey")
	publickey, err := wallet.ImportKey(name, privateKey)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	publicKey := OneKey{publickey}
	c.JSON(http.StatusOK, publicKey)
}

func removeKey(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")
	publickey := c.PostForm("publickey")
	err := wallet.RemoveKey(name, []byte(password), publickey)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func listKey(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")
	accounts, err := wallet.ListKeys(name, []byte(password))
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
	strInterval := c.PostForm("interval")
	interval, err := strconv.ParseInt(strInterval, 10, 64)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	err = wallet.SetTimeout(interval)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}
