package committee

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
)

func (c *committee) processFinalBlockMsg(final *cs.FinalBlock) {
	c.fsm.Execute(ActLedgerBlockMsg, final)
}
