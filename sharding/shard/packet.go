package shard

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/net"
	"github.com/ecoball/go-ecoball/sharding/simulate"
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
	} else if p.BlockType == sc.SD_MINOR_BLOCK {
		csp = s.ns.VerifyMinorPacket(p)
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

func (s *shard) recvConsensusPacket(packet *sc.CsPacket) {
	s.fsm.Execute(ActRecvConsensusPacket, packet)
}

func (s *shard) recvShardingPacket(packet *sc.CsPacket) {
	s.fsm.Execute(ActRecvShardingPacket, packet)
}

func (s *shard) processShardingPacket(csp interface{}) {
	p := csp.(*sc.CsPacket)
	switch p.BlockType {
	case sc.SD_CM_BLOCK:
		cm := p.Packet.(*cs.CMBlock)
		last := s.ns.GetLastCMBlock()
		if last != nil {
			if last.Height >= cm.Height {
				log.Debug("old cm packet ", cm.Height)
				return
			}
		}

		simulate.TellBlock(cm)
		s.ns.SaveLastCMBlock(cm)
		s.broadcastCommitteePacket(p)

		s.fsm.Execute(ActProductMinorBlock, nil)
	case sc.SD_FINAL_BLOCK:
		final := p.Packet.(*cs.FinalBlock)
		last := s.ns.GetLastFinalBlock()
		if last != nil {
			if last.Height >= final.Height {
				log.Debug("old final packet ", final.Height)
				return
			}
		}

		simulate.TellBlock(final)
		s.ns.SaveLastFinalBlock(final)
		s.broadcastCommitteePacket(p)

		if final.Height%sc.DefaultEpochFinalBlockNumber != 0 {
			s.fsm.Execute(ActProductMinorBlock, nil)
		}
	case sc.SD_VIEWCHANGE_BLOCK:
		vc := p.Packet.(*cs.ViewChangeBlock)
		last := s.ns.GetLastViewchangeBlock()
		if last != nil {
			if last.FinalBlockHeight > vc.FinalBlockHeight ||
				(last.FinalBlockHeight == vc.FinalBlockHeight ||
					last.Round >= vc.Round) {
				log.Debug("old vc packet ", vc.FinalBlockHeight, " ", vc.Round)
				return
			}
		}

		simulate.TellBlock(vc)
		s.ns.SaveLastViewchangeBlock(vc)
		s.broadcastCommitteePacket(p)
	case sc.SD_MINOR_BLOCK:
		minor := p.Packet.(*cs.MinorBlock)
		if !s.ns.SaveMinorBlockToPool(minor) {
			return
		}
		simulate.TellBlock(minor)
		net.Np.TransitBlock(p)

	default:
		log.Error("block type error ", p.BlockType)
		return
	}

}

func (s *shard) broadcastCommitteePacket(p *sc.CsPacket) {
	net.Np.TransitBlock(p)
}
