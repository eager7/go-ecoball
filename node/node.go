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
	"github.com/ecoball/go-ecoball/mobsync"
	"os"
	"os/signal"
	"syscall"

	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/consensus/solo"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/http/rpc"
	"github.com/ecoball/go-ecoball/txpool"
	"github.com/urfave/cli"

	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/lib-p2p"
	"github.com/ecoball/go-ecoball/spectator"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"fmt"
	"github.com/ecoball/go-ecoball/sharding"
)

var (
	RunCommand = cli.Command{
		Name:   "run",
		Usage:  "run node",
		Action: runNode,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "bits, b",
				Usage: "Number of bits to use in the generated RSA private key.",
				Value: 2048,
			},
			cli.BoolFlag{
				Name:  "empty-repo, e",
				Usage: "Don't add and pin help files to the local storage.",
			},
			cli.StringFlag{
				Name:  "profile, p",
				Usage: "Apply profile settings to config. Multiple profiles can be separated by ','",
			},
			cli.StringFlag{
				Name:  "routing",
				Usage: "Overrides the routing option",
				Value: "default",
			},
			cli.BoolFlag{
				Name:  "mount",
				Usage: "Mounts this node to the filesystem",
			},
			cli.BoolFlag{
				Name:  "writable",
				Usage: "Enable writing objects (with POST, PUT and DELETE)",
			},
			cli.StringFlag{
				Name:  "mount-ipfs",
				Usage: "Path to the mountpoint for IPFS (if using --mount). Defaults to config setting.",
			},
			cli.StringFlag{
				Name:  "mount-ipns",
				Usage: "Path to the mountpoint for IPNS (if using --mount). Defaults to config setting.",
			},
			cli.BoolFlag{
				Name:  "unrestricted-api",
				Usage: "Allow API access to unlisted hashes",
			},
			cli.BoolFlag{
				Name:  "disable-transport-encryption",
				Usage: "Disable transport encryption (for debugging protocols)",
			},
			cli.BoolFlag{
				Name:  "enable-gc",
				Usage: "Enable automatic periodic repo garbage collection",
			},
			cli.BoolFlag{
				Name:  "manage-fdlimit",
				Usage: "Check and raise file descriptor limits if needed",
			},
			cli.BoolFlag{
				Name:  "offline",
				Usage: "Run offline. Do not connect to the rest of the network but provide local API.",
			},
			cli.BoolFlag{
				Name:  "migrate",
				Usage: "If true, assume yes at the migrate prompt. If false, assume no.",
			},
			cli.BoolFlag{
				Name:  "enable-pubsub-experiment",
				Usage: "Instantiate the ipfs daemon with the experimental pubsub feature enabled.",
			},
			cli.BoolFlag{
				Name:  "enable-namesys-pubsub",
				Usage: "Enable IPNS record distribution through pubsub; enables pubsub.",
			},
			cli.BoolFlag{
				Name:  "enable-mplex-experiment",
				Usage: "Add the experimental 'go-multiplex' stream muxer to libp2p on construction.",
			},
		},
	}
)

func runNode(c *cli.Context) error {
	shutdown := make(chan bool, 1)
	ecoballGroup, ctx := errgroup.WithContext(context.Background())

	event.InitMsgDispatcher()
	p2p.InitNetWork(ctx)

	log.Info("Build Geneses Block")
	var err error
	ledger.L, err = ledgerimpl.NewLedger(config.RootDir+store.PathBlock, config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey))
	if err != nil {
		log.Fatal(err)
	}
	
	var sdActor *sharding.ShardingActor
	//if !config.DisableSharding {
	if (false) {
		log.Info("start sharding")
		sdActor, _ = sharding.NewShardingActor(ledger.L)
		sdActor.SetNet()
	}
	fmt.Println(sdActor)


	txPool, err := txpool.Start(ctx, ledger.L)

	if err != nil {
		log.Fatal("start txPool error, ", err.Error())
		os.Exit(1)
	}

	log.Info("consensus", config.ConsensusAlgorithm)
	//start consensus
	switch config.ConsensusAlgorithm {
	case "SOLO":
		_, _ = solo.NewSoloConsensusServer(ledger.L, txPool, config.User)
	default:
		log.Fatal("unsupported consensus algorithm:", config.ConsensusAlgorithm)
	}
	if err := mobsync.NewSyncEngine(ctx, ledger.L); err != nil {
		log.Error(err)
		return err
	}
	//start block chain browser
	ecoballGroup.Go(func() error {
		errChan := make(chan error, 1)
		go func() {
			if err := spectator.Bystander(ledger.L); nil != err {
				errChan <- err
			}
		}()

		select {
		case <-ctx.Done():
		case <-shutdown:
		case err := <-errChan:
			log.Error("goroutine spectator error exit: ", err)
			return err
		}

		return nil
	})

	//start http server
	ecoballGroup.Go(func() error {
		errChan := make(chan error, 1)
		go func() {
			if err := rpc.StartHttpServer(); nil != err {
				errChan <- err
			}
		}()

		select {
		case <-ctx.Done():
		case <-shutdown:
		case err := <-errChan:
			log.Error("goroutine start http server error exit: ", err)
			return err
		}

		return nil
	})
	//capture single
	//go dsn.DsnHttpServ()
	go wait(shutdown)

	//Wait for each sub goroutine to exit
	if err := ecoballGroup.Wait(); err != nil {
		log.Error(err)
	}
	return nil
}

//capture single
func wait(shutdown chan bool) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	event.EventStop()
	log.Info("ecoball received signal:", sig)
	close(shutdown)
}
