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
package network

import (
	"github.com/ecoball/go-ecoball/net/util"
	"github.com/ecoball/go-ecoball/net/message"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

func (net *NetImpl)ConnectToPeer(ip, port, pubKey string, isPermanent bool) error {
	addrInfo := util.ConstructAddrInfo(ip, port)
	pi, err := constructPeerInfo(addrInfo, pubKey)
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

func (net *NetImpl)ClosePeer(pubKey string) error {
	id, err := IdFromConfigEncodePublickKey(pubKey)
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

func (net *NetImpl)SendMsgToPeer(ip, port, pubKey string, msg message.EcoBallNetMsg) error {
	addrInfo := util.ConstructAddrInfo(ip, port)
	peer, err := constructPeerInfo(addrInfo, pubKey)
	if err != nil {
		return err
	}
/*
	if addr := net.host.Peerstore().Addrs(peer.ID); len(addr) == 0 {
		return fmt.Errorf("connection have not created for %s", peer.ID.Pretty())
	}
*/
	sendJob := &SendMsgJob{
		[]*peerstore.PeerInfo{&peer},
		msg,
	}
	net.AddMsgJob(sendJob)

	return nil
}

func (net *NetImpl)SendMsgToPeerWithId(id peer.ID, msg message.EcoBallNetMsg) error {
	peer := &peerstore.PeerInfo{ID:id}
	sendJob := &SendMsgJob{
		[]*peerstore.PeerInfo{peer},
		msg,
	}
	net.AddMsgJob(sendJob)

	return nil
}

func (net *NetImpl)SendMsgToPeersWithId(pid []peer.ID, msg message.EcoBallNetMsg) error {
	peers := []*peerstore.PeerInfo{}
	for _, id := range pid {
		peers = append(peers, &peerstore.PeerInfo{ID:id})
	}

	sendJob := &SendMsgJob{
		peers,
		msg,
	}
	net.AddMsgJob(sendJob)

	return nil
}

func (net *NetImpl)BroadcastMessage(msg message.EcoBallNetMsg) error {
	peers := []*peerstore.PeerInfo{}
	conns := net.host.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		if !net.receiver.IsNotMyShard(pid) {
			peers = append(peers, &peerstore.PeerInfo{ID:pid})
		}
	}

	if len(peers) > 0 {
		sendJob := &SendMsgJob{
			peers,
			msg,
		}
		net.AddMsgJob(sendJob)
	}

	return nil
}

func constructPeerInfo(addrInfo, pubKey string) (peerstore.PeerInfo, error) {
	pma, err := ma.NewMultiaddr(addrInfo)
	if err != nil {
		return peerstore.PeerInfo{}, err
	}

	id, err := IdFromConfigEncodePublickKey(pubKey)
	if err != nil {
		return peerstore.PeerInfo{}, err
	}

	peer := peerstore.PeerInfo{id, []ma.Multiaddr{pma}}

	return peer, nil
}

func IdFromConfigEncodePublickKey(pubKey string) (peer.ID, error) {
	key, err := ic.ConfigDecodeKey(pubKey)
	if err != nil {
		return "", err
	}
	pk, err := ic.UnmarshalPublicKey(key)
	if err != nil {
		return "", err
	}

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return "", err
	}

	return id, nil
}

func IdFromProtoserPublickKey(pubKey []byte) (peer.ID, error) {
	pk, err := ic.UnmarshalPublicKey(pubKey)
	if err != nil {
		return "", err
	}

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return "", err
	}

	return id, nil
}