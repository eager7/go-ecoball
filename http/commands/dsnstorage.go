package commands

import (
	"os"
	"net/http"
	"net"
	"runtime"
	"github.com/gin-gonic/gin"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common"
	dsnComm "github.com/ecoball/go-ecoball/dsn/common"
	stm "github.com/ecoball/go-ecoball/dsn/audit"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/http/response"
	"github.com/ecoball/go-ecoball/dsn/host"
	"github.com/ecoball/go-ecoball/http/request"
	"github.com/oschwald/geoip2-golang"
	"context"
	"strconv"
 	dsncli "github.com/ecoball/go-ecoball/dsn/client"
)

func TotalHandler(c *gin.Context)  {
	dkey := []byte(stm.KeyStorageTotal)
	total, err := ledger.L.StoreGet(config.ChainHash, common.NameToIndex(dsnComm.RootAccount), dkey)
	var du stm.DiskResource
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
	} else {
		err = encoding.Unmarshal(total, &du)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"result": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"result": "success", "total": du.TotalCapacity, "used": du.UsedCapacity, "hosts": du.Hosts})
		}
	}
	
}

func EraCoding(c *gin.Context)  {
	var req dsnComm.RscReq
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnEraCoding{ Code: response.CODEPARAMSERR, Msg: err.Error(), Cid: ""})
	}

	cid, err := host.RscCoding(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnEraCoding{ Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: ""})
	} else {
		c.JSON(http.StatusOK, response.DsnEraCoding {Code: response.CODENOMAL, Msg:"success", Cid: cid })
	}
	
}

func EraDecoding(c *gin.Context)  {
	cid , exsited:= c.GetQuery("cid")
	if !exsited {
		c.JSON(http.StatusInternalServerError, response.DsnEraDecoding{ Code: response.CODESERVERINNERERR, Msg: "can not find cid" })
	} else {
		r, err := host.RscDecoding(cid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.DsnEraDecoding{ Code: response.CODESERVERINNERERR, Msg: err.Error()})
		} else {
			c.JSON(http.StatusOK, response.DsnEraDecoding{ Code: response.CODESERVERINNERERR, Msg: "can not find cid", Reader: r })
		}
	}
}


func DsnaddfileCid(c *gin.Context)  {
	
	cid, _ := c.GetQuery("cid")
	filesize, _ := c.GetQuery("filesize")
	filesizeInt, _ := strconv.ParseInt(filesize, 10, 64)
	cbtx := context.Background()
	dclient := dsncli.NewRcWithDefaultConf(cbtx)
	newCid, err := dclient.RscCodingReqWeb(filesizeInt, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: newCid })
		return 
	}
	c.JSON(http.StatusOK, response.DsnAddFileResponse{	Code: response.CODENOMAL, Msg:"success", Cid: newCid })

}


func DsnGetIpInfo(c *gin.Context)  {
	var req request.DsnIpInfoReq
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnEraCoding{ Code: response.CODEPARAMSERR, Msg: err.Error()})
	}
	var ostype = runtime.GOOS

	db := &geoip2.Reader{}
    if ostype == "windows"{
		db, err = geoip2.Open("..\\dsn\\GeoLite2-City.mmdb")
    }else if ostype == "linux"{
        db, err = geoip2.Open("../dsn/GeoLite2-City.mmdb")
    }

    if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnIpInfoRep{ Code: response.CODEPARAMSERR, Msg: err.Error()})
		return
	}


	ipInfoLists := make ([]response.DsnIpInfo , len(req.Iplists))
	for i := 0; i < len(req.Iplists); i++ {

		ip := net.ParseIP(req.Iplists[i])
		record, err := db.City(ip)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.DsnIpInfoRep{ Code: response.CODEPARAMSERR, Msg: err.Error()})
		}
		ipInfoLists[i].City =  record.City.Names["en"]
		if(len(record.Subdivisions)>0){
			ipInfoLists[i].Subdivision = record.Subdivisions[0].Names["en"]
		}else{
			ipInfoLists[i].Subdivision = ""
		}
		ipInfoLists[i].Country = record.Country.Names["en"]
		ipInfoLists[i].Countrycode = record.Country.IsoCode
		ipInfoLists[i].Timezone =  record.Location.TimeZone
		ipInfoLists[i].Latitude = record.Location.Latitude
		ipInfoLists[i].Longitude = record.Location.Longitude

	}
	c.JSON(http.StatusOK, response.DsnIpInfoRep{Code: response.CODENOMAL, Msg: "success", IpInfoLists: ipInfoLists})
}

func GetProjectPath() string{
    var projectPath string
    projectPath, _ = os.Getwd()
    return projectPath
}

