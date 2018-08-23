package commands

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ecoball/go-ecoball/client/rpc"
	innercommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/urfave/cli"
)

var (
	CreateCommands = cli.Command{
		Name:     "create",
		Usage:    "create operations",
		Category: "Create",
		Subcommands: []cli.Command{
			{
				Name:   "account",
				Usage:  "create account",
				Action: newAccount,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "creator, c",
						Usage: "creator name",
					},
					cli.StringFlag{
						Name:  "name, n",
						Usage: "account name",
					},
					cli.StringFlag{
						Name:  "owner, o",
						Usage: "owner public key",
					},
					cli.StringFlag{
						Name:  "active, a",
						Usage: "active public key",
					},
				},
			},
		},
	}
)

func newAccount(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//creator
	creator := c.String("creator")
	if creator == "" {
		fmt.Println("Invalid creator name")
		return errors.New("Invalid creator name")
	}

	//name
	name := c.String("name")
	if name == "" {
		fmt.Println("Invalid account name")
		return errors.New("Invalid account name")
	}

	if err := innercommon.AccountNameCheck(name); nil != err {
		fmt.Println(err)
		return err
	}

	//owner key
	owner := c.String("owner")
	if "" == owner {
		fmt.Println("Invalid owner key")
		return errors.New("Invalid owner key")
	}

	//active key
	active := c.String("active")
	if "" == active {
		active = owner
	}

	creatorAccount := innercommon.NameToIndex(creator)
	timeStamp := time.Now().Unix()

	invoke, err := types.NewInvokeContract(creatorAccount, creatorAccount, config.ChainHash, "owner", "new_account",
		[]string{name, innercommon.AddressFromPubKey(innercommon.FromHex(owner)).HexString()}, 0, timeStamp)
	invoke.SetSignature(&config.Root)

	//rpc call
	//resp, err := rpc.Call("createAccount", []interface{}{creator, name, owner, active})
	data, err := invoke.Serialize()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(len(data), data)
	resp, err := rpc.NodeCall("createAccount", []interface{}{string(data)})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	//result
	return rpc.EchoResult(resp)
}
