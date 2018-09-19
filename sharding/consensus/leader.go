package consensus

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

func (c *Consensus) startBlockConsensusLeader(instance sc.ConsensusInstance) {
	c.view = instance.GetCsView()
	log.Debug("currenet view ", c.view.EpochNo, " ", c.view.FinalHeight, " ", c.view.MinorHeight, " ", c.view.Round)
	c.instance = instance

	c.sendPrepare()
}

func (c *Consensus) sendPrepare() {
	log.Debug("send prepare")
	c.step = StepPrePare
	c.sendCsPacket()
	c.retransTimer(true)
}

func (c *Consensus) prepareRsp(csp *sc.CsPacket) {
	log.Debug("prepare response")
	counter := c.instance.PrepareRsp()
	if c.isVoteEnough(counter) {
		c.sendPreCommit()
	}
}

func (c *Consensus) sendPreCommit() {
	log.Debug("send precommit")
	c.step = StepPreCommit
	c.sendCsPacket()
	c.retransTimer(true)
}

func (c *Consensus) precommitRsp(csp *sc.CsPacket) {
	log.Debug("precommit response")
	counter := c.instance.PrecommitRsp()
	if c.isVoteEnough(counter) {
		c.retransTimer(false)
		c.sendCommit()
	}
}

func (c *Consensus) sendCommit() {
	log.Debug("send commit")
	c.step = StepCommit
	c.sendCsPacket()

	c.csComplete()
}

func (c *Consensus) processPacketByLeader(csp *sc.CsPacket) {
	if c.step != csp.Step {
		log.Error("packet round error leader round ", c.step, " packet round ", csp.Step)
		return
	}

	switch c.step {
	case StepPrePare:
		c.prepareRsp(csp)
	case StepPreCommit:
		c.precommitRsp(csp)
	case StepCommit:
		log.Error("leader didn't need recevie commit message")
	default:
		log.Error("leader round error ", c.step)
		panic("leader round error ")
	}
}
