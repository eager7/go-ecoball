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
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/net/message"
	"gx/ipfs/QmdbxjQWogRCHRaxhhGnYdT1oQJzL9GdqSKzCdqWr85AP2/pubsub"
)

const (
	bufferSize = 16
	errorStr = "dispatcher is not ready"
)

var (
	log = elog.NewLogger("disp", elog.DebugLog)
	dispatcher *Dispatcher
)

func InitMsgDispatcher(){
	if dispatcher == nil {
		dispatcher = &Dispatcher{
			pubsub.New(bufferSize),
		}
	}
}

type Dispatcher struct {
	ps     *pubsub.PubSub
}

func (ds *Dispatcher) publish(msg message.EcoBallNetMsg) {
	if message.MessageToStr[msg.Type()] == "" {
		log.Error("failed to find message type ", msg.Type())
		return
	}
	ds.ps.Pub(msg, message.MessageToStr[msg.Type()])
}

func (ds *Dispatcher) subscribe(msgs ...uint32) <-chan interface{} {
	var msgstr []string
	for _, msg := range msgs {
		if message.MessageToStr[msg] == "" {
			log.Error("failed to find message type ", msg)
			continue
		}
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
		if message.MessageToStr[msg] == "" {
			log.Error("failed to find message type ", msg)
			continue
		}
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