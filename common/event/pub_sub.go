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
package event

import (
	"errors"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"gx/ipfs/QmdbxjQWogRCHRaxhhGnYdT1oQJzL9GdqSKzCdqWr85AP2/pubsub"
)

const (
	channelSize = 16
)

var errorStr = errors.New("dispatcher is not ready")

var (
	dispatcher *Dispatcher
)

func InitMsgDispatcher() {
	if dispatcher == nil {
		dispatcher = &Dispatcher{
			pubsub.New(channelSize),
		}
	}
}

type Dispatcher struct {
	ps *pubsub.PubSub
}

func (ds *Dispatcher) subscribe(topics ...mpb.Identify) chan interface{} {
	var strings []string
	for _, topic := range topics {
		strings = append(strings, topic.String())
	}
	if len(strings) > 0 {
		return ds.ps.Sub(strings...)
	}
	return nil
}

func (ds *Dispatcher) subscribeOnce(topics ...mpb.Identify) chan interface{} {
	var strings []string
	for _, topic := range topics {
		strings = append(strings, topic.String())
	}
	if len(strings) > 0 {
		return ds.ps.SubOnce(strings...)
	}
	return nil
}

func (ds *Dispatcher) unsubscribe(chn chan interface{}, topics ...mpb.Identify) {
	var strings []string
	for _, topic := range topics {
		strings = append(strings, topic.String())
	}
	ds.ps.Unsub(chn, strings...)
}

// Not safe to call more than once.
func (ds *Dispatcher) shutdown() {
	// shutdown the pub sub.
	ds.ps.Shutdown()
}

/*订阅消息,返回一个channel,循环接收此channel,将返回mpb.Message{Identify:0, Payload:nil,}类型数据,根据类型进行解析即可*/
func Subscribe(topics ...mpb.Identify) (chan interface{}, error) {
	if dispatcher == nil {
		return nil, errorStr
	}
	return dispatcher.subscribe(topics...), nil
}

/*订阅消息,返回一个channel,循环接收此channel,将返回mpb.Message{Identify:0, Payload:nil,}类型数据,根据类型进行解析即可,channel只能接收一次*/
func SubOnceEach(topics ...string) (chan interface{}, error) {
	if dispatcher == nil {
		return nil, errorStr
	}
	return dispatcher.ps.SubOnceEach(topics...), nil
}

func UnSubscribe(chn chan interface{}, topics ...mpb.Identify) error {
	if dispatcher == nil {
		return errorStr
	}
	dispatcher.unsubscribe(chn, topics...)
	return nil
}

func Publish(msg *mpb.Message, topics ...mpb.Identify) error {
	if dispatcher == nil {
		return errorStr
	}
	var strings []string
	for _, topic := range topics {
		strings = append(strings, topic.String())
	}
	if len(strings) > 0 {
		dispatcher.ps.Pub(msg, strings...)
	}
	return nil
}

func PublishCustom(msg interface{}, topics ...string) error {
	if dispatcher == nil {
		return errorStr
	}
	dispatcher.ps.Pub(msg, topics...)
	return nil
}
