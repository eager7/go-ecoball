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
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
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

func (ds *Dispatcher) publish(id common.Hash, msg string) {
	ds.ps.Pub(msg, id.HexString())
}

func (ds *Dispatcher) subscribe(msgs ...common.Hash) chan interface{} {
	var msgstr []string
	for _, msg := range msgs {
		msgstr = append(msgstr, msg.HexString())
	}
	if len(msgstr) > 0 {
		return ds.ps.Sub(msgstr...)
	}

	return nil
}

func (ds *Dispatcher) subscribeOnce(msgs ...common.Hash) chan interface{} {
	var msgstr []string
	for _, msg := range msgs {
		msgstr = append(msgstr, msg.HexString())
	}
	if len(msgstr) > 0 {
		return ds.ps.SubOnce(msgstr...)
	}

	return nil
}

func (ds *Dispatcher) subscribeOnceEach(msgs ...common.Hash) chan interface{} {
	var msgstr []string
	for _, msg := range msgs {
		msgstr = append(msgstr, msg.HexString())
	}
	if len(msgstr) > 0 {
		return ds.ps.SubOnceEach(msgstr...)
	}

	return nil
}

func (ds *Dispatcher) unsubscribe(chn chan interface{}, msgType ...common.Hash) {
	var msgstr []string
	for _, msg := range msgType {
		msgstr = append(msgstr, msg.HexString())
	}

	ds.ps.Unsub(chn, msgstr...)
}

// Not safe to call more than once.
func (ds *Dispatcher) shutdown() {
	// shutdown the pubsub.
	ds.ps.Shutdown()
}

func Subscribe (msgs ...common.Hash) (chan interface{}, error) {
	if dispatcher == nil {
		return nil, fmt.Errorf(errorStr)
	}
	return dispatcher.subscribe(msgs...), nil
}

func SubscribeOnceEach (msgs ...common.Hash) (chan interface{}, error) {
	if dispatcher == nil {
		return nil, fmt.Errorf(errorStr)
	}
	return dispatcher.subscribeOnceEach(msgs...), nil
}

func UnSubscribe (chn chan interface{}, msgs ...common.Hash) error {
	if dispatcher == nil {
		return fmt.Errorf(errorStr)
	}
	dispatcher.unsubscribe(chn, msgs...)

	return nil
}

func PublishTrxRes (id common.Hash, msg string) error {
	if dispatcher == nil {
		return fmt.Errorf(errorStr)
	}
	dispatcher.publish(id, msg)

	return nil
}