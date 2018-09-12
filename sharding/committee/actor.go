package committee

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
)

func (c *committee) processSyncComplete(msg interface{}) {
	lastCmBlock := simulate.GetLastCMBlock()
	if lastCmBlock == nil {
		c.fsm.Execute(ActProductCommitteeBlock, msg)
		return
	}

	c.ns.SyncCMBlockComplete(lastCmBlock)

	lastFinalBlock := simulate.GetLastFinalBlock()
	if lastFinalBlock == nil {
		c.fsm.Execute(ActWaitMinorBlock, msg)
		return
	}

	c.ns.SetLastFinalBlock(lastFinalBlock)

	if lastCmBlock.Height > lastFinalBlock.CMEpochNo {
		c.fsm.Execute(ActWaitMinorBlock, msg)
		return
	} else if lastCmBlock.Height < lastFinalBlock.CMEpochNo {
		panic("wrong sync status")
		log.Panic("wrong sync status, cm block height: %d, final block number: %d", lastCmBlock.Height, lastFinalBlock.CMEpochNo)
		return
	}

	if lastFinalBlock.Height%sc.DefaultEpochFinalBlockNumber == 0 {
		c.fsm.Execute(ActProductCommitteeBlock, msg)
		return
	}

	blocks := simulate.GetMinorBlockPool()
	c.ns.SyncMinorsBlockToPool(blocks)

	/*haven't collect enough shard's minor block, the wait time will be longer than default configure when we enter
	  WaitMinorBlock status, maybe we can recalculate the left time by check the minor block's timestamps */
	if c.ns.GetMinorBlockPoolCount() < uint16(len(lastCmBlock.Shards)*sc.DefaultThresholdOfMinorBlock/100) {
		c.fsm.Execute(ActWaitMinorBlock, msg)
		return
	} else {
		c.fsm.Execute(ActProductFinalBlock, msg)
		return
	}
}
