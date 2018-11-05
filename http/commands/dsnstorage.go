package commands

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
//	rtypes "github.com/ecoball/go-ecoball/dsn/renter"
	"github.com/ecoball/go-ecoball/http/request"
	"context"
	dsncli "github.com/ecoball/go-ecoball/dsn/renter/client"
	//"github.com/ecoball/go-ecoball/dsn/renter"
	"github.com/ecoball/go-ecoball/client/commands"
	clientCommon "github.com/ecoball/go-ecoball/client/common"
	
//	"net/url"
	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/ecoball/go-ecoball/dsn"
	"github.com/ecoball/go-ecoball/http/response"
)

// func DsnHttpServ()  {
// 	router := gin.Default()
// 	router.GET("/dsn/total", totalHandler)
// 	router.POST("/dsn/eracode", eraCoding)
// 	router.GET("/dsn/eradecode/:cid", eraDecoding)
// 	router.GET("/dsn/accountstake", accountStake)
// 	router.POST("/dsn/dsnaddfile", Dsnaddfile)
// 	//TODO listen port need to be moved to config
// 	http.ListenAndServe(":9000", router)
// }

func TotalHandler(c *gin.Context)  {

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

func EraCoding(c *gin.Context)  {

	var req request.DsnAddFileReq
	buf := make([]byte,c.Request.ContentLength)
    _ , err := c.Request.Body.Read(buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnEraCoding{ Code: response.CODEPARAMSERR, Msg: err.Error(), Cid: ""})
	}

	err = json.Unmarshal(buf,&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnEraCoding{ Code: response.CODEPARAMSERR, Msg: err.Error(), Cid: ""})
	} 

	cid, err := dsn.RscCoding(&req)
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
		r, err := dsn.RscDecoding(cid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.DsnEraDecoding{ Code: response.CODESERVERINNERERR, Msg: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.DsnEraDecoding{ Code: response.CODESERVERINNERERR, Msg: "can not find cid", Reader: r })
		}
	}
}

func AccountStake(c *gin.Context)  {

	//name, exsited := c.Params.Get("name")
	name , exsited:= c.GetQuery("name")
	if !exsited {
		c.JSON(http.StatusInternalServerError, response.DsnAccountStake{ Code: response.CODEPARAMSERR, Msg: "params not found"})
		return
	}
	chainId, exsited := c.GetQuery("chainid")
	if !exsited {
		c.JSON(http.StatusInternalServerError, response.DsnAccountStake{ Code: response.CODEPARAMSERR, Msg: "params not found"})
		return
	}
	sacc, err := ledger.L.AccountGet(common.HexToHash(chainId), common.NameToIndex(name))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnAccountStake{ Code: response.CODESERVERINNERERR, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.DsnAccountStake{ Code: response.CODENOMAL, Msg: err.Error(), AccountStake:  sacc.Resource.Votes.Staked})

}


func Dsnaddfile(c *gin.Context)  {


	file, err := c.FormFile("file")
    if err != nil {
	//	c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODEPARAMSERR, Msg: err.Error(), Cid: ""})
		return
	}

	accountName, _ := c.GetQuery("accountname")
	hashcode , _:= c.GetQuery("hash")
	fmt.Println(hashcode)
	cbtx := context.Background()
	dclient := dsncli.NewRcWithDefaultConf(cbtx)
	cid, _, err := dclient.HttpAddFile(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: cid })
		return
	}
	fmt.Println("added ",  cid)
	newCid, err := dclient.RscCodingReqWeb(file.Size, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: newCid })
		return 
	}

	fmt.Println("addednew ",  newCid)
	transaction, err := dclient.InvokeFileContractWeb(newCid, uint64(file.Size), newCid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: newCid })
		return 
	}



	chainId, err := commands.GetChainId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: newCid })
		return 
	}

	pkKeys, err := commands.GetPublicKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: newCid })
		return 
	}

	reqKeys, err := commands.GetRequiredKeys(chainId, "owner", accountName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: newCid })
		return 
	}

	publickeys := clientCommon.IntersectionKeys(pkKeys, reqKeys)
	if 0 == len(publickeys.KeyList) {

		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: "no publickeys", Cid: newCid })
	}

	
	data, err := commands.SignTransaction(chainId, publickeys, transaction.Hash[:])
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: newCid })
		return
	}

	for _, v := range data.Signature {
		transaction.AddSignature(v.PublicKey.Key, v.SignData)
	}

	var result rpc.SimpleResult
	err = rpc.NodePost("/invokeContract", transaction, &result)
	if nil == err {
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: newCid })
		return 
	}
	
	///////////////////////////////////////////////////
	payTrn, err := dclient.PayForFile(newCid, newCid)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: newCid })
		return 
	}



	dataPays, err := commands.SignTransaction(chainId, publickeys, payTrn.Hash[:])
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: newCid })
		return 
	}

	for _, v := range dataPays.Signature {
		transaction.AddSignature(v.PublicKey.Key, v.SignData)
	}

	var resultPays rpc.SimpleResult
	err = rpc.NodePost("/invokeContract", transaction, &resultPays)
	if nil == err {
		c.JSON(http.StatusInternalServerError, response.DsnAddFileResponse{	Code: response.CODESERVERINNERERR, Msg: err.Error(), Cid: newCid })
		return 
	}
	
	c.JSON(http.StatusOK, response.DsnAddFileResponse{	Code: response.CODENOMAL, Msg: "success", Cid: newCid })
	
}
