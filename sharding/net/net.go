package net

import (
	"github.com/ecoball/go-ecoball/common/elog"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"math"
	"math/rand"
	"time"
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
	if worker == nil {
		log.Error("leader is nil")
		return
	}

	go simulate.Sendto(worker.Address, worker.Port, packet)
}

func (n *net) GossipBlock(packet *sc.NetPacket) {
	log.Debug("gossip block")

	works := n.ns.GetWorks()
	if works == nil {
		log.Error("works is nil")
		return
	}

	if len(works) < 5 {
		if len(works) > 1 {
			rand.Seed(time.Now().UnixNano())
			r := rand.Int31n(int32(len(works)))
			log.Debug("gossip to peer address ", works[r].Address, " port ", works[r].Port)
			go simulate.Sendto(works[r].Address, works[r].Port, packet)
		}
		return
	}

	var index int
	if n.ns.NodeType == sc.NodeCommittee {
		index = 0
	} else if n.ns.NodeType == sc.NodeShard {
		index = int(n.ns.CalcShardLeader(len(works), false))
	} else {
		return
	}

	var peers []*sc.Worker
	if index == 0 {
		peers = works[1:]
	} else {
		peers = append(peers, works[0:index]...)
		if index < len(works)-1 {
			peers = append(peers, works[index+1:]...)
		}
	}

	size := int(len(peers))
	number := int(math.Sqrt(float64(size)))

	arr := make([]int, 0, 3*size)
	for i := 0; i < 3; i++ {
		for j := 0; j < size; j++ {
			arr = append(arr, j)
		}
	}

	var indexs []int
	for i, peer := range peers {
		if n.ns.Self.Equal(peer) {
			indexs = append(indexs, arr[i+size-number/2:i+size]...)
			indexs = append(indexs, arr[i+size+1:i+size-number/2+number]...)
			break
		}
	}

	//
	//rand.Seed(time.Now().UnixNano())
	//var indexs []int32
	//for i := 0; i < number; i++ {
	//	r := rand.Int31n(total)
	//	if r == 0 {
	//		continue
	//	}
	//
	//	same := false
	//	for _, index := range indexs {
	//		if index == r {
	//			same = true
	//			break
	//		}
	//	}
	//
	//	if same {
	//		continue
	//	}
	//
	//	indexs = append(indexs, r)
	//}

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

func (n *net) SendBlockToShards(packet *sc.NetPacket) {
	/*only leader and backup send*/
	leader := n.ns.IsLeader()
	bakcup := n.ns.IsBackup()
	if !leader && !bakcup {
		return
	}

	log.Debug("send block to shard")

	sp := &sc.NetPacket{}
	sp.DupHeader(packet)
	sp.PacketType = pb.MsgType_APP_MSG_SHARDING_PACKET
	sp.Packet = packet.Packet

	cm := n.ns.GetLastCMBlock()

	var bfinal = false
	if sp.BlockType == sc.SD_FINAL_BLOCK {
		bfinal = true
	}

	for _, shard := range cm.Shards {
		if leader {
			i := n.ns.CalcShardLeader(len(shard.Member), bfinal)
			go simulate.Sendto(shard.Member[i].Address, shard.Member[i].Port, sp)
			if n.ns.GetWorksCounter() == 1 && len(shard.Member) > 1 {
				j := n.ns.CalcShardBackup(len(shard.Member), bfinal)
				go simulate.Sendto(shard.Member[j].Address, shard.Member[j].Port, sp)
			}
		} else if bakcup {
			if len(shard.Member) > 1 {
				i := n.ns.CalcShardBackup(len(shard.Member), bfinal)
				go simulate.Sendto(shard.Member[i].Address, shard.Member[i].Port, sp)
			}
		}
	}

}

func (n *net) SendBlockToCommittee(packet *sc.NetPacket) {
	/*only leader and backup send*/
	leader := n.ns.IsLeader()
	bakcup := n.ns.IsBackup()
	if !leader && !bakcup {
		return
	}

	log.Debug("send block to committee")

	sp := &sc.NetPacket{}
	sp.DupHeader(packet)
	sp.PacketType = pb.MsgType_APP_MSG_SHARDING_PACKET
	sp.Packet = packet.Packet

	cm := n.ns.GetCmWorks()
	if leader {
		go simulate.Sendto(cm[0].Address, cm[0].Port, sp)
	} else if bakcup {
		if len(cm) > 1 {
			go simulate.Sendto(cm[1].Address, cm[1].Port, sp)
		}
	}

	//send block to other shard
	cmb := n.ns.GetLastCMBlock()
	for i, shard := range cmb.Shards {
		if n.ns.Shardid == uint16(i+1) {
			continue
		}

		if leader {
			i := n.ns.CalcShardLeader(len(shard.Member), false)
			go simulate.Sendto(shard.Member[i].Address, shard.Member[i].Port, sp)
		} else if bakcup {
			if len(shard.Member) > 1 {
				i := n.ns.CalcShardBackup(len(shard.Member), false)
				go simulate.Sendto(shard.Member[i].Address, shard.Member[i].Port, sp)
			}
		}
	}

}

func (n *net) TransitBlock(p *sc.CsPacket) {
	log.Debug("transit block")

	leader := n.ns.IsLeader()
	bakcup := n.ns.IsBackup()
	if !leader && !bakcup {
		return
	}

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

	n.BroadcastBlock(sp)
}
