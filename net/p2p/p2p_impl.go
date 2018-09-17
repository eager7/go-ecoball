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
package p2p

import (
	"time"
	"fmt"
	"io"
	"context"
	"sync"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/util"
	"github.com/ecoball/go-ecoball/common/config"
	kb "gx/ipfs/QmesQqwonP618R7cJZoFfA4ioYhhMKnDmtUxcAvvxEEGnw/go-libp2p-kbucket"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	ggio "gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/io"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	"gx/ipfs/QmXuucFcuvAWYAJfhHV2h4BYreHEAsLSsiquosiXeuduTN/go-libp2p-interface-connmgr"
	"gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p/p2p/discovery"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"

)

const (
	// K is the maximum number of requests to perform before returning failure.
	KValue                 = 20
	// Alpha is the concurrency factor for asynchronous requests.
	AlphaValue             = 3

	sendMessageChanBuff    = 1024
	sendMessageTimeout     = time.Minute * 10

	discoveryConnTimeout   = time.Second * 30

	ServiceTag                 = "_net-discovery._udp"

	ProtocolP2pV1  protocol.ID = "/ecoball/app/1.0.0"
)

var (
	log = elog.NewLogger("p2p", elog.DebugLog)
	netImpl *impl
)

func NewNetwork(ctx context.Context, host host.Host) EcoballNetwork {
	if netImpl != nil {
		return netImpl
	}
	netImpl = &impl{
		ctx:          ctx,
		host:         host,
		sendJbQueue:  make(chan interface{}, sendMessageChanBuff),
	}
	netImpl.routingTable = initRoutingTable(host)

	host.SetStreamHandler(ProtocolP2pV1, netImpl.handleNewStream)
	host.Network().Notify((*netNotifiee)(netImpl))
	// TODO: StopNotify.
	return netImpl
}

func GetPeerID() (peer.ID, error) {
	if netImpl == nil {
		return "", fmt.Errorf(networkError)
	}
	return netImpl.host.ID(), nil
}

func GetRandomPeers(k int) []peer.ID {
	if netImpl == nil {
		return []peer.ID{}
	}
	return netImpl.selectRandomPeers(k)
}

func initRoutingTable(host host.Host) (table *kb.RoutingTable) {
	peerID := kb.ConvertPeerID(host.ID())

	rt := kb.NewRoutingTable(
		KValue,
		peerID,
		time.Minute,   //TOD, should come from config file
		host.Peerstore())
	cmgr := host.ConnManager()
	rt.PeerAdded = func(p peer.ID) {
		cmgr.TagPeer(p, "kbucket", 5)
	}
	rt.PeerRemoved = func(p peer.ID) {
		cmgr.UntagPeer(p, "kbucket")
	}

	return rt
}

// impl transforms the network interface, which sends and receives
// NetMessage objects, into the ecoball network interface.
type impl struct {
	ctx          context.Context
	host         host.Host

	// inbound messages from the network are forwarded to the receiver
	receiver     Receiver

	sendJbQueue  chan interface{}

	routingTable *kb.RoutingTable
	rtLock       sync.Mutex
}

type streamMessageSender struct {
	s inet.Stream
}

func (s *streamMessageSender) Close() error {
	return inet.FullClose(s.s)
}

func (s *streamMessageSender) Reset() error {
	return s.s.Reset()
}

func (s *streamMessageSender) SendMsg(ctx context.Context, msg message.EcoBallNetMsg) error {
	return msgToStream(ctx, s.s, msg)
}

func msgToStream(ctx context.Context, s inet.Stream, msg message.EcoBallNetMsg) error {
	deadline := time.Now().Add(sendMessageTimeout)
	if dl, ok := ctx.Deadline(); ok {
		deadline = dl
	}

	if err := s.SetWriteDeadline(deadline); err != nil {
		log.Warn("error setting deadline: ", err)
	}

	switch s.Protocol() {
	case ProtocolP2pV1:
		if err := msg.ToNetV1(s); err != nil {
			log.Debug("error: ", err)
			return err
		}
	default:
		return fmt.Errorf("unrecognized protocol on remote: %s", s.Protocol())
	}

	if err := s.SetWriteDeadline(time.Time{}); err != nil {
		log.Warn("error resetting deadline: ", err)
	}
	return nil
}

func (bsnet *impl) Host() host.Host {
	return bsnet.host
}

func (bsnet *impl) NewMessageSender(ctx context.Context, p peer.ID) (MessageSender, error) {
	s, err := bsnet.newStreamToPeer(ctx, p)
	if err != nil {
		return nil, err
	}
	return &streamMessageSender{s: s}, nil
}

func (bsnet *impl) newStreamToPeer(ctx context.Context, p peer.ID) (inet.Stream, error) {
	return bsnet.host.NewStream(ctx, p, ProtocolP2pV1)
}

func (bsnet *impl) sendMessage(
	ctx context.Context,
	p peer.ID,
	outgoing message.EcoBallNetMsg) error {

	s, err := bsnet.newStreamToPeer(ctx, p)
	if err != nil {
		return err
	}

	if err = msgToStream(ctx, s, outgoing); err != nil {
		s.Reset()
		return err
	}
	log.Debug("send msg to ", p.Pretty())
	return inet.FullClose(s)
}

func (bsnet *impl) SetDelegate(r Receiver) {
	bsnet.receiver = r
}

func (bsnet *impl) ConnectTo(ctx context.Context, p peer.ID) error {
	return bsnet.host.Connect(ctx, pstore.PeerInfo{ID: p})
}

func (bsnet *impl) FindPeer(ctx context.Context, id peer.ID) (pstore.PeerInfo, error) {
	// Check if were already connected to them
	if pi := bsnet.FindLocal(id); pi.ID != "" {
		return pi, nil
	}

	peers := bsnet.routingTable.NearestPeers(kb.ConvertPeerID(id), AlphaValue)
	if len(peers) == 0 {
		return pstore.PeerInfo{}, kb.ErrLookupFailure
	}

	for _, p := range peers {
		if p == id {
			log.Debug("found target peer in list of closest peers...")
			return bsnet.host.Peerstore().PeerInfo(p), nil
		}
	}

	return pstore.PeerInfo{}, kb.ErrLookupFailure
}

func (bsnet *impl) FindLocal(id peer.ID) pstore.PeerInfo {
	switch bsnet.host.Network().Connectedness(id) {
	case inet.Connected, inet.CanConnect:
		return bsnet.host.Peerstore().PeerInfo(id)
	default:
		return pstore.PeerInfo{}
	}
}

func (bsnet *impl) ConnectionManager() ifconnmgr.ConnManager {
	return bsnet.host.ConnManager()
}

// select randomly k peers from remote peers and returns them.
func (bsnet *impl) selectRandomPeers(k int) []peer.ID {
	if netImpl == nil {
		return []peer.ID{}
	}
	conns := netImpl.host.Network().Conns()
	if len(conns) < k {
		k = len(conns)
	}
	indices := util.GetRandomIndices(k, len(conns)-1)
	peers := make([]peer.ID, len(indices))
	for i, j := range indices {
		pid := conns[j].RemotePeer()
		peers[i] = pid
	}

	return peers
}

func (bsnet *impl) update(p peer.ID) {
	bsnet.rtLock.Lock()
	defer bsnet.rtLock.Unlock()
	bsnet.routingTable.Update(p)
}

func (bsnet *impl) remove(p peer.ID) {
	bsnet.rtLock.Lock()
	defer bsnet.rtLock.Unlock()
	bsnet.routingTable.Remove(p)
}

func (bsnet *impl) nearestPeersToQuery(id peer.ID, count int) []peer.ID {
	closer := bsnet.routingTable.NearestPeers(kb.ConvertKey(id.String()), count)
	return closer
}

func (bsnet *impl) handleNewStream(s inet.Stream) {
	go bsnet.handleNewStreamMsg(s)
}

func (bsnet *impl) handleNewStreamMsg(s inet.Stream) {
	defer s.Close()
	if bsnet.receiver == nil {
		s.Reset()
		return
	}

	reader := ggio.NewDelimitedReader(s, inet.MessageSizeMax)
	for {
		received, err := message.FromPBReader(reader)
		if err != nil {
			if err != io.EOF {
				s.Reset()
				go bsnet.receiver.ReceiveError(err)
				log.Debug("p2p net handleNewStream from %s error: ", s.Conn().RemotePeer(), err)
			}
			return
		}

		p := s.Conn().RemotePeer()
		ctx := context.Background()
		log.Debug("p2p net handleNewStream from ", s.Conn().RemotePeer())
		bsnet.update(p)
		bsnet.receiver.ReceiveMessage(ctx, p, received)
	}
}

func (bsnet *impl) SendMsgJob(job *message.SendMsgJob) {
	bsnet.sendJbQueue <- job
}

func (bsnet *impl) handleSendJob() {
	go func() {
		for {
			select {
			case job, ok := <- bsnet.sendJbQueue:
				if !ok {
					log.Error("chan for sending job queue was closed")
					return
				}
				sendJb, ok := job.(*message.SendMsgJob)
				if ok {
					for _, pi := range sendJb.Peers {
						if pi.ID == bsnet.host.ID() {
							continue
						}
						addr := bsnet.host.Peerstore().Addrs(pi.ID)
						if len(addr) == 0 && len(pi.Addrs) >0 {
							if err := bsnet.host.Connect(bsnet.ctx, *pi); err != nil {
								log.Error(err)
								continue
							}
						}
						if err:= bsnet.sendMessage(bsnet.ctx, pi.ID, sendJb.Msg); err != nil {
							log.Error("send message to ", pi.ID.Pretty(), err)
						}
					}
				}
			}
		}
	}()
}



func (bsnet *impl) StartLocalDiscovery() error {
	service, err := discovery.NewMdnsService(bsnet.ctx, bsnet.host, 10*time.Second, ServiceTag)
	if err != nil {
		return fmt.Errorf("net discovery error,", err)
	}
	service.RegisterNotifee((*netNotifiee)(bsnet))

	return nil
}

func (bsnet *impl) Start() {
	// it is up to the requirement of network sharding,
	if config.EnableLocalDiscovery {
		if err := bsnet.StartLocalDiscovery(); err != nil {
			log.Error(err)
		}
		log.Debug("start p2p local discovery")
	}

	bsnet.handleSendJob()
}