package consensus

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

func (c *Consensus) startBlockConsensusVoter(instance sc.ConsensusInstance) {
	c.view = instance.GetCsView()
	c.instance = instance
	c.round = RoundPrePare
}

func (c *Consensus) processPacketByVoter(csp *sc.CsPacket) {
	if c.round == RoundPrePare {
		if csp.Round == RoundPrePare {
			c.processPrepare(csp)
		} else if csp.Round == RoundPreCommit {
			c.processPrecommit(csp)
		} else if csp.Round == RoundCommit {
			c.processCommit(csp)
		} else {
			log.Error("prepare receive packet round ", csp.Round)
		}
	} else if c.round == RoundPreCommit {
		if csp.Round == RoundPreCommit {
			c.processPrecommit(csp)
		} else if csp.Round == RoundCommit {
			c.processCommit(csp)
		} else {
			log.Error("precommit receive packet round ", csp.Round)
		}
	} else {
		log.Error("voter round error ", c.round)
		panic("voter round error ")
	}
}

func (c *Consensus) processPrepare(csp *sc.CsPacket) {
	c.instance.UpdateBlock(csp)
	c.sendPrepareRsp()
}

func (c *Consensus) sendPrepareRsp() {
	c.sendCsPacket()
}

func (c *Consensus) processPrecommit(csp *sc.CsPacket) {
	c.instance.UpdateBlock(csp)
	c.round = RoundPreCommit

	c.sendPrecommitRsp(csp)
}

func (c *Consensus) sendPrecommitRsp(csp *sc.CsPacket) {
	c.sendCsPacket()
}

func (c *Consensus) processCommit(csp *sc.CsPacket) {
	c.instance.UpdateBlock(csp)
	c.round = RoundPreCommit

	c.csComplete()

}
