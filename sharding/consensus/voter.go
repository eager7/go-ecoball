package consensus

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

func (c *Consensus) StartBlockConsensusVoter(instance sc.ConsensusInstance) {
	c.view = instance.GetCsView()
	c.instance = instance
	c.round = RoundPrePare
}
