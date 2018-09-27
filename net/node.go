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

package net

import (
	"context"
	"fmt"
	"os"
	"time"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/net/dispatcher"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/p2p"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/AsynkronIT/protoactor-go/actor"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmYAL9JsqVVPFWwM1ZzHNsofmTzRYQHJ2KqQaBmFJjJsNx/go-libp2p-connmgr"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p/p2p/host/basic"
	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	circuit "gx/ipfs/QmcQ56iqKP8ZRhRGLe5EReJVvrJZDaGzkuatrPv4Z1B6cG/go-libp2p-circuit"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	mafilter "gx/ipfs/QmSW4uNHbvQia8iZDXzbwjiyHQtnyo9aFqfQAMasj3TJ6Y/go-maddr-filter"
	mamask "gx/ipfs/QmSMZwvs3n4GBikZ7hKzT17c3bk65FmyZo2JqtJ16swqCv/multiaddr-filter"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
)

var (
	log = elog.NewLogger("net", elog.DebugLog)

	ecoballChainId uint32 = 1

	netNode  *NetNode
)

type NetNode struct {
	ctx         context.Context
	self        peer.ID
	network     p2p.EcoballNetwork
	broadCastCh chan message.EcoBallNetMsg
	handlers    map[uint32]message.HandlerFunc
	actorId     *actor.PID
	listen      []string
	//pubSub      *floodsub.PubSub
}

func constructPeerHost(ctx context.Context, id peer.ID, ps peerstore.Peerstore, options ...libp2p.Option) (host.Host, error) {
	pkey := ps.PrivKey(id)
	if pkey == nil {
		return nil, fmt.Errorf("missing private key for node ID: %s", id.Pretty())
	}
	options = append([]libp2p.Option{libp2p.Identity(pkey), libp2p.Peerstore(ps)}, options...)
	return libp2p.New(ctx, options...)
}

func makeAddrsFactory(cfg config.SwarmConfigInfo) (basichost.AddrsFactory, error) {
	var annAddrs []ma.Multiaddr
	for _, addr := range cfg.AnnounceAddr {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		annAddrs = append(annAddrs, maddr)
	}

	filters := mafilter.NewFilters()
	noAnnAddrs := map[string]bool{}
	for _, addr := range cfg.NoAnnounceAddr {
		f, err := mamask.NewMask(addr)
		if err == nil {
			filters.AddDialFilter(f)
			continue
		}
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		noAnnAddrs[maddr.String()] = true
	}

	return func(allAddrs []ma.Multiaddr) []ma.Multiaddr {
		var addrs []ma.Multiaddr
		if len(annAddrs) > 0 {
			addrs = annAddrs
		} else {
			addrs = allAddrs
		}

		var out []ma.Multiaddr
		for _, maddr := range addrs {
			// check for exact matches
			ok, _ := noAnnAddrs[maddr.String()]
			// check for /ipcidr matches
			if !ok && !filters.AddrBlocked(maddr) {
				out = append(out, maddr)
			}
		}
		return out
	}, nil
}

func filterRelayAddrs(addrs []ma.Multiaddr) []ma.Multiaddr {
	var raddrs []ma.Multiaddr
	for _, addr := range addrs {
		_, err := addr.ValueForProtocol(circuit.P_CIRCUIT)
		if err == nil {
			continue
		}
		raddrs = append(raddrs, addr)
	}
	return raddrs
}

func composeAddrsFactory(f, g basichost.AddrsFactory) basichost.AddrsFactory {
	return func(addrs []ma.Multiaddr) []ma.Multiaddr {
		return f(g(addrs))
	}
}

//func New(parent context.Context, privKey ic.PrivKey, listen []string) (*NetNode, error) {
func New(parent context.Context) (*NetNode, error) {
	var privKey ic.PrivKey
	dsnCfg, err := fsrepo.ConfigAt(config.IpfsDir)
	if err != nil {
		privKey, _, err = ic.GenerateKeyPair(ic.RSA, 2048)
		if err != nil {
			return nil, err
		}
	} else {
		privKey, err = dsnCfg.Identity.DecodePrivateKey("passphrase todo!")
		if err != nil {
			return nil, err
		}
	}

	id, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("error for getting id from key,", err)
	}

	netNode := &NetNode{
		ctx:         parent,
		self:        id,
		broadCastCh: make(chan message.EcoBallNetMsg, 4*1024), //TODO move to config
		handlers:    message.MakeHandlers(),
		//pubSub:      ipfs.Floodsub,
	}

	var libp2pOpts []libp2p.Option

	addrsFactory, err := makeAddrsFactory(config.SwarmConfig)
	if err != nil {
		return nil, err
	}
	if !config.SwarmConfig.DisableRelay {
		addrsFactory = composeAddrsFactory(addrsFactory, filterRelayAddrs)
	}
	libp2pOpts = append(libp2pOpts, libp2p.AddrsFactory(addrsFactory))

	if !config.SwarmConfig.DisableNatPortMap {
		libp2pOpts = append(libp2pOpts, libp2p.NATPortMap())
	}

	if !config.SwarmConfig.DisableRelay {
		var opts []circuit.RelayOpt
		if config.SwarmConfig.EnableRelayHop {
			opts = append(opts, circuit.OptHop)
		}
		libp2pOpts = append(libp2pOpts, libp2p.EnableRelay(opts...))
	}

	period :=time.Duration(config.SwarmConfig.ConnGracePeriod) * time.Second
	grace, err := time.ParseDuration(period.String())
	if err != nil {
		return nil, err
	}
	mgr := connmgr.NewConnManager(config.SwarmConfig.ConnLowWater, config.SwarmConfig.ConnHighWater, grace)
	libp2pOpts = append(libp2pOpts, libp2p.ConnectionManager(mgr))


	peerStore := peerstore.NewPeerstore()
	peerStore.AddPrivKey(id, privKey)
	peerStore.AddPubKey(id, privKey.GetPublic())
	h, err := constructPeerHost(parent, id, peerStore, libp2pOpts...)
	if err != nil {
		return nil, fmt.Errorf("error for constructing host,", err)
	}

	network := p2p.NewNetwork(parent, h)
	network.SetDelegate(netNode)

	netNode.network = network
	netNode.listen = config.SwarmConfig.ListenAddress

	dispatcher.InitMsgDispatcher()

	return netNode, nil
}

func (nn *NetNode) Start() error {
	multiaddrs := make([]ma.Multiaddr, len(nn.listen))
	for idx, v := range nn.listen {
		addr, err := ma.NewMultiaddr(v)
		if err != nil {
			return err
		}

		multiaddrs[idx] = addr
	}

	host := nn.network.Host()
	if err := host.Network().Listen(multiaddrs...); err != nil {
		host.Close()
		return fmt.Errorf("error for listening,",err)
	}

	addrs, err := host.Network().InterfaceListenAddresses()
	if err != nil {
		return err
	}

	log.Info("netnode listening on:", addrs)

	nn.network.Start()

	nn.broadcastLoop()

	return nil
}

func (nn *NetNode) SendBroadcastMsg(msg message.EcoBallNetMsg) {
	nn.broadCastCh <- msg
}

func (nn *NetNode) broadcastLoop() {
	go func() {
		for {
			select {
			case msg := <-nn.broadCastCh:
				log.Debug("broadCastCh receive msg:", message.MessageToStr[msg.Type()])
				//TODO cache check
				//node.netMsgCache.Add(msg.DataSum, msg.Size)
				nn.broadcastMessage(msg)
			}
		}
	}()
}

func (nn *NetNode) broadcastMessage(msg message.EcoBallNetMsg) {
	// In case of network sharding, should send the message to the shard internal peers
	peers := nn.connectedPeerIds()

	err := nn.network.SendMsg2PeersWithId(peers, msg)
	if err != nil {
		log.Error("failed to send msg to network,", err)
	}
}

func (nn *NetNode) connectedPeerIds() []peer.ID {
	peers := []peer.ID{}
	host := nn.network.Host()
	conns := host.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		peers = append(peers, pid)
	}
	return peers
}

func (nn *NetNode) ReceiveMessage(ctx context.Context, p peer.ID, incoming message.EcoBallNetMsg) {
	log.Debug("receive msg:", message.MessageToStr[incoming.Type()], "from ", p.Pretty())
	if incoming.Type() > message.APP_MSG_MAX {
		log.Error("receive a invalid message ", message.MessageToStr[incoming.Type()])
		return
	}

	handler, ok := nn.handlers[incoming.Type()]
	if ok {
		err := handler(incoming.Data())
		if err != nil {
			log.Error(err.Error())
			return
		}
		if err := dispatcher.Publish(incoming); err != nil {
			log.Error(err)
		}
	} else {
		dispatcher.Publish(incoming)
		log.Debug("publish msg ", incoming.Type())
		return
	}
}

func (nn *NetNode) ReceiveError(err error) {
	// TODO log the network error
	// TODO bubble the network error up to the parent context/error logger
}

func (nn *NetNode) PeerConnected(p peer.ID) {
	// TOD
}

func (nn *NetNode) PeerDisconnected(p peer.ID) {
	// TOD
}

func (nn *NetNode) SelfId() string {
	return nn.self.Pretty()
}

func (nn *NetNode) SelfRawId() peer.ID {
	return nn.self
}

func (nn *NetNode) Nbrs() []string {
	peers := []string{}

	host := nn.network.Host()
	conns := host.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		peers = append(peers, pid.Pretty())
	}
	return peers
}

func (nn *NetNode) SetActorPid(pid *actor.PID) {
	nn.actorId = pid
}

func (nn *NetNode) GetActorPid() *actor.PID {
	return nn.actorId
}

func SetChainId(id uint32) {
	ecoballChainId = id
}

func GetChainId() uint32 {
	return ecoballChainId
}

func InitNetWork(ctx context.Context) {
	var err error
	netNode, err = New(ctx)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.Info("i am ", netNode.SelfId())
}

func StartNetWork() {
	netActor := NewNetActor(netNode)
	actorId, _ := netActor.Start()
	netNode.SetActorPid(actorId)

	if err := netNode.Start(); err != nil {
		log.Error("error for starting netnode,", err)
		os.Exit(1)
	}
	log.Info(netNode.SelfId(), " is running.")
}
