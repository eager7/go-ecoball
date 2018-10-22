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

package rpc

import (
	//"fmt"
	"net/http"
	"strings"

	"github.com/ecoball/go-ecoball/common/config"
	"github.com/gin-gonic/gin"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/core/state"
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/event"
)

func StartHttpServer() (err error) {
	//get router instance
	router := gin.Default()

	//register handle
	router.POST("/getAccountInfo", getAccountInfo)
	router.GET("/getInfo", getInfo)
	router.POST("/get_required_keys", get_required_keys)
	router.POST("/invokeContract", invokeContract)
	router.POST("/setContract", setContract)
	router.POST("/getContract", getContract)
	router.POST("/storeGet", storeGet)
	router.POST("/transfer", transfer)

	http.ListenAndServe(":20681", router)
	return nil
}

func getAccountInfo(c *gin.Context) {
	name := c.PostForm("name")
	chainId_str := c.PostForm("chainId")
	hash := new(innerCommon.Hash)

	data, err := ledger.L.AccountGet(hash.FormHexString(chainId_str), innerCommon.NameToIndex(name))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": data.JsonString(false)})
}

func getInfo(c *gin.Context) {
	var height uint64 = 1
	blockInfo, errcode := ledger.L.GetTxBlockByHeight(config.ChainHash, height)
	if errcode != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": errcode.Error()})
		return
	}
	
	data, errs := blockInfo.Serialize()
	if errs != nil{
		c.JSON(http.StatusBadRequest, gin.H{"result": errs.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": innerCommon.ToHex(data)})
}

func get_required_keys(c *gin.Context) {
	chainId := c.PostForm("chainId")
	required_keys := c.PostForm("keys")
	permission := c.PostForm("permission")
	transaction_data := c.PostForm("transaction")

	key_datas := strings.Split(required_keys, "\n")
	Transaction := new(types.Transaction)
	if err := Transaction.Deserialize(innerCommon.FromHex(transaction_data)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}
	//signTransaction, err := wallet.SignTransaction(inner.FromHex(transaction_data), datas)
	hash := new(innerCommon.Hash)
	chainids := hash.FormHexString(chainId)
	data, err := ledger.L.FindPermission(chainids, Transaction.From, permission)
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	permission_datas := []state.Permission{}
	if err := json.Unmarshal([]byte(data), &permission_datas); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	public_address := []innerCommon.Address{}
	for _, v := range permission_datas {
		for _, value:= range v.Keys{
			public_address = append(public_address, value.Actor)
		}
	}

	publickeys := ""
	for _, v := range key_datas {
		addr := innerCommon.AddressFromPubKey(innerCommon.FromHex(v))
		for _, vv := range public_address {
			if addr == vv {
				publickeys += v
				publickeys += "\n"
				break
			}
		}
	}
	if "" != publickeys {
		publickeys = strings.TrimSuffix(publickeys, "\n")
		c.JSON(http.StatusOK, gin.H{"result": publickeys})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"result": "no required_keys"})
}

func invokeContract(c *gin.Context) {
	invoke := new(types.Transaction)//{
	transaction_data := c.PostForm("transaction")
	
	if err := invoke.Deserialize(innerCommon.FromHex(transaction_data)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return 
	}
	
	//send to txpool
	err := event.Send(event.ActorNil, event.ActorTxPool, invoke)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return 
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func setContract(c *gin.Context) {
	deploy := new(types.Transaction)//{
	transaction_data := c.PostForm("transaction")
	
	if err := deploy.Deserialize(innerCommon.FromHex(transaction_data)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return 
	}
	
	//send to txpool
	err := event.Send(event.ActorNil, event.ActorTxPool, deploy)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return 
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func getContract(c *gin.Context) {
	chainId := c.PostForm("chainId")
	accountName := c.PostForm("contractName")
	hash := new(innerCommon.Hash)

	chainids := hash.FormHexString(chainId)
	contract, err := ledger.L.GetContract(chainids, innerCommon.NameToIndex(accountName))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return 
	}

	data, err := contract.Serialize()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return 
	}

	c.JSON(http.StatusOK, gin.H{"result": innerCommon.ToHex(data)})
}

func storeGet(c *gin.Context) {
	chainId := c.PostForm("chainId")
	accountName := c.PostForm("contractName")
	key := c.PostForm("key")

	hash := new(innerCommon.Hash)
	chainids := hash.FormHexString(chainId)
	storage, err := ledger.L.StoreGet(chainids, innerCommon.NameToIndex(accountName), innerCommon.FromHex(key))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return 
	}

	c.JSON(http.StatusOK, gin.H{"result": innerCommon.ToHex(storage)})
}

func transfer(c *gin.Context) {
	transfer := new(types.Transaction)//{
	transaction_data := c.PostForm("transfer")
		
	if err := transfer.Deserialize(innerCommon.FromHex(transaction_data)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return 
	}
		
	//send to txpool
	err := event.Send(event.ActorNil, event.ActorTxPool, transfer)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return 
	}
	
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

