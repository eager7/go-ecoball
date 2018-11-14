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

package common

import (
	"fmt"
	"os"

	innerCommon "github.com/ecoball/go-ecoball/common"
	walletHttp "github.com/ecoball/go-ecoball/walletserver/http"
	"github.com/urfave/cli"
)

var (
	NodeIp     string
	NodePort   = "20678"
	WalletIp   string
	WalletPort = "20679"
)

func RpcAddress() string {
	address := "http://" + NodeIp + ":" + NodePort
	return address
}

func WalletRpcAddress() string {
	address := "http://" + WalletIp + ":" + WalletPort
	return address
}

// FileExisted checks whether filename exists in filesystem
func FileExisted(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

//default action function
func DefaultAction(c *cli.Context) error {
	args := c.Args()
	if args.Present() {
		if err := cli.ShowCommandHelp(c, args.First()); nil != err {
			fmt.Fprintln(os.Stderr, err)
		}
		return nil
	}

	if err := cli.ShowSubcommandHelp(c); nil != err {
		fmt.Fprintln(os.Stderr, err)
	}
	return nil
}

//get real public key
func IntersectionKeys(allKeys walletHttp.Keys, permission []innerCommon.Address) walletHttp.Keys {
	result := walletHttp.Keys{KeyList: []walletHttp.OneKey{}}
	for _, v := range allKeys.KeyList {
		addr := innerCommon.AddressFromPubKey(v.Key)
		for _, one := range permission {
			if addr == one {
				result.KeyList = append(result.KeyList, v)
			}
		}
	}

	return result
}
