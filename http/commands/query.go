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

package commands

import (
	"encoding/json"
	"net/http"
	"strconv"

	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/gin-gonic/gin"
)

func GetMainChainHash(c *gin.Context) {
	//response
	data, err := json.Marshal(&config.ChainHash)
	if nil != err {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"result": string(data)})
	}
}

func GetAllChainInfo(c *gin.Context) {
	//Gets the child chain under the creation chain
	chainList, errcode := ledger.L.GetChainList(config.ChainHash)
	if nil != errcode {
		c.JSON(http.StatusInternalServerError, gin.H{"message": errcode.Error()})
		return
	}

	//Gets all chain under the creation chain and child chain
	allChainInfo := chainList
	for {
		chainListTemp := []state.Chain{}
		for _, oneChain := range chainList {
			oneChainList, err := ledger.L.GetChainList(oneChain.Hash)
			if nil != err {
				continue
			}
			chainListTemp = append(chainListTemp, oneChainList...)
		}

		if 0 == len(chainListTemp) {
			break
		} else {
			allChainInfo = append(allChainInfo, chainListTemp...)
			chainList = chainListTemp
		}
	}

	//response
	data, err := json.Marshal(&allChainInfo)
	if nil != err {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"result": string(data)})
	}
}

func GetAccountInfo(c *gin.Context) {
	name := c.PostForm("name")
	chainHashStr := c.PostForm("chainHash")
	hash := new(innerCommon.Hash)

	data, err := ledger.L.AccountGet(hash.FormHexString(chainHashStr), innerCommon.NameToIndex(name))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": data.JsonString(true)})
}

func GetTokenInfo(c *gin.Context) {
	name := c.PostForm("name")
	chainHashStr := c.PostForm("chainHash")
	hash := new(innerCommon.Hash)

	data, err := ledger.L.GetTokenInfo(hash.FormHexString(chainHashStr), name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": data.JsonString(true)})
}

func GetBlockInfo(c *gin.Context) {
	heightStr := c.PostForm("height")
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	blockInfo, errcode := ledger.L.GetTxBlockByHeight(config.ChainHash, height)
	if errcode != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": errcode.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": blockInfo.JsonString(true)})
}

func GetTransaction(c *gin.Context) {
	hashHex := c.PostForm("hash")
	hash := new(innerCommon.Hash)
	trx, errcode := ledger.L.GetTransaction(config.ChainHash, hash.FormHexString(hashHex))
	if errcode != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": errcode.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": trx.JsonString()})
}

func GetRequiredKeys(c *gin.Context) {
	chainHashStr := c.PostForm("chainHash")
	permission := c.PostForm("permission")
	accountName := c.PostForm("name")

	hash := new(innerCommon.Hash)
	chainHash := hash.FormHexString(chainHashStr)
	data, err := ledger.L.FindPermission(chainHash, innerCommon.NameToIndex(accountName), permission)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	permissionDatas := []state.Permission{}
	if err := json.Unmarshal([]byte(data), &permissionDatas); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	publicAddress := []innerCommon.Address{}
	for _, v := range permissionDatas {
		for _, value := range v.Keys {
			publicAddress = append(publicAddress, value.Actor)
		}
	}

	keys, err := json.Marshal(&publicAddress)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": string(keys)})
}

func GetContract(c *gin.Context) {
	chainId := c.PostForm("chainId")
	accountName := c.PostForm("contractName")
	hash := new(innerCommon.Hash)

	chainids := hash.FormHexString(chainId)
	contract, err := ledger.L.GetContract(chainids, innerCommon.NameToIndex(accountName))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	data, err := contract.Serialize()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": innerCommon.ToHex(data)})
}

func StoreGet(c *gin.Context) {
	chainId := c.PostForm("chainId")
	accountName := c.PostForm("contractName")
	key := c.PostForm("key")

	hash := new(innerCommon.Hash)
	chainids := hash.FormHexString(chainId)
	storage, err := ledger.L.StoreGet(chainids, innerCommon.NameToIndex(accountName), innerCommon.FromHex(key))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": innerCommon.ToHex(storage)})
}
