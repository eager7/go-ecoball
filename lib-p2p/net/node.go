package net

import (
	"context"
	"fmt"

	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/lib-p2p/address"
	"gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p"
	"gx/ipfs/QmYAL9JsqVVPFWwM1ZzHNsofmTzRYQHJ2KqQaBmFJjJsNx/go-libp2p-connmgr"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/io"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	"strings"
	"sync"
	"time"
)

var log = elog.Log

const (
	Protocol           = "/eager7/test/1.0.0"
	sendMessageTimeout = time.Minute * 10
)

type Instance struct {
	ctx       context.Context
	Host      host.Host
	ID        peer.ID
	Address   string
	Peers     PeerMap
	ShardInfo address.ShardInfo
	lock      sync.RWMutex
}

func NewInstance(ctx context.Context, b64Pri, address string) (*Instance, error) {
	i := new(Instance)
	i.ShardInfo.Initialize()
	if err := i.initialize(ctx, b64Pri, address); err != nil {
		return nil, err
	}
	return i, nil
}

func (i *Instance) initialize(ctx context.Context, b64Pri, address string) error {
	i.Peers.Initialize()
	if ctx == nil {
		ctx = context.Background()
	}
	i.ctx = ctx
	i.Address = address
	return i.initNetwork(b64Pri)
}

func (i *Instance) initNetwork(b64Pri string) (err error) {
	var private crypto.PrivKey
	var public crypto.PubKey
	if b64Pri == "" {
		private, public, err = crypto.GenerateKeyPair(crypto.RSA, 1024)
		if err != nil {
			return err
		}
		b, _ := private.Bytes()
		log.Info("generate private b64 key:", crypto.ConfigEncodeKey(b))
		b, _ = public.Bytes()
		log.Info("generate public b64 key:", crypto.ConfigEncodeKey(b))
	} else {
		key, err := crypto.ConfigDecodeKey(b64Pri)
		if err != nil {
			return err
		}
		private, err = crypto.UnmarshalPrivateKey(key)
		if err != nil {
			return err
		}
	}

	i.ID, err = peer.IDFromPrivateKey(private)
	if err != nil {
		return err
	}
	log.Info("this node id is :", i.ID.String())

	var options []libp2p.Option
	options = append(options, libp2p.Identity(private))

	period := time.Duration(20) * time.Second
	grace, err := time.ParseDuration(period.String())
	if err != nil {
		return errors.New(err.Error())
	}
	mgr := connmgr.NewConnManager(600, 900, grace)
	options = append(options, libp2p.ConnectionManager(mgr))

	ps := peerstore.NewPeerstore()
	if err := ps.AddPrivKey(i.ID, private); err != nil {
		return errors.New(err.Error())
	}
	if err := ps.AddPubKey(i.ID, private.GetPublic()); err != nil {
		return errors.New(err.Error())
	}
	options = append(options, libp2p.Peerstore(ps))

	i.Host, err = libp2p.New(i.ctx, options...)
	if err != nil {
		return err
	}

	i.Host.SetStreamHandler(Protocol, i.NetworkHandler)
	i.Host.Network().Notify(i)

	mAddr, err := multiaddr.NewMultiaddr(i.Address)
	if err != nil {
		return err
	}
	err = i.Host.Network().Listen([]multiaddr.Multiaddr{mAddr}...)
	if err != nil {
		return err
	}
	log.Debug(i.Host.Network().InterfaceListenAddresses())

	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", i.Host.ID().Pretty()))
	addresses := i.Host.Addrs()
	var addrM multiaddr.Multiaddr
	for _, i := range addresses {
		if strings.HasPrefix(i.String(), "/ip4") {
			addrM = i
			break
		}
	}
	fullAddr := addrM.Encapsulate(hostAddr)
	log.Debug("I am ", fullAddr, i.Host.Peerstore().Addrs(i.ID))
	return nil
}

//每个连接只会触发一次这个回调函数，之后需要在线程中做收发
func (i *Instance) NetworkHandler(s net.Stream) {
	log.Debug("receive connect peer from:", s.Conn().RemotePeer().Pretty(), s.Conn().RemoteMultiaddr(), "| topic is:", s.Protocol(), s)
	//i.Peers.Add(s.Conn().RemotePeer(), s, s.Conn().RemoteMultiaddr())

	go i.ReceiveMessage(s)
}

func (i *Instance) StreamConnect(b64Pub, addr, port string) (net.Stream, error) {
	id, err := address.IdFromPublicKey(b64Pub)
	if err != nil {
		return nil, err
	}
	if p := i.Peers.Get(id); p != nil {
		log.Warn("the stream is created:", p.s)
		return p.s, nil
	}

	multiAddr, err := multiaddr.NewMultiaddr(address.NewAddrInfo(addr, port))
	if err != nil {
		return nil, errors.New(err.Error())
	}
	//i.Host.Peerstore().AddAddr(id, addr, peerstore.PermanentAddrTTL)
	if len(i.Host.Peerstore().Addrs(id)) == 0 {
		i.Host.Peerstore().AddAddr(id, multiAddr, time.Minute*10)
	}
	log.Info("create new stream:", id.Pretty(), multiAddr, port)
	s, err := i.Host.NewStream(i.ctx, id, Protocol)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	log.Info("add stream:", s)
	i.Peers.Add(id, s, multiAddr)
	//go i.ReceiveMessage(s)
	return s, nil
}

/**
 *  @brief 建立和对端的连接，这个函数会调用Dial函数拨号，可以节省ConnectPeer函数执行时间，因为如果已经拨号成功，那么创建流时就不需要再次拨号，因此此函数可以作为ping函数使用，实时去刷新和节点间的连接
 *  @param b64Pub - the public key
 *  @param address - the address of ip
 *  @param port - the port of ip
 */
func (i *Instance) Connect(b64Pub, addr, port string) error {
	id, err := address.IdFromPublicKey(b64Pub)
	if err != nil {
		return err
	}
	multiAddr, err := multiaddr.NewMultiaddr(address.NewAddrInfo(addr, port))
	if err != nil {
		return errors.New(err.Error())
	}
	pi := peerstore.PeerInfo{ID: id, Addrs: []multiaddr.Multiaddr{multiAddr}}
	if err := i.Host.Connect(i.ctx, pi); err != nil {
		return errors.New(err.Error())
	}
	return nil
}

func (i *Instance) ReceiveMessage(s net.Stream) {
	log.Debug("start receive thread")
	reader := io.NewDelimitedReader(s, net.MessageSizeMax)
	for {
		msg := mpb.Message{}
		err := reader.ReadMsg(&msg)
		if err != nil {
			log.Error("the peer ", s.Conn().RemotePeer().Pretty(), i.Host.Peerstore().Addrs(s.Conn().RemotePeer()), "is disconnected:", err)
			s.Reset()
			i.Peers.Del(s.Conn().RemotePeer())
			return
		}
		log.Info("receive msg:", msg.String())
		if err := event.Publish(msg, msg.Identify); err != nil {
			return
		}
	}
}

func (i *Instance) SendMessage(b64Pub, addr, port string, message *mpb.Message) error {
	id, err := address.IdFromPublicKey(b64Pub)
	if err != nil {
		return errors.New(err.Error())
	}
	var s net.Stream
	info := i.Peers.Get(id)
	if info == nil {
		if s, err = i.StreamConnect(b64Pub, addr, port); err != nil {
			return err
		}
	} else {
		s = info.s
	}

	deadline := time.Now().Add(sendMessageTimeout)
	if dl, ok := i.ctx.Deadline(); ok {
		deadline = dl
		log.Info("set deal line:", deadline)
	}
	if err := s.SetWriteDeadline(deadline); err != nil {
		return errors.New(err.Error())
	}

	writer := io.NewDelimitedWriter(s)
	err = writer.WriteMsg(message)
	if err != nil {
		return errors.New(err.Error())
	}
	if err := s.SetWriteDeadline(time.Time{}); err != nil {
		log.Warn("error resetting deadline: ", err)
	}
	log.Info("send message finished:", message)
	return nil
}

func (i *Instance) ResetStream(s net.Stream) error {
	id := s.Conn().RemotePeer()
	if err := s.Reset(); err != nil {
		return err
	}
	i.Peers.Del(id)
	return nil
}

func (i *Instance) BroadcastToShard(shardId uint32, msg message.EcoMessage) error {
	peerMap := i.ShardInfo.GetShardNodes(shardId)
	if peerMap == nil {
		return errors.New(fmt.Sprintf("can't find shard[%d] nodes", shardId))
	}
	for node := range peerMap.Iterator() {
		data, err := msg.Serialize()
		if err != nil {
			return err
		}
		if err := i.SendMessage(node.Pubkey, node.Address, node.Port, &mpb.Message{Identify: msg.Identify(), Payload: data}); err != nil {
			log.Error(err)
		}
	}
	return nil
}
