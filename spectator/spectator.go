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

package spectator

import (
	"errors"
	"fmt"
	"net"

	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/spectator/connect"
	"github.com/ecoball/go-ecoball/spectator/notify"
)

var (
	log = elog.NewLogger("spectator", elog.DebugLog)
)

func Bystander(l ledger.Ledger) error {
	notify.CoreLedger = l
	listener, err := net.Listen("tcp", "127.0.0.1:9000")
	if nil != err {
		log.Error("explorer server net.Listen error: ", err)
		return errors.New("explorer server net.Listen error: " + fmt.Sprintf("%v", err))
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if nil != err {
			log.Error("explorer server net.Accept error: ", err)
			return errors.New("explorer server net.Accept error: " + fmt.Sprintf("%v", err))
		}

		connect.Onlookers.Add(conn)

		go notify.ReceiveNotify(conn)
	}
}
