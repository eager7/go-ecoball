package consensus

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/net"
)

func (c *Consensus) startBlockConsensusVoter(instance sc.ConsensusInstance) {
	c.view = instance.GetCsView()
	log.Debug("currenet view ", c.view.EpochNo, " ", c.view.FinalHeight, " ", c.view.MinorHeight, " ", c.view.Round)

	c.instance = instance
	c.step = StepPrePare
}

func (c *Consensus) processPacketByVoter(csp *sc.CsPacket) {
	if c.step == StepPrePare {
		if csp.Step == StepPrePare {
			c.processPrepare(csp)
		} else if csp.Step == StepPreCommit {
			c.processPrecommit(csp)
		} else if csp.Step == StepCommit {
			c.processCommit(csp)
		} else {
			log.Error("prepare receive packet round ", csp.Step)
		}
	} else if c.step == StepPreCommit {
		if csp.Step == StepPreCommit {
			c.processPrecommit(csp)
		} else if csp.Step == StepCommit {
			c.processCommit(csp)
		} else {
			log.Error("precommit receive packet round ", csp.Step)
		}
	} else {
		log.Error("voter round error ", c.step)
		panic("voter round error ")
	}
}

func (c *Consensus) processPrepare(csp *sc.CsPacket) {
	log.Debug("process prepare")
	c.sendPrepareRsp()
}

func (c *Consensus) sendPrepareRsp() {
	log.Debug("send prepare response")
	c.sendCsRspPacket()
}

func (c *Consensus) processPrecommit(csp *sc.CsPacket) {
	log.Debug("process precommit")
	c.step = StepPreCommit

	c.sendPrecommitRsp(csp)
}

func (c *Consensus) sendPrecommitRsp(csp *sc.CsPacket) {
	log.Debug("send precommit response")
	c.sendCsRspPacket()
}

func (c *Consensus) processCommit(csp *sc.CsPacket) {
	log.Debug("process commit")

	c.step = StepCommit
	packet := c.instance.MakeNetPacket(c.step)
	if packet.BlockType == sc.SD_CM_BLOCK {
		c.csComplete()

		net.Np.GossipBlock(packet)
		if c.ns.NodeType == sc.NodeCommittee {
			net.Np.SendBlockToShards(packet)
		} else if c.ns.NodeType == sc.NodeShard {
			net.Np.SendBlockToCommittee(packet)
		}

	} else {
		net.Np.GossipBlock(packet)
		if c.ns.NodeType == sc.NodeCommittee {
			net.Np.SendBlockToShards(packet)
		} else if c.ns.NodeType == sc.NodeShard {
			net.Np.SendBlockToCommittee(packet)
		}

		c.csComplete()
	}

}
