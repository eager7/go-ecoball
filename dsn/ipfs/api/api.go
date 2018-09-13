package api

import (
	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
)

var DsnIpfsApi coreiface.CoreAPI

func StartDsnIpfsApi(node *core.IpfsNode)  {
	DsnIpfsApi = coreapi.NewCoreAPI(node)
}