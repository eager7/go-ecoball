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
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
)

const (
	OutBoxChanBufSize = 0
	InBoxChanBufSize  = 1024
)

type SendFunc func(pstore.PeerInfo, message.EcoBallNetMsg) error

type SendMsgJob struct {
	Peers    []*pstore.PeerInfo
	Msg      message.EcoBallNetMsg
}

type MsgWrapper struct {
	pi        pstore.PeerInfo
	emsg      message.EcoBallNetMsg
}

type MsgEngine struct {
	ctx         context.Context
	id          peer.ID
	//contains outgoing messages to peers
	outbox      chan (<-chan *MsgWrapper)
	//enqueue a msg job from servise
	inbox       chan interface{}

	quitworker  chan bool
}

func NewMsgEngine(ctx context.Context, id peer.ID) *MsgEngine {
	me := &MsgEngine{
		ctx:       ctx,
		id:        id,
		outbox:    make(chan (<-chan *MsgWrapper), OutBoxChanBufSize),
		inbox:     make(chan interface{}, InBoxChanBufSize),
	}
	go me.taskWorker(ctx)
	return me
}

func (me *MsgEngine)Outbox() <-chan (<-chan *MsgWrapper) {
	return me.outbox
}

func (me *MsgEngine)PushJob(job *SendMsgJob) {
	for _, peer := range job.Peers {
		if peer.ID == me.id {
			continue
		}
		me.inbox <- &MsgWrapper{*peer, job.Msg}
	}
}

func (me *MsgEngine)Stop() {
	me.quitworker <- true
}

func (me *MsgEngine) taskWorker(ctx context.Context) {
	defer close(me.outbox)
	for {
		oneTimeUse := make(chan *MsgWrapper, 1)
		select {
		case <-ctx.Done():
			return
		case <- me.quitworker:
			return
		case me.outbox <- oneTimeUse:
		}

		envelope, err := me.nextMsgWrapper(ctx)
		if err != nil {
			close(oneTimeUse)
			return // ctx cancelled
		}
		oneTimeUse <- envelope
		close(oneTimeUse)
	}
}

func (me *MsgEngine) nextMsgWrapper(ctx context.Context) (*MsgWrapper, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case mw := <-me.inbox:
			w, ok := mw.(*MsgWrapper)
			if ok {
				return w, nil
			}
		}
	}
}