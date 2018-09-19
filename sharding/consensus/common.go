package consensus

import (
	"github.com/ecoball/go-ecoball/sharding/cell"
	"github.com/ecoball/go-ecoball/sharding/common"
)

func (c *Consensus) isVoteEnough(counter uint16) bool {
	if counter > c.ns.GetWorksCounter()*common.DefaultThresholdOfConsensus/1000+1 {
		return true
	} else {
		return false
	}
}

func (c *Consensus) sendCsPacket() {
	csp := c.instance.MakeNetPacket(c.step)

	c.BroadcastBlock(csp)
}

func (c *Consensus) sendCsRspPacket() {
	csp := c.instance.MakeNetPacket(c.step)

	candiate := c.instance.GetCandidate()
	if candiate != nil {
		worker := &cell.Worker{}
		worker.InitWork(candiate)
		c.sendToPeer(csp, worker)
	} else {
		leader := c.ns.GetLeader()
		c.sendToPeer(csp, leader)
	}

}

func (c *Consensus) reset() {
	c.step = StepNIL
	c.instance = nil
	c.view = nil
}

func (c *Consensus) csComplete() {
	bl := c.instance.GetCsBlock()
	c.reset()
	c.completeCb(bl)
}
