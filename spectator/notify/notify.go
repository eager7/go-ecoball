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

package notify

import (
	"net"

	"github.com/ecoball/go-ecoball/spectator/connect"
	"github.com/ecoball/go-ecoball/spectator/info"
)

func ReceiveNotify(conn net.Conn) {
	for {
		buf, n, err := info.ReadData(conn)
		if nil != err {
			log.Error("explorer server read data error: ", err)
			break
		}

		one := info.OneNotify{info.InfoNil, []byte{}, 0}
		if err := one.Deserialize(buf[:n]); nil != err {
			log.Error("explorer server notify.Deserialize error: ", err)
			continue
		}
		go dispatch(conn, one)
	}

	connect.Onlookers.Delete(conn)
}

func dispatch(conn net.Conn, one info.OneNotify) {
	switch one.InfoType {
	case info.SynBlock:
		if err := HandleSynBlock(conn, one); nil != err {
			log.Error("handleBlock error: ", err)
		}
	default:

	}
}
