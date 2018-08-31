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

package dispatcher

import (
	"fmt"
	"github.com/ecoball/go-ecoball/net/message"
	"gx/ipfs/QmdbxjQWogRCHRaxhhGnYdT1oQJzL9GdqSKzCdqWr85AP2/pubsub"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
)

const (
	bufferSize = 16
	errorStr = "dispatcher is not ready"
)

var (
	dispatcher *Dispatcher
)

func InitMsgDispatcher(sender MsgNode){
	if dispatcher == nil {
		dispatcher = &Dispatcher{
			pubsub.New(bufferSize),
			sender,
		}
	}
}

type Dispatcher struct {
	ps     *pubsub.PubSub
	sender MsgNode
}

func (ds *Dispatcher) publish(msg message.EcoBallNetMsg) {
	ds.ps.Pub(msg, message.MessageToStr[msg.Type()])
}

func (ds *Dispatcher) subscribe(msgs ...uint32) <-chan interface{} {
	var msgstr []string
	for _, msg := range msgs {
		msgstr = append(msgstr, message.MessageToStr[msg])
	}
	if len(msgstr) > 0 {
		return ds.ps.Sub(msgstr...)
	}

	return nil
}

func (ds *Dispatcher) unsubscribe(chn chan interface{}, msgType ...uint32) {
	var msgstr []string
	for _, msg := range msgType {
		msgstr = append(msgstr, message.MessageToStr[msg])
	}

	ds.ps.Unsub(chn, msgstr...)
}

// Not safe to call more than once.
func (ds *Dispatcher) shutdown() {
	// shutdown the pubsub.
	ds.ps.Shutdown()
}

func Subscribe (msgs ...uint32) (<-chan interface{}, error) {
	if dispatcher == nil {
		return nil, fmt.Errorf(errorStr)
	}
	return dispatcher.subscribe(msgs...), nil
}

func UnSubscribe (chn chan interface{}, msgs ...uint32) error {
	if dispatcher == nil {
		return fmt.Errorf(errorStr)
	}
	dispatcher.unsubscribe(chn, msgs...)

	return nil
}

func Publish (msg message.EcoBallNetMsg) error {
	if dispatcher == nil {
		return fmt.Errorf(errorStr)
	}
	dispatcher.publish(msg)

	return nil
}

func SendMessage (pid peer.ID, msg message.EcoBallNetMsg) error {
	if dispatcher == nil {
		return fmt.Errorf(errorStr)
	}
	dispatcher.sender.SendMsg2Peer(pid, msg)

	return nil
}

func SendMsgToRandomPeers (peerCounts int, msg message.EcoBallNetMsg) error {
	if dispatcher == nil {
		return fmt.Errorf(errorStr)
	}
	dispatcher.sender.SendMsg2RandomPeers(peerCounts, msg)

	return nil
}

func BroadcastMessage (msg message.EcoBallNetMsg) error {
	if dispatcher == nil {
		return fmt.Errorf(errorStr)
	}
	dispatcher.sender.SendBroadcastMsg(msg)

	return nil
}

func GetPeerID () (peer.ID, error) {
	if dispatcher == nil {
		return "", fmt.Errorf(errorStr)
	}
	return dispatcher.sender.SelfRawId(), nil
}

func GetRandomPeers (k int) ([]peer.ID, error) {
	if dispatcher == nil {
		return []peer.ID{}, fmt.Errorf(errorStr)
	}
	return dispatcher.sender.SelectRandomPeers(k), nil
}