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

package p2p

import (
	"context"
	"github.com/ecoball/go-ecoball/net/message"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/QmZ383TySJVeZWzGnWui6pRcKyYZk9VkKTuW7tmKRWk5au/go-libp2p-routing"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
)

type EcoballNetwork interface {
	Host() host.Host

	// SetDelegate registers the Reciver to handle messages received from the network.
	SetDelegate(Receiver)

	Start()
	Stop()

	CommAPI
	routing.PeerRouting
}

type CommAPI interface {
	//addrInfo example: ipv4:"/ip4/192.168.1.2/tcp/1234", ipv6:"/ip6/::/tcp/1234"
	ConnectToPeer(addrInfo string, pubKey []byte, isPermanent bool) error
	ClosePeer(pubKey []byte) error

	//Send a message to the peer with the ip/port/pubkey info
	//addrInfo example: ipv4:"/ip4/192.168.1.2/tcp/1234", ipv6:"/ip6/::/tcp/1234"
	SendMsgToPeer(addrInfo string, pubKey []byte, msg message.EcoBallNetMsg) error

	//Gossip a message to random peers
	GossipMsg(msg message.EcoBallNetMsg) error

	/*Send a message to a connected peer*/
	SendMsgToPeerWithId(peer.ID, message.EcoBallNetMsg) error
	/*Send a message to some connected peers*/
	SendMsgToPeersWithId([]peer.ID, message.EcoBallNetMsg) error

	/*Broadcast message to the connected peers*/
	BroadcastMessage(message.EcoBallNetMsg) error
}

// Implement Receiver to receive messages from the EcoBallNetwork
type Receiver interface {
	ReceiveMessage(
		ctx context.Context,
		sender peer.ID,
		incoming message.EcoBallNetMsg)

	ReceiveError(error)

	IsValidRemotePeer(peer.ID) bool

	// Connected/Disconnected warns net about peer connections
	PeerConnected(peer.ID)
	PeerDisconnected(peer.ID)
}