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
	"net/http"
	"strconv"
	"strings"
	"time"
	//"os"
	//"io/ioutil"
	"encoding/base64"

	"encoding/json"

	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/http/commands"
	"github.com/ecoball/go-ecoball/http/common/abi"
	"github.com/gin-gonic/gin"
)

func StartHttpServer() (err error) {
	//get router instance
	router := gin.Default()

	//register handle
	router.GET("/getHeadBlock", getHeadBlock)
	router.POST("/invokeContract", invokeContract)
	router.POST("/setContract", setContract)
	router.POST("/getContract", getContract)
	router.POST("/storeGet", storeGet)
	router.POST("/newInvokeContract", newInvokeContract)
	router.POST("/newDeployContract", newDeployContract)
	//for invokContract
	router.POST("/newContract", newContract)
	//router.POST("/recieveFile", recieveFile)

	//attach
	router.GET("/attach", attach)
	router.GET("/getInfo", getInfo)

	//query information
	router.GET("/query/mainChainHash", commands.GetMainChainHash)
	router.GET("/query/allChainInfo", commands.GetAllChainInfo)
	router.POST("/query/getAccountInfo", commands.GetAccountInfo)
	router.POST("/query/getTokenInfo", commands.GetTokenInfo)
	router.POST("/query/getBlockInfo", commands.GetBlockInfo)
	router.POST("/query/getTransaction", commands.GetTransaction)
	router.POST("/query/getRequiredKeys", commands.GetRequiredKeys)

	//transfer
	router.POST("/transfer", commands.Transfer)

	http.ListenAndServe(":"+config.HttpLocalPort, router)
	return nil
}

func attach(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func getInfo(c *gin.Context) {
	var height uint64 = 1
	blockInfo, errcode := ledger.L.GetTxBlockByHeight(config.ChainHash, height)
	if errcode != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": errcode.Error()})
		return
	}

	data, errs := blockInfo.Serialize()
	if errs != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": errs.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": innerCommon.ToHex(data)})
}

func invokeContract(c *gin.Context) {
	invoke := new(types.Transaction) //{
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
	deploy := new(types.Transaction) //{
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

//use for scan
func getHeadBlock(c *gin.Context) {
	var height uint64 = 1
	blockInfo, errcode := ledger.L.GetTxBlockByHeight(config.ChainHash, height)
	if errcode != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": errcode.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success", "chainId": blockInfo.ChainID.HexString()})
}

func newInvokeContract(c *gin.Context) {
	chainId_str := c.PostForm("chainId")
	accountName := c.PostForm("accountName")
	creator := c.PostForm("creator")
	owner := c.PostForm("owner")

	if "" == chainId_str || "" == accountName || "" == creator || "" == owner {
		c.JSON(http.StatusBadRequest, gin.H{"result": "invalid params"})
		return
	}

	max_cpu_usage_ms, err := strconv.ParseFloat(c.PostForm("max-cpu-usage-ms"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	max_net_usage, err := strconv.ParseFloat(c.PostForm("max-net-usage"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	creatorAccount := innerCommon.NameToIndex(creator)
	timeStamp := time.Now().UnixNano()

	hash := new(innerCommon.Hash)
	chainId := hash.FormHexString(chainId_str)

	invoke, err := types.NewInvokeContract(creatorAccount, creatorAccount, chainId, "owner", "new_account",
		[]string{accountName, innerCommon.AddressFromPubKey(innerCommon.FromHex(owner)).HexString()}, 0, timeStamp)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	invoke.Receipt.Cpu = max_cpu_usage_ms
	invoke.Receipt.Net = max_net_usage

	data, err := invoke.Serialize()
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"result": "success", "invoke": innerCommon.ToHex(data)})
}

func newContract(c *gin.Context) {
	chainId_str := c.PostForm("chainId")
	contractName := c.PostForm("name")
	contractMethod := c.PostForm("method")
	contractParam := c.PostForm("params")

	if "" == chainId_str || "" == contractName || "" == contractMethod ||
		"" == contractParam {
		c.JSON(http.StatusBadRequest, gin.H{"result": "invalid params"})
		return
	}

	hash := new(innerCommon.Hash)
	chainId := hash.FormHexString(chainId_str)

	var parameters []string
	if "new_account" == contractMethod {
		parameter := strings.Split(contractParam, ",")
		for _, v := range parameter {
			if strings.Contains(v, "0x") {
				parameters = append(parameters, innerCommon.AddressFromPubKey(innerCommon.FromHex(v)).HexString())
			} else {
				parameters = append(parameters, v)
			}
		}
	} else if "pledge" == contractMethod || "reg_prod" == contractMethod || "vote" == contractMethod {
		parameters = strings.Split(contractParam, ",")
	} else if "set_account" == contractMethod {
		parameters = strings.Split(contractParam, "--")
	} else if "reg_chain" == contractMethod {
		parameter := strings.Split(contractParam, ",")
		if len(parameter) == 3 {
			parameters = append(parameters, parameter[0])
			parameters = append(parameters, parameter[1])
			parameters = append(parameters, innerCommon.AddressFromPubKey(innerCommon.FromHex(parameter[2])).HexString())
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"result": "Invalid parameters"})
			return
		}
	} else {
		contract, err := ledger.L.GetContract(chainId, innerCommon.NameToIndex(contractName))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
			return
		}

		var abiDef abi.ABI
		err = abi.UnmarshalBinary(contract.Abi, &abiDef)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
			return
		}

		//log.Debug("contractParam: ", contractParam)
		argbyte, err := abi.CheckParam(abiDef, contractMethod, []byte(contractParam))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
			return
		}

		parameters = append(parameters, string(argbyte[:]))
		abi.GetContractTable(contractName, "root", abiDef, "Account")
	}

	//time
	time := time.Now().UnixNano()
	transaction, err := types.NewInvokeContract(innerCommon.NameToIndex("root"), innerCommon.NameToIndex(contractName), chainId, "owner", contractMethod, parameters, 0, time)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	trx_data, err := transaction.Serialize()
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"result": "success", "invoke": innerCommon.ToHex(trx_data)})
}

func newDeployContract(c *gin.Context) {
	chainId_str := c.PostForm("chainId")
	contractName := c.PostForm("name")
	description := c.PostForm("description")
	contract_data := c.PostForm("contract_data")
	abi_data := c.PostForm("abi_data")

	if "" == chainId_str || "" == contractName || "" == description || contract_data == "" || abi_data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"result": "invalid params"})
		return
	}

	hash := new(innerCommon.Hash)
	chainId := hash.FormHexString(chainId_str)

	data, err := base64.StdEncoding.DecodeString(contract_data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	var contractAbi abi.ABI
	if err := json.Unmarshal([]byte(abi_data), &contractAbi); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	abibyte, err := abi.MarshalBinary(contractAbi)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	time := time.Now().UnixNano()
	transaction, err := types.NewDeployContract(innerCommon.NameToIndex(contractName), innerCommon.NameToIndex(contractName), chainId, "owner", types.VmWasm, description, data, abibyte, 0, time)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	trx_data, err := transaction.Serialize()
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"result": "success", "invoke": innerCommon.ToHex(trx_data)})
}
