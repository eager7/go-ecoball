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
	"github.com/ecoball/go-ecoball/net/message/pb"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
)

const defGossipPeerCount = 5

func (net *NetImpl)ConnectToPeer(addrInfo string, pubKey []byte, isPermanent bool) error {
	pi, err := net.constructPeerInfo(addrInfo, pubKey)
	if err != nil {
		return err
	}

	if isPermanent {
		net.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
	}

	if err := net.host.Connect(net.ctx, pi); err != nil {
		return err
	}

	return nil
}

func (net *NetImpl)ClosePeer(pubKey []byte) error {
	id, err := peer.IDFromBytes(pubKey)
	if err != nil {
		return err
	}

	conns := net.host.Network().ConnsToPeer(id)

	var streams []inet.Stream
	for _, conn := range conns {
		streams = append(streams, conn.GetStreams()...)
	}

	net.strmlk.Lock()
	defer  net.strmlk.Unlock()

	for _, stream := range streams {
		stream.Close()
	}

	delete(net.strmap, id)
	return net.host.Network().ClosePeer(id)
}

func (net *NetImpl)SendMsgToPeer(addrInfo string, pubKey []byte, msg message.EcoBallNetMsg) error {
	peer, err := net.constructPeerInfo(addrInfo, pubKey)
	if err != nil {
		return err
	}
/*
	if addr := net.host.Peerstore().Addrs(peer.ID); len(addr) == 0 {
		return fmt.Errorf("connection have not created for %s", peer.ID.Pretty())
	}
*/
	sendJob := &message.SendMsgJob{
		[]*peerstore.PeerInfo{&peer},
		msg,
	}
	net.SendMsgJob(sendJob)

	return nil
}

func (net *NetImpl)GossipMsg(msg message.EcoBallNetMsg) error {
	// wrap the message by the gossip msg type
	gossipMsg, err := net.warpMsgByGossip(msg)
	if err != nil {
		return err
	}
	return net.sendMsgToRandomPeers(defGossipPeerCount, gossipMsg)
}

func (net *NetImpl)ForwardMsg(msg message.EcoBallNetMsg, filtPeers[]peer.ID) error {
	peers := net.selectRandomPeers(defGossipPeerCount, filtPeers)
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

func (net *NetImpl) warpMsgByGossip(msg message.EcoBallNetMsg) (message.EcoBallNetMsg, error) {
	pbMsg := msg.ToProtoV1()
	wrapData, err := pbMsg.Marshal()
	if err != nil {
		return nil, err
	}
	// wrap the message by the gossip msg type
	gossipMsg := message.New(message.APP_MSG_GOSSIP, wrapData)
	return gossipMsg, nil
}

func (net *NetImpl)unwarpGossipMsg(msg message.EcoBallNetMsg) (message.EcoBallNetMsg, error) {
	if msg.Type() != message.APP_MSG_GOSSIP {
		return nil, fmt.Errorf("unwrap an invalid gossip message")
	}
	oriPbMsg := pb.Message{}
	err := oriPbMsg.Unmarshal(msg.Data())
	if err != nil {
		return nil, fmt.Errorf("error for unmarshal a gossip message data")
	} else {
		msg, _ := message.NewMessageFromProto(oriPbMsg)
		return msg, nil
	}
}

func (net *NetImpl)sendMsgToRandomPeers(peerCounts int, msg message.EcoBallNetMsg) error {
	peers := net.selectRandomPeers(peerCounts, []peer.ID{})
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

func (net *NetImpl)SendMsgToPeerWithId(id peer.ID, msg message.EcoBallNetMsg) error {
	peer := &peerstore.PeerInfo{ID:id}
	sendJob := &message.SendMsgJob{
		[]*peerstore.PeerInfo{peer},
		msg,
	}
	net.SendMsgJob(sendJob)

	return nil
}

func (net *NetImpl)SendMsgToPeersWithId(pid []peer.ID, msg message.EcoBallNetMsg) error {
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

func (net *NetImpl)constructPeerInfo(addrInfo string, pubKey []byte) (peerstore.PeerInfo, error) {
	pma, err := ma.NewMultiaddr(addrInfo)
	if err != nil {
		return peerstore.PeerInfo{}, err
	}
	id, err := peer.IDFromBytes(pubKey)
	if err != nil {
		return peerstore.PeerInfo{}, err
	}

	peer := peerstore.PeerInfo{id, []ma.Multiaddr{pma}}

	return peer, nil
}