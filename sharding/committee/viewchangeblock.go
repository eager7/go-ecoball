package committee

import (
	"github.com/ecoball/go-ecoball/common/etime"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"time"
)

func (c *committee) productViewChangeBlock(msg interface{}) {
	etime.StopTime(c.stateTimer)

	if c.ns.IsCmCandidateLeader() {
		if c.newViewChangeBlock() {
			log.Debug("failed to new view change block")
			panic("failed to new view change block")
			return
		}
	}

	c.stateTimer.Reset(sc.DefaultProductViewChangeBlockTimer * time.Second)
}

func (c *committee) newViewChangeBlock() bool {
	return true
}
