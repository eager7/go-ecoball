package shard

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/core/types"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
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
	if p.BlockType == sc.SD_CM_BLOCK {
		s.verifyCmBlock(p)
	} else if p.BlockType == sc.SD_FINAL_BLOCK {
		s.verifyFinalBlock(p)
	} else if p.BlockType == sc.SD_VIEWCHANGE_BLOCK {
		s.verifyViewChangeBlock(p)
	} else {
		log.Error("wrong block type")
		return
	}
}

func (s *shard) verifyMinorBlock(p *sc.NetPacket) {
	var minor types.MinorBlock
	err := json.Unmarshal(p.Packet, &minor)
	if err != nil {
		log.Error("minor block unmarshal error ", err)
		return
	}

	last := s.ns.GetLastCMBlock()
	if last == nil {
		log.Debug("wait cm packet")
		return
	}

	if last.Height != minor.CMEpochNo {
		log.Debug("old cm packet")
		return
	}

	/*missing_func need verify signature here*/

	var csp sc.CsPacket
	(&csp).Copyhead(p)
	(&csp).Packet = &minor

	s.ppc <- &csp
}

func (s *shard) verifyCmBlock(p *sc.NetPacket) {
}

func (s *shard) verifyFinalBlock(p *sc.NetPacket) {
}

func (s *shard) verifyViewChangeBlock(p *sc.NetPacket) {
}

func (s *shard) consensusCb(bl interface{}) {
	switch blockType := bl.(type) {
	case *types.MinorBlock:
		s.recvCommitMinorBlock(bl.(*types.MinorBlock))
	default:
		log.Error("consensus call back wrong packet type ", blockType)
	}
}

func (s *shard) processPacket(packet *sc.CsPacket) {
	switch packet.PacketType {
	case netmsg.APP_MSG_CONSENSUS_PACKET:
		s.processConsensusPacket(packet)
	case netmsg.APP_MSG_SHARDING_PACKET:
		s.processShardingPacket(packet)
	default:
		log.Error("wrong packet")
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
		cm := p.Packet.(*types.CMBlock)
		s.ns.SaveLastCMBlock(cm)
		s.fsm.Execute(ActProductMinorBlock, nil)
	case sc.SD_FINAL_BLOCK:
		final := p.Packet.(*types.FinalBlock)
		s.ns.SaveLastFinalBlock(final)
		if final.Height%sc.DefaultEpochFinalBlockNumber != 0 {
			s.fsm.Execute(ActProductMinorBlock, nil)
		}
	case sc.SD_VIEWCHANGE_BLOCK:
		vc := p.Packet.(*types.ViewChangeBlock)
		s.ns.SaveLastViewchangeBlock(vc)
	default:
		log.Error("block type error ", p.BlockType)
		return
	}
}
