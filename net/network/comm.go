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
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/util"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

func (net *NetWork) ConnectToPeer(ip, port, pubKey string, isPermanent bool) error {
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
	log.Debug("connect peer finished:", ip, port, common.AddressFromPubKey([]byte(pubKey)).HexString(), isPermanent)

	return nil
}

func (net *NetWork) ClosePeer(pubKey string) error {
	id, err := IdFromConfigEncodePublicKey(pubKey)
	if err != nil {
		return err
	}

	var streams []inet.Stream
	for _, conn := range net.host.Network().ConnsToPeer(id) {
		streams = append(streams, conn.GetStreams()...)
	}

	for _, stream := range streams {
		stream.Close()
	}

	net.SenderMap.Del(id)
	log.Info("close the peer", id.String())
	return net.host.Network().ClosePeer(id)
}

func (net *NetWork) SendMsgToPeer(ip, port, pubKey string, msg message.EcoBallNetMsg) error {
	addrInfo := util.ConstructAddrInfo(ip, port)
	p, err := constructPeerInfo(addrInfo, pubKey)
	if err != nil {
		return err
	}
	sendJob := &SendMsgJob{Peers: []*peerstore.PeerInfo{&p}, Msg: msg}
	net.AddMsgJob(sendJob)

	return nil
}

func (net *NetWork) SendMsgToPeerWithPeerInfo(info []*peerstore.PeerInfo, msg message.EcoBallNetMsg) error {
	sendJob := &SendMsgJob{info, msg}
	net.AddMsgJob(sendJob)

	return nil
}

//sync send msg to peer by id
/*func (net *NetWork) SendMsgSyncToPeerWithId(id peer.ID, msg message.EcoBallNetMsg) error {
	p := peerstore.PeerInfo{ID: id}
	if err := net.sendMessage(p, msg); err != nil {
		log.Error("send message to ", p.ID.Pretty(), err)
	}

	return nil
}*/

//async send msg to peer by id
func (net *NetWork) SendMsgToPeerWithId(id peer.ID, msg message.EcoBallNetMsg) error {
	p := &peerstore.PeerInfo{ID: id}
	sendJob := &SendMsgJob{[]*peerstore.PeerInfo{p}, msg}
	net.AddMsgJob(sendJob)

	return nil
}

func (net *NetWork) SendMsgToPeersWithId(pid []peer.ID, msg message.EcoBallNetMsg) error {
	var peers []*peerstore.PeerInfo
	for _, id := range pid {
		peers = append(peers, &peerstore.PeerInfo{ID: id})
	}

	sendJob := &SendMsgJob{peers, msg}
	net.AddMsgJob(sendJob)

	return nil
}


func (net *NetWork) GetPeerStoreConnectStatus() []peer.ID {

	return net.host.Network().Peers()

}

func constructPeerInfo(addrInfo, pubKey string) (peerstore.PeerInfo, error) {
	pma, err := ma.NewMultiaddr(addrInfo)
	if err != nil {
		return peerstore.PeerInfo{}, err
	}

	id, err := IdFromConfigEncodePublicKey(pubKey)
	if err != nil {
		return peerstore.PeerInfo{}, err
	}

	p := peerstore.PeerInfo{ID: id, Addrs: []ma.Multiaddr{pma}}

	return p, nil
}

func IdFromConfigEncodePublicKey(pubKey string) (peer.ID, error) {
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

func IdFromProtoserPublicKey(pubKey []byte) (peer.ID, error) {
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
