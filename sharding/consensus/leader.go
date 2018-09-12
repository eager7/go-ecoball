package consensus

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

func (c *Consensus) StartBlockConsensusLeader(instance sc.ConsensusInstance) {
	c.view = instance.GetCsView()
	c.instance = instance
	c.round = RoundPrePare

	c.sendPrepare()
}

func (c *Consensus) sendPrepare() {
	c.sendCsPacket()
}

func (c *Consensus) prepareRsp(csp *sc.CsPacket) {
	if c.round != RoundPrePare {
		log.Error("wrong consensus status")
		return
	}

	counter := c.instance.PrepareRsp(csp)
	if c.isVoteEnough(counter) {
		c.sendPreCommit()
	}
}

func (c *Consensus) sendPreCommit() {
	c.sendCsPacket()
}

func (c *Consensus) precommitRsp(csp *sc.CsPacket) {
	if c.round != RoundPrePare {
		log.Error("wrong consensus status")
		return
	}

	counter := c.instance.PrecommitRsp(csp)
	if c.isVoteEnough(counter) {
		c.sendCommit()
	}
}

func (c *Consensus) sendCommit() {
	c.sendCsPacket()

	//c.completeCb(c.instance.get)

}
