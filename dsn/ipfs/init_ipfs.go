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

	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"context"
	"sort"
	"path/filepath"

	"github.com/ipfs/go-ipfs/assets"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/namesys"
	"github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/urfave/cli"
	ecoballConfig "github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/dsn/cmd"
)

const (
	nBitsForKeypairDefault = 2048
)

type IpfsCtrl struct {
	IpfsNode      *core.IpfsNode
	//RepoStat      *StoreStatMonitor
}

var errRepoExists = errors.New(`ipfs configuration file already exists!
Reinitializing would overwrite your keys.
`)

var ipfsCtrl *IpfsCtrl
var IpfsNode *core.IpfsNode

func initWithDefaults(out io.Writer, repoRoot string, profile string) error {
	var profiles []string
	if profile != "" {
		profiles = strings.Split(profile, ",")
	}

	return doInit(out, repoRoot, false, nBitsForKeypairDefault, profiles, nil)
}

func doInit(out io.Writer, repoRoot string, empty bool, nBitsForKeypair int, confProfiles []string, conf *config.Config) error {
	if _, err := fmt.Fprintf(out, "initializing IPFS node at %s\n", repoRoot); err != nil {
		return err
	}

	if err := checkWritable(repoRoot); err != nil {
		return err
	}

	if fsrepo.IsInitialized(repoRoot) {
		return errRepoExists
	}

	if conf == nil {
		var err error
		conf, err = config.Init(out, nBitsForKeypair)
		if err != nil {
			return err
		}
	}

	for _, profile := range confProfiles {
		transformer, ok := config.Profiles[profile]
		if !ok {
			return fmt.Errorf("invalid configuration profile: %s", profile)
		}

		if err := transformer.Transform(conf); err != nil {
			return err
		}
	}

	if err := fsrepo.Init(repoRoot, conf); err != nil {
		return err
	}

	if !empty {
		if err := addDefaultAssets(out, repoRoot); err != nil {
			return err
		}
	}

	return initializeIpnsKeyspace(repoRoot)
}

func checkWritable(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		// dir exists, make sure we can write to it
		testfile := path.Join(dir, "test")
		fi, err := os.Create(testfile)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("%s is not writeable by the current user", dir)
			}
			return fmt.Errorf("unexpected error while checking writeablility of repo root: %s", err)
		}
		fi.Close()
		return os.Remove(testfile)
	}

	if os.IsNotExist(err) {
		// dir doesn't exist, check that we can create it
		return os.Mkdir(dir, 0775)
	}

	if os.IsPermission(err) {
		return fmt.Errorf("cannot write to %s, incorrect permissions", err)
	}

	return err
}

func addDefaultAssets(out io.Writer, repoRoot string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := fsrepo.Open(repoRoot)
	if err != nil { // NB: repo is owned by the node
		return err
	}

	nd, err := core.NewNode(ctx, &core.BuildCfg{Repo: r})
	if err != nil {
		return err
	}
	defer nd.Close()

	dkey, err := assets.SeedInitDocs(nd)
	if err != nil {
		return fmt.Errorf("init: seeding init docs failed: %s", err)
	}
	fmt.Printf("init: seeded init docs %s", dkey)

	if _, err = fmt.Fprintf(out, "to get started, enter:\n"); err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "\n\tipfs cat /ipfs/%s/readme\n\n", dkey)
	return err
}

func initializeIpnsKeyspace(repoRoot string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := fsrepo.Open(repoRoot)
	if err != nil { // NB: repo is owned by the node
		return err
	}

	nd, err := core.NewNode(ctx, &core.BuildCfg{Repo: r})
	if err != nil {
		return err
	}
	defer nd.Close()

	err = nd.SetupOfflineRouting()
	if err != nil {
		return err
	}

	return namesys.InitializeKeyspace(ctx, nd.Namesys, nd.Pinning, nd.PrivateKey)
}

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

func initIpfsNode(path string) (*core.IpfsNode, error) {
	//open debug
	//u.Debug = true
	//logging.SetDebugLogging()

	if !fsrepo.IsInitialized(path) {
		//TODO
		initWithDefaults(os.Stdout, path, "")
	}

	// acquire the repo lock _before_ constructing a node. we need to make
	// sure we are permitted to access the resources (datastore, etc.)
	repo, err := fsrepo.Open(path)
	switch err {
	default:
		//TODO
	case fsrepo.ErrNeedMigration:
		//TODO
	case nil:
		break
	}

	cfg, err := fsrepo.ConfigAt(path)
	if err != nil {
		//TODO
	}

	//offline := false
	ipnsps := true
	pubsub := true
	mplex := true

	// Start assembling node config
	ncfg := &core.BuildCfg{
		Repo:      repo,
		Permanent: true, // It is temporary way to signify that node is permanent
		Online:    true,
		ExtraOpts: map[string]bool{
			"pubsub": pubsub,
			"ipnsps": ipnsps,
			"mplex":  mplex,
		},
		//TODO(Kubuxu): refactor Online vs Offline by adding Permanent vs Ephemeral
	}

	loadIpldPlugin()

	//rcfg, err := repo.Config()
	//if err != nil {
	//re.SetError(err, cmdkit.ErrNormal)
	//return
	//}

	ncfg.Routing = core.DHTOption

	//???
	//cctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	node, err := core.NewNode(context.Background(), ncfg)
	if err != nil {
		fmt.Printf("error from node construction: ", err)
		//re.SetError(err, cmdkit.ErrNormal)
		return nil, err
	}
	node.SetLocal(false)

	printSwarmAddrs(node)

	if node.PNetFingerprint != nil {
		//fmt.Println("Swarm is limited to private network of peers with the swarm key")
		//fmt.Printf("Swarm key fingerprint: %x\n", node.PNetFingerprint)
	}

	//TODO serveHTTPApi(req, cctx)

	//var gwErrc <-chan error
	if len(cfg.Addresses.Gateway) > 0 {

	}

	return node, nil
}

func InitAndRunIpfs(path string) (*IpfsCtrl, error) {
	var err error
	IpfsNode, err = initIpfsNode(path)
	if (err != nil) {
		//log.Error("error for initializing ipfs node:", err)
		return nil, err
	}

	//storeStatEngine := NewStoreStatMonitor()

	ipfsCtrl = &IpfsCtrl{
		IpfsNode,
		//storeStatEngine,
	}

	return ipfsCtrl, nil
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
func Initialize(c *cli.Context) error {
	if fsrepo.IsInitialized(ecoballConfig.IpfsDir) {
		return nil
	}
	cmd.Root.Subcommands["init"] = initCmd
	os.Args[1] = "init"
	return cmd.StorageFun()
}

//start storage
func DaemonRun(c *cli.Context) error {
	cmd.Root.Subcommands["daemon"] = daemonCmd
	os.Args[1] = "daemon"
	return cmd.StorageFun()
}