package net

import (
	"context"
	"fmt"

	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/lib-p2p/address"
	"gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p"
	"gx/ipfs/QmYAL9JsqVVPFWwM1ZzHNsofmTzRYQHJ2KqQaBmFJjJsNx/go-libp2p-connmgr"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	"gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/io"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
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
	ctx          context.Context
	Host         host.Host
	ID           peer.ID
	Address      []string
	BootStrapper *BootStrap
	senderMap    address.SenderMap
	lock         sync.RWMutex
}

func NewInstance(ctx context.Context, b64Pri string, address ...string) (*Instance, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	i := &Instance{ctx: ctx}
	i.senderMap.Initialize()
	i.Address = append(i.Address, address...)
	if err := i.initNetwork(b64Pri); err != nil {
		return nil, err
	}
	return i, nil
}

func (i *Instance) initNetwork(b64Pri string) (err error) {
	private, err := address.GetNodePrivateKey(b64Pri)
	if err != nil {
		return err
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

	mAddresses := make([]multiaddr.Multiaddr, len(i.Address))
	for idx, v := range i.Address {
		addr, err := multiaddr.NewMultiaddr(v)
		if err != nil {
			return err
		}
		mAddresses[idx] = addr
	}

	err = i.Host.Network().Listen(mAddresses...)
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
	log.Debug("start bootstrap nodes...")
	i.BootStrapper = i.bootStrapInitialize(config.SwarmConfig.BootStrapAddr)
	return nil
}

func (i *Instance) StopNetwork() {
	if i.BootStrapper != nil {
		i.BootStrapper.closer.Close()
	}
	i.Host.Network().StopNotify(i)
}

//多次发送只会触发一次这个回调函数，之后需要在线程中做收发
func (i *Instance) NetworkHandler(s net.Stream) {
	log.Debug("receive connect peer from:", s.Conn().RemotePeer().Pretty(), s.Conn().RemoteMultiaddr(), "| topic is:", s.Protocol(), s)
	go i.receive(s)
}

func (i *Instance) SendMessage(b64Pub, addr, port string, msg types.EcoMessage) error {
	data, err := msg.Serialize()
	if err != nil {
		return err
	}
	sendMsg := &mpb.Message{Identify: msg.Identify(), Payload: data}

	id, err := address.IdFromPublicKey(b64Pub)
	if err != nil {
		return errors.New(err.Error())
	}
	var s net.Stream
	info := i.senderMap.Get(id)
	if info == nil || info.Stream == nil {
		multiAddr, err := multiaddr.NewMultiaddr(address.NewAddrInfo(addr, port))
		if err != nil {
			return errors.New(err.Error())
		}
		id, err := address.IdFromPublicKey(b64Pub)
		if err != nil {
			return err
		}
		if s, err = i.newStream(id, multiAddr); err != nil {
			return err
		}
	} else {
		s = info.Stream
	}
	return i.transmit(s, sendMsg)
}

func (i *Instance) BroadcastToNeighbors(msg types.EcoMessage) error {
	data, err := msg.Serialize()
	if err != nil {
		return err
	}
	sendMsg := &mpb.Message{Identify: msg.Identify(), Payload: data}
	for _, c := range i.Host.Network().Conns() {
		id := c.RemotePeer()
		var s net.Stream
		info := i.senderMap.Get(id)
		if info == nil {
			log.Error(fmt.Sprintf("the node is not connected:%s", id.Pretty()))
		} else {
			if info.Stream != nil {
				s = info.Stream
			} else if len(info.PeerInfo.Addrs) > 0 {
				if s, err = i.newStream(info.ID, info.PeerInfo.Addrs[0]); err != nil {
					log.Error("new stream error:", err)
				}
			}
		}
		if err := i.transmit(s, sendMsg); err != nil {
			log.Error("transmit err:", err)
		}
	}
	return nil
}

/**
 *  @brief 建立和对端的连接，这个函数会调用Dial函数拨号，可以节省ConnectPeer函数执行时间，因为如果已经拨号成功，
 *  那么创建流时就不需要再次拨号，因此此函数可以作为ping函数使用，实时去刷新和节点间的连接
 *  @param b64Pub - the public key
 *  @param address - the address of ip
 *  @param port - the port of ip
 */
func (i *Instance) connect(b64Pub, addr, port string) error {
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
	i.senderMap.Add(id, nil, multiAddr)
	return nil
}

func (i *Instance) newStream(id peer.ID, multiAddr multiaddr.Multiaddr) (net.Stream, error) {
	if p := i.senderMap.Get(id); p != nil && p.Stream != nil {
		log.Warn("the stream is created:", p.Stream)
		return p.Stream, nil
	}

	if len(i.Host.Peerstore().Addrs(id)) == 0 {
		i.Host.Peerstore().AddAddr(id, multiAddr, time.Minute*10)
	}
	s, err := i.Host.NewStream(i.ctx, id, Protocol)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	log.Info("add stream:", s)
	i.senderMap.Add(id, s, multiAddr)
	return s, nil
}

func (i *Instance) receive(s net.Stream) {
	log.Debug("start receive thread")
	reader := io.NewDelimitedReader(s, net.MessageSizeMax)
	for {
		msg := &mpb.Message{}
		err := reader.ReadMsg(msg)
		if err != nil {
			log.Error("the peer ", s.Conn().RemotePeer().Pretty(), i.Host.Peerstore().Addrs(s.Conn().RemotePeer()), "is disconnected:", err)
			s.Reset()
			return
		}
		log.Info("receive msg:", msg.Identify.String())
		if err := event.Publish(msg, msg.Identify); err != nil {
			log.Error("event publish error:", err)
			return
		}
	}
}

func (i *Instance) transmit(s net.Stream, sendMsg *mpb.Message) error {
	deadline := time.Now().Add(sendMessageTimeout)
	if dl, ok := i.ctx.Deadline(); ok {
		deadline = dl
		log.Info("set deal line:", deadline)
	}
	if err := s.SetWriteDeadline(deadline); err != nil {
		return errors.New(err.Error())
	}

	writer := io.NewDelimitedWriter(s)
	err := writer.WriteMsg(sendMsg)
	if err != nil {
		s.Reset()
		i.senderMap.Del(s.Conn().RemotePeer())
		return errors.New(err.Error())
	}
	if err := s.SetWriteDeadline(time.Time{}); err != nil {
		log.Warn("error resetting deadline: ", err)
	}
	log.Info("transmit message finished:", sendMsg.Identify)
	return nil
}
