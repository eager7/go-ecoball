package consensus

import (
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/net"
)

func (c *Consensus) isVoteEnough(counter uint32) bool {
	if counter >= c.ns.GetWorksCounter()*sc.DefaultThresholdOfConsensus/1000+1 {
		return true
	} else {
		return false
	}
}

func (c *Consensus) isVoteOnThreshold(counter uint32) bool {
	if counter == c.ns.GetWorksCounter()*sc.DefaultThresholdOfConsensus/1000+1 {
		return true
	} else {
		return false
	}
}

func (c *Consensus) isVoteFull(counter uint32) bool {
	if counter == c.ns.GetWorksCounter() {
		return true
	} else {
		return false
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
		worker := &cell.Worker{}
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
}

func (c *Consensus) csComplete() {
	bl := c.instance.GetCsBlock()
	c.Reset()
	c.ccb(bl)
}
