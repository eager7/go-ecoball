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

// Implement the network communication APIs
package p2p

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/net/message"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
)

const (
	networkError = "network is not ready"
)

func (net *NetImpl)SendMsg2Peer(peer peerstore.PeerInfo, msg message.EcoBallNetMsg) error {
	if addr := net.host.Peerstore().Addrs(peer.ID); len(addr) == 0 {
		return fmt.Errorf("connection have not created for %s", peer.ID.Pretty())
	}

	sendJob := &message.SendMsgJob{
		[]*peerstore.PeerInfo{&peer},
		msg,
	}
	net.SendMsgJob(sendJob)

	return nil
}

func (net *NetImpl)SendMsgToPeers (peers []*peerstore.PeerInfo, msg message.EcoBallNetMsg) error {
	sendJob := &message.SendMsgJob{
		peers,
		msg,
	}
	net.SendMsgJob(sendJob)

	return nil
}

func (net *NetImpl)sendMsg2RandomPeers(peerCounts int, msg message.EcoBallNetMsg) error {
	peers := net.selectRandomPeers(peerCounts)
	if len(peers) == 0 {
		return errors.New(log,"failed to select random peers")
	}
	peerInfo := []*peerstore.PeerInfo{}
	for _, id := range peers {
		peerInfo = append(peerInfo, &peerstore.PeerInfo{ID:id})
	}
	sendJob := &message.SendMsgJob{
		peerInfo,
		msg,
	}
	net.SendMsgJob(sendJob)

	return nil
}

func (net *NetImpl)SendMsg2PeerWithId(id peer.ID, msg message.EcoBallNetMsg) error {
	peer := &peerstore.PeerInfo{ID:id}
	sendJob := &message.SendMsgJob{
		[]*peerstore.PeerInfo{peer},
		msg,
	}
	net.SendMsgJob(sendJob)

	return nil
}

func (net *NetImpl)SendMsg2PeersWithId (pid []peer.ID, msg message.EcoBallNetMsg) error {
	peers := []*peerstore.PeerInfo{}
	for _, id := range pid {
		peers = append(peers, &peerstore.PeerInfo{ID:id})
	}

	sendJob := &message.SendMsgJob{
		peers,
		msg,
	}
	net.SendMsgJob(sendJob)

	return nil
}

func (net *NetImpl)BroadcastMessage(msg message.EcoBallNetMsg) error {
	peers := []*peerstore.PeerInfo{}
	conns := net.host.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		peers = append(peers, &peerstore.PeerInfo{ID:pid})
	}

	sendJob := &message.SendMsgJob{
		peers,
		msg,
	}
	net.SendMsgJob(sendJob)

	return nil
}