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

// Implement the network notification APIs

package network

import (
	"context"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
)

func (net *NetWork) HandlePeerFound(p peerstore.PeerInfo) {
	if config.DisableLocalDisLog {
		log.SetLogLevel(elog.InfoLog)
	}
	log.Debug("trying peer info: ", p)
	ctx, cancel := context.WithTimeout(net.ctx, discoveryConnTimeout)
	defer cancel()
	if err := net.host.Connect(ctx, p); err != nil {
		log.Debug("Failed to connect to peer found by discovery: ", err)
	} else {
		log.Debug("connected to peer ", p)
	}
	log.SetLogLevel(elog.DebugLog)
}

func (net *NetWork) Listen(n inet.Network, a multiaddr.Multiaddr)      {}
func (net *NetWork) ListenClose(n inet.Network, a multiaddr.Multiaddr) {}
func (net *NetWork) Connected(n inet.Network, v inet.Conn) {
	log.Info("connected peer:", v.RemotePeer().Pretty(), v.RemoteMultiaddr().String())
	id := v.RemotePeer()
	if net.IsValidRemotePeer(id) {
		//net.PeerConnected(v.RemotePeer())
		if net.host.Network().Connectedness(id) == inet.Connected {
			net.routingTable.update(id)
		}
	} else {
		log.Warn("close invalid connection:", v.RemotePeer().Pretty(), v.RemoteMultiaddr().String())
		//v.Close() // invalid connection, close it...
	}
}
func (net *NetWork) Disconnected(n inet.Network, v inet.Conn) {
	log.Info("disconnected peer:", v.RemotePeer().Pretty(), v.RemoteMultiaddr().String())
	//net.PeerDisconnected(v.RemotePeer())
	id := v.RemotePeer()
	if net.host.Network().Connectedness(id) == inet.Connected {
		// We're still connected.
		return
	}
	net.routingTable.remove(id)
}
func (net *NetWork) OpenedStream(n inet.Network, v inet.Stream) {
	log.Info("OpenedStream peer:", v.Conn().RemotePeer().Pretty(), v.Conn().RemoteMultiaddr().String())
}
func (net *NetWork) ClosedStream(n inet.Network, v inet.Stream) {
	log.Info("ClosedStream peer:", v.Conn().RemotePeer().Pretty(), v.Conn().RemoteMultiaddr().String())
}
