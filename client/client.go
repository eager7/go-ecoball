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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/ecoball/go-ecoball/client/commands"
	"github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/dsn/cmd"
	"github.com/peterh/liner"
	"github.com/urfave/cli"
)

const STORAGE = "storage"

var (
	historyFilePath = filepath.Join(os.TempDir(), ".ecoclient_history")
	commandName     = []string{"contract", "transfer", "wallet", "query", "attach", STORAGE}
)

func newClientApp() *cli.App {
	app := cli.NewApp()

	//set attribute of client
	app.Name = "ecoclient"
	app.Version = config.EcoVersion
	app.HelpName = "ecoclient"
	app.Usage = "command line tool for ecoball"
	app.UsageText = "ecoclient [global options] command [command options] [args]"
	app.Copyright = "2018 ecoball. All rights reserved"
	app.Author = "ecoball"
	app.Email = "service@ecoball.org"
	app.HideHelp = true
	app.HideVersion = true

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "console",
			Usage: "open ecoball client console",
		},
	}

	//commands
	app.Commands = []cli.Command{
		commands.ContractCommands,
		commands.TransferCommands,
		commands.WalletCommands,
		commands.QueryCommands,
		commands.AttachCommands,
		commands.CreateCommands,
		commands.NetworkCommand,
		commands.StorageCommands,
	}

	//set default action
	app.Action = common.DefaultAction

	sort.Sort(cli.CommandsByName(app.Commands))

	return app
}

func main() {
	//interrupt handle
	go interruptHandle()

	//client
	app := newClientApp()

	//console
	app.After = func(c *cli.Context) error {
		if c.Bool("console") {
			newConsole()
		}
		return nil
	}

	//run
	var result int
	if err := appRun(app); nil != err {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		result = 1
	} else {
		result = 0
	}

	os.Exit(result)
}

func newConsole() {
	state := liner.NewLiner()
	defer func() {
		state.Close()
		if err := recover(); err != nil {
			fmt.Println("panic occur:", err)
		}
	}()

	//set attribute of console
	state.SetCtrlCAborts(true)
	state.SetTabCompletionStyle(liner.TabPrints)
	state.SetMultiLineMode(true)
	state.SetCompleter(func(line string) (c []string) {
		for _, n := range commandName {
			if strings.HasPrefix(n, strings.ToLower(line)) {
				c = append(c, n)
			}
		}
		return
	})

	//read history info
	var historyFile *os.File
	var err error
	if common.FileExisted(historyFilePath) {
		historyFile, err = os.Open(historyFilePath)
		_, err = state.ReadHistory(historyFile)
	} else {
		historyFile, err = os.Create(historyFilePath)
	}

	if nil != err {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		state.WriteHistory(historyFile)
		historyFile.Close()
	}()

	//new console
	scheduler := make(chan string)

	go func() {
		for {
			info := <-scheduler
			line, errLine := state.Prompt(info)
			if errLine == nil {
				state.AppendHistory(line)
				scheduler <- line
			} else if errLine == liner.ErrPromptAborted {
				fmt.Println("Aborted")
				close(scheduler)
				return
			} else {
				fmt.Println("Error reading line: ", errLine)
				close(scheduler)
				return
			}
		}
	}()

	//single abort
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	scheduler <- "ecoclient: \\>"
	for {
		select {
		case <-sig:
			fmt.Println("exit signal")
			return
		case line, ok := <-scheduler:
			if ok {
				if "exit" == line {
					return
				} else {
					if err := handleLine(line); nil != err {
						fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
					}
				}
				scheduler <- "ecoclient: \\>"
			}
		}
	}
}

func handleLine(line string) error {
	args := []string{os.Args[0]}
	lines := strings.Fields(line)
	args = append(args, lines...)
	os.Args = args

	//run
	return appRun(newClientApp())
}

func appRun(app *cli.App) (err error) {
	if len(os.Args) >= 2 && os.Args[1] == STORAGE {
		temp := make([]string, 0, len(os.Args))
		temp = append(temp, os.Args[0])
		temp = append(temp, os.Args[2:]...)
		os.Args = temp
		return cmd.StorageFun()
	}
	return app.Run(os.Args)
}

func interruptHandle() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	fmt.Println("ecoclient received signal:", sig)
	os.Exit(1)
}
