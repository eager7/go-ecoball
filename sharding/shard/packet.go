package shard

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/net"
)

func (s *shard) verifyPacket(p *sc.NetPacket) {
	log.Debug("verify packet ", p.BlockType)
	if p.PacketType == netmsg.APP_MSG_CONSENSUS_PACKET {
		s.verifyConsensusPacket(p)
	} else if p.PacketType == netmsg.APP_MSG_SHARDING_PACKET {
		s.verifyShardingPacket(p)
	} else {
		log.Error("wrong packet type")
		return
	}
}

func (s *shard) verifyConsensusPacket(p *sc.NetPacket) {
	if p.Step >= consensus.StepNIL || p.Step < consensus.StepPrePare {
		log.Error("wrong step ", p.Step)
		return
	}

	var csp *sc.CsPacket

	if p.BlockType == sc.SD_MINOR_BLOCK {
		csp = s.ns.VerifyMinorPacket(p)
	} else {
		log.Error("wrong block type")
		return
	}

	if csp != nil {
		s.ppc <- csp
	}
}

func (s *shard) verifyShardingPacket(p *sc.NetPacket) {
	var csp *sc.CsPacket

	if p.BlockType == sc.SD_CM_BLOCK {
		csp = s.ns.VerifyCmPacket(p)
	} else if p.BlockType == sc.SD_FINAL_BLOCK {
		csp = s.ns.VerifyFinalPacket(p)
	} else if p.BlockType == sc.SD_VIEWCHANGE_BLOCK {
		csp = s.ns.VerifyViewChangePacket(p)
	} else {
		log.Error("wrong block type")
		return
	}

	if csp != nil {
		s.ppc <- csp
	}
}

func (s *shard) consensusCb(bl interface{}) {
	switch blockType := bl.(type) {
	case *cs.MinorBlock:
		s.commitMinorBlock(bl.(*cs.MinorBlock))
	default:
		log.Error("consensus call back wrong packet type ", blockType)
	}
}

func (s *shard) processRetransTimeout() {
	s.cs.ProcessRetransPacket()
}

func (s *shard) processConsensusPacket(packet *sc.CsPacket) {
	s.fsm.Execute(ActRecvConsensusPacket, packet)
}

func (s *shard) processShardingPacket(p *sc.CsPacket) {
	switch p.BlockType {
	case sc.SD_CM_BLOCK:
		cm := p.Packet.(*cs.CMBlock)
		s.ns.SaveLastCMBlock(cm)
		s.broadcastShardingPacket(p)

		s.fsm.Execute(ActProductMinorBlock, nil)
	case sc.SD_FINAL_BLOCK:
		final := p.Packet.(*cs.FinalBlock)
		s.ns.SaveLastFinalBlock(final)
		s.broadcastShardingPacket(p)

		if final.Height%sc.DefaultEpochFinalBlockNumber != 0 {
			s.fsm.Execute(ActProductMinorBlock, nil)
		}
	case sc.SD_VIEWCHANGE_BLOCK:
		vc := p.Packet.(*cs.ViewChangeBlock)
		s.ns.SaveLastViewchangeBlock(vc)
		s.broadcastShardingPacket(p)
	default:
		log.Error("block type error ", p.BlockType)
		return
	}

}

func (s *shard) broadcastShardingPacket(p *sc.CsPacket) {
	net.Np.TransitBlock(p)
}
