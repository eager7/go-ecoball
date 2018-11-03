
package response

import (
	"io"
)

const CODENOMAL int = 0  //返回结果正常
const CODEPARAMSERR int = 40002   //返回结果为参数传递错误
const CODESERVERINNERERR int = 50000  //返回结果为服务端内部处理错误

//無值時返回結構
type DsnBasicResponse struct {
	
	Code   int    //返回码
	Msg    string //返回消息 

}

//添加文件返回结果
type DsnAddFileResponse struct {

	Code   int
	Msg    string  
	Cid    string  

}

//添加冗余接口
type DsnEraCoding struct {

	Code   int
	Msg    string  
	Cid    string  

}

//解析冗余接口
type DsnEraDecoding struct {

	Code   int
	Msg    string  
	Reader io.Reader  

}

//获取accountstake
type DsnAccountStake struct {

	Code   int
	Msg    string  
	AccountStake uint64  

}
