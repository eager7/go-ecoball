package commands

import (
	"io/ioutil"
	"github.com/ecoball/go-ecoball/http/common"
	"github.com/ecoball/go-ecoball/dsn"
	"github.com/ecoball/go-ecoball/dsn/renter"
	"github.com/ecoball/go-ecoball/dsn/renter/backend"
	"fmt"
)

func DsnAddFile(params []interface{})  *common.Response {
	if len(params) < 1 {
		log.Error("invalid arguments")
		return common.NewResponse(common.INVALID_PARAMS, "type not ok")
	}
	req := params[0].(renter.RscReq)
	fmt.Println("-------------DsnAddFile")
	fmt.Println(req)
	cid, err := backend.EraCoding(&req)
	if err != nil {
		return common.NewResponse(common.INVALID_PARAMS, "DsnAddFile faild")
	}
	return common.NewResponse(common.SUCCESS, cid)
}

func DsnCatFile(params []interface{})  *common.Response {

	if len(params) < 1 {
		log.Error("invalid arguments")
	}


	readerResult, err := dsn.CatFile(params[3].(string))
	if err != nil {
		return common.NewResponse(common.INVALID_PARAMS, "DsnGetFile faild")
	}

	d, err := ioutil.ReadAll(readerResult)
	if err != nil {
		return common.NewResponse(common.INVALID_PARAMS, "readerResult.Read(p) faild")
	}
	
//	ioutil.WriteFile("E:\\临时\\test3.txt", d , os.ModeAppend)
	return common.NewResponse(common.SUCCESS,string(d[:]))


}