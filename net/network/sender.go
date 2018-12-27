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
	"gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/io"
	"gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	"sync"
	"time"
)

const connectedAddrTTL = time.Minute * 10
const connectTry = 5

type messageSender struct {
	stream   net.Stream
	lock     sync.Mutex
	peerInfo peerstore.PeerInfo
	net      *NetWork
}

func NewMsgSender(pi peerstore.PeerInfo, s net.Stream, p2pNet *NetWork) *messageSender {
	return &messageSender{stream: s, lock: sync.Mutex{}, peerInfo: pi, net: p2pNet}
}

func (sender *messageSender) newStream() (err error) {
	if sender.stream != nil {
		return nil
	}
	if len(sender.net.host.Peerstore().Addrs(sender.peerInfo.ID)) == 0 && len(sender.peerInfo.Addrs) > 0 {
		sender.net.host.Peerstore().AddAddrs(sender.peerInfo.ID, sender.peerInfo.Addrs, connectedAddrTTL)
	}
	log.Info("connect to peer:", sender.String())
	if sender.stream, err = sender.net.host.NewStream(sender.net.ctx, sender.peerInfo.ID, ProtocolP2pV1); err != nil { //basic_host.go
		return errors.New(err.Error())
	}

	return nil
}

func (sender *messageSender) SendMessage(ctx context.Context, msg message.EcoBallNetMsg) error {
	sender.lock.Lock()
	defer sender.lock.Unlock()
	if err := sender.send(ctx, msg); err != nil {
		go net.FullClose(sender.stream)
		sender.stream = nil
		return err
	}
	log.Debug(fmt.Sprintf("success send msg %s to peer:", msg.Type().String()), sender.peerInfo)

	return nil
}

func (net *NetWork) NewMessageSender(p peerstore.PeerInfo) (*messageSender, error) {
	sender := net.SenderMap.Get(p.ID)
	if sender != nil {
		return sender, nil
	}
	sender = NewMsgSender(p, nil, net)

	var i = 0
RETRY:
	if err := sender.newStream(); err != nil {
		log.Error("new stream failed:", err.Error())
		if i >= connectTry {
			return nil, errors.New(fmt.Sprintf("can't create new stream:%s", err.Error()))
		}
		i += 1
		goto RETRY
	}
	net.SenderMap.Add(p.ID, sender)
	go net.HandleNewStream(sender.stream) /*当本节点先和对端建立连接时，对端再次连接将无法触发handler函数，因此需要在此启动接收线程*/

	return sender, nil
}

func (sender *messageSender) send(ctx context.Context, msg message.EcoBallNetMsg) error {
	deadline := time.Now().Add(sendMessageTimeout)
	if dl, ok := ctx.Deadline(); ok {
		deadline = dl
	}

	if err := sender.stream.SetWriteDeadline(deadline); err != nil {
		log.Warn("error setting deadline: ", err)
	}

	switch sender.stream.Protocol() {
	case ProtocolP2pV1:
		pbw := io.NewDelimitedWriter(sender.stream)
		if err := pbw.WriteMsg(msg.ToProtoV1()); err != nil {
			return errors.New(err.Error())
		}
	default:
		return fmt.Errorf("unrecognized protocol on remote: %s", sender.stream.Protocol())
	}

	if err := sender.stream.SetWriteDeadline(time.Time{}); err != nil {
		log.Warn("error resetting deadline: ", err)
	}
	return nil
}

func (sender *messageSender) String() string {
	ret := sender.peerInfo.ID.Pretty()
	for _, addr := range sender.peerInfo.Addrs {
		ret += addr.String()
	}
	return ret
}
