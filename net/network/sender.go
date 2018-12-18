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
	"sync"
	"fmt"
	"time"
	"context"
	"github.com/ecoball/go-ecoball/net/message"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common"
)

const connectedAddrTTL = time.Minute * 10

type messageSender struct {
	s         inet.Stream
	lk        sync.Mutex
	p         peerstore.PeerInfo
	net       *NetImpl
	invalid   bool
}

func NewMsgSender(pi peerstore.PeerInfo, p2pNet *NetImpl) *messageSender {
	return &messageSender{p: pi, net: p2pNet}
}

func (ms *messageSender) invalidate() {
	ms.invalid = true
	if ms.s != nil {
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

func (ms *messageSender) prep() error {
	if ms.invalid {
		return errors.New(log, "message sender has been invalidated")
	}
	if ms.s != nil {
		return nil
	}

	addr := ms.net.host.Peerstore().Addrs(ms.p.ID)
	if len(addr) == 0 && len(ms.p.Addrs) >0 {
		ms.net.host.Peerstore().AddAddrs(ms.p.ID, ms.p.Addrs, connectedAddrTTL)
	}

	nstr, err := ms.newStreamToPeer(ms.net.ctx, ms.p.ID, ProtocolP2pV1)
	if err != nil {
		return errors.New(log, err.Error())
	}

	ms.s = nstr

	return nil
}

func (ms *messageSender) newStreamToPeer(ctx context.Context, p peer.ID, pids ...protocol.ID) (inet.Stream, error) {
	log.Info(ms.net.host.Peerstore().Addrs(p), common.JsonString(pids))
	return ms.net.host.NewStream(ctx, p, pids...) //basic_host.go
}

func (ms *messageSender) SendMsg(ctx context.Context, msg message.EcoBallNetMsg) error {
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

	log.Debug(fmt.Sprintf("send msg %s to peer:", msg.Type().String()), ms.p, "data:", string(msg.Data()))

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
			return errors.New(log, err.Error())
		}
	default:
		return fmt.Errorf("unrecognized protocol on remote: %s", s.Protocol())
	}

	if err := s.SetWriteDeadline(time.Time{}); err != nil {
		log.Warn("error resetting deadline: ", err)
	}
	return nil
}