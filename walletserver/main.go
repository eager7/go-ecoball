// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/walletserver/http"
	"github.com/ecoball/go-ecoball/walletserver/wallet"
	"github.com/urfave/cli"
)

var (
	log          = elog.NewLogger("wallet", elog.DebugLog)
	startCommand = cli.Command{
		Name:   "start",
		Usage:  "start ecowallet service",
		Action: startServive,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "path, p",
				Usage: "Wallet file store path, the default path is the wallet directory of the current working directory",
			},
		},
	}
)

func main() {
	app := cli.NewApp()

	//set attribute of EcoBall
	app.Name = "ecowallet"
	app.HelpName = "ecowallet"
	app.Usage = "Ecowallet from QuakerChain Technology"
	app.UsageText = "Ecowallet is a secure, independently deployed wallet system"
	app.Copyright = "2018 ecoball. All rights reserved"
	app.Author = "EcoBall"
	app.Email = "service@ecoball.org"
	app.HideHelp = true
	app.HideVersion = true

	//commands
	app.Commands = []cli.Command{
		startCommand,
	}

	//run
	app.Run(os.Args)
}

func startServive(c *cli.Context) error {
	walletPath := c.String("path")

	if "" == walletPath {
		rootDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		rootDir = strings.Replace(rootDir, "\\", "/", -1)
		wallet.Dir = path.Join(rootDir, "wallet/")
	} else {
		wallet.Dir = path.Join(walletPath, "wallet/")
	}

	if _, err := os.Stat(wallet.Dir); os.IsNotExist(err) {
		if err := os.MkdirAll(wallet.Dir, 0777); err != nil {
			fmt.Println("could not create directory:", wallet.Dir, err)
		}
	}

	// http server
	go http.StartHttpServer()

	wait()
	return nil
}

//capture single
func wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	log.Info("ecowallet received signal:", sig)
}
