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
	"path/filepath"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	"github.com/ipfs/go-ipfs/plugin/loader"
	ecoballConfig "github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/dsn/host/cmd"
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