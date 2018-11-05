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

// Implement the gossip pull mediator

package gossip

import (
	"time"
	"sync"
	"context"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/network"
	"github.com/ecoball/go-ecoball/net/dispatcher"
	gpb "github.com/ecoball/go-ecoball/net/gossip/protos"
	mpb "github.com/ecoball/go-ecoball/net/message/pb"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
)

type Mediator interface {
	Stop()
	HandleMessage(msg message.EcoBallNetMsg)
}

type PullConfig struct {
	ChainId           uint16
	PullPeersCount    uint16
	PullInterval      time.Duration
	MsgType           gpb.PullMsgType
}

type pullMediator struct {
	config       PullConfig
	netinst      network.EcoballNetwork
	pullEngine   *GspPullEngine

	msgSubChan   <- chan interface{}

	stopCh       chan struct{}
	stopOnce     sync.Once

	Receiver
}

func NewPullMediator(ctx context.Context, cfg PullConfig, receiver Receiver) Mediator {
	inst, err := network.GetNetInstance()
	if err != nil {
		log.Error(err)
		return nil
	}

	pm := &pullMediator{
		config:       cfg,
		netinst:      inst,
		stopCh:       make(chan struct{}),
		Receiver:     receiver,
	}

	pm.pullEngine = newGspPullEngine(pm)

	pm.start(ctx)

	return pm
}

func(pm *pullMediator) start(ctx context.Context) {
	var err error
	if pm.msgSubChan, err = dispatcher.Subscribe(mpb.MsgType_APP_MSG_GOSSIP_PULL); err != nil {
		log.Error(err)
		return
	}

	pm.pullEngine.start(ctx, pm.config.PullInterval)

	go func() {
		defer log.Debug("a gossip pull engine exit")
		for {
			select {
			case <-ctx.Done():
				return
			case <-pm.stopCh:
				pm.pullEngine.Stop()
				return
			case m, ok := <-pm.msgSubChan:
				if !ok {
					continue
				}
				msg := m.(message.EcoBallNetMsg)
				pm.HandleMessage(msg)
			}
		}
	}()
}

func(pm *pullMediator) Stop() {
	stopFunc := func() {
		close(pm.stopCh)
	}
	pm.stopOnce.Do(stopFunc)
}

func(pm *pullMediator) Hello(id peer.ID) error {
	hello := new(GspPullHello)
	hello.SenderId = pm.netinst.Host().ID()
	hello.MsgType = pm.config.MsgType

	pullmsg := new(GossipPullMsg)
	pullmsg.SubMsg = hello
	data, err := pullmsg.Serialize()
	if err != nil {
		return err
	}

	msg := message.New(mpb.MsgType_APP_MSG_GOSSIP_PULL, data)
	log.Debug("send hello,", id,  hello)
	return pm.netinst.SendMsgToPeerWithId(id, msg)
}

func(pm *pullMediator) SendDigest(id peer.ID, digest interface{}) error {
	dig := new(GspPullDigest)
	dig.MsgType = pm.config.MsgType
	dig.SenderId = pm.netinst.Host().ID()
	dig.Digests = digest.([]string)

	pullmsg := new(GossipPullMsg)
	pullmsg.SubMsg = dig
	data, err := pullmsg.Serialize()
	if err != nil {
		return err
	}

	msg := message.New(mpb.MsgType_APP_MSG_GOSSIP_PULL, data)
	log.Debug("send digest,", id,  dig)
	return pm.netinst.SendMsgToPeerWithId(id, msg)
}

func(pm *pullMediator) SendRequest(id peer.ID, request interface{}) error {
	req := new(GspPullRequest)
	req.MsgType = pm.config.MsgType
	req.Asker = pm.netinst.Host().ID()
	req.ReqItems = request.([]string)

	pullmsg := new(GossipPullMsg)
	pullmsg.SubMsg = req
	data, err := pullmsg.Serialize()
	if err != nil {
		return err
	}

	msg := message.New(mpb.MsgType_APP_MSG_GOSSIP_PULL, data)
	log.Debug("send request,", id,  req)
	return pm.netinst.SendMsgToPeerWithId(id, msg)
}

func(pm *pullMediator) SendResponse(id peer.ID, response interface{}) error {
	res := new(GspPullReqAck)
	res.MsgType = pm.config.MsgType
	res.Responser = pm.netinst.Host().ID()

	for _, item := range response.([]string) {
		data := pm.GetItemData(item)
		env := &GspDataEnv{
			Data:  data,
		}
		res.Payload = append(res.Payload, env)
	}

	pullmsg := new(GossipPullMsg)
	pullmsg.SubMsg = res
	data, err := pullmsg.Serialize()
	if err != nil {
		return err
	}

	msg := message.New(mpb.MsgType_APP_MSG_GOSSIP_PULL, data)
	log.Debug("send response,", id,  res)
	return pm.netinst.SendMsgToPeerWithId(id, msg)
}

func (pm *pullMediator) HandleMessage(msg message.EcoBallNetMsg) {
	pullmsg := new(GossipPullMsg)
	if err := pullmsg.Deserialize(msg.Data()); err != nil {
		log.Error("failed to deserialize gossip pull msg")
		return
	}
	switch pullmsg.SubMsg.(type) {
	case *GspPullHello:
		pm.pullEngine.OnHello(pullmsg.SubMsg.(*GspPullHello))
	case *GspPullDigest:
		pm.pullEngine.OnDigest(pullmsg.SubMsg.(*GspPullDigest))
	case *GspPullRequest:
		pm.pullEngine.OnRequest(pullmsg.SubMsg.(*GspPullRequest))
	case *GspPullReqAck:
		pm.pullEngine.OnResponse(pullmsg.SubMsg.(*GspPullReqAck))
	default:
		log.Error("gossip pull engine receive an invalid message")
	}
}

func (pm *pullMediator) SelectRemotePeers() []peer.ID {
	return pm.netinst.SelectRandomPeers(pm.config.PullPeersCount)
}