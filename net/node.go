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

package net

import (
	"context"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	inCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/net/dispatcher"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/net/network"
	"github.com/ecoball/go-ecoball/net/util"
	"github.com/ecoball/go-ecoball/sharding/common"
	mamask "gx/ipfs/QmSMZwvs3n4GBikZ7hKzT17c3bk65FmyZo2JqtJ16swqCv/multiaddr-filter"
	mafilter "gx/ipfs/QmSW4uNHbvQia8iZDXzbwjiyHQtnyo9aFqfQAMasj3TJ6Y/go-maddr-filter"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p/p2p/host/basic"
	"gx/ipfs/QmYAL9JsqVVPFWwM1ZzHNsofmTzRYQHJ2KqQaBmFJjJsNx/go-libp2p-connmgr"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	circuit "gx/ipfs/QmcQ56iqKP8ZRhRGLe5EReJVvrJZDaGzkuatrPv4Z1B6cG/go-libp2p-circuit"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"os"
	"sync"
	"time"
	"github.com/ecoball/go-ecoball/net/address"
)

const (
	OtherMember = iota
	CommitteeLeader
	CommitteeBackup
	ShardLeader
	ShardBackup
)


var (
	log = elog.NewLogger("net", elog.DebugLog)

	ecoballChainId uint32 = 1

	defaultNode *netNode
)

type ShardingInfo struct {
	shardId   uint16
	role      int
	peersInfo [][]peer.ID
	info      map[uint16]map[peer.ID]ma.Multiaddr // to accelerate the finding speed
	rwLock    sync.RWMutex
}

type netNode struct {
	ctx           context.Context
	self          peer.ID
	network       network.EcoballNetwork
	broadCastCh   chan message.EcoBallNetMsg
	handlers      map[pb.MsgType]message.HandlerFunc
	actorId       *actor.PID
	listen        []string
	shardingSubCh <-chan interface{}
	shardingInfo  *ShardingInfo

	network.Receiver
}

func constructPeerHost(ctx context.Context, id peer.ID, ps peerstore.Peerstore, options ...libp2p.Option) (host.Host, error) {
	key := ps.PrivKey(id)
	if key == nil {
		return nil, fmt.Errorf("missing private key for node ID: %s", id.Pretty())
	}
	options = append([]libp2p.Option{libp2p.Identity(key), libp2p.Peerstore(ps)}, options...)
	return libp2p.New(ctx, options...)
}

func makeAddressesFactory(cfg config.SwarmConfigInfo) (basichost.AddrsFactory, error) {
	var annAdds []ma.Multiaddr
	for _, addr := range cfg.AnnounceAddr {
		mAddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		annAdds = append(annAdds, mAddr)
	}

	filters := mafilter.NewFilters()
	noAnnAdds := map[string]bool{}
	for _, addr := range cfg.NoAnnounceAddr {
		f, err := mamask.NewMask(addr)
		if err == nil {
			filters.AddDialFilter(f)
			continue
		}
		mAddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		noAnnAdds[mAddr.String()] = true
	}

	return func(allAddrs []ma.Multiaddr) []ma.Multiaddr {
		var adds []ma.Multiaddr
		if len(annAdds) > 0 {
			adds = annAdds
		} else {
			adds = allAddrs
		}

		var out []ma.Multiaddr
		for _, mAddr := range adds {
			// check for exact matches
			ok, _ := noAnnAdds[mAddr.String()]
			// check for /ipcidr matches
			if !ok && !filters.AddrBlocked(mAddr) {
				out = append(out, mAddr)
			}
		}
		return out
	}, nil
}

func filterRelayAddrs(addrs []ma.Multiaddr) []ma.Multiaddr {
	var rAdds []ma.Multiaddr
	for _, addr := range addrs {
		_, err := addr.ValueForProtocol(circuit.P_CIRCUIT)
		if err == nil {
			continue
		}
		rAdds = append(rAdds, addr)
	}
	return rAdds
}



func composeAddrsFactory(f, g basichost.AddrsFactory) basichost.AddrsFactory {
	return func(addrs []ma.Multiaddr) []ma.Multiaddr {
		return f(g(addrs))
	}
}

func NewNetNode(parent context.Context) (*netNode, error) {
	private, err := address.GetNodePrivateKey()
	if err != nil {
		return nil, err
	}

	id, err := peer.IDFromPrivateKey(private)
	if err != nil {
		return nil, fmt.Errorf("error for generate id from key,%s", err.Error())
	}
	netNode := &netNode{
		ctx:           parent,
		self:          id,
		broadCastCh:   make(chan message.EcoBallNetMsg, 4*1024), //TODO move to config
		handlers:      message.MakeHandlers(),
		shardingInfo:  new(ShardingInfo),
		shardingSubCh: make(<-chan interface{}, 1),
	}

	netNode.shardingInfo.peersInfo = make([][]peer.ID, 0)
	netNode.shardingInfo.info = make(map[uint16]map[peer.ID]ma.Multiaddr)

	var libP2pOpts []libp2p.Option

	addsFactory, err := makeAddressesFactory(config.SwarmConfig)
	if err != nil {
		return nil, err
	}
	if !config.SwarmConfig.DisableRelay {
		addsFactory = composeAddrsFactory(addsFactory, filterRelayAddrs)
	}
	libP2pOpts = append(libP2pOpts, libp2p.AddrsFactory(addsFactory))

	if !config.SwarmConfig.DisableNatPortMap {
		libP2pOpts = append(libP2pOpts, libp2p.NATPortMap())
	}

	if !config.SwarmConfig.DisableRelay {
		var opts []circuit.RelayOpt
		if config.SwarmConfig.EnableRelayHop {
			opts = append(opts, circuit.OptHop)
		}
		libP2pOpts = append(libP2pOpts, libp2p.EnableRelay(opts...))
	}

	period := time.Duration(config.SwarmConfig.ConnGracePeriod) * time.Second
	grace, err := time.ParseDuration(period.String())
	if err != nil {
		return nil, err
	}
	mgr := connmgr.NewConnManager(config.SwarmConfig.ConnLowWater, config.SwarmConfig.ConnHighWater, grace)
	libP2pOpts = append(libP2pOpts, libp2p.ConnectionManager(mgr))

	peerStore := peerstore.NewPeerstore()
	peerStore.AddPrivKey(id, private)
	peerStore.AddPubKey(id, private.GetPublic())
	h, err := constructPeerHost(parent, id, peerStore, libP2pOpts...) //basic_host.go
	if err != nil {
		return nil, fmt.Errorf("error for constructing host, %s", err.Error())
	}

	n := network.NewNetwork(parent, h)
	n.SetDelegate(netNode)

	netNode.network = n
	netNode.listen = config.SwarmConfig.ListenAddress

	dispatcher.InitMsgDispatcher()

	return netNode, nil
}

func (nn *netNode) Start() error {
	multiAddresses := make([]ma.Multiaddr, len(nn.listen))
	for idx, v := range nn.listen {
		addr, err := ma.NewMultiaddr(v)
		if err != nil {
			return err
		}

		multiAddresses[idx] = addr
	}

	h := nn.network.Host()
	if err := h.Network().Listen(multiAddresses...); err != nil {
		h.Close()
		return fmt.Errorf("error for listening,%s", err)
	}

	addresses, err := h.Network().InterfaceListenAddresses()
	if err != nil {
		return err
	}

	log.Info("net node listening on:", addresses)
	nn.network.Start()
	nn.nativeMessageLoop()

	return nil
}

//连接本shard内的节点
func (nn *netNode) connectToShardingPeers() {
	nn.shardingInfo.rwLock.RLock()
	defer nn.shardingInfo.rwLock.RUnlock()
	works := nn.shardingInfo.info[nn.shardingInfo.shardId]
	h := nn.network.Host()
	var wg sync.WaitGroup
	for id, w := range works {
		if id == h.ID() {
			continue
		}
		wg.Add(1)
		go func(p peer.ID, addr ma.Multiaddr) {
			log.Info("start host connect thread:", p, addr)
			defer wg.Done()
			h.Peerstore().AddAddrs(p, []ma.Multiaddr{addr}, peerstore.PermanentAddrTTL)
			pi := peerstore.PeerInfo{ID: p, Addrs: []ma.Multiaddr{addr}}
			if err := h.Connect(nn.ctx, pi); err != nil {
				log.Error("failed to connect peer ", pi, err)
			} else {
				log.Debug("succeed to connect peer ", pi)
			}
		}(id, w)
	}
	wg.Wait()
	log.Debug("finish connecting to sharding peers exit...")
}

func (nn *netNode) updateShardingInfo(info *common.ShardingTopo) {
	log.Info(inCommon.JsonString(info))
	nn.shardingInfo.rwLock.Lock()
	nn.shardingInfo.shardId = info.ShardId
	for sid, shard := range info.ShardingInfo {
		for i, member := range shard {
			log.Info(sid, i)
			if sid == 0 && i == 0 {
				nn.shardingInfo.role = CommitteeLeader
			} else if sid == 0 && i == 1 {
				nn.shardingInfo.role = CommitteeBackup
			} else if sid > 0 && i == 0 {
				nn.shardingInfo.role = ShardLeader
			} else if sid > 0 && i == 1 {
				nn.shardingInfo.role = ShardBackup
			}

			id, err := network.IdFromConfigEncodePublicKey(member.Pubkey)
			if err != nil {
				log.Error("error for getting id from public key")
				continue
			}

			//var addr ma.Multiaddr
			addInfo := util.ConstructAddrInfo(member.Address, member.Port)
			addr, err := ma.NewMultiaddr(addInfo)
			if err != nil {
				log.Error("error for create ip addr from member info", err)
				continue
			}

			if _, ok := nn.shardingInfo.info[uint16(sid)]; !ok {
				idAddr := make(map[peer.ID]ma.Multiaddr)
				idAddr[id] = addr
				nn.shardingInfo.info[uint16(sid)] = idAddr

				nn.shardingInfo.peersInfo = append(nn.shardingInfo.peersInfo, []peer.ID{})
				nn.shardingInfo.peersInfo[sid] = append(nn.shardingInfo.peersInfo[sid], id)
			} else {
				nn.shardingInfo.info[uint16(sid)][id] = addr
				nn.shardingInfo.peersInfo[sid] = append(nn.shardingInfo.peersInfo[sid], id)
			}
		}
	}
	nn.shardingInfo.rwLock.Unlock()
	nn.connectToShardingPeers()
}

func (nn *netNode) GetShardAddress(id peer.ID) peerstore.PeerInfo {
	nn.shardingInfo.rwLock.RLock()
	defer nn.shardingInfo.rwLock.RUnlock()
	for _, addr := range nn.shardingInfo.info {
		for i, m := range addr {
			if i.String() == id.String() {
				return peerstore.PeerInfo{ID: i, Addrs: []ma.Multiaddr{m}}
			}
		}
	}
	return peerstore.PeerInfo{}
}

func (nn *netNode) nativeMessageLoop() {
	go func() {
		for {
			select {
			case info := <-nn.shardingSubCh:
				sInfo, ok := info.(*common.ShardingTopo)
				if !ok {
					log.Error("unsupported info from sharding.")
					continue
				}
				log.Debug("receive a update sharding message, my shard:", sInfo.ShardId)
				go nn.updateShardingInfo(sInfo)
			case msg := <-nn.broadCastCh:
				log.Debug("broadCastCh receive msg:", msg.Type().String())
				nn.network.BroadcastMessage(msg)
			}
		}
	}()
}

func (nn *netNode) ReceiveMessage(ctx context.Context, p peer.ID, incoming message.EcoBallNetMsg) {
	log.Debug(fmt.Sprintf("receive msg %s from peer", incoming.Type().String()), nn.GetShardAddress(p))
	if incoming.Type() >= pb.MsgType_APP_MSG_UNDEFINED {
		log.Error("receive a invalid message ", incoming.Type().String())
		return
	}

	handler, ok := nn.handlers[incoming.Type()] //go-ecoball/net/message/handler.go:MakeHandlers()
	if ok {
		err := handler(incoming.Data())
		if err != nil {
			log.Error(err.Error())
			return
		}
		if err := dispatcher.Publish(incoming); err != nil {
			log.Error(err)
		}
	} else {
		dispatcher.Publish(incoming)
		return
	}
}

func (nn *netNode) ReceiveError(err error) {
	//TOD
}

func (nn *netNode) IsValidRemotePeer(p peer.ID) bool {
	if !config.DisableSharding {
		nn.shardingInfo.rwLock.RLock()
		defer nn.shardingInfo.rwLock.RUnlock()

		for _, shard := range nn.shardingInfo.info {
			if _, ok := shard[p]; ok {
				return true
			}
		}

		log.Error(nn.shardingInfo.shardId, p, nn.network.Host().Peerstore().Addrs(p))
		return false
	}

	return true
}

func (nn *netNode) IsNotMyShard(p peer.ID) bool {
	if !config.DisableSharding {
		nn.shardingInfo.rwLock.RLock()
		defer nn.shardingInfo.rwLock.RUnlock()

		if _, ok := nn.shardingInfo.info[nn.shardingInfo.shardId][p]; !ok {
			return true
		}
	}

	return false
}

func (nn *netNode) IsLeaderOrBackup() bool {
	nn.shardingInfo.rwLock.RLock()
	defer nn.shardingInfo.rwLock.RUnlock()

	if nn.shardingInfo.role != OtherMember {
		return true
	}

	return false
}

func (nn *netNode) GetShardLeader(shardId uint16) (*peerstore.PeerInfo, error) {
	nn.shardingInfo.rwLock.RLock()
	defer nn.shardingInfo.rwLock.RUnlock()

	if int(shardId) > (len(nn.shardingInfo.peersInfo) - 1) {
		return nil, fmt.Errorf("invalid shard id:%d(shard len:%d)", shardId, len(nn.shardingInfo.peersInfo))
	}

	id := nn.shardingInfo.peersInfo[shardId][0]
	pi := &peerstore.PeerInfo{ID: id, Addrs: []ma.Multiaddr{nn.shardingInfo.info[shardId][id]}}
	return pi, nil
}

func (nn *netNode) GetShardMembersToReceiveCBlock() [][]*peerstore.PeerInfo {
	nn.shardingInfo.rwLock.RLock()
	defer nn.shardingInfo.rwLock.RUnlock()

	if len(nn.shardingInfo.peersInfo) <= 1 {
		return [][]*peerstore.PeerInfo{}
	}

	var peers = make([][]*peerstore.PeerInfo, len(nn.shardingInfo.peersInfo)-1)
	// the algo may be changed according to the requirement
	for id, shard := range nn.shardingInfo.peersInfo[1:] {
		if nn.shardingInfo.role == CommitteeLeader {
			addrInfo := []ma.Multiaddr{nn.shardingInfo.info[uint16(id)][shard[0]]}
			pi := &peerstore.PeerInfo{ID: shard[0], Addrs: addrInfo}
			peers[id] = append(peers[id], pi)
		} else if nn.shardingInfo.role == CommitteeBackup {
			if len(shard) > 1 {
				addrInfo := []ma.Multiaddr{nn.shardingInfo.info[uint16(id)][shard[1]]}
				pi := &peerstore.PeerInfo{ID: shard[0], Addrs: addrInfo}
				peers[id] = append(peers[id], pi)
			}
		}
	}
	return peers
}

func (nn *netNode) GetCMMembersToReceiveSBlock() []*peerstore.PeerInfo {
	nn.shardingInfo.rwLock.RLock()
	defer nn.shardingInfo.rwLock.RUnlock()

	var peers []*peerstore.PeerInfo
	if nn.shardingInfo.role == ShardLeader {
		id := nn.shardingInfo.peersInfo[0][0]
		addrInfo := []ma.Multiaddr{nn.shardingInfo.info[0][id]}
		pi := &peerstore.PeerInfo{id, addrInfo}
		peers = append(peers, pi)
	} else if nn.shardingInfo.role == ShardBackup {
		if len(nn.shardingInfo.peersInfo[0]) > 1 {
			id := nn.shardingInfo.peersInfo[0][1]
			addrInfo := []ma.Multiaddr{nn.shardingInfo.info[0][id]}
			pi := &peerstore.PeerInfo{ID: id, Addrs: addrInfo}
			peers = append(peers, pi)
		}
	}

	return peers
}

func (nn *netNode) PeerConnected(p peer.ID) {
	// TOD
}

func (nn *netNode) PeerDisconnected(p peer.ID) {
	// TOD
}

func (nn *netNode) SelfId() string {
	return nn.self.Pretty()
}

func (nn *netNode) SelfRawId() peer.ID {
	return nn.self
}

func (nn *netNode) Neighbors() []string {
	var peers []string

	host := nn.network.Host()
	conns := host.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		peers = append(peers, pid.Pretty())
	}
	return peers
}

func (nn *netNode) SetActorPid(pid *actor.PID) {
	nn.actorId = pid
}

func (nn *netNode) GetActorPid() *actor.PID {
	return nn.actorId
}

func (nn *netNode) SetShardingSubCh(ch <-chan interface{}) {
	nn.shardingSubCh = ch
}

func SetChainId(id uint32) {
	ecoballChainId = id
}

func GetChainId() uint32 {
	return ecoballChainId
}

func InitNetWork(ctx context.Context) {
	var err error
	defaultNode, err = NewNetNode(ctx)
	if err != nil {
		log.Panic(err)
	}
}

func StartNetWork(cShard <-chan interface{}) {
	netActor := NewNetActor(defaultNode)
	actorId, _ := netActor.Start()
	defaultNode.SetActorPid(actorId)

	if cShard != nil {
		defaultNode.SetShardingSubCh(cShard)
	}

	if err := defaultNode.Start(); err != nil {
		log.Error("error for starting net node,", err)
		os.Exit(1)
	}

	log.Info(fmt.Sprintf("peer(self) %s is running", defaultNode.SelfId()))
}
