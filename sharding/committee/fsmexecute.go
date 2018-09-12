package committee

import (
	"github.com/ecoball/go-ecoball/common/etime"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

func (c *committee) doBlockSync(msg interface{}) {
	etime.StopTime(c.stateTimer)
	c.stateTimer.Reset(sc.DefaultSyncBlockTimer * time.Second)
}

func (c *committee) processBlockSyncTimeout(msg interface{}) {
	log.Debug("processBlockSyncTimeout")
	if simulate.GetSyncStatus() {
		log.Error("complete sync , maybe lose message")
		c.processSyncComplete(nil)
	} else {
		log.Debug("didn't complete sync , wait next time")
		c.stateTimer.Reset(sc.DefaultSyncBlockTimer * time.Second)
	}
}

func (c *committee) waitMinorBlock(msg interface{}) {
	etime.StopTime(c.stateTimer)
	c.stateTimer.Reset(sc.DefaultWaitMinorBlockTimer * time.Second)
}
