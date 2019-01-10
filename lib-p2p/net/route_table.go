package net

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/common/utils"
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

type RelayMessage struct {
	id  peer.ID
	msg *mpb.Message
}

func (r *RelayMessage) Identify() mpb.Identify {
	return mpb.Identify_APP_MSG_LIB_P2P_RELAY
}
func (r *RelayMessage) GetInstance() interface{} {
	return r
}
func (r *RelayMessage) Serialize() ([]byte, error) {
	//TODO
	return nil, nil
}
func (r *RelayMessage) Deserialize(data []byte) error {
	//TODO
	return nil
}
func (r *RelayMessage) String() string {
	return ""
}

func (i *Instance) SendRelayMessage(id peer.ID, msg *mpb.Message) (err error) {
	des := i.RouteTable.FindNearestPeer(id)
	desInfo := i.senderMap.Get(des.ID)
	if desInfo == nil {
		return errors.New("can't find nearest node")
	}
	if desInfo.Stream == nil && len(des.Addrs) != 0 {
		if desInfo.Stream, err = i.newStream(id, des.Addrs[0]); err != nil {
			return errors.New(fmt.Sprintf("new stream error:%s", err))
		}
	}
	relayMsg := &RelayMessage{id: id, msg: msg}
	data, err := relayMsg.Serialize()
	if err != nil {
		return err
	}
	sendMsg := &mpb.Message{Nonce: utils.RandomUint64(), Identify: relayMsg.Identify(), Payload: data}
	return i.transmit(desInfo.Stream, sendMsg)
}
