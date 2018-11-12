package consensus

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/net"
)

func (c *Consensus) checkCosign(step uint16) bool {
	cosign := c.instance.GetCosign()
	if step == StepPreCommit {
		log.Debug("cosign step1 ", cosign.Step1)
		return c.ns.IsVoteEnough(cosign.Step1)
	} else if step == StepCommit {
		log.Debug("cosign step1 ", cosign.Step1)
		log.Debug("cosign step2 ", cosign.Step2)
		return c.ns.IsVoteEnough(cosign.Step1) && c.ns.IsVoteEnough(cosign.Step2)
	} else {
		log.Debug("wrong step")
		return false
	}
}

func (c *Consensus) setCosign() {
	cosign := c.instance.GetCosign()
	if c.step == StepPrePare {
		sign := c.ns.GetSignBit()
		cosign.Step1 = sign
	} else if c.step == StepPreCommit {
		sign := c.ns.GetSignBit()
		cosign.Step2 = sign
	} else {
		panic("wrong step")
	}
}

func (c *Consensus) sendCsPacket(packet *sc.NetPacket) {
	net.Np.BroadcastBlock(packet)

	if c.step >= StepCommit {
		if c.ns.NodeType == sc.NodeCommittee {
			net.Np.SendBlockToShards(packet)
		} else if c.ns.NodeType == sc.NodeShard {
			net.Np.SendBlockToCommittee(packet)
		}
	}
}

func (c *Consensus) sendCsRspPacket() {
	csp := c.instance.MakeNetPacket(c.step)

	candiate := c.instance.GetCandidate()
	if candiate != nil {
		worker := &sc.Worker{}
		worker.InitWork(candiate)
		net.Np.SendToPeer(csp, worker)
	} else {
		leader := c.ns.GetLeader()
		net.Np.SendToPeer(csp, leader)
	}

}

func (c *Consensus) Reset() {
	c.step = StepNIL
	c.instance = nil
	c.view = nil
	c.rcb(false, sc.DefaultBlockWindow)
	c.fcb(false)
}

func (c *Consensus) csComplete() {
	bl := c.instance.GetCsBlock()
	c.Reset()
	c.ccb(bl)
}
