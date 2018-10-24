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
	"net/url"
	//"strings"

	//"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	outerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/urfave/cli"
)

var (
	WalletCommands = cli.Command{
		Name:        "wallet",
		Usage:       "wallet operation",
		Category:    "Wallet",
		Description: "wallet operate",
		ArgsUsage:   "[args]",
		Action:      common.DefaultAction,
		Subcommands: []cli.Command{
			{
				Name:   "attach",
				Usage:  "hang different wallet nodes",
				Action: attachWallet,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "ip",
						Usage: "node's ip address",
						Value: "localhost",
					},
					cli.StringFlag{
						Name:  "port",
						Usage: "node's RPC port",
						Value: "20679",
					},
				},
			},
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
				Flags:  []cli.Flag{},
			},
		},
	}
)

func attachWallet(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//ip address
	ip := c.String("ip")
	if "" != ip {
		common.WalletIp = ip
	}

	//port
	port := c.String("port")
	if "" != port {
		common.WalletPort = port
	}

	//rpc call
	var result common.SimpleResult
	err := rpc.WalletGet("/wallet/attach", &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
}

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

	var result common.SimpleResult
	values := url.Values{}
	values.Set("name", name)
	values.Set("password", passwd)
	err := rpc.WalletPost("/wallet/create", values.Encode(), &result)

	if nil == err {
		fmt.Println(result.Result)
	}
	return err
}

func GetPublicKeys() (string, error) {
	var result common.SimpleResult
	err := rpc.WalletGet("/wallet/getPublicKeys", &result)
	if nil == err {
		return result.Result, nil
	}
	return "", err
}

func sign_transaction(chainId outerCommon.Hash, required_keys string, trx *types.Transaction) error {
	data, err := trx.Serialize()
	if err != nil {
		return err
	}
	var result common.SimpleResult
	values := url.Values{}
	values.Set("keys", required_keys)
	values.Set("transaction", outerCommon.ToHex(data))
	err = rpc.WalletPost("/wallet/signTransaction", values.Encode(), &result)
	if nil == err {
		trx.Deserialize(outerCommon.FromHex(result.Result))
	}
	return err
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

	var result common.SimpleResult
	values := url.Values{}
	values.Set("name", name)
	err := rpc.WalletPost("/wallet/createKey", values.Encode(), &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
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

	var result common.SimpleResult
	values := url.Values{}
	values.Set("name", name)
	values.Set("password", passwd)
	err := rpc.WalletPost("/wallet/openWallet", values.Encode(), &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
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

	var result common.SimpleResult
	values := url.Values{}
	values.Set("name", name)
	err := rpc.WalletPost("/wallet/lockWallet", values.Encode(), &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
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

	var result common.SimpleResult
	values := url.Values{}
	values.Set("name", name)
	values.Set("password", passwd)
	err := rpc.WalletPost("/wallet/unlockWallet", values.Encode(), &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
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

	privateKey := c.String("private")
	if "" == privateKey {
		fmt.Println("Invalid private key")
		return errors.New("Invalid private key")
	}

	var result common.SimpleResult
	values := url.Values{}
	values.Set("name", name)
	values.Set("privateKey", privateKey)
	err := rpc.WalletPost("/wallet/importKey", values.Encode(), &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
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

	var result common.SimpleResult
	values := url.Values{}
	values.Set("name", name)
	values.Set("password", password)
	values.Set("publickey", publicKey)
	err := rpc.WalletPost("/wallet/removeKey", values.Encode(), &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
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

	var result common.SimpleResult
	values := url.Values{}
	values.Set("name", name)
	values.Set("password", passwd)
	err := rpc.WalletPost("/wallet/listKey", values.Encode(), &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
}

func listWallets(c *cli.Context) error {
	var result common.SimpleResult
	err := rpc.WalletGet("/wallet/listWallets", &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
}
