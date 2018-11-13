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
	//"fmt"
	"net/http"
	"encoding/hex"

	"github.com/ecoball/go-ecoball/core/types"
	"github.com/gin-gonic/gin"

	"github.com/ecoball/go-ecoball/common/event"
	"time"
)

func InvokeContract(c *gin.Context) {
	var trx string
	var oneTransaction types.Transaction
	if err := c.BindJSON(&trx); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	data, err := hex.DecodeString(trx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = oneTransaction.Deserialize(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	//send to txpool
	err = event.Send(event.ActorNil, event.ActorTxPool, &oneTransaction)
	if nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// wait for trx handle result
	var result string
	cmsg, err := event.SubscribeOnceEach(oneTransaction.Hash)
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(time.Second * 10)
		timeout <- true
	}()
	select {
		case msg := <-cmsg:
			result = msg.(string) + "\nwarning: transaction executed locally, but may not be confirmed by the network yet"
		case <-timeout:
			result = "trx handle timeout, maybe it had handled in other shard, please check it later"
	}
	//event.UnSubscribe(cmsg, oneTransaction.Hash)

	c.JSON(http.StatusOK, gin.H{"result": result})
}
