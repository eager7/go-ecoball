package commands

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	clientCommon "github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	dsncli "github.com/ecoball/go-ecoball/dsn/renter/client"
	"github.com/urfave/cli"
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

func dsnAddFile(ctx *cli.Context) error {
	cbtx := context.Background()
	dclient := dsncli.NewRcWithDefaultConf(cbtx)
	file := os.Args[3]
	ok := dclient.CheckCollateral()
	if !ok {
		return errors.New("Checking collateral failed")
	}
	cid, err := dclient.AddFile(file)
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

	chainId, err := getMainChainHash()
	if err != nil {
		return err
	}

	pkKeys, err := getPublicKeys()
	if err != nil {
		return err
	}

	reqKeys, err := getRequiredKeys(chainId, "owner", transaction.From.String())
	if err != nil {
		return err
	}

	err = SignTransaction(chainId, "", transaction)
	if err != nil {
		return err
	}

	data, err := transaction.Serialize()
	if err != nil {
		return err
	}

	var retContract clientCommon.SimpleResult
	ctcv := url.Values{}
	ctcv.Set("transaction", common.ToHex(data))
	err = rpc.NodePost("/invokeContract", ctcv.Encode(), &retContract)
	fmt.Println("fileContract: ", retContract.Result)

	///////////////////////////////////////////////////
	payTrn, err := dclient.PayForFile(file, newCid)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	reqKeys, err = getRequiredKeys(chainId, "owner", payTrn.From.String())
	if err != nil {
		return err
	}
	err = SignTransaction(chainId, "", payTrn)
	if err != nil {
		return err
	}
	data, err = payTrn.Serialize()
	if err != nil {
		return err
	}

	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("transfer", common.ToHex(data))
	err = rpc.NodePost("/transfer", values.Encode(), &result)
	fmt.Println("pay: ", result.Result)

	return nil
}

func dsnCatFile(ctx *cli.Context) {
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
	info, err := getInfo()
	return info.ChainID, err
}

func GetRequiredKeys(chainId common.Hash, required_keys, permission string, trx *types.Transaction) (string, error) {
	/*data, err := trx.Serialize()
	if err != nil {
		return "", err
	}*/

	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("permission", permission)
	values.Set("chainId", chainId.HexString())
	values.Set("keys", required_keys)
	values.Set("name", trx.From.String())
	err := rpc.NodePost("/get_required_keys", values.Encode(), &result)
	if nil == err {
		return result.Result, nil
	}
	return "", err
}

func SignTransaction(chainId common.Hash, required_keys string, trx *types.Transaction) error {
	data, err := trx.Serialize()
	if err != nil {
		return err
	}
	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("keys", required_keys)
	values.Set("transaction", common.ToHex(data))
	err = rpc.WalletPost("/wallet/signTransaction", values.Encode(), &result)
	if nil == err {
		trx.Deserialize(common.FromHex(result.Result))
	}
	return err
}

func TxTransaction(trx *types.Transaction) error {
	chainId, err := getMainChainHash()
	if err != nil {
		return err
	}

	pkKeys, err := getPublicKeys()
	if err != nil {
		return err
	}

	reqKeys, err := getRequiredKeys(chainId, "owner", trx.From.String())
	if err != nil {
		return err
	}
	err = SignTransaction(chainId, "", trx)
	if err != nil {
		return err
	}
	data, err := trx.Serialize()
	if err != nil {
		return err
	}

	var result clientCommon.SimpleResult
	values := url.Values{}
	values.Set("transfer", common.ToHex(data))
	err = rpc.NodePost("/transfer", values.Encode(), &result)
	fmt.Println("tx transaction: ", result.Result)
	return err
}

func InvokeContract(trx *types.Transaction) error {
	chainId, err := getMainChainHash()
	if err != nil {
		return err
	}

	pkKeys, err := getPublicKeys()
	if err != nil {
		return err
	}

	reqKeys, err := getRequiredKeys(chainId, "owner", trx.From.String())
	if err != nil {
		return err
	}

	err = SignTransaction(chainId, "", trx)
	if err != nil {
		return err
	}

	data, err := trx.Serialize()
	if err != nil {
		return err
	}

	var retContract clientCommon.SimpleResult
	ctcv := url.Values{}
	ctcv.Set("transaction", common.ToHex(data))
	err = rpc.NodePost("/invokeContract", ctcv.Encode(), &retContract)
	fmt.Println("Contract: ", retContract.Result)
	return err
}
