package shard

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/net/message/pb"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/net"
	"github.com/ecoball/go-ecoball/sharding/simulate"
)

func (s *shard) verifyPacket(p *sc.NetPacket) {
	log.Debug("verify packet ", p.BlockType)
	if p.PacketType == pb.MsgType_APP_MSG_CONSENSUS_PACKET {
		s.verifyConsensusPacket(p)
	} else if p.PacketType == pb.MsgType_APP_MSG_SHARDING_PACKET {
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
			} else if cm.Height > last.Height+1 {
				log.Debug("cm block last ", last.Height, " recv ", cm.Height, " need sync")
				s.fsm.Execute(ActChainNotSync, nil)
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
			} else if final.Height > last.Height+1 {
				log.Debug("final block last ", last.Height, " recv ", final.Height, " need sync")
				s.fsm.Execute(ActChainNotSync, nil)
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
		lastfinal := s.ns.GetLastFinalBlock()
		if lastfinal == nil || last == nil {
			panic("last block is nil")
			return
		}

		if last.Height >= vc.Height {
			log.Debug("old vc packet height ", vc.Height, " last height ", last.Height)
			return
		} else if vc.Height > last.Height+1 {
			log.Debug("vc packet height ", vc.Height, " last height ", last.Height, " need sync")
			s.fsm.Execute(ActChainNotSync, nil)
		}

		if vc.FinalBlockHeight < lastfinal.Height {
			log.Debug("wrong vc packet final height ", vc.FinalBlockHeight, " last final height", lastfinal.Height)
			return
		} else if vc.FinalBlockHeight == lastfinal.Height {
			if last.FinalBlockHeight == vc.FinalBlockHeight {
				if last.Round >= vc.Round {
					log.Debug("old vc packet vc round ", vc.Round, " last round ", last.Round)
					return
				} else if vc.Round > last.Round+1 {
					log.Debug("vc round ", vc.Round, " last round ", last.Round, " need sync")
					s.fsm.Execute(ActChainNotSync, nil)
					return
				}
			} else {
				if vc.Round > 1 {
					log.Debug("vc round ", vc.Round, " need sync")
					s.fsm.Execute(ActChainNotSync, nil)
					return
				} else if vc.Round < 1 {
					log.Debug("wrong round ", vc.Round)
					return
				}
			}
		} else {
			log.Debug("last final ", lastfinal.Height, " recv view change final height ", vc.FinalBlockHeight, " need sync")
			s.fsm.Execute(ActChainNotSync, nil)
			return
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
