package net

import (
	"github.com/ecoball/go-ecoball/common/elog"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"math"
	"math/rand"
	"time"
)

var (
	log = elog.NewLogger("sdnet", elog.DebugLog)
)

type net struct {
	ns *cell.Cell
}

var Np *net

func MakeNet(ns *cell.Cell) {
	Np = &net{ns: ns}
	return
}

func (n *net) SendToPeer(packet *sc.NetPacket, worker *cell.Worker) {
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

	peers := works[1:]

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
	log.Debug("send block to shard")

	/*only leader and backup send*/
	leader := n.ns.IsLeader()
	bakcup := n.ns.IsBackup()
	if !leader && !bakcup {
		return
	}

	sp := &sc.NetPacket{}
	sp.DupHeader(packet)
	sp.PacketType = netmsg.APP_MSG_SHARDING_PACKET
	sp.Packet = packet.Packet

	cm := n.ns.GetLastCMBlock()
	for _, shard := range cm.Shards {
		if leader {
			go simulate.Sendto(shard.Member[0].Address, shard.Member[0].Port, sp)
		} else if bakcup {
			if len(shard.Member) > 1 {
				go simulate.Sendto(shard.Member[1].Address, shard.Member[1].Port, sp)
			}
		}
	}

}

func (n *net) SendBlockToCommittee(packet *sc.NetPacket) {
	log.Debug("send block to committee")

	/*only leader and backup send*/
	leader := n.ns.IsLeader()
	bakcup := n.ns.IsBackup()
	if !leader && !bakcup {
		return
	}

	sp := &sc.NetPacket{}
	sp.DupHeader(packet)
	sp.PacketType = netmsg.APP_MSG_SHARDING_PACKET
	sp.Packet = packet.Packet

	cm := n.ns.GetCmWorks()
	if leader {
		go simulate.Sendto(cm[0].Address, cm[0].Port, sp)
	} else if bakcup {
		if len(cm) > 1 {
			go simulate.Sendto(cm[1].Address, cm[1].Port, sp)
		}
	}

}
