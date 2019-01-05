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
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"
	//"io/ioutil"
	"encoding/base64"

	"encoding/json"

	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
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
	router.POST("/setContract", setContract)
	router.POST("/newInvokeContract", newInvokeContract)
	router.POST("/newDeployContract", newDeployContract)
	//for invokContract
	//router.POST("/newContract", newContract)
	router.POST("/getRequiredKeys", get_required_keys)
	router.POST("/invokeContractForScan", invokeContractForScan)
	//router.POST("/recieveFile", recieveFile)

	//attach
	router.GET("/attach", attach)

	//query information
	router.GET("/query/mainChainHash", commands.GetMainChainHash)
	router.GET("/query/allChainInfo", commands.GetAllChainInfo)
	router.POST("/query/getAccountInfo", commands.GetAccountInfo)
	router.POST("/query/getTokenInfo", commands.GetTokenInfo)
	router.POST("/query/getBlockInfo", commands.GetBlockInfo)
	router.POST("/query/getTransaction", commands.GetTransaction)
	router.POST("/query/getRequiredKeys", commands.GetRequiredKeys)
	router.POST("/query/getContract", commands.GetContract)
	router.POST("//query/storeGet", commands.StoreGet)

	//contract
	router.POST("/invokeContract", commands.InvokeContract)

	http.ListenAndServe(":"+config.HttpLocalPort, router)
	return nil
}

func attach(c *gin.Context) {
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

	invoke, err := types.NewInvokeContract(creatorAccount, innerCommon.NameToIndex("root"), chainId, "owner", "new_account",
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

	c.JSON(http.StatusOK, gin.H{"result": "success", "invoke": hex.EncodeToString(data)})
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

	abiData, err := base64.StdEncoding.DecodeString(abi_data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	var contractAbi abi.ABI
	if err := json.Unmarshal(abiData, &contractAbi); err != nil {
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

	c.JSON(http.StatusOK, gin.H{"result": "success", "invoke": hex.EncodeToString(trx_data)})
}

func get_required_keys(c *gin.Context) {
	chainId := c.PostForm("chainId")
	required_keys := c.PostForm("keys")
	permission := c.PostForm("permission")
	from := c.PostForm("name")

	key_datas := strings.Split(required_keys, ",")

	//signTransaction, err := wallet.SignTransaction(inner.FromHex(transaction_data), datas)
	hash := new(innerCommon.Hash)
	chainids := hash.FormHexString(chainId)
	data, err := ledger.L.FindPermission(chainids, innerCommon.NameToIndex(from), permission)
	if err != nil {
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
		for _, value := range v.Keys {
			public_address = append(public_address, value.Actor)
		}
	}

	publickeys := ""
	for _, v := range key_datas {
		pubKey, _ := hex.DecodeString(v)
		addr := innerCommon.AddressFromPubKey(pubKey)
		for _, vv := range public_address {
			if addr == vv {
				publickeys += v
				publickeys += ","
				break
			}
		}
	}
	if "" != publickeys {
		publickeys = strings.TrimSuffix(publickeys, ",")
		c.JSON(http.StatusOK, gin.H{"result": publickeys})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"result": "no required_keys"})
}

func invokeContractForScan(c *gin.Context) {
	invoke := new(types.Transaction) //{
	transaction_data := c.PostForm("transaction")

	bytes, err := hex.DecodeString(transaction_data)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	if err := invoke.Deserialize(bytes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	//send to txpool
	err = event.Send(event.ActorNil, event.ActorTxPool, invoke)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	// wait for trx handle result
	var result string
	cmsg, err := event.SubOnceEach(invoke.Hash.String())
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(time.Second * 30)
		timeout <- true
	}()
	select {
	case msg := <-cmsg:
		result = msg.(string) + "\nwarning: transaction executed locally, but may not be confirmed by the network yet"
	case <-timeout:
		result = "trx handle timeout"
	}
	//event.UnSubscribe(cmsg, oneTransaction.Hash)

	c.JSON(http.StatusOK, gin.H{"result": result})
}
