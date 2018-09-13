package consensus

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

func (c *Consensus) startBlockConsensusVoter(instance sc.ConsensusInstance) {
	c.view = instance.GetCsView()
	c.instance = instance
	c.step = RoundPrePare
}

func (c *Consensus) processPacketByVoter(csp *sc.CsPacket) {
	if c.step == RoundPrePare {
		if csp.Round == RoundPrePare {
			c.processPrepare(csp)
		} else if csp.Round == RoundPreCommit {
			c.processPrecommit(csp)
		} else if csp.Round == RoundCommit {
			c.processCommit(csp)
		} else {
			log.Error("prepare receive packet round ", csp.Round)
		}
	} else if c.step == RoundPreCommit {
		if csp.Round == RoundPreCommit {
			c.processPrecommit(csp)
		} else if csp.Round == RoundCommit {
			c.processCommit(csp)
		} else {
			log.Error("precommit receive packet round ", csp.Round)
		}
	} else {
		log.Error("voter round error ", c.step)
		panic("voter round error ")
	}
}

func (c *Consensus) processPrepare(csp *sc.CsPacket) {
	c.instance.UpdateBlock(csp)
	c.sendPrepareRsp()
}

func (c *Consensus) sendPrepareRsp() {
	c.sendCsRspPacket()
}

func (c *Consensus) processPrecommit(csp *sc.CsPacket) {
	c.instance.UpdateBlock(csp)
	c.step = RoundPreCommit

	c.sendPrecommitRsp(csp)
}

func (c *Consensus) sendPrecommitRsp(csp *sc.CsPacket) {
	c.sendCsRspPacket()
}

func (c *Consensus) processCommit(csp *sc.CsPacket) {
	c.instance.UpdateBlock(csp)
	c.step = RoundPreCommit

	c.csComplete()

}
