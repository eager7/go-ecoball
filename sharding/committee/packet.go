package committee

import (
	"github.com/ecoball/go-ecoball/common/etime"
	cs "github.com/ecoball/go-ecoball/core/shard"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/net"
	"time"
)

func (c *committee) verifyPacket(p *sc.NetPacket) {
	log.Debug("verify packet ", p.BlockType)
	if p.PacketType == netmsg.APP_MSG_CONSENSUS_PACKET {
		c.verifyConsensusPacket(p)
	} else if p.PacketType == netmsg.APP_MSG_SHARDING_PACKET {
		c.verifyShardingPacket(p)
	} else {
		log.Error("wrong packet type ", p.PacketType)
		return
	}
}

func (c *committee) verifyConsensusPacket(p *sc.NetPacket) {
	if p.Step >= consensus.StepNIL || p.Step < consensus.StepPrePare {
		log.Error("wrong step ", p.Step)
		return
	}

	var csp *sc.CsPacket

	if p.BlockType == sc.SD_CM_BLOCK {
		csp = c.ns.VerifyCmPacket(p)
	} else if p.BlockType == sc.SD_FINAL_BLOCK {
		csp = c.ns.VerifyFinalPacket(p)
	} else if p.BlockType == sc.SD_VIEWCHANGE_BLOCK {
		csp = c.ns.VerifyViewChangePacket(p)
	} else {
		log.Error("wrong block type")
		return
	}

	if csp != nil {
		c.ppc <- csp
	}
}

func (c *committee) verifyShardingPacket(p *sc.NetPacket) {
	var csp *sc.CsPacket

	if p.BlockType == sc.SD_MINOR_BLOCK {
		csp = c.ns.VerifyMinorPacket(p)
	} else {
		log.Error("wrong block type")
		return
	}

	if csp != nil {
		c.ppc <- csp
	}
}

func (c *committee) setRetransTimer(bStart bool) {
	etime.StopTime(c.retransTimer)

	if bStart {
		c.retransTimer.Reset(sc.DefaultRetransTimer * time.Second)
	}
}

func (c *committee) processRetransTimeout() {
	c.cs.ProcessRetransPacket()
}

func (c *committee) consensusCb(bl interface{}) {
	switch blockType := bl.(type) {
	case *cs.CMBlock:
		c.commitCmBlock(bl.(*cs.CMBlock))
	case *cs.FinalBlock:
		c.commitFinalBlock(bl.(*cs.FinalBlock))
	case *cs.ViewChangeBlock:
		c.commitViewchangeBlock(bl.(*cs.ViewChangeBlock))
	default:
		log.Error("consensus call back wrong packet type ", blockType)
	}
}

func (c *committee) processShardBlockOnWaitStatus(p interface{}) {
	csp := p.(*sc.CsPacket)

	if csp.BlockType != sc.SD_MINOR_BLOCK {
		log.Error("block type error ", csp.BlockType)
		return
	}

	minor := csp.Packet.(*cs.MinorBlock)
	if !c.ns.SaveMinorBlockToPool(minor) {
		return
	}

	net.Np.TransitBlock(csp)

	if c.ns.IsMinorBlockEnoughInPool() {
		c.fsm.Execute(ActProductFinalBlock, nil)
	}
}
