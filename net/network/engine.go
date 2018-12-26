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

// Implement the network output message engine

package network

import (
	"context"
	"github.com/ecoball/go-ecoball/net/message"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
)

const (
	OutBoxChanBufSize = 0
	InBoxChanBufSize  = 1024
)

type SendMsgJob struct {
	Peers []*peerstore.PeerInfo
	Msg   message.EcoBallNetMsg
}

func (s *SendMsgJob) String() string {
	var ret string
	for _, p := range s.Peers {
		ret += p.ID.Pretty()
		for _, addr := range p.Addrs {
			ret += addr.String()
		}
	}
	ret += "-" + s.Msg.Type().String()
	return ret
}

type MsgWrapper struct {
	pi   peerstore.PeerInfo
	eMsg message.EcoBallNetMsg
}

type MsgEngine struct {
	ctx        context.Context
	id         peer.ID
	outbox     chan (<-chan *MsgWrapper) //contains outgoing messages to peers
	inbox      chan interface{}          //enqueue a msg job from service
	quitWorker chan bool
}

func NewMsgEngine(ctx context.Context, id peer.ID) *MsgEngine {
	me := &MsgEngine{
		ctx:    ctx,
		id:     id,
		outbox: make(chan (<-chan *MsgWrapper), OutBoxChanBufSize),
		inbox:  make(chan interface{}, InBoxChanBufSize),
	}
	go me.taskWorker(ctx)
	return me
}

func (m *MsgEngine) taskWorker(ctx context.Context) {
	defer close(m.outbox)
	for {
		oneTimeUse := make(chan *MsgWrapper, 1)
		select {
		case <-ctx.Done():
			return
		case <-m.quitWorker:
			return
		case m.outbox <- oneTimeUse:
		}

		envelope, err := m.nextMsgWrapper(ctx)
		if err != nil {
			close(oneTimeUse)
			return // ctx cancelled
		}
		oneTimeUse <- envelope
		close(oneTimeUse)
	}
}
func (m *MsgEngine) nextMsgWrapper(ctx context.Context) (*MsgWrapper, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case mw := <-m.inbox:
			w, ok := mw.(*MsgWrapper)
			if ok {
				return w, nil
			}
		}
	}
}
func (m *MsgEngine) Outbox() <-chan (<-chan *MsgWrapper) {
	return m.outbox
}
func (m *MsgEngine) PushJob(job *SendMsgJob) {
	for _, p := range job.Peers {
		if p.ID == m.id {
			continue
		}
		m.inbox <- &MsgWrapper{*p, job.Msg}
	}
}
func (m *MsgEngine) Stop() {
	m.quitWorker <- true
}

func (net *NetWork) nativeMessageLoop() {
	go func() {
		for {
			select {
			case msg := <-net.BroadCastCh:
				log.Debug("BroadCastCh receive msg:", msg.Type().String())
				net.BroadcastMessageToNeighbors(msg)
			}
		}
	}()
}
func (net *NetWork) startSendWorkers() {
	for i := 0; i < sendWorkerCount; i++ {
		i := i
		go net.sendWorker(i)
	}
}
func (net *NetWork) AddMsgJob(job *SendMsgJob) {
	log.Debug("put msg in send pool:", job.String())
	net.engine.PushJob(job)
}
func (net *NetWork) sendWorker(id int) {
	defer log.Debug("network send message worker ", id, " shutting down.")
	for {
		select {
		case nextWrapper := <-net.engine.Outbox():
			select {
			case wrapper, ok := <-nextWrapper:
				if !ok {
					continue
				}
				if err := net.sendMessage(wrapper.pi, wrapper.eMsg); err != nil {
					log.Error("send message to ", wrapper.pi, net.host.Peerstore().Addrs(wrapper.pi.ID), err)
				}
			case <-net.ctx.Done():
				return
			}
		case <-net.ctx.Done():
			return
		}
	}
}
func (net *NetWork) sendMessage(p peerstore.PeerInfo, outgoing message.EcoBallNetMsg) error {
	log.Info("send message to", p.ID.Pretty(), p.Addrs)
	sender, err := net.NewMessageSender(p)
	if err != nil {
		return err
	}
	return sender.SendMessage(net.ctx, outgoing)
}

func (net *NetWork) BroadcastMessageToNeighbors(msg message.EcoBallNetMsg) error {
	peers := net.Neighbors()
	if len(peers) > 0 {
		sendJob := &SendMsgJob{peers, msg}
		net.AddMsgJob(sendJob)
	}

	return nil
}
