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

// Implement the message sender

package network

import (
	"context"
	"fmt"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/net/message"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"sync"
	"time"
)

const connectedAddrTTL = time.Minute * 10

type messageSender struct {
	s       inet.Stream
	lk      sync.Mutex
	p       peerstore.PeerInfo
	net     *NetWork
	invalid bool
}

func NewMsgSender(pi peerstore.PeerInfo, p2pNet *NetWork) *messageSender {
	return &messageSender{p: pi, net: p2pNet}
}

func (ms *messageSender) invalidate() {
	ms.invalid = true
	if ms.s != nil {
		log.Error("reset stream")
		ms.s.Reset()
		ms.s = nil
	}
}

func (ms *messageSender) prepOrInvalidate() error {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	if err := ms.prep(); err != nil {
		ms.invalidate()
		return err
	}
	return nil
}

func (ms *messageSender) prep() (err error) {
	if ms.invalid {
		return errors.New("message sender has been invalidated")
	}
	if ms.s != nil {
		return nil
	}

	addr := ms.net.host.Peerstore().Addrs(ms.p.ID)
	if len(addr) == 0 && len(ms.p.Addrs) > 0 {
		ms.net.host.Peerstore().AddAddrs(ms.p.ID, ms.p.Addrs, connectedAddrTTL)
	}

	var stream inet.Stream
	retry := 0
	for retry < 3 {
		if stream, err = ms.net.host.NewStream(ms.net.ctx, ms.p.ID, ProtocolP2pV1); err != nil { //basic_host.go
			err = errors.New(err.Error())
			retry += 1
		} else {
			break
		}
		if retry == 3 {
			return err
		}
	}

	ms.s = stream

	return nil
}

func (ms *messageSender) SendMessage(ctx context.Context, msg message.EcoBallNetMsg) error {
	ms.lk.Lock()
	defer ms.lk.Unlock()

	if err := ms.prep(); err != nil {
		return err
	}

	if err := msgToStream(ctx, ms.s, msg); err != nil {
		go inet.FullClose(ms.s)
		ms.s = nil
		log.Warn(err)
		return err
	}

	log.Debug(fmt.Sprintf("success send msg %s to peer:", msg.Type().String()), ms.p)

	return nil
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
			return errors.New(err.Error())
		}
	default:
		return fmt.Errorf("unrecognized protocol on remote: %s", s.Protocol())
	}

	if err := s.SetWriteDeadline(time.Time{}); err != nil {
		log.Warn("error resetting deadline: ", err)
	}
	return nil
}

func (net *NetWork) NewMessageSender(p pstore.PeerInfo) (*messageSender, error) {
	sender := net.SenderMap.Get(p.ID)
	if sender != nil {
		return sender, nil
	}
	sender = NewMsgSender(p, net)

	if err := sender.prepOrInvalidate(); err != nil {
		return nil, err
	}
	net.SenderMap.Add(p.ID, sender)

	return sender, nil
}
