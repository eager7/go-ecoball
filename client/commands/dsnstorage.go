package commands

import (
	"fmt"
	"os"
   "github.com/urfave/cli"
   dsncli "github.com/ecoball/go-ecoball/dsn/renter/client"
	"context"
	"io/ioutil"
//	"net/url"
	"github.com/ecoball/go-ecoball/common"
	 clientCommon "github.com/ecoball/go-ecoball/client/common"
	 innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
//	"github.com/ecoball/go-ecoball/client/rpc"
	"errors"
	walletHttp "github.com/ecoball/go-ecoball/walletserver/http"
	"github.com/ecoball/go-ecoball/client/rpc"
)
var (
	DsnStorageCommands = cli.Command{
		Name:     "dsnstorage",
		Usage:    "Distributed storage  interaction",
		Category: "dsnstorage",
		Subcommands: []cli.Command{
			{
				Name:   "add",
				Usage:  "add file",
				Action: dsnAddFile,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "add_-1, 1",
						Usage: "add -1",
						Value: "-1",
					},
				},

			},
			{
				Name:   "cat",
				Usage:  "cat file",
				Action: dsnCatFile,
	
			},
		},
		
	}
	
)

func GetPublicKeys() (walletHttp.Keys, error) {

//	return "",nil
	return getPublicKeys();
}

func dsnAddFile(ctx *cli.Context) error {

	cbtx := context.Background()
	dclient := dsncli.NewRcWithDefaultConf(cbtx)
	file := os.Args[3]
	accountName := os.Args[4]
	ok := dclient.CheckCollateralParams(accountName)
	if !ok {
		return errors.New("Checking collateral failed")
	}
	cid, _, err := dclient.AddFile(file)
	if err != nil {
		return err
	}
	
	newCid, err := dclient.RscCodingReq(file, cid)
	if err != nil {
		return err
	}
	fmt.Println("added ", file, newCid)
	transaction, err := dclient.InvokeFileContract(file, newCid)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	chainId, err := GetChainId()
	if err != nil {
		return err
	}

	pkKeys, err := GetPublicKeys()
	if err != nil {
		return err
	}

	//reqKeys, err := getRequiredKeys(chainId, pkKeys, "owner", transaction)
	reqKeys, err := GetRequiredKeys(chainId, "owner", accountName)
	if err != nil {
		return err
	}

	publickeys := clientCommon.IntersectionKeys(pkKeys, reqKeys)
	if 0 == len(publickeys.KeyList) {
		fmt.Println("no publickeys")
		return errors.New("no publickeys")
	}

	data, err := SignTransaction(chainId, publickeys, transaction.Hash[:])
	if err != nil {
		return err
	}

	for _, v := range data.Signature {
		transaction.AddSignature(v.PublicKey.Key, v.SignData)
	}

	var result rpc.SimpleResult
	err = rpc.NodePost("/invokeContract", transaction, &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	

	payTrn, err := dclient.PayForFile(file, newCid)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}


	dataPays, err := SignTransaction(chainId, publickeys, payTrn.Hash[:])
	if err != nil {
		return err
	}

	for _, v := range dataPays.Signature {
		transaction.AddSignature(v.PublicKey.Key, v.SignData)
	}

	var resultPays rpc.SimpleResult
	err = rpc.NodePost("/invokeContract", transaction, &resultPays)
	if nil == err {
		fmt.Println(resultPays.Result)
	}
	
	/*payTrn, err := dclient.PayForFile(file, newCid)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	reqKeys, err = GetRequiredKeys(chainId, pkKeys, "owner", payTrn)
	if err != nil {
		return err
	}
	err = SignTransaction(chainId, reqKeys, payTrn)
	if err != nil {
		return err
	}
	data, err = payTrn.Serialize()
	if err != nil {
		return err
	}*/

	// var result clientCommon.SimpleResult
	// values := url.Values{}
	// values.Set("transfer", common.ToHex(data))
	// err = rpc.NodePost("/transfer", values.Encode(), &result)
	// fmt.Println("pay: ", result.Result)

	return nil
}


func dsnCatFile (ctx *cli.Context)  {
	cbtx := context.Background()
	dclient := dsncli.NewRcWithDefaultConf(cbtx)
	//dclient.CheckCollateral()
	cid := os.Args[3]
	r, err := dclient.CatFile(cid)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	d, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(d))
}

func dsnGetFile(ctx *cli.Context) error {
	cid := os.Args[3]
	outPath := os.Args[4]
	cbtx := context.Background()
	dclient := dsncli.NewRcWithDefaultConf(cbtx)
	return dclient.GetFile(cid, outPath)
}

func GetChainId() (common.Hash, error) {
	return getMainChainHash()
}

func GetRequiredKeys(chainHash innerCommon.Hash, permission string, account string) ([]innerCommon.Address, error) {

	return getRequiredKeys(chainHash, permission, account)
	

}

func SignTransaction(chainHash innerCommon.Hash, publickeys walletHttp.Keys, rawData []byte) (walletHttp.SignTransaction, error) {
	
	return signTransaction(chainHash, publickeys, rawData)
	
}

func TxTransaction(trx *types.Transaction, accountName string) error {

	chainId, err := GetChainId()
	if err != nil {
		return err
	}

	pkKeys, err := GetPublicKeys()
	if err != nil {
		return err
	}

	//reqKeys, err := getRequiredKeys(chainId, pkKeys, "owner", transaction)
	reqKeys, err := GetRequiredKeys(chainId, "owner", accountName)
	if err != nil {
		return err
	}

	publickeys := clientCommon.IntersectionKeys(pkKeys, reqKeys)
	if 0 == len(publickeys.KeyList) {
		fmt.Println("no publickeys")
		return errors.New("no publickeys")
	}

	data, err := SignTransaction(chainId, publickeys, trx.Hash[:])
	if err != nil {
		return err
	}

	for _, v := range data.Signature {
		trx.AddSignature(v.PublicKey.Key, v.SignData)
	}

	var result rpc.SimpleResult
	err = rpc.NodePost("/invokeContract", trx, &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	// reqKeys, err := GetRequiredKeys(chainId, pkKeys, "owner", trx)
	// if err != nil {
	// 	return err
	// }
	// err = SignTransaction(chainId, reqKeys, trx)
	// if err != nil {
	// 	return err
	// }
	// data, err := trx.Serialize()
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(data)
	// var result clientCommon.SimpleResult
	// values := url.Values{}
	// values.Set("transfer", common.ToHex(data))
	// err = rpc.NodePost("/transfer", values.Encode(), &result)
	// fmt.Println("tx transaction: ", result.Result)
	return err
}

func InvokeContract(trx *types.Transaction, accountName string) error {
	chainId, err := GetChainId()
	if err != nil {
		return err
	}

	pkKeys, err := GetPublicKeys()
	if err != nil {
		return err
	}


	//reqKeys, err := getRequiredKeys(chainId, pkKeys, "owner", transaction)
	reqKeys, err := GetRequiredKeys(chainId, "owner", accountName)
	if err != nil {
		return err
	}

	publickeys := clientCommon.IntersectionKeys(pkKeys, reqKeys)
	if 0 == len(publickeys.KeyList) {
		fmt.Println("no publickeys")
		return errors.New("no publickeys")
	}

	data, err := SignTransaction(chainId, publickeys, trx.Hash[:])
	if err != nil {
		return err
	}

	for _, v := range data.Signature {
		trx.AddSignature(v.PublicKey.Key, v.SignData)
	}

	var result rpc.SimpleResult
	err = rpc.NodePost("/invokeContract", trx, &result)
	if nil == err {
		fmt.Println(result.Result)
	}


	// reqKeys, err := GetRequiredKeys(chainId, pkKeys, "owner", trx)
	// if err != nil {
	// 	return err
	// }

	// err = SignTransaction(chainId, reqKeys, trx)
	// if err != nil {
	// 	return err
	// }

	// data, err := trx.Serialize()
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(data)
	// var retContract clientCommon.SimpleResult
	// ctcv := url.Values{}
	// ctcv.Set("transaction", common.ToHex(data))
	// err = rpc.NodePost("/invokeContract", ctcv.Encode(), &retContract)
	// fmt.Println("fileContract: ", retContract.Result)
	return err
}