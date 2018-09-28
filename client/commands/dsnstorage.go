package commands

import (
	"fmt"
	"os"
   "github.com/urfave/cli"
  // "github.com/ecoball/go-ecoball/dsn"
   "github.com/ecoball/go-ecoball/client/rpc"
)
var (
	DsnStorageCommands = cli.Command{
		Name:     "dsnstorage",
		Usage:    "operations for query state",
		Category: "Get",
		Subcommands: []cli.Command{
			{
				Name:   "dsnadd",
				Usage:  "dsnadd file",
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
				Name:   "dsncat",
				Usage:  "dsncat file",
				Action: dsnGetFile,
	
			},
		},
		
		
	}
)
func dsnAddFile(ctx *cli.Context) error {

	
	//dsn.AddFile(os.Args[3],0)
	resp, err := rpc.NodeCall("DsnAddFile", []interface{}{os.Args[0],os.Args[1],os.Args[2],os.Args[3],os.Args[4]})
	if nil != resp["result"] {
		switch resp["result"].(type) {
		case string:
			data := resp["result"].(string)
			fmt.Println(data)
			return nil
		default:
		}
	}
	
	return err
}

func dsnGetFile(ctx *cli.Context) error {
	
	//dsn.AddFile(os.Args[3],0)
	resp, err := rpc.NodeCall("DsnCatFile", []interface{}{os.Args[0],os.Args[1],os.Args[2],os.Args[3]})
	if nil != resp["result"] {
		switch resp["result"].(type) {
		case string:
			data := resp["result"].(string)
			fmt.Println(data)
			return nil
		default:
		}
	}
	
	return err

}