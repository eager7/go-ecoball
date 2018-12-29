package net

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/event"
	cm "github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/net/network"
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
	n  network.EcoballNetwork
}

var Np *net

func MakeNet(ns *cell.Cell, n network.EcoballNetwork) {
	Np = &net{ns: ns, n: n}
	return
}

func (n *net) SendToPeer(packet *sc.NetPacket, worker *sc.Worker) {
	log.Debug("send to peer")

	if worker == nil {
		log.Error("leader is nil")
		return
	}

	n.sendto(worker.Address, worker.Port, worker.Pubkey, packet)
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
		n.sendto(peers[index].Address, peers[index].Port, peers[index].Pubkey, packet)
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

		n.sendto(work.Address, work.Port, work.Pubkey, packet)
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

func (n *net) SendSyncResponse(packet *sc.NetPacket, work *sc.WorkerId) {
	log.Debug("send sync response")
	go n.sendto(work.Address, work.Port, work.Pubkey, packet)
}

func (n *net) SendSyncMessage(packet *sc.NetPacket) {
	log.Debug("send sync message")

	works := n.ns.GetWorks()
	if works == nil {
		log.Error("works is nil")
		return
	}

	rand.Seed(time.Now().UnixNano())
	log.Info("worker size = ", len(works))
	var r int32
	if len(works) <= 1 {
		r = 0
	} else {
		r = rand.Int31n(int32(len(works) - 1))
	}

	var i int32 = 0
	for _, work := range works {
		if n.ns.Self.Equal(work) && len(works) > 1 {
			continue
		}
		if i == r {
			go n.sendto(work.Address, work.Port, work.Pubkey, packet)
			break
		}
		i++
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
			n.sendto(shard.Member[begin+i].Address, shard.Member[begin+i].Port, string(shard.Member[begin+i].PublicKey), sp)
		}
	}

	candidates := n.ns.GetCandidateWorks()
	shardSize := len(candidates)
	bSend, begin, count := CalcCrossShardIndex(si, int(selfSize), shardSize)
	if !bSend {
		return
	}

	log.Debug("send block to candidate ")
	for i := 0; i < count; i++ {
		n.sendto(candidates[begin].Address, candidates[begin].Port, string(candidates[begin].Pubkey), sp)
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
			n.sendto(cm[begin+i].Address, cm[begin+i].Port, cm[begin+i].Pubkey, sp)
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
			n.sendto(shard.Member[i+begin].Address, shard.Member[i+begin].Port, string(shard.Member[i+begin].PublicKey), sp)
		}
	}

	candidates := n.ns.GetCandidateWorks()
	shardSize := len(candidates)
	bSend, begin, count = CalcCrossShardIndex(si, int(selfSize), shardSize)
	if !bSend {
		return
	}

	log.Debug("send block to candidate ")
	for i := 0; i < count; i++ {
		n.sendto(candidates[begin].Address, candidates[begin].Port, string(candidates[begin].Pubkey), sp)
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

func (n *net) Subscribe(port string, chanSize uint16) (rcv <-chan interface{}, err error) {
	if n.n == nil {
		rcv, err = simulate.Subscribe(port, chanSize)
		if err != nil {
			log.Panic("simulate error ", err)
			return
		}
		return
	} else {
		msg := []mpb.Identify{mpb.Identify_APP_MSG_SHARDING_PACKET, mpb.Identify_APP_MSG_CONSENSUS_PACKET,
			mpb.Identify_APP_MSG_SYNC_REQUEST, mpb.Identify_APP_MSG_SYNC_RESPONSE}
		rcv, err = event.Subscribe(msg...)
		if err != nil {
			log.Error("Subscribe error ", err)
			panic("Subscribe error ")
		}
		return
	}

}

func (n *net) sendto(addr string, port string, pubKey string, packet *sc.NetPacket) error {
	if n.n == nil {
		go simulate.Sendto(addr, port, packet)
		return nil
	} else {
		/*data, err := json.Marshal(packet)
		if err != nil {
			log.Error("wrong packet")
			return err
		}*/

		log.Debug("p2p net send to peer ", addr, " port ", port, " packet type ", packet.PacketType, " block type ", packet.BlockType)

		//msg := message.New(packet.PacketType, data)
		//n.n.SendMsgToPeer(addr, port, pubKey, msg)
		event.Send(event.ActorSharding, event.ActorP2P, cm.NetPacket{
			Address:   addr,
			Port:      port,
			PublicKey: pubKey,
			Message:   nil, //TODO
		})
		return nil
	}
}

func (n *net) RecvNetMsg(msg interface{}) (packet *sc.NetPacket, err error) {
	err = nil
	if n.n == nil {
		log.Debug("recv net message ")

		packet = msg.(*sc.NetPacket)
		return
	} else {
		log.Debug("recv p2p net message ")

		emsg := msg.(message.EcoBallNetMsg)
		var np sc.NetPacket
		err = json.Unmarshal(emsg.Data(), &np)
		packet = &np
		return
	}
}
