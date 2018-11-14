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
	"net"

	"github.com/ecoball/go-ecoball/client/commands"
	"github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/dsn/host/cmd"
	"github.com/peterh/liner"
	"github.com/urfave/cli"
)

const (
	SUCCESS = 0
	FAILED  = 1
)

var (
	historyFilePath = filepath.Join(os.TempDir(), ".ecoclient_history")
	commandName     = []string{"storage"}
	commandMap      = make(map[string][]string)
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
	app.HideHelp = false
	app.HideVersion = false

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "console",
			Usage: "open ecoball client console",
		},
	}

	//commands
	app.Commands = []cli.Command{
		commands.ContractCommands,
		commands.SystemCommands,
		commands.TransferCommands,
		commands.WalletCommands,
		commands.QueryCommands,
		commands.AttachCommands,
		commands.CreateCommands,
		commands.DsnStorageCommands,
	}

	//set default action
	app.Action = common.DefaultAction

	sort.Sort(cli.CommandsByName(app.Commands))

	return app
}

func getLocalIP() error{
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				common.NodeIp = ipnet.IP.String()
				common.WalletIp = common.NodeIp
				break
			}
		}
	}

	return nil
}

func main() {
	if err := getLocalIP(); nil != err{
		return
	}

	//interrupt handle
	go interruptHandle()

	//client
	app := newClientApp()

	//Collect command
	for _, command := range app.Commands {
		commandName = append(commandName, command.Name)
		commandMap[command.Name] = []string{}
		if nil != command.Subcommands && 0 != len(command.Subcommands) {
			for _, subCommand := range command.Subcommands {
				commandMap[command.Name] = append(commandMap[command.Name], subCommand.Name)
			}
		}
	}

	//console
	app.After = func(c *cli.Context) error {
		var err error
		if c.Bool("console") {
			err = newConsole()
		}
		return err
	}

	//run
	var result int
	if err := appRun(app); nil != err {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		result = FAILED
	} else {
		result = SUCCESS
	}

	os.Exit(result)
}

func newConsole() error {
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

		subLineTemp := strings.Fields(line)
		subLine := make([]string, 0, len(subLineTemp))
		for _, oneTemp := range subLineTemp {
			one := strings.Trim(oneTemp, " ")
			if "" != one {
				subLine = append(subLine, one)
			}
		}
		if 2 == len(subLine) {
			for command, subCommand := range commandMap {
				if strings.ToLower(subLine[0]) == command {
					for _, onecommand := range subCommand {
						if strings.HasPrefix(onecommand, strings.ToLower(subLine[1])) {
							subLine[1] = onecommand
							realLine := strings.Join(subLine, " ")
							c = append(c, realLine)
						}
					}
				}
			}
		}

		return
	})

	//read history info
	if historyFile, err := os.Open(historyFilePath); nil == err {
		state.ReadHistory(historyFile)
		historyFile.Close()
	}

	defer func() {
		if historyFile, err := os.Create(historyFilePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing history file: %s\n", err.Error())
		} else {
			state.WriteHistory(historyFile)
			historyFile.Close()
		}
	}()

	//console
	prompt := "ecoclient: \\>"
	for {
		line, errLine := state.Prompt(prompt)
		if nil != errLine {
			return errLine
		} else {
			state.AppendHistory(line)
			if line != "exit" {
				if err := handleLine(line); nil != err {
					fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
				}
			} else {
				return nil
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
	if len(os.Args) >= 2 && os.Args[1] == "storage" {
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
