// cmd/ipfs implements the primary CLI binary for ipfs
package cmd

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	oldcmds "github.com/ipfs/go-ipfs/commands"
	core "github.com/ipfs/go-ipfs/core"
	corecmds "github.com/ipfs/go-ipfs/core/commands"
	corehttp "github.com/ipfs/go-ipfs/core/corehttp"
	loader "github.com/ipfs/go-ipfs/plugin/loader"
	repo "github.com/ipfs/go-ipfs/repo"
	config "github.com/ipfs/go-ipfs/repo/config"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	"gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds"
	"gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds/cli"
	"gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds/http"
	u "gx/ipfs/QmPdKqUcHGFdeSpvjVoaTRPPstGif9GBZb5Q56RVw9o69A/go-ipfs-util"
	loggables "gx/ipfs/QmRPkGkHLB72caXgdDYnoaWigXNWx95BcYDKV1n3KTEpaG/go-libp2p-loggables"
	manet "gx/ipfs/QmV6FjemM1K8oXjrvuq3wuVWWoU2TLDPmNnKrxHzY3v6Ai/go-multiaddr-net"
	osh "gx/ipfs/QmXuBJ7DR6k3rmUEKtvVMhwjmXDuJgXXPUt4LQXKBMsU93/go-os-helper"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	logging "gx/ipfs/QmcVVHfdyv15GVPk7NrxdWjh2hLVccXnoD8j2tyQShiXJb/go-log"
)

// log is the command logger
var log = logging.Logger("cmd/ipfs")

var errRequestCanceled = errors.New("request canceled")

func StorageFun() error {
	rand.Seed(time.Now().UnixNano())
	ctx := logging.ContextWithLoggable(context.Background(), loggables.Uuid("session"))

	// Handle `ipfs help'
	if len(os.Args) == 2 {
		if os.Args[1] == "help" {
			os.Args[1] = "-h"
		} else if os.Args[1] == "--version" {
			os.Args[1] = "version"
		}
	}

	// output depends on executable name passed in os.Args
	// so we need to make sure it's stable
	os.Args[0] = "storage"

	buildEnv := func(ctx context.Context, req *cmds.Request) (cmds.Environment, error) {
		checkDebug(req)
		repoPath, err := getRepoPath(req)
		if err != nil {
			return nil, err
		}
		log.Debugf("config path is %s", repoPath)

		// this sets up the function that will initialize the node
		// this is so that we can construct the node lazily.
		return &oldcmds.Context{
			ConfigRoot: repoPath,
			LoadConfig: loadConfig,
			ReqLog:     &oldcmds.ReqLog{},
			ConstructNode: func() (n *core.IpfsNode, err error) {
				if req == nil {
					return nil, errors.New("constructing node without a request")
				}

				r, err := fsrepo.Open(repoPath)
				if err != nil { // repo is owned by the node
					return nil, err
				}

				// ok everything is good. set it on the invocation (for ownership)
				// and return it.
				n, err = core.NewNode(ctx, &core.BuildCfg{
					Repo: r,
				})
				if err != nil {
					return nil, err
				}

				n.SetLocal(true)
				return n, nil
			},
		}, nil
	}

	err := cli.Run(ctx, Root, os.Args, os.Stdin, os.Stdout, os.Stderr, buildEnv, makeExecutor)
	return err
}

func checkDebug(req *cmds.Request) {
	// check if user wants to debug. option OR env var.
	debug, _ := req.Options["debug"].(bool)
	if debug || os.Getenv("IPFS_LOGGING") == "debug" {
		u.Debug = true
		logging.SetDebugLogging()
	}
	if u.GetenvBool("DEBUG") {
		u.Debug = true
	}
}

func makeExecutor(req *cmds.Request, env interface{}) (cmds.Executor, error) {
	details := commandDetails(req.Path)
	client, err := commandShouldRunOnDaemon(*details, req, env.(*oldcmds.Context))
	if err != nil {
		return nil, err
	}

	var exctr cmds.Executor
	if client != nil && !req.Command.External {
		exctr = client.(cmds.Executor)
	} else {
		cctx := env.(*oldcmds.Context)
		pluginpath := filepath.Join(cctx.ConfigRoot, "plugins")

		// check if repo is accessible before loading plugins
		ok, err := checkPermissions(cctx.ConfigRoot)
		if err != nil {
			return nil, err
		}
		if ok {
			if _, err := loader.LoadPlugins(pluginpath); err != nil {
				log.Error("error loading plugins: ", err)
			}
		}

		exctr = cmds.NewExecutor(req.Root)
	}

	return exctr, nil
}

func checkPermissions(path string) (bool, error) {
	_, err := os.Open(path)
	if os.IsNotExist(err) {
		// repo does not exist yet - don't load plugins, but also don't fail
		return false, nil
	}
	if os.IsPermission(err) {
		// repo is not accessible. error out.
		return false, fmt.Errorf("error opening repository at %s: permission denied", path)
	}

	return true, nil
}

// commandDetails returns a command's details for the command given by |path|.
func commandDetails(path []string) *cmdDetails {
	var details cmdDetails
	// find the last command in path that has a cmdDetailsMap entry
	for i := range path {
		if cmdDetails, found := cmdDetailsMap[strings.Join(path[:i+1], "/")]; found {
			details = cmdDetails
		}
	}
	return &details
}

// commandShouldRunOnDaemon determines, from command details, whether a
// command ought to be executed on an ipfs daemon.
//
// It returns a client if the command should be executed on a daemon and nil if
// it should be executed on a client. It returns an error if the command must
// NOT be executed on either.
func commandShouldRunOnDaemon(details cmdDetails, req *cmds.Request, cctx *oldcmds.Context) (http.Client, error) {
	path := req.Path
	// root command.
	if len(path) < 1 {
		return nil, nil
	}

	if details.cannotRunOnClient && details.cannotRunOnDaemon {
		return nil, fmt.Errorf("command disabled: %s", path[0])
	}

	if details.doesNotUseRepo && details.canRunOnClient() {
		return nil, nil
	}

	// at this point need to know whether api is running. we defer
	// to this point so that we don't check unnecessarily

	// did user specify an api to use for this command?
	apiAddrStr, _ := req.Options[corecmds.ApiOption].(string)

	client, err := getApiClient(cctx.ConfigRoot, apiAddrStr)
	if err == repo.ErrApiNotRunning {
		if apiAddrStr != "" {
			// if user SPECIFIED an api, and this cmd is not daemon
			// we MUST use it. so error out.
			return nil, err
		}

		// ok for api not to be running
	} else if err != nil { // some other api error
		return nil, err
	}

	if client != nil {
		if details.cannotRunOnDaemon {
			// check if daemon locked. legacy error text, for now.
			log.Debugf("Command cannot run on daemon. Checking if daemon is locked")
			if daemonLocked, _ := fsrepo.LockedByOtherProcess(cctx.ConfigRoot); daemonLocked {
				return nil, cmds.ClientError("ipfs daemon is running. please stop it to run this command")
			}
			return nil, nil
		}

		return client, nil
	}

	if details.cannotRunOnClient {
		return nil, cmds.ClientError("must run on the ipfs daemon")
	}

	return nil, nil
}

func getRepoPath(req *cmds.Request) (string, error) {
	repoOpt, found := req.Options["config"].(string)
	if found && repoOpt != "" {
		return repoOpt, nil
	}

	repoPath := "."

	return repoPath, nil
}

func loadConfig(path string) (*config.Config, error) {
	return fsrepo.ConfigAt(path)
}

var apiFileErrorFmt string = `Failed to parse '%[1]s/api' file.
	error: %[2]s
If you're sure go-ipfs isn't running, you can just delete it.
`
var checkIPFSUnixFmt = "Otherwise check:\n\tps aux | grep ipfs"
var checkIPFSWinFmt = "Otherwise check:\n\ttasklist | findstr ipfs"

// getApiClient checks the repo, and the given options, checking for
// a running API service. if there is one, it returns a client.
// otherwise, it returns errApiNotRunning, or another error.
func getApiClient(repoPath, apiAddrStr string) (http.Client, error) {
	var apiErrorFmt string
	switch {
	case osh.IsUnix():
		apiErrorFmt = apiFileErrorFmt + checkIPFSUnixFmt
	case osh.IsWindows():
		apiErrorFmt = apiFileErrorFmt + checkIPFSWinFmt
	default:
		apiErrorFmt = apiFileErrorFmt
	}

	var addr ma.Multiaddr
	var err error
	if len(apiAddrStr) != 0 {
		addr, err = ma.NewMultiaddr(apiAddrStr)
		if err != nil {
			return nil, err
		}
		if len(addr.Protocols()) == 0 {
			return nil, fmt.Errorf("multiaddr doesn't provide any protocols")
		}
	} else {
		addr, err = fsrepo.APIAddr(repoPath)
		if err == repo.ErrApiNotRunning {
			return nil, err
		}

		if err != nil {
			return nil, fmt.Errorf(apiErrorFmt, repoPath, err.Error())
		}
	}
	if len(addr.Protocols()) == 0 {
		return nil, fmt.Errorf(apiErrorFmt, repoPath, "multiaddr doesn't provide any protocols")
	}
	return apiClientForAddr(addr)
}

func apiClientForAddr(addr ma.Multiaddr) (http.Client, error) {
	_, host, err := manet.DialArgs(addr)
	if err != nil {
		return nil, err
	}

	return http.NewClient(host, http.ClientWithAPIPrefix(corehttp.APIPath)), nil
}
