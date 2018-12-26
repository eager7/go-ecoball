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
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/net/address"
	"github.com/ecoball/go-ecoball/net/dispatcher"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/net/network"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p"
	"gx/ipfs/QmYAL9JsqVVPFWwM1ZzHNsofmTzRYQHJ2KqQaBmFJjJsNx/go-libp2p-connmgr"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	circuit "gx/ipfs/QmcQ56iqKP8ZRhRGLe5EReJVvrJZDaGzkuatrPv4Z1B6cG/go-libp2p-circuit"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	"os"
	"time"
)

var (
	log = elog.NewLogger("net", elog.DebugLog)
)

type Node struct {
	ctx         context.Context
	self        peer.ID
	network     *network.NetImpl
	listen      []string

	network.Receiver
}

func constructPeerHost(ctx context.Context, id peer.ID, private crypto.PrivKey) (host.Host, error) {
	addsFactory, err := address.MakeAddressesFactory(config.SwarmConfig)
	if err != nil {
		return nil, err
	}
	addsFactory = address.ComposeAddressesFactory(addsFactory, address.FilterRelayAddresses)

	var options []libp2p.Option
	options = append(options, libp2p.Identity(private))
	options = append(options, libp2p.AddrsFactory(addsFactory))
	if !config.SwarmConfig.DisableNatPortMap {
		options = append(options, libp2p.NATPortMap())
	}
	if !config.SwarmConfig.DisableRelay {
		var opts []circuit.RelayOpt
		if config.SwarmConfig.EnableRelayHop {
			opts = append(opts, circuit.OptHop)
		}
		options = append(options, libp2p.EnableRelay(opts...))
	}

	period := time.Duration(config.SwarmConfig.ConnGracePeriod) * time.Second
	grace, err := time.ParseDuration(period.String())
	if err != nil {
		return nil, errors.New(err.Error())
	}
	mgr := connmgr.NewConnManager(config.SwarmConfig.ConnLowWater, config.SwarmConfig.ConnHighWater, grace)
	options = append(options, libp2p.ConnectionManager(mgr))

	ps := peerstore.NewPeerstore()
	ps.AddPrivKey(id, private)
	ps.AddPubKey(id, private.GetPublic())
	options = append(options, libp2p.Peerstore(ps))
	return libp2p.New(ctx, options...)
}

func InitNetWork(ctx context.Context) *Node {
	node, err := newNetNode(ctx)
	if err != nil {
		log.Panic(err)
	}

	if err := node.Start(); err != nil {
		log.Error("error for starting net node,", err)
		os.Exit(1)
	}

	if err := NewNetActor(&netActor{network: node.network, ctx: node.ctx}); err != nil {
		log.Panic(err)
	}
	log.Info(fmt.Sprintf("peer(self) %s is running", node.SelfRawId().Pretty()))
	return node
}

func newNetNode(parent context.Context) (*Node, error) {
	private, err := address.GetNodePrivateKey()
	if err != nil {
		return nil, err
	}

	id, err := peer.IDFromPrivateKey(private)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error for generate id from key,%s", err.Error()))
	}
	netNode := &Node{
		ctx:         parent,
		self:        id,
		network:     nil,
		listen:      config.SwarmConfig.ListenAddress,
		Receiver:    nil,
	}

	h, err := constructPeerHost(parent, id, private) //basic_host.go
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error for constructing host, %s", err.Error()))
	}

	netNode.network = network.NewNetwork(parent, h, netNode)
	dispatcher.InitMsgDispatcher()

	return netNode, nil
}

func (nn *Node) Start() error {
	multiAddresses := make([]ma.Multiaddr, len(nn.listen))
	for idx, v := range nn.listen {
		addr, err := ma.NewMultiaddr(v)
		if err != nil {
			return err
		}
		multiAddresses[idx] = addr
	}

	h := nn.network.Host()
	if err := h.Network().Listen(multiAddresses...); err != nil {
		h.Close()
		return fmt.Errorf("error for listening,%s", err)
	}

	addresses, err := h.Network().InterfaceListenAddresses()
	if err != nil {
		return err
	}

	log.Info("net node listening on:", addresses)
	nn.network.Start()

	return nil
}

func (nn *Node) ReceiveMessage(ctx context.Context, p peer.ID, incoming message.EcoBallNetMsg) {
	log.Debug(fmt.Sprintf("receive msg %s from peer", incoming.Type().String()), nn.network.Host().Peerstore().Addrs(p))
	if incoming.Type() >= pb.MsgType_APP_MSG_UNDEFINED {
		log.Error("receive a invalid message ", incoming.Type().String())
		return
	}

	if err := dispatcher.Publish(incoming); err != nil {
		log.Error(err)
	}
}

func (nn *Node) ReceiveError(err error) {
	//TOD
}

func (nn *Node) GetShardMembersToReceiveCBlock() [][]*peerstore.PeerInfo {
	var peers = make([][]*peerstore.PeerInfo, 1)
	return peers
}

func (nn *Node) GetCMMembersToReceiveSBlock() []*peerstore.PeerInfo {
	var peers []*peerstore.PeerInfo
	return peers
}

func (nn *Node) PeerConnected(p peer.ID) {
	// TOD
}

func (nn *Node) PeerDisconnected(p peer.ID) {
	// TOD
}

func (nn *Node) SelfRawId() peer.ID {
	return nn.self
}

func (nn *Node) Neighbors() (peers []string) {
	h := nn.network.Host()
	cs := h.Network().Conns()
	for _, c := range cs {
		pid := c.RemotePeer()
		peers = append(peers, pid.Pretty())
	}
	return peers
}
