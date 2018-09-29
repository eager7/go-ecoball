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
		Usage:    "Distributed storage  interaction",
		Category: "dsnstorage",
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

	var resp map[string]interface{}
	var err error
	//dsn.AddFile(os.Args[3],0)
	if len(os.Args) <= 3  {
		fmt.Println("error error need more than three Args")
	}

	if len(os.Args) == 4{
		resp, err = rpc.NodeCall("DsnAddFile", []interface{}{os.Args[0],os.Args[1],os.Args[2],os.Args[3],"-1"})
		fmt.Println("only 4 args, dsnadd file use default value -1")
	}else if len(os.Args) == 5{
		resp, err = rpc.NodeCall("DsnAddFile", []interface{}{os.Args[0],os.Args[1],os.Args[2],os.Args[3],os.Args[4]})
	}
	
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