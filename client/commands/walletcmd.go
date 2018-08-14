// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball.
//
// The go-ecoball is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.
package commands

import (
	"errors"
	"fmt"
	"os"

	//"github.com/ecoball/go-ecoball/account"
	//"github.com/ecoball/go-ecoball/common"
	"github.com/urfave/cli"
	"github.com/ecoball/go-ecoball/client/rpc"
	innerCommon "github.com/ecoball/go-ecoball/http/common"
)

var (
	WalletCommands = cli.Command{
		Name:        "wallet",
		Usage:       "wallet operation",
		Category:    "Wallet",
		Description: "wallet operate",
		ArgsUsage:   "[args]",
		Subcommands: []cli.Command{
			{
				Name:   "create",
				Usage:  "create wallet",
				Action: createWallet,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "wallet name",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
				},
			},
			{
				Name:   "open",
				Usage:  "open wallet",
				Action: openWallet,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "wallet name",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
				},
			},
			{
				Name:   "createkey",
				Usage:  "create key",
				Action: createKey,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "wallet name",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
				},
			},
			{
				Name:   "lock",
				Usage:  "lock wallet",
				Action: lockWallet,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "wallet name",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
				},
			},
			{
				Name:   "import",
				Usage:  "import private key",
				Action: importKey,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "wallet name",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
					cli.StringFlag{
						Name:  "private, k",
						Usage: "private key",
					},
				},
			},
			{
				Name:   "remove",
				Usage:  "remove private key",
				Action: removeKey,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "wallet name",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
					cli.StringFlag{
						Name:  "public, k",
						Usage: "public key",
					},
				},
			},
			{
				Name:   "unlock",
				Usage:  "unlock wallet",
				Action: unlockWallet,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "wallet name",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
				},
			},
			{
				Name:   "list_keys",
				Usage:  "list keys",
				Action: listAccount,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "wallet name",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "wallet password",
					},
				},
			},
			{
				Name:   "list",
				Usage:  "list wallets",
				Action: listWallets,
				Flags: []cli.Flag{
				},
			},
		},
	}
)

func createWallet(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	name := c.String("name")
	if "" == name {
		fmt.Println("Invalid wallet name")
		return errors.New("Invalid wallet name")
	}

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	resp, err := rpc.Call("createWallet", []interface{}{name, passwd})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	//result
	rpc.EchoResult(resp)
	if int64(innerCommon.SUCCESS) == int64(resp["errorCode"].(float64)) {
		fmt.Println("wallet file path:", name)
	}
	return nil
}

func createKey(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	name := c.String("name")
	if "" == name {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	//Check the number of flags
	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	resp, err := rpc.Call("createKey", []interface{}{name, passwd})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	rpc.EchoResult(resp)
	return nil
}

func openWallet(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	name := c.String("name")
	if "" == name {
		fmt.Println("Invalid wallet name")
		return errors.New("Invalid wallet name")
	}

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	resp, err := rpc.Call("openWallet", []interface{}{name, passwd})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	rpc.EchoResult(resp)
	if int64(innerCommon.SUCCESS) == int64(resp["errorCode"].(float64)) {
		fmt.Println("open wallet success, wallet file path:", name)
	}
	return nil
}

func lockWallet(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	name := c.String("name")
	if "" == name {
		fmt.Println("Invalid wallet name")
		return errors.New("Invalid wallet name")
	}

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	resp, err := rpc.Call("lockWallet", []interface{}{name, passwd})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	rpc.EchoResult(resp)
	return nil
}

func unlockWallet(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	name := c.String("name")
	if "" == name {
		fmt.Println("Invalid wallet name")
		return errors.New("Invalid wallet name")
	}

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	resp, err := rpc.Call("unlockWallet", []interface{}{name, passwd})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	rpc.EchoResult(resp)
	return nil
}

func importKey(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	name := c.String("name")
	if "" == name {
		fmt.Println("Invalid wallet name")
		return errors.New("Invalid wallet name")
	}

	password := c.String("password")
	if "" == password {
		fmt.Println("Invalid wallet password")
		return errors.New("Invalid password")
	}

	privateKey := c.String("private")
	if "" == privateKey {
		fmt.Println("Invalid private key")
		return errors.New("Invalid private key")
	}

	resp, err := rpc.Call("importKey", []interface{}{name, password, privateKey})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	rpc.EchoResult(resp)
	return nil
}

func removeKey(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	name := c.String("name")
	if "" == name {
		fmt.Println("Invalid wallet name")
		return errors.New("Invalid wallet name")
	}

	password := c.String("password")
	if "" == password {
		fmt.Println("Invalid wallet password")
		return errors.New("Invalid password")
	}

	publicKey := c.String("public")
	if "" == publicKey {
		fmt.Println("Invalid private key")
		return errors.New("Invalid private key")
	}

	resp, err := rpc.Call("removeKey", []interface{}{name, password, publicKey})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	rpc.EchoResult(resp)
	return nil
}

func listAccount(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	name := c.String("name")
	if "" == name {
		fmt.Println("Invalid wallet name")
		return errors.New("Invalid wallet name")
	}

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid password")
		return errors.New("Invalid password")
	}

	resp, err := rpc.Call("list_keys", []interface{}{name, passwd})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	rpc.EchoResult(resp)
	return nil
}

func listWallets(c *cli.Context) error {
	resp, err := rpc.Call("list_wallets", []interface{}{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	rpc.EchoResult(resp)
	return nil
}
