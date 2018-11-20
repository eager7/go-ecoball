package net

import (
	"github.com/ecoball/go-ecoball/common/elog"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"math"
)

var (
	log = elog.NewLogger("sharding", elog.DebugLog)
)

type net struct {
	ns *cell.Cell
}

var Np *net

func MakeNet(ns *cell.Cell) {
	Np = &net{ns: ns}
	return
}

func (n *net) SendToPeer(packet *sc.NetPacket, worker *sc.Worker) {
	log.Debug("send to peer")

	if worker == nil {
		log.Error("leader is nil")
		return
	}

	go simulate.Sendto(worker.Address, worker.Port, packet)
}

func CalcGossipIndex(size int, i int) (indexs []int) {
	if size < 2 || size <= i {
		return
	}

	number := int(math.Sqrt(float64(size)))

	arr := make([]int, 0, 3*size)
	for i := 0; i < 3; i++ {
		for j := 0; j < size; j++ {
			arr = append(arr, j)
		}
	}

	indexs = append(indexs, arr[size+i+1:i+size+1+number]...)

	return
}

func (n *net) GossipBlock(packet *sc.NetPacket) {
	log.Debug("gossip block")

	works := n.ns.GetWorks()
	if works == nil {
		log.Error("works is nil")
		return
	}

	if len(works) < 2 {
		return
	}

	var peers []*sc.Worker
	if packet.PacketType == pb.MsgType_APP_MSG_CONSENSUS_PACKET {
		var index int
		if n.ns.NodeType == sc.NodeCommittee {
			index = 0
		} else if n.ns.NodeType == sc.NodeShard {
			index = int(n.ns.CalcShardLeader(len(works)))
		} else {
			return
		}

		if index == 0 {
			peers = works[1:]
		} else {
			peers = append(peers, works[0:index]...)
			if index < len(works)-1 {
				peers = append(peers, works[index+1:]...)
			}
		}
	} else {
		peers = works
	}

	size := int(len(peers))
	pos := 0
	for i, peer := range peers {
		if n.ns.Self.Equal(peer) {
			pos = i
			break
		}
	}

	indexs := CalcGossipIndex(size, pos)
	for _, index := range indexs {
		log.Debug("gossip to peer address ", peers[index].Address, " port ", peers[index].Port)
		go simulate.Sendto(peers[index].Address, peers[index].Port, packet)
	}
}

func (n *net) BroadcastBlock(packet *sc.NetPacket) {
	log.Debug("broadcast block")

	works := n.ns.GetWorks()
	if works == nil {
		log.Error("works is nil")
		return
	}

	for _, work := range works {
		if n.ns.Self.Equal(work) {
			continue
		}

		go simulate.Sendto(work.Address, work.Port, packet)
	}
}

func CalcCrossShardIndex(si int, ourSize int, shardSize int) (bSend bool, begin int, count int) {
	bSend = false

	if ourSize >= shardSize {
		if si >= shardSize {
			bSend = false
			return
		} else {
			bSend = true
			begin = si
			count = 1
			return
		}
	} else {
		cover := shardSize / ourSize
		part := shardSize % ourSize

		if si < part {
			begin = (cover + 1) * si
			count = cover + 1
		} else {
			begin = cover*si + part
			count = cover
		}

		return true, begin, count
	}

}

func (n *net) SendBlockToShards(packet *sc.NetPacket) {
	si := n.ns.SelfIndex()
	selfSize := n.ns.GetWorksCounter()

	sp := &sc.NetPacket{}
	sp.DupHeader(packet)
	sp.PacketType = pb.MsgType_APP_MSG_SHARDING_PACKET
	sp.Packet = packet.Packet

	cm := n.ns.GetLastCMBlock()

	for j, shard := range cm.Shards {
		shardSize := len(shard.Member)
		bSend, begin, count := CalcCrossShardIndex(si, int(selfSize), shardSize)
		if !bSend {
			return
		}

		log.Debug("send block to shard ", j+1)

		for i := 0; i < count; i++ {
			go simulate.Sendto(shard.Member[begin+i].Address, shard.Member[begin+i].Port, sp)
		}

	}

}

func (n *net) SendBlockToCommittee(packet *sc.NetPacket) {
	sp := &sc.NetPacket{}
	sp.DupHeader(packet)
	sp.PacketType = pb.MsgType_APP_MSG_SHARDING_PACKET
	sp.Packet = packet.Packet

	si := n.ns.SelfIndex()
	selfSize := n.ns.GetWorksCounter()

	cmSize := n.ns.GetCmWorksCounter()
	bSend, begin, count := CalcCrossShardIndex(si, int(selfSize), cmSize)
	if bSend {
		log.Debug("send block to committee")
		cm := n.ns.GetCmWorks()
		for i := 0; i < count; i++ {
			go simulate.Sendto(cm[begin+i].Address, cm[begin+i].Port, sp)
		}
	}

	//send block to other shard
	cmb := n.ns.GetLastCMBlock()
	for i, shard := range cmb.Shards {
		if n.ns.Shardid == uint16(i+1) {
			continue
		}

		shardSize := len(shard.Member)
		bSend, begin, count = CalcCrossShardIndex(si, int(selfSize), shardSize)
		if !bSend {
			continue
		}

		log.Debug("send block other shard, id:  ", i+1)
		for i := 0; i < count; i++ {
			go simulate.Sendto(shard.Member[i+begin].Address, shard.Member[i+begin].Port, sp)
		}
	}

}

func (n *net) TransitBlock(p *sc.CsPacket) {
	leader := n.ns.IsLeader()

	log.Debug("transit block")

	sp := &sc.NetPacket{}
	sp.CopyHeader(p)

	switch p.Packet.(type) {
	case *cs.CMBlock:
		cm := p.Packet.(*cs.CMBlock)
		packet, err := cm.Serialize()
		if err != nil {
			log.Error("transit cm block packet Serialize  error ", err)
			return
		}
		sp.Packet = packet
	case *cs.FinalBlock:
		final := p.Packet.(*cs.FinalBlock)
		packet, err := final.Serialize()
		if err != nil {
			log.Error("transit final block packet Serialize error ", err)
			return
		}
		sp.Packet = packet
	case *cs.MinorBlock:
		minor := p.Packet.(*cs.MinorBlock)
		packet, err := minor.Serialize()
		if err != nil {
			log.Error("transit minor block packet Serialize error ", err)
			return
		}
		sp.Packet = packet
	case *cs.ViewChangeBlock:
		vc := p.Packet.(*cs.ViewChangeBlock)
		packet, err := vc.Serialize()
		if err != nil {
			log.Error("transit block packet Marshal error ", err)
			return
		}
		sp.Packet = packet
	default:
		log.Error("transit block wrong block type ")
		return
	}

	if leader {
		n.BroadcastBlock(sp)
	} else {
		n.GossipBlock(sp)
	}
}
