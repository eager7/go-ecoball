package net

import (
	"context"
	"fmt"
	"gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmSF8fPo3jgVBAy8fpdjjYqgG87dkJgUprRBHRd2tmfgpP/goprocess"
	ptx "gx/ipfs/QmSF8fPo3jgVBAy8fpdjjYqgG87dkJgUprRBHRd2tmfgpP/goprocess/context"
	"gx/ipfs/QmSF8fPo3jgVBAy8fpdjjYqgG87dkJgUprRBHRd2tmfgpP/goprocess/periodic"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmYyFh6g1C9uieTpH8CR8PpWBUQjvMDJTsRhJWx5qkXy39/go-ipfs-config"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
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

type BootStrap struct {
	closer  io.Closer
	bsPeers []config.BootstrapPeer
}

func (i *Instance) bootStrapInitialize(bsAddress []string) *BootStrap {
	bsPeers, err := config.ParseBootstrapPeers(bsAddress)
	if err != nil || len(bsPeers) == 0 {
		log.Error("failed to parse bootstrap address:", err)
		return nil
	}
	connected := i.Host.Network().Peers()
	if len(connected) > minPeerThreshold {
		log.Warn("this node was connected with network already, bootstrap skipped")
		return nil
	}
	numToDial := minPeerThreshold - len(connected)
	doneWithRound := make(chan struct{})
	periodic := func(worker goprocess.Process) {
		ctx := ptx.OnClosedContext(worker)
		if err := i.bootStrapConnect(ctx, bsPeers, numToDial); err != nil {
			log.Error(i.Host.ID().Pretty(), "bootstrap error:", err)
		}
		<-doneWithRound
	}
	process := periodicproc.Tick(bootStrapInterval, periodic)
	process.Go(periodic)
	doneWithRound <- struct{}{}
	close(doneWithRound)
	return &BootStrap{closer: process, bsPeers: bsPeers}
}

func (i *Instance) bootStrapConnect(ctx context.Context, bsPeers []config.BootstrapPeer, numToDial int) error {
	ctx, cancel := context.WithTimeout(ctx, bootStrapTimeOut)
	defer cancel()

	var notConnected []peerstore.PeerInfo
	for _, p := range bsPeers {
		if i.Host.Network().Connectedness(p.ID()) != net.Connected {
			protocols := len(p.Multiaddr().Protocols())
			sep := "/" + p.Multiaddr().Protocols()[protocols-1].Name
			addr, _ := multiaddr.NewMultiaddr(strings.Split(p.String(), sep)[0])
			peerInfo := peerstore.PeerInfo{ID: p.ID(), Addrs: []multiaddr.Multiaddr{addr}}
			notConnected = append(notConnected, peerInfo)
		}
	}
	if len(notConnected) < 1 {
		log.Warn("not enough bootstrap peers to connect")
		return nil
	}

	peers := randomPickPeers(notConnected, numToDial)
	var wg sync.WaitGroup
	for _, p := range peers {
		wg.Add(1)
		go func(p peerstore.PeerInfo) {
			defer wg.Done()
			log.Debug(fmt.Sprintf("%s bootstrapping to %s", i.Host.ID().Pretty(), p.ID.Pretty()))
			if err := i.Host.Connect(i.ctx, p); err != nil {
				log.Error("failed to bootstrap with:", p.ID.Pretty(), err)
				return
			}
			log.Info("bootstrapped successfully with:", p.ID.Pretty())
		}(p)
	}
	wg.Wait()
	return nil
}

func randomPickPeers(in []peerstore.PeerInfo, max int) (out []peerstore.PeerInfo) {
	n := func(x, y int) int {
		if x < y {
			return x
		}
		return y
	}(max, len(in))
	for _, val := range rand.Perm(len(in)) {
		out = append(out, in[val])
		if len(out) >= n {
			break
		}
	}
	return
}
