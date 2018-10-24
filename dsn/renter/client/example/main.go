package cli

import (
	"context"
	rclient "github.com/ecoball/go-ecoball/dsn/renter/client"
	"os"
	"fmt"
	"io/ioutil"
)

func add()  {
	conf := rclient.InitDefaultConf()
	ctx := context.Background()
	appClient := rclient.NewRenter(ctx, conf)
	file := os.Args[3]
	ok := appClient.CheckCollateral()
	if !ok {
		fmt.Println("Checking collateral failed")
		return
	}
	cid, err := appClient.AddFile(file)
	if err != nil {
		panic(err)
	}
	cid, err = appClient.RscCodingReq(file, cid)
	if err != nil {
		panic(err)
	}
	appClient.InvokeFileContract(file, cid)
	appClient.PayForFile(file, cid)
}

func cat()  {
	conf := rclient.InitDefaultConf()
	ctx := context.Background()
	appClient := rclient.NewRenter(ctx, conf)
	ok := appClient.CheckCollateral()
	if !ok {
		fmt.Println("Checking collateral failed")
		return
	}
	cid := os.Args[3]
	r, err := appClient.CatFile(cid)
	if err != nil {
		appClient.RscDecodingReq(cid)
	}
	d, err := ioutil.ReadAll(r)
	fmt.Println(d)
}

func main()  {
	add()
}