package net

import (
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"time"
)

type RouteTable struct {
	host  host.Host
	route *kbucket.RoutingTable
}

func RouteInitialize(host host.Host) *RouteTable {
	id := kbucket.ConvertPeerID(host.ID())
	route := kbucket.NewRoutingTable(20, id, time.Minute, host.Peerstore())
	route.PeerAdded = func(id peer.ID) {
		//当更新一个节点信息时的回调
	}
	route.PeerRemoved = func(id peer.ID) {
		//当移除一个节点信息时的回调
	}
	return &RouteTable{host: host, route: route}
}

func (r *RouteTable) RouteUpdate(id peer.ID) {
	r.route.Update(id)
}

func (r *RouteTable) RouteRemove(id peer.ID) {
	r.route.Remove(id)
}

func (r *RouteTable) FindPeerStore(id peer.ID) peerstore.PeerInfo {
	switch r.host.Network().Connectedness(id) {
	case net.Connected, net.CanConnect:
		return r.host.Peerstore().PeerInfo(id)
	default:
		return peerstore.PeerInfo{}
	}
}

func (r *RouteTable) FindPeer(id peer.ID) (peerstore.PeerInfo, error) {
	if node := r.FindPeerStore(id); node.ID != "" {
		return node, nil
	}
	peers := r.route.NearestPeers(kbucket.ConvertPeerID(id), 3)
	for _, p := range peers {
		if p == id {
			return r.host.Peerstore().PeerInfo(p), nil
		}
	}
	return peerstore.PeerInfo{}, kbucket.ErrLookupFailure
}

func (r *RouteTable) FindNearestPeer(id peer.ID) peerstore.PeerInfo {
	if node := r.FindPeerStore(id); node.ID != "" {
		return node
	}
	return r.host.Peerstore().PeerInfo(r.route.Find(id))
}

