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

// Implement a simple gossip push function
package network

import (
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/net/util"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
)

const (
	GossipPeerCount = 5
)

type RoutingFilter func(id peer.ID) bool

var (
	NullFilter = func(peer.ID) bool {
		return false
	}
)

func CombineRoutingFilters(filters ...RoutingFilter) RoutingFilter {
	return func(id peer.ID) bool {
		for _, filter := range filters {
			if !filter(id) {
				return false
			}
		}
		return true
	}
}

func (net *NetImpl) GossipMsg(msg message.EcoBallNetMsg) error {
	// wrap the message by the gossip msg type
	gossipMsg, err := net.warpMsgByGossip(msg)
	if err != nil {
		return err
	}

	if net.gossipStore.Add(gossipMsg) {
		return net.sendMsgToRandomPeers(GossipPeerCount, gossipMsg)
	}

	return fmt.Errorf("duplicated msg in gossip store")
}

func (net *NetImpl) sendMsgToRandomPeers(peerCounts int, msg message.EcoBallNetMsg) (err error) {
	peers := net.getRandomPeers(peerCounts, net.receiver.IsNotMyShard)
	if len(peers) == 0 {
		err = errors.New("failed to select random peers")
		log.Error(err)
		return err
	}
	var peerInfo []*peerstore.PeerInfo
	for _, id := range peers {
		peerInfo = append(peerInfo, &peerstore.PeerInfo{ID: id})
	}
	sendJob := &SendMsgJob{
		peerInfo,
		msg,
	}
	net.AddMsgJob(sendJob)

	return nil
}

func (net *NetImpl) forwardMsg(msg message.EcoBallNetMsg, peers []peer.ID) {
	var peerInfo []*peerstore.PeerInfo
	for _, id := range peers {
		peerInfo = append(peerInfo, &peerstore.PeerInfo{ID: id})
	}
	if len(peerInfo) == 0 {
		return
	}
	sendJob := &SendMsgJob{
		peerInfo,
		msg,
	}
	net.AddMsgJob(sendJob)
}

func (net *NetImpl) warpMsgByGossip(msg message.EcoBallNetMsg) (message.EcoBallNetMsg, error) {
	pbMsg := msg.ToProtoV1()
	wrapData, err := pbMsg.Marshal()
	if err != nil {
		return nil, err
	}
	// wrap the message by the gossip msg type
	gossipMsg := message.New(pb.MsgType_APP_MSG_GOSSIP, wrapData)
	return gossipMsg, nil
}

func (net *NetImpl) unWarpGossipMsg(msg message.EcoBallNetMsg) (message.EcoBallNetMsg, error) {
	if msg.Type() != pb.MsgType_APP_MSG_GOSSIP {
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

func (net *NetImpl) getRandomPeers(k int, filter RoutingFilter) []peer.ID {
	var filtedConns []inet.Conn
	conns := net.host.Network().Conns()
	for _, conn := range conns {
		if !filter(conn.RemotePeer()) {
			filtedConns = append(filtedConns, conn)
		}
	}
	if len(filtedConns) < k {
		k = len(filtedConns)
	}
	indices := util.GetRandomIndices(k, len(filtedConns)-1)
	peers := make([]peer.ID, len(indices))
	for i, j := range indices {
		pid := filtedConns[j].RemotePeer()
		peers[i] = pid
	}

	return peers
}
