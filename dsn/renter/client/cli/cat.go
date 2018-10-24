package cli

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"time"

	oldcmds "github.com/ipfs/go-ipfs/commands"
	core "github.com/ipfs/go-ipfs/core"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	"gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds"
	loggables "gx/ipfs/QmRPkGkHLB72caXgdDYnoaWigXNWx95BcYDKV1n3KTEpaG/go-libp2p-loggables"
	logging "gx/ipfs/QmcVVHfdyv15GVPk7NrxdWjh2hLVccXnoD8j2tyQShiXJb/go-log"
	dsncmd "github.com/ecoball/go-ecoball/dsn/cmd"
)

func CliCatFile() error {
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
		repoPath, err := getRepoPath(req)
		if err != nil {
			return nil, err
		}

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

	err := catRun(ctx, dsncmd.Root, os.Args, os.Stdin, os.Stdout, os.Stderr, buildEnv, makeExecutor)
	if err != nil {
		return  nil
	}
	return nil
}
