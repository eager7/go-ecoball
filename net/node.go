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
	"os"
	"time"
	"github.com/urfave/cli"
	"github.com/AsynkronIT/protoactor-go/actor"
	ecoballConfig "github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/net/cmd"
	"github.com/ecoball/go-ecoball/net/dispatcher"
	"github.com/ecoball/go-ecoball/net/ipfs"
	"github.com/ecoball/go-ecoball/net/message"

	"github.com/ecoball/go-ecoball/net/p2p"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmYAL9JsqVVPFWwM1ZzHNsofmTzRYQHJ2KqQaBmFJjJsNx/go-libp2p-connmgr"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p"
	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

type NetCtrl struct {
	IpfsCtrl *ipfs.IpfsCtrl
	NetNode  *NetNode
	actor    *NetActor
}

const (
	DefaultConnMgrHighWater   = 900
	DefaultConnMgrLowWater    = 600
	DefaultConnMgrGracePeriod = time.Second * 20

	p2pListenP2PPort          = 4013
)

var log = elog.NewLogger("net", elog.DebugLog)

//TODO move to config
var ecoballChainId uint32 = 1

var netCtrl *NetCtrl

type NetNode struct {
	ctx         context.Context
	self        peer.ID
	network     p2p.EcoballNetwork
	broadCastCh chan message.EcoBallNetMsg
	handlers    map[uint32]message.HandlerFunc
	actorId     *actor.PID
	listen      []string
	//pubSub      *floodsub.PubSub

	//TODO cache check
	//netMsgCache  *lru.Cache
}

func constructPeerHost(ctx context.Context, id peer.ID, ps peerstore.Peerstore, options ...libp2p.Option) (host.Host, error) {
	pkey := ps.PrivKey(id)
	if pkey == nil {
		return nil, fmt.Errorf("missing private key for node ID: %s", id.Pretty())
	}
	options = append([]libp2p.Option{libp2p.Identity(pkey), libp2p.Peerstore(ps)}, options...)
	return libp2p.New(ctx, options...)
}

func New(parent context.Context, privKey ic.PrivKey, listen []string) (*NetNode, error) {
	id, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("error for getting id from key,", err)
	}

	netNode := &NetNode{
		ctx:         parent,
		self:        id,
		broadCastCh: make(chan message.EcoBallNetMsg, 4*1024), //TODO move to config
		handlers:    message.MakeHandlers(),
		//pubSub:      ipfs.Floodsub,
	}

	var libp2pOpts []libp2p.Option
	grace, err := time.ParseDuration(DefaultConnMgrGracePeriod.String())
	if err != nil {
		return nil, err
	}
	mgr := connmgr.NewConnManager(DefaultConnMgrLowWater, DefaultConnMgrHighWater, grace)
	libp2pOpts = append(libp2pOpts, libp2p.ConnectionManager(mgr))

	peerStore := peerstore.NewPeerstore()
	peerStore.AddPrivKey(id, privKey)
	peerStore.AddPubKey(id, privKey.GetPublic())
	h, err := constructPeerHost(parent, id, peerStore, libp2pOpts...)
	if err != nil {
		return nil, fmt.Errorf("error for constructing host,", err)
	}

	network := p2p.NewNetwork(parent, h)
	network.SetDelegate(netNode)

	netNode.network = network
	netNode.listen = listen

	dispatcher.InitMsgDispatcher()

	return netNode, nil
}

func (node *NetNode) Start() error {
	multiaddrs := make([]multiaddr.Multiaddr, len(node.listen))
	for idx, v := range node.listen {
		addr, err := multiaddr.NewMultiaddr(v)
		if err != nil {
			return err
		}

		multiaddrs[idx] = addr
	}

	host := node.network.Host()
	if err := host.Network().Listen(multiaddrs...); err != nil {
		host.Close()
		return fmt.Errorf("error for listening,",err)
	}

	addrs, err := host.Network().InterfaceListenAddresses()
	if err != nil {
		return err
	}

	log.Info("netnode listening on:", addrs)

	node.network.Start()

	node.broadcastLoop()

	return nil
}

func (node *NetNode) SendBroadcastMsg(msg message.EcoBallNetMsg) {
	node.broadCastCh <- msg
}

func (node *NetNode) broadcastLoop() {
	go func() {
		for {
			select {
			case msg := <-node.broadCastCh:
				log.Debug("broadCastCh receive msg:", message.MessageToStr[msg.Type()])
				//TODO cache check
				//node.netMsgCache.Add(msg.DataSum, msg.Size)
				node.broadcastMessage(msg)
			}
		}
	}()
}

func (node *NetNode) broadcastMessage(msg message.EcoBallNetMsg) {
	// In case of network sharding, should send the message to the shard internal peers
	peers := node.connectedPeerIds()

	err := p2p.SendMsg2PeersWithId(peers, msg)
	if err != nil {
		log.Error("failed to send msg to network,", err)
	}
}

func (node *NetNode) connectedPeerIds() []peer.ID {
	peers := []peer.ID{}
	host := node.network.Host()
	conns := host.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		peers = append(peers, pid)
	}
	return peers
}

func (bs *NetNode) ReceiveMessage(ctx context.Context, p peer.ID, incoming message.EcoBallNetMsg) {
	log.Debug("receive msg:", message.MessageToStr[incoming.Type()], "from ", p.Pretty())
	if incoming.Type() > message.APP_MSG_MAX {
		log.Error("receive a invalid message ", message.MessageToStr[incoming.Type()])
		return
	}

	handler, ok := bs.handlers[incoming.Type()]
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
		log.Debug("publish msg ", incoming.Type())
		return
	}
}

func (bs *NetNode) ReceiveError(err error) {
	// TODO log the network error
	// TODO bubble the network error up to the parent context/error logger
}

func (bs *NetNode) PeerConnected(p peer.ID) {
	// TOD
}

func (bs *NetNode) PeerDisconnected(p peer.ID) {
	// TOD
}

func (node *NetNode) SelfId() string {
	return node.self.Pretty()
}

func (node *NetNode) SelfRawId() peer.ID {
	return node.self
}

func (node *NetNode) Nbrs() []string {
	peers := []string{}

	host := node.network.Host()
	conns := host.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		peers = append(peers, pid.Pretty())
	}
	return peers
}

func (node *NetNode) SetActorPid(pid *actor.PID) {
	node.actorId = pid
}

func (node *NetNode) GetActorPid() *actor.PID {
	return node.actorId
}

func SetChainId(id uint32) {
	ecoballChainId = id
}

func GetChainId() uint32 {
	return ecoballChainId
}

func InitNetWork() {
	//TODO load config
	//configFile, err := ioutil.ReadFile(ConfigFile)
	//if err != nil {
	//
	//}
	//TODO move to config file
	//InitIpfsConfig(path)
	var path = ecoballConfig.IpfsDir

	ipfsCtrl, err := ipfs.InitAndRunIpfs(path)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	ipfsNode := ipfsCtrl.IpfsNode

	privkey := ipfsNode.PrivateKey
	ipv4ListenAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", p2pListenP2PPort)//should come from config file
	ipv6ListenAddr := fmt.Sprintf("/ip6/::/tcp/%d", p2pListenP2PPort)
	listenAddr := []string{ipv4ListenAddr, ipv6ListenAddr}

	netNode, err := New(context.Background(), privkey, listenAddr)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	netCtrl = &NetCtrl{
		IpfsCtrl: ipfsCtrl,
		NetNode:  netNode,
	}

	log.Info("i am ", netNode.self.Pretty())
}

func StartNetWork() {
	netActor := NewNetActor(netCtrl.NetNode)
	actorId, _ := netActor.Start()
	netCtrl.NetNode.SetActorPid(actorId)

	if err := netCtrl.NetNode.Start(); err != nil {
		log.Error("error for starting netnode,", err)
		os.Exit(1)
	}

	//start store repo stat engine
	netCtrl.IpfsCtrl.RepoStat.Start()

	log.Info(netCtrl.NetNode.SelfId(), " is running.")
}

//initialize
func Initialize(c *cli.Context) error {
	if fsrepo.IsInitialized(ecoballConfig.IpfsDir) {
		return nil
	}
	cmd.Root.Subcommands["init"] = initCmd
	os.Args[1] = "init"
	return cmd.StorageFun()
}

//start storage
func DaemonRun(c *cli.Context) error {
	cmd.Root.Subcommands["daemon"] = daemonCmd
	os.Args[1] = "daemon"
	return cmd.StorageFun()
}
