package commands

import (
	"fmt"
	"os"
	"context"
	"io/ioutil"
	"errors"
	"github.com/urfave/cli"
	wc "github.com/ecoball/go-ecoball/client/walletclient"
	fc "github.com/ecoball/go-ecoball/dsn/client"
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
			{
				Name:   "get",
				Usage:  "get file",
				Action: dsnGetFile,
	
			},
		},
		
	}
	
)


func dsnAddFile(ctx *cli.Context) error {

	if len(os.Args) < 4{
		return errors.New("please input dsnstorage add filepath")
	}

	cbtx := context.Background()
	file := os.Args[3]

	walletName := "dsnwallet"
	accountName := "dsn"
	collateral := 0

	wClient := wc.NewWalletClient(accountName, walletName, collateral)
	ok := wClient.CheckCollateral()
	if !ok {
		return errors.New("Checking account's collateral failed")
	}

	dclient := fc.NewRcWithDefaultConf(cbtx)
	//Add file to ipfs network
	cid, _, err := dclient.AddFile(file)
	if err != nil {
		return err
	}
	//erasure coding
	newCid, err := dclient.RscCodingReq(file, cid)
	if err != nil {
		return err
	}
	fmt.Println("added ", file, newCid)
	//pay for file
	payTrn, err := dclient.PayForFile(file)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	trnID, err := wClient.Transer(payTrn)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Println("payed for file, id: ", payTrn.Hash.HexString())
	//Invoke file contract
	transaction, err := dclient.InvokeFileContract(file, newCid, trnID)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	_, err = wClient.InvokeContract(transaction)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return err
}


func dsnCatFile (ctx *cli.Context) error {

	
	if len(os.Args) < 4{
		return errors.New("please input dsnstorage cat cid")
	}

	cbtx := context.Background()
	walletName := "dsnwallet"
	accountName := "dsn"
	collateral := 0
	wClient := wc.NewWalletClient(accountName, walletName, collateral)
	ok := wClient.CheckCollateral()
	if !ok {
		return errors.New("Checking account's collateral failed")
	}
	dclient := fc.NewRcWithDefaultConf(cbtx)
	//dclient.CheckCollateral()
	cid := os.Args[3]

	r, err := dclient.CatFile(cid)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	d, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	
	 payTrn, err := dclient.PayForFileSize(int64(len(d)))
	 if err != nil {
	 	fmt.Println(err.Error())
	 	return err
	}
	trnID, err := wClient.Transer(payTrn)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Println("payed for file, id: ", payTrn.Hash.HexString())
	//Invoke file contract
	transaction, err := dclient.InvokeFileContractWeb("cat" + cid, uint64(len(d)), cid, trnID)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	_, err = wClient.InvokeContract(transaction)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Println(string(d))
	return nil
}

func dsnGetFile(ctx *cli.Context) error {

	if len(os.Args) < 5{
		return errors.New("please input dsnstorage get cid localfilepath")
	}

	cid := os.Args[3]
	outPath := os.Args[4]
	cbtx := context.Background()
	walletName := "dsnwallet"
	accountName := "dsn"
	collateral := 0
	wClient := wc.NewWalletClient(accountName, walletName, collateral)
	ok := wClient.CheckCollateral()
	if !ok {
		return errors.New("Checking account's collateral failed")
	}
	dclient := fc.NewRcWithDefaultConf(cbtx)
	err :=  dclient.GetFile(cid, outPath)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}