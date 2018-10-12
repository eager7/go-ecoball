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

	"github.com/ecoball/go-ecoball/common/config"
	"github.com/gin-gonic/gin"
	innercommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
)

func StartHttpServer() (err error) {
	//get router instance
	router := gin.Default()

	//register handle
	router.GET("/transfer", transfer)

	router.POST("/getAccountInfo", getAccountInfo)

	http.ListenAndServe(":20681", router)
	return nil
}

func transfer(c *gin.Context) {}

func getAccountInfo(c *gin.Context) {
	name := c.PostForm("name")

	data, err := ledger.L.AccountGet(config.ChainHash, innercommon.NameToIndex(name))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"account": data})
}
