package consensus

import (
	"github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/sharding/simulate"
)

func (c *Consensus) sendToLeader(packet message.EcoBallNetMsg) {
	work := c.ns.GetLeader()
	if work == nil {
		log.Error("leader is nil")
		return
	}

	simulate.Sendto(work.Address, work.Port, packet)
}

func (c *Consensus) GossipBlock(packet message.EcoBallNetMsg) {
	works := c.ns.GetWorks()
	if works == nil {
		log.Error("works is nil")
		return
	}

	for _, work := range works {
		if c.ns.Self.Equal(work) {
			continue
		}

		simulate.Sendto(work.Address, work.Port, packet)
	}
}

func (c *Consensus) BroadcastBlock(packet message.EcoBallNetMsg) {
	works := c.ns.GetWorks()
	if works == nil {
		log.Error("works is nil")
		return
	}

	for _, work := range works {
		if c.ns.Self.Equal(work) {
			continue
		}

		simulate.Sendto(work.Address, work.Port, packet)
	}
}
