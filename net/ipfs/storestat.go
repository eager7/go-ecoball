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

package ipfs

import (
	"time"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/corerepo"
	"github.com/ecoball/go-ecoball/net/dispatcher"
	"github.com/ecoball/go-ecoball/net/util"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/core/types"
	dag "gx/ipfs/QmRy4Qk9hbgFX9NGJRm8rBThrA8PZhNCitMgeRYyZ67s59/go-merkledag"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	"gx/ipfs/QmVzK524a2VWLqyvtBeiHKsUAWYgeAk4DBeZoY7vpNPNRx/go-block-format"
	"gx/ipfs/QmYKNMEUK7nCVAefgXF1LVtZEZg3uRmBqiae4FJRXDNAyJ/go-path"
	"github.com/ecoball/go-ecoball/net/ipfs/ipld"
)

const (
	randomBlksDutyShift = 4  //2^4  ???
)

type StoreStatMonitor struct {
	Interval   time.Duration
	msgbuf     <-chan interface{}
}

func NewStoreStatMonitor() (*StoreStatMonitor) {
	return &StoreStatMonitor{
		Interval: time.Hour * 24,
	}
}

func (pe *StoreStatMonitor) Start() {
	chn, err := dispatcher.Subscribe(message.APP_MSG_STORE_STAT)
	if err != nil {
		log.Error(err)
		return
	}

	pe.msgbuf = chn

	go pe.run()
}

func (pe *StoreStatMonitor) run() {
	if pe.msgbuf == nil {
		log.Error("subscribe message before running dispatcher")
		return
	}

	timer := time.NewTimer(pe.Interval)
	for {
		select {
		case <- timer.C:
			go pe.collectStoreStat()
			timer.Reset(pe.Interval)
		case msg := <- pe.msgbuf:
			// handle the message
			ecomsg, ok := msg.(message.EcoBallNetMsg)
			if ok {
				if ecomsg.Type() == message.APP_MSG_STORE_STAT {
					go pe.handlemsg(ecomsg)
				}
			} else {
				log.Error("receive an invalid message")
			}
		}
	}
}

func (pe *StoreStatMonitor) handlemsg(msg message.EcoBallNetMsg) {
	repoStat := new(types.StoreRepoStat)
	repoStat.Deserialize(msg.Data())

	// verify the rand blocks
	for _, info := range repoStat.RandBlkInfo {
		p := path.Path(info.Cid.String())
		dn, err := core.Resolve(ipfsCtrl.IpfsNode.Context(), ipfsCtrl.IpfsNode.Namesys, ipfsCtrl.IpfsNode.Resolver, p)
		if err != nil {
			log.Error("error for resolving node ", p, err)
			return
		}

		rawSize := uint64(len(dn.RawData()))
		if info.RawSize != rawSize {
			log.Error("error for verifing node ", p)
			return
		}
	}

	// try to get the back blocks
	badBlkSize := uint64(0)
	for _, cid := range repoStat.BadBlocks {
		p := path.Path(cid.String())
		dn, err := core.Resolve(ipfsCtrl.IpfsNode.Context(), ipfsCtrl.IpfsNode.Namesys, ipfsCtrl.IpfsNode.Resolver, p)
		if err != nil {
			log.Error("error for resolving node ", p, err)
			return
		}

		switch dn := dn.(type) {
		case *dag.ProtoNode:
			size, err := dn.Size()
			if err != nil {
				log.Error("error for getting node size: ", err)
				return
			}
			badBlkSize += size
		case *ipldecoball.EcoballRawData:
			badBlkSize += uint64(len(dn.RawData()))
		case *ipldecoball.EcoballShard:
			badBlkSize += uint64(len(dn.RawData()))
		default:
			log.Error("unknow node type...")
			return
		}
	}

	// yeah, get here for getting store incentives
	// TOD
	log.Debug("receive a store repo stat messageg from ", repoStat.Peer.String())
}

func (pe *StoreStatMonitor) collectStoreStat() {
	cctx := ipfsCtrl.IpfsNode.Context()

	sizeStat, err := corerepo.RepoSize(cctx, ipfsCtrl.IpfsNode)
	if err != nil {
		log.Error("error for getting ipfs repo size:", err)
		return
	}

	bs := ipfsCtrl.IpfsNode.Blockstore
	allKeys, err := bs.AllKeysChan(cctx)
	if err != nil {
		log.Error("error for getting local keys:", err)
		return
	}

	numObjects := uint64(0)
	var badCids []*cid.Cid
	var blks []blocks.Block
	// Verify all the blocks via the get operation
	for cid := range allKeys {
		blk, err := bs.Get(cid)
		if err != nil {
			badCids = append(badCids, cid)
		} else {
			blks = append(blks, blk)
		}
		numObjects++

	}

	randObject := numObjects >> randomBlksDutyShift
	if randObject == 0 {
		randObject = 1
	}
	indices := util.GetRandomIndices(int(randObject), int(numObjects-1))
	randBlks := make([]*types.ShardInfo, len(indices))
	for i, j := range indices {
		blk := blks[j]
		blkInfo := &types.ShardInfo{
			blk.Cid(),
			uint64(len(blk.RawData())),
		}
		randBlks[i] = blkInfo
	}

	peer, err := dispatcher.GetPeerID()
	if err != nil {
		log.Error(err)
		return
	}
	repoStat := types.StoreRepoStat{
		Peer:        peer,
		ChainID:     1,
		RepoSize:    sizeStat.RepoSize,
		StorageMax:  sizeStat.StorageMax,
		NumObjects:  numObjects,
		RandBlkInfo: randBlks,
		BadBlocks:   badCids,
	}

	pe.sendRepoStatMsg(repoStat)
}

func (pe *StoreStatMonitor) sendRepoStatMsg(repoStat types.StoreRepoStat) {
	data, err:= repoStat.Serialize()
	if err != nil {
		log.Error("error for serializing store stat message")
		return
	}

	netMsg := message.New(message.APP_MSG_STORE_STAT, data)
	dispatcher.BroadcastMessage(netMsg)

	log.Debug("broadcast a store stat message")
}