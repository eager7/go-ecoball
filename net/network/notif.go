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
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
)

type netNotifiee NetImpl

func (nn *netNotifiee) impl() *NetImpl {
	return (*NetImpl)(nn)
}

func (nn *netNotifiee) Connected(n inet.Network, v inet.Conn) {
	p := v.RemotePeer()
	if nn.impl().receiver.IsValidRemotePeer(p) {
		nn.impl().receiver.PeerConnected(v.RemotePeer())

		if nn.impl().host.Network().Connectedness(p) == inet.Connected {
			nn.impl().routingTable.update(p)
		}
	} else {
		// invalid connection, close it...
		v.Close()
	}
}

func (nn *netNotifiee) Disconnected(n inet.Network, v inet.Conn) {
	nn.impl().receiver.PeerDisconnected(v.RemotePeer())

	p := v.RemotePeer()
	if nn.impl().host.Network().Connectedness(p) == inet.Connected {
		// We're still connected.
		return
	}
	nn.impl().routingTable.remove(p)
}

func (nn *netNotifiee) HandlePeerFound(p pstore.PeerInfo) {
	if config.DisableLocalDisLog {
		log.SetLogLevel(elog.InfoLog)
	}
	log.Debug("trying peer info: ", p)
	ctx, cancel := context.WithTimeout(nn.ctx, discoveryConnTimeout)
	defer cancel()
	if err := nn.host.Connect(ctx, p); err != nil {
		log.Debug("Failed to connect to peer found by discovery: ", err)
	} else {
		log.Debug("connected to peer ", p)
	}
	log.SetLogLevel(elog.DebugLog)
}

func (nn *netNotifiee) OpenedStream(n inet.Network, v inet.Stream) {}
func (nn *netNotifiee) ClosedStream(n inet.Network, v inet.Stream) {}
func (nn *netNotifiee) Listen(n inet.Network, a ma.Multiaddr)      {}
func (nn *netNotifiee) ListenClose(n inet.Network, a ma.Multiaddr) {}
