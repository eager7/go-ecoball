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
	pmsg "github.com/ecoball/go-ecoball/net/message"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/QmZ383TySJVeZWzGnWui6pRcKyYZk9VkKTuW7tmKRWk5au/go-libp2p-routing"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	"gx/ipfs/QmXuucFcuvAWYAJfhHV2h4BYreHEAsLSsiquosiXeuduTN/go-libp2p-interface-connmgr"
)

type EcoballNetwork interface {
	Host() host.Host

	// SetDelegate registers the Reciver to handle messages received from the network.
	SetDelegate(Receiver)

	ConnectTo(context.Context, peer.ID) error

	// Start the local discovery and messgage pump
	Start()

	NewMessageSender(context.Context, peer.ID) (MessageSender, error)

	ConnectionManager() ifconnmgr.ConnManager

	routing.PeerRouting
}

type MessageSender interface {
	SendMsg(context.Context, pmsg.EcoBallNetMsg) error
	Close() error
	Reset() error
}

// Implement Receiver to receive messages from the EcoBallNetwork
type Receiver interface {
	ReceiveMessage(
		ctx context.Context,
		sender peer.ID,
		incoming pmsg.EcoBallNetMsg)

	ReceiveError(error)

	// Connected/Disconnected warns net about peer connections
	PeerConnected(peer.ID)
	PeerDisconnected(peer.ID)
}