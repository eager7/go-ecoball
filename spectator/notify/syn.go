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

	scanSyn "github.com/ecoball/eballscan/syn"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/spectator/info"
)

var (
	CoreLedger ledger.Ledger
	log        = elog.NewLogger("notify", elog.DebugLog)
)

func HandleSynBlock(conn net.Conn, one info.OneNotify) error {
	if config.ConsensusAlgorithm == "SOLO" {
		if one.BlockType == 0 {
			var blockHight scanSyn.BlockHeight
			if err := blockHight.Deserialize(one.Info); nil != err {
				return err
			}
			hight := uint64(blockHight)

			nowHight := CoreLedger.GetCurrentHeight(config.ChainHash)
			for hight < nowHight {
				hight++

				block, err := CoreLedger.GetTxBlockByHeight(config.ChainHash, hight)
				if nil != err {
					log.Error("GetTxBlockByHeight error: ", err)
					continue
				}

				if err := send_message(info.InfoBlock, conn, block); nil != err {
					log.Error("send_message error: ", err)
				}
			}
		}
	}

	return nil
}

func send_message(oneType info.NotifyType, conn net.Conn, message info.NotifyInfo) error {
	notify, err := info.NewOneNotify(oneType, message)
	if nil != err {
		log.Error("NewOneNotify error: ", err)
		return err
	}

	data, err := notify.Serialize()
	if nil != err {
		log.Error("Serialize error: ", err)
		return err
	}

	data = info.MessageDecorate(data)
	if _, err := conn.Write(data); nil != err {
		addr := conn.RemoteAddr().String()
		log.Warn(addr, " disconnect")
		return err
	}
	return nil
}
