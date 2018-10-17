package api

import (
	"fmt"
	"encoding/json"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common"
	dsnComm "github.com/ecoball/go-ecoball/dsn/common"
	stm "github.com/ecoball/go-ecoball/dsn/settlement"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	rbd "github.com/ecoball/go-ecoball/dsn/renter/backend"
	rtypes "github.com/ecoball/go-ecoball/dsn/renter"
)

func DsnHttpServ()  {
	router := gin.Default()
	router.GET("/dsn/total", totalHandler)
	router.POST("/dsn/eracode", eraCoding)
	router.GET("/dsn/eradecode/:cid", eraDecoding)
	//TODO listen port need to be moved to config
	http.ListenAndServe(":9000", router)
}

func totalHandler(c *gin.Context)  {
	dkey := []byte(stm.KeyStorageTotal)
	total, err := ledger.L.StoreGet(config.ChainHash, common.NameToIndex(dsnComm.RootAccount), dkey)
	var du stm.DiskResource
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
	} else {
		err = encoding.Unmarshal(total, &du)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"result": "success", "total": du.TotalCapacity, "used": du.UsedCapacity, "hosts": du.Hosts})
		}
	}
}

func eraCoding(c *gin.Context)  {
	
	var req rtypes.RscReq
	buf := make([]byte,c.Request.ContentLength)
    _ , err := c.Request.Body.Read(buf)
	if err != nil {
 
	}
	err = json.Unmarshal(buf,&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
	} else {
			fmt.Println(req)
		}
	cid, err := rbd.EraCoding(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"result": "success", "cid": cid})
	}

}

func eraDecoding(c *gin.Context)  {
	cid, exsited := c.Params.Get("cid")
	if !exsited {
		c.JSON(http.StatusOK, gin.H{"result": "param err"})
	} else {
		r, err := rbd.EraDecoding(cid)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"result": "success", "data": r})
		}
	}
}