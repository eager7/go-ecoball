package consensus

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

func (c *Consensus) startBlockConsensusLeader(instance sc.ConsensusInstance) {
	c.view = instance.GetCsView()
	c.instance = instance

	c.sendPrepare()
}

func (c *Consensus) sendPrepare() {
	c.round = RoundPrePare
	c.sendCsPacket()
}

func (c *Consensus) prepareRsp(csp *sc.CsPacket) {
	counter := c.instance.PrepareRsp()
	if c.isVoteEnough(counter) {
		c.sendPreCommit()
	}
}

func (c *Consensus) sendPreCommit() {
	c.round = RoundPreCommit
	c.sendCsPacket()
}

func (c *Consensus) precommitRsp(csp *sc.CsPacket) {
	counter := c.instance.PrecommitRsp()
	if c.isVoteEnough(counter) {
		c.sendCommit()
	}
}

func (c *Consensus) sendCommit() {
	c.round = RoundCommit
	c.sendCsPacket()

	c.csComplete()
}

func (c *Consensus) processPacketByLeader(csp *sc.CsPacket) {
	if c.round != csp.Round {
		log.Error("packet round error leader round ", c.round, " packet round ", csp.Round)
		return
	}

	switch c.round {
	case RoundPrePare:
		c.prepareRsp(csp)
	case RoundPreCommit:
		c.precommitRsp(csp)
	case RoundCommit:
		log.Error("leader didn't need recevie commit message")
	default:
		log.Error("leader round error ", c.round)
		panic("leader round error ")
	}
}
