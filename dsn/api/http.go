package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common"
	dsnComm "github.com/ecoball/go-ecoball/dsn/common"
	stm "github.com/ecoball/go-ecoball/dsn/settlement"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
)

func DsnHttpServ()  {
	router := gin.Default()
	router.GET("/dsn/total", totalHandler)
	router.POST("/dsn/eracode", eraCoding)
	router.POST("/dsn/eradecode", eraDecoding)
	router.Run(":8086")
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

}

func eraDecoding(c *gin.Context)  {

}