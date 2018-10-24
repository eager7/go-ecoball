package commands

import (
	"fmt"
	"os"
   "github.com/urfave/cli"
   "github.com/ecoball/go-ecoball/client/rpc"
   dsncli "github.com/ecoball/go-ecoball/dsn/renter/client/cli"
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
	_, err := dsncli.CliAddFile()
	return err
}

func dsnGetFile(ctx *cli.Context) error {
	var resp map[string]interface{}
	var err error

	if len(os.Args) == 4{
		resp, err = rpc.NodeCall("DsnCatFile", []interface{}{os.Args[0],os.Args[1],os.Args[2],os.Args[3]})
	}else{
		fmt.Println("only input 4 args")
	}

	if nil != resp["result"] {
		
		switch resp["result"].(type) {

		case string:
			data := resp["result"].(string)
			fmt.Println("catResult:",data)
			return nil
		default:
		}
	}
	
	return err

}

func dsnCatFile (ctx *cli.Context)  {
	dsncli.CliCatFile()
}