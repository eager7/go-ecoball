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
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/spectator/info"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/common/message/mpb"
)

var (
	CoreLedger ledger.Ledger
	log        = elog.NewLogger("notify", elog.DebugLog)
)

func HandleSynBlock(conn net.Conn, one info.OneNotify) error {
	if config.ConsensusAlgorithm == "SOLO"{
		if one.BlockType == 0{
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

	if config.ConsensusAlgorithm == "SHARD"{
		switch one.BlockType {
		case 1:
			var Height scanSyn.CommitteeHeight
			if err := Height.Deserialize(one.Info); nil != err {
				return err
			}
			height := uint64(Height)

			synShardBlock(height, mpb.Identify_APP_MSG_CM_BLOCK, conn)
			break
		case 2:
			var Height scanSyn.FinalHeight
			if err := Height.Deserialize(one.Info); nil != err {
				return err
			}
			height := uint64(Height)

			synShardBlock(height, mpb.Identify_APP_MSG_FINAL_BLOCK, conn)
			break
		case 3:
			break
		case 4:
			var Height scanSyn.ViewChangeHeight
			if err := Height.Deserialize(one.Info); nil != err {
				return err
			}
			height := uint64(Height)

			synShardBlock(height, mpb.Identify_APP_MSG_VC_BLOCK, conn)
			break
		default:
		}
	}

	return nil
}

func synShardBlock(height uint64, typ mpb.Identify, conn net.Conn) error{
	block, _, err := CoreLedger.GetLastShardBlock(config.ChainHash, typ)
	if nil != err {
		log.Error("GetLastShardBlock error: ", err)
	}

	for height < block.GetHeight(){
		height++

		block, _, err := CoreLedger.GetShardBlockByHeight(config.ChainHash, typ, height, 0)
		if nil != err {
			log.Error("GetTxBlockByHeight error: ", err)
			continue
		}

		if err := send_message(info.ShardBlock, conn, block); nil != err {
			log.Error("send_message error: ", err)
		}

		if mpb.Identify_APP_MSG_FINAL_BLOCK == typ {
			data, err := block.Serialize()
			if nil != err {
				continue
			}

			final := new(shard.FinalBlock)
			err = final.Deserialize(data)
			if nil != err {
				continue
			}

			if len(final.MinorBlocks) > 0 {
				for _, v := range final.MinorBlocks{
					minorblock, _, err := CoreLedger.GetShardBlockByHash(config.ChainHash, mpb.Identify_APP_MSG_MINOR_BLOCK, v.Hash(), true)
					if nil != err {
						log.Error("GetShardBlockByHash error: ", err)
						continue
					}

					if err := send_message(info.ShardBlock, conn, minorblock); nil != err{
						log.Error("send_message error: ", err)
					}
				}
			}
		}
	}

	return nil
}

func send_message(oneType info.NotifyType, conn net.Conn, message info.NotifyInfo) error{
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
