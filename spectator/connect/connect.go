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

package connect

import (
	"net"
	"sync"

	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/spectator/info"
)

var (
	Onlookers = onlooker{connects: make([]net.Conn, 0, 10)}
	log       = elog.NewLogger("connect", elog.DebugLog)
)

type onlooker struct {
	connects []net.Conn
	sync.Mutex
}

func (this *onlooker) Add(conn net.Conn) {
	this.Lock()
	defer this.Unlock()

	for _, v := range this.connects {
		if conn == v {
			return
		}
	}

	this.connects = append(this.connects, conn)
}

func (this *onlooker) Delete(conn net.Conn) {
	this.Lock()
	defer this.Unlock()

	for k, v := range this.connects {
		if conn == v {
			this.connects = append(this.connects[:k], this.connects[k+1:]...)
			break
		}
	}
}

func (this *onlooker) notify(message []byte) {
	this.Lock()
	defer this.Unlock()

	for k, v := range this.connects {
		addr := v.RemoteAddr().String()

		if _, err := v.Write(message); nil != err {
			log.Warn(addr, " disconnect")
			this.connects = append(this.connects[:k], this.connects[k+1:]...)
		}

		log.Info("send to addres: ", addr, " message: ", string(message))
	}
}

func Notify(infoType info.NotifyType, message info.NotifyInfo) error {
	one, err := info.NewOneNotify(infoType, message)
	if nil != err {
		return err
	}

	data, err := one.Serialize()
	if nil != err {
		return nil
	}
	data = info.MessageDecorate(data)

	//fmt.Println("new block: ", string(data))
	Onlookers.notify(data)
	return nil
}
