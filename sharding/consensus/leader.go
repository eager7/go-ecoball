package consensus

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"time"
)

func (c *Consensus) startBlockConsensusLeader(instance sc.ConsensusInstance, d time.Duration) {
	c.view = instance.GetCsView()
	log.Debug("currenet view ", c.view.EpochNo, " ", c.view.FinalHeight, " ", c.view.MinorHeight, " ", c.view.Round)
	c.instance = instance

	c.sendPrepare(d)
}

func (c *Consensus) sendPrepare(d time.Duration) {
	log.Debug("send prepare")
	c.step = StepPrePare
	//packet := c.instance.MakeNetPacket(c.step)
	//c.sendCsPacket(packet)
	c.rcb(true, d)
}

func (c *Consensus) prepareRsp(csp *sc.CsPacket) {
	log.Debug("prepare response")
	counter := c.instance.PrepareRsp()
	if c.isVoteFull(counter) {
		c.fcb(false)
		c.sendPreCommit()
	} else if c.isVoteOnThreshold(counter) {
		c.fcb(true)
	}
}

func (c *Consensus) sendPreCommit() {
	log.Debug("send precommit")
	c.step = StepPreCommit
	c.rcb(true, sc.DefaultRetransTimer*time.Second)

	packet := c.instance.MakeNetPacket(c.step)
	c.sendCsPacket(packet)

}

func (c *Consensus) precommitRsp(csp *sc.CsPacket) {
	log.Debug("precommit response")
	counter := c.instance.PrecommitRsp()
	if c.isVoteFull(counter) {
		c.fcb(false)
		c.sendCommit()
	} else if c.isVoteOnThreshold(counter) {
		c.fcb(true)
	}
}

func (c *Consensus) sendCommit() {
	log.Debug("send commit")
	c.step = StepCommit
	c.rcb(false, sc.DefaultBlockWindow)

	packet := c.instance.MakeNetPacket(c.step)
	//we need save cm block before we send it to peer because the shards is change
	if packet.BlockType == sc.SD_CM_BLOCK {
		c.csComplete()
		c.sendCsPacket(packet)
	} else {
		c.sendCsPacket(packet)
		c.csComplete()
	}
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
