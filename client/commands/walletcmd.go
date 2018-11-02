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
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	walletHttp "github.com/ecoball/go-ecoball/walletserver/http"
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
				Name:   "listkeys",
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
			{
				Name:   "setTimeout",
				Usage:  "Set the lock interval of wallet",
				Action: setTimeout,
				Flags: []cli.Flag{
					cli.IntFlag{
						Name:  "interval, i",
						Usage: "the lock interval of wallet.",
					},
				},
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
	requestData := walletHttp.WalletNamePassword{Name: name, Password: passwd}
	err := rpc.WalletPost("/wallet/create", &requestData, &result)
	if nil == err {
		fmt.Println(result.Result)
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

	var result walletHttp.KeyPair
	requestData := walletHttp.WalletName{Name: name}
	err := rpc.WalletPost("/wallet/createKey", &requestData, &result)
	if nil == err {
		fmt.Println("PrivateKey:", hex.EncodeToString(result.PrivateKey))
		fmt.Println("PublicKey:", hex.EncodeToString(result.PublicKey))
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
	requestData := walletHttp.WalletNamePassword{Name: name, Password: passwd}
	err := rpc.WalletPost("/wallet/openWallet", &requestData, &result)
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
	requestData := walletHttp.WalletName{Name: name}
	err := rpc.WalletPost("/wallet/lockWallet", &requestData, &result)
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
	requestData := walletHttp.WalletNamePassword{Name: name, Password: passwd}
	err := rpc.WalletPost("/wallet/unlockWallet", &requestData, &result)
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

	privateKeyStr := c.String("private")
	if "" == privateKeyStr {
		fmt.Println("Invalid private key")
		return errors.New("Invalid private key")
	}

	privateKey, err := hex.DecodeString(privateKeyStr)
	if nil != err {
		fmt.Println(err)
		return err
	}

	var result walletHttp.OneKey
	oneKey := walletHttp.OneKey{privateKey}
	requestData := walletHttp.WalletImportKey{Name: name, PriKey: oneKey}
	err = rpc.WalletPost("/wallet/importKey", &requestData, &result)
	if nil == err {
		fmt.Println("PublicKey:", hex.EncodeToString(result.Key))
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

	passwd := c.String("password")
	if "" == passwd {
		fmt.Println("Invalid wallet password")
		return errors.New("Invalid password")
	}

	publicKeyStr := c.String("public")
	if "" == publicKeyStr {
		fmt.Println("Invalid public key")
		return errors.New("Invalid public key")
	}

	publicKey, err := hex.DecodeString(publicKeyStr)
	if nil != err {
		fmt.Println(err)
		return err
	}

	var result common.SimpleResult
	oneKey := walletHttp.OneKey{publicKey}
	oneWallet := walletHttp.WalletNamePassword{Name: name, Password: passwd}
	requestData := walletHttp.WalletRemoveKey{NamePassword: oneWallet, PubKey: oneKey}
	err = rpc.WalletPost("/wallet/removeKey", &requestData, &result)
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

	var result walletHttp.KeyPairs
	requestData := walletHttp.WalletNamePassword{Name: name, Password: passwd}
	err := rpc.WalletPost("/wallet/listKey", &requestData, &result)
	if nil == err {
		for _, v := range result.Pairs {
			fmt.Println("PrivateKey:", hex.EncodeToString(v.PrivateKey), "/tPublicKey:", hex.EncodeToString(v.PublicKey))
		}
	}
	return err
}

func listWallets(c *cli.Context) error {
	var result walletHttp.Wallets
	err := rpc.WalletGet("/wallet/listWallets", &result)
	if nil == err {
		for _, oneWallet := range result.NameList {
			fmt.Println(oneWallet)
		}
	}
	return err
}

func setTimeout(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	interval := c.Int64("interval")
	if interval <= 0 {
		fmt.Println("Invalid lock interval of wallet(greater than 0)")
		return errors.New("Invalid lock interval of wallet")
	}

	var result common.SimpleResult
	requestData := walletHttp.WalletTimeout{Interval: interval}
	err := rpc.WalletPost("/wallet/setTimeout", &requestData, &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
}

//other wallet method
func getPublicKeys() (walletHttp.Keys, error) {
	var result walletHttp.Keys
	err := rpc.WalletGet("/wallet/getPublicKeys", &result)
	if nil == err {
		return result, nil
	}
	return walletHttp.Keys{}, err
}

func signTransaction(publickeys walletHttp.Keys, rawData []byte) (walletHttp.TransactionData, error) {
	var result walletHttp.TransactionData
	oneTransaction := TransactionData{rawData}
	requestData := RawTransactionData{PublicKeys: publickeys, RawTransaction: oneTransaction}
	err = rpc.WalletPost("/wallet/signTransaction", &requestData, &result)
	if nil == err {
		return result, nil
	}
	return walletHttp.TransactionData{}, err
}
