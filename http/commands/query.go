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

	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/http/request"
	"github.com/gin-gonic/gin"
)

func GetMainChainHash(c *gin.Context) {
	//response
	c.JSON(http.StatusOK, config.ChainHash)
}

func GetAllChainInfo(c *gin.Context) {
	//Gets the child chain under the creation chain
	chainList := []state.Chain{}
	mainChain := state.Chain{Hash: config.ChainHash}
	chainList = append(chainList, mainChain)

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
	c.JSON(http.StatusOK, allChainInfo)
}

func GetAccountInfo(c *gin.Context) {
	var oneAccount request.AccountName
	if err := c.BindJSON(&oneAccount); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	data, err := ledger.L.AccountGet(oneAccount.ChainHash, innerCommon.NameToIndex(oneAccount.Name))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, *data)
}

func GetTokenInfo(c *gin.Context) {
	var oneToken request.TokenName
	if err := c.BindJSON(&oneToken); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	data, err := ledger.L.GetTokenInfo(oneToken.ChainHash, oneToken.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, *data)
}

func GetBlockInfo(c *gin.Context) {
	var oneHeight request.BlockHeight
	if err := c.BindJSON(&oneHeight); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	blockInfo, errcode := ledger.L.GetTxBlockByHeight(oneHeight.ChainHash, oneHeight.Height)
	if errcode != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": errcode.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": blockInfo.JsonString(true)})
}

func GetTransaction(c *gin.Context) {
	var oneHash request.TransactionHash
	if err := c.BindJSON(&oneHash); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	trx, errcode := ledger.L.GetTransaction(oneHash.ChainHash, oneHash.Hash)
	if errcode != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": errcode.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": trx.JsonString()})
}

func GetRequiredKeys(c *gin.Context) {
	var onePermission request.PermissionPublicKeys
	if err := c.BindJSON(&onePermission); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	data, err := ledger.L.FindPermission(onePermission.ChainHash, innerCommon.NameToIndex(onePermission.Name), onePermission.Permission)
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

	c.JSON(http.StatusOK, publicAddress)
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
