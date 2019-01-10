package net

import (
	"context"
	"fmt"
	"github.com/ecoball/go-ecoball/lib-p2p/go-ipfs-addr"
	"github.com/jbenet/goprocess"
	ptx "github.com/jbenet/goprocess/context"
	"github.com/jbenet/goprocess/periodic"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
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
	bsPeers []ipfsaddr.IPFSAddr
}

func (i *Instance) bootStrapInitialize(bsAddress []string) *BootStrap {
	var bsPeers []ipfsaddr.IPFSAddr
	for _, addr := range bsAddress {
		if bsPeer, err := ipfsaddr.ParseString(addr); err != nil {
			log.Error("failed to parse bootstrap address:", addr, err)
			return nil
		} else {
			bsPeers = append(bsPeers, bsPeer)
		}
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

func (i *Instance) bootStrapConnect(ctx context.Context, bsPeers []ipfsaddr.IPFSAddr, numToDial int) error {
	ctx, cancel := context.WithTimeout(ctx, bootStrapTimeOut)
	defer cancel()

	log.Debug("bootstrap connect:", bsPeers)
	var notConnected []peerstore.PeerInfo
	for _, p := range bsPeers {
		if p.ID() == i.ID {
			log.Debug("skip self address")
			continue
		}
		if i.Host.Network().Connectedness(p.ID()) != net.Connected {
			protocols := len(p.Multiaddr().Protocols())
			sep := "/" + p.Multiaddr().Protocols()[protocols-1].Name
			addr, _ := multiaddr.NewMultiaddr(strings.Split(p.String(), sep)[0])
			peerInfo := peerstore.PeerInfo{ID: p.ID(), Addrs: []multiaddr.Multiaddr{addr}}
			notConnected = append(notConnected, peerInfo)
		}
	}
	if len(notConnected) < 1 {
		log.Warn("the bootstrap peers were connected already")
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
				log.Error("failed to bootstrap with:", p.ID.Pretty(), p.Addrs, err)
				i.senderMap.Del(p.ID)
				return
			}
			log.Info("bootstrapped successfully with:", p.ID.Pretty())
			i.senderMap.Add(p.ID, nil, p.Addrs[0])
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
