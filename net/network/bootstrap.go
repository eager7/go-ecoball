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

// Implement a simple bootstrap function

package network

import (
	"context"
	"fmt"
	cfg "github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/thirdparty/math2"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmSF8fPo3jgVBAy8fpdjjYqgG87dkJgUprRBHRd2tmfgpP/goprocess"
	procctx "gx/ipfs/QmSF8fPo3jgVBAy8fpdjjYqgG87dkJgUprRBHRd2tmfgpP/goprocess/context"
	"gx/ipfs/QmSF8fPo3jgVBAy8fpdjjYqgG87dkJgUprRBHRd2tmfgpP/goprocess/periodic"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const (
	minPeerThreshold  = 4
	bootStrapInterval = 30 * time.Second
	bootStrapTimeOut  = bootStrapInterval / 3
)

type BootStrapper struct {
	closer  io.Closer
	bsPeers []cfg.BootstrapPeer
}

func (net *NetImpl) bootstrap(bsAddress []string) *BootStrapper {
	bsPeers, err := cfg.ParseBootstrapPeers(bsAddress)
	if err != nil {
		log.Error("failed to parse bootstrap address", err)
		return nil
	}
	if len(bsPeers) == 0 {
		return nil
	}

	connected := net.host.Network().Peers()
	if len(connected) >= minPeerThreshold {
		log.Debug("connected to the network,bootstrap skipped")
		return nil
	}
	numToDial := minPeerThreshold - len(connected)

	doneWithRound := make(chan struct{})
	periodic := func(worker goprocess.Process) {
		ctx := procctx.OnClosingContext(worker)

		if err := net.bootstrapConnect(ctx, bsPeers, numToDial); err != nil {
			log.Error(fmt.Sprintf("%s bootstrap error: %s", net.host.ID(), err))
		}
		<-doneWithRound
	}

	process := periodicproc.Tick(bootStrapInterval, periodic)
	process.Go(periodic) // run one right now.

	doneWithRound <- struct{}{}
	close(doneWithRound) // it no longer blocks periodic

	return &BootStrapper{process, bsPeers}
}

func (net *NetImpl) bootstrapConnect(ctx context.Context, bsPeers []cfg.BootstrapPeer, numToDial int) error {
	ctx, cancel := context.WithTimeout(ctx, bootStrapTimeOut)
	defer cancel()

	var notConnected []pstore.PeerInfo
	for _, p := range bsPeers {
		if net.host.Network().Connectedness(p.ID()) != inet.Connected {
			protoNum := len(p.Multiaddr().Protocols())
			sep := "/" + p.Multiaddr().Protocols()[protoNum-1].Name
			addr, _ := ma.NewMultiaddr(strings.Split(p.String(), sep)[0])
			peerInfo := pstore.PeerInfo{ID: p.ID(), Addrs: []ma.Multiaddr{addr}}
			notConnected = append(notConnected, peerInfo)
		}
	}

	if len(notConnected) < 1 {
		log.Error("not enough bootstrap peers to bootstrap")
	}

	peers := randomSubsetOfPeers(notConnected, numToDial)
	errs := make(chan error, len(peers))
	var wg sync.WaitGroup
	for _, p := range peers {
		wg.Add(1)
		go func(p pstore.PeerInfo) {
			defer wg.Done()
			log.Debug(fmt.Sprintf("%s bootstrapping to %s", net.host.ID(), p.ID))

			//bsnet.host.Peerstore().AddAddrs(p.ID, p.Addrs, pstore.PermanentAddrTTL)
			if err := net.host.Connect(net.ctx, p); err != nil {
				log.Error(fmt.Sprintf("failed to bootstrap with %v: %s", p.ID, err))
				errs <- err
				return
			}
			log.Debug(fmt.Sprintf("bootstrapped with %v", p.ID))
		}(p)
	}

	wg.Wait()
	close(errs)
	count := 0
	var err error
	for err = range errs {
		if err != nil {
			count++
		}
	}
	if len(peers) > 0 && count == len(peers) {
		return fmt.Errorf("failed to bootstrap. %s", err)
	}

	return nil
}

func randomSubsetOfPeers(in []pstore.PeerInfo, max int) []pstore.PeerInfo {
	n := math2.IntMin(max, len(in))
	var out []pstore.PeerInfo
	for _, val := range rand.Perm(len(in)) {
		out = append(out, in[val])
		if len(out) >= n {
			break
		}
	}
	return out
}
