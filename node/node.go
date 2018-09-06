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
	"syscall"

	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/consensus/solo"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/http/rpc"
	"github.com/ecoball/go-ecoball/net"
	"github.com/ecoball/go-ecoball/txpool"
	"github.com/urfave/cli"

	"github.com/ecoball/go-ecoball/consensus/dpos"

	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/consensus/ababft"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common"
)

var (
	RunCommand = cli.Command{
		Name:   "run",
		Usage:  "run node",
		Action: runNode,
	}
)

func runNode(c *cli.Context) error {
	net.InitNetWork()
	shutdown := make(chan bool, 1)
	ecoballGroup, ctx := errgroup.WithContext(context.Background())

	fmt.Println("Run Node")
	log.Info("Build Geneses Block")
	var err error
	ledger.L, err = ledgerimpl.NewLedger(store.PathBlock, config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey))
	if err != nil {
		log.Fatal(err)
	}

	//start transaction pool
	txPool, err := txpool.Start(ledger.L)
	if err != nil {
		log.Fatal("start txPool error, ", err.Error())
		os.Exit(1)
	}

	log.Info("consensus", config.ConsensusAlgorithm)
	//start consensus
	switch config.ConsensusAlgorithm {
	case "SOLO":
		solo.NewSoloConsensusServer(ledger.L, txPool, config.User)
		event.Send(event.ActorNil, event.ActorConsensusSolo, &message.RegChain{ChainID: config.ChainHash, Address: common.AddressFromPubKey(config.Root.PublicKey)})
	case "DPOS":
		log.Info("Start DPOS consensus")

		c, _ := dpos.NewDposService()
		c.Setup(ledger.L, txPool)
		c.Start()

	case "ABABFT":
		var acc account.Account
		acc = config.Worker
		serviceConsensus, _ := ababft.ServiceABABFTGen(ledger.L, txPool, &acc)
		println("build the ababft service")
		serviceConsensus.Start()
		println("start the ababft service")
		if ledger.L.StateDB(config.ChainHash).RequireVotingInfo() {
			event.Send(event.ActorNil, event.ActorConsensus, message.ABABFTStart{config.ChainHash})
		}
	default:
		log.Fatal("unsupported consensus algorithm:", config.ConsensusAlgorithm)
	}

	// do something before start the network
	//TOD

	net.StartNetWork()

	//start blockchain browser
	/*ecoballGroup.Go(func() error {
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
	})*/

	//start http server
	ecoballGroup.Go(func() error {
		errChan := make(chan error, 1)
		go func() {
			if err := rpc.StartRPCServer(); nil != err {
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
	log.Info("ecoball received signal:", sig)
	close(shutdown)
}
