package commands

import (
	"fmt"
	"os"
   "github.com/urfave/cli"
   dsncli "github.com/ecoball/go-ecoball/dsn/renter/client"
	"context"
	"io/ioutil"
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
	//dclient.CheckCollateral()
	cid, err := dclient.AddFile(file)
	if err != nil {
		return err
	}
	fmt.Println(cid)
	newCid, err := dclient.RscCodingReq(file, cid)
	if err != nil {
		return err
	}
	fmt.Println("added ", file, newCid)
	//dclient.InvokeFileContract(file, newCid)
	//dclient.PayForFile(file, newCid)
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