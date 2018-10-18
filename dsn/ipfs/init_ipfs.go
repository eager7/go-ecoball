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

package ipfs

import (
	"fmt"
	"os"
	"sort"
	"path/filepath"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/urfave/cli"
	ecoballConfig "github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/dsn/cmd"
)
// load ecoball ipfs ipld format plugin
func loadIpldPlugin() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	dir = filepath.Join(dir, "/plugins")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Errorf("Missing Ecoball ipld plugin file!")
	}

	if _, err := loader.LoadPlugins(dir); err != nil {
		fmt.Println("error loading plugins: ", err)
	}
}

// printSwarmAddrs prints the addresses of the host
func printSwarmAddrs(node *core.IpfsNode) {
	if !node.OnlineMode() {
		fmt.Println("Swarm not listening, running in offline mode.")
		return
	}

	var lisAddrs []string
	ifaceAddrs, err := node.PeerHost.Network().InterfaceListenAddresses()
	if err != nil {
		//log.Error("failed to read listening addresses: %s", err)
	}
	for _, addr := range ifaceAddrs {
		lisAddrs = append(lisAddrs, addr.String())
	}
	sort.Sort(sort.StringSlice(lisAddrs))
	for _, addr := range lisAddrs {
		fmt.Printf("Swarm listening on %s\n", addr)
	}
}
//initialize
func Initialize() error {
	if fsrepo.IsInitialized(ecoballConfig.IpfsDir) {
		return nil
	}
	cmd.Root.Subcommands["init"] = initCmd
	os.Args[1] = "init"
	return cmd.StorageFun()
}

//start storage
func DaemonRun() error {
	cmd.Root.Subcommands["daemon"] = daemonCmd
	os.Args[1] = "daemon"
	return cmd.StorageFun()
}