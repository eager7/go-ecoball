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
	"fmt"

	"github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/urfave/cli"
)

var (
	AttachCommands = cli.Command{
		Name:     "attach",
		Usage:    "hang different nodes",
		Category: "Attach",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "ip, i",
				Usage: "node's ip address",
				Value: "localhost",
			},
			cli.StringFlag{
				Name:  "port, p",
				Usage: "node's RPC port",
				Value: "20678",
			},
		},
		Action: attach,
	}
)

func attach(c *cli.Context) error {
	//Check the number of flags
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	//ip address
	ip := c.String("ip")
	if "" != ip {
		common.NodeIp = ip
	}

	//port
	port := c.String("port")
	if "" != port {
		common.NodePort = port
	}

	var result rpc.SimpleResult
	err := rpc.NodeGet("/attach", &result)
	if nil == err {
		fmt.Println(result.Result)
	}
	return err
}
