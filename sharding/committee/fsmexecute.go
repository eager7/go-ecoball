package committee

import (
	"github.com/ecoball/go-ecoball/common/etime"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

func (c *committee) processStateTimeout() {
	c.setRetransTimer(false)
	c.fsm.Execute(ActStateTimeout, nil)
}

func (c *committee) processConsensusPacket(packet *sc.CsPacket) {
	c.fsm.Execute(ActRecvConsensusPacket, packet)
}

func (c *committee) processSyncComplete(msg interface{}) {
	lastCmBlock := simulate.GetLastCMBlock()
	if lastCmBlock == nil {
		c.fsm.Execute(ActProductCommitteeBlock, msg)
		return
	}

	c.ns.SyncCmBlockComplete(lastCmBlock)

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
		log.Panic("wrong sync status, cm block height ", lastCmBlock.Height, " final block number ", lastFinalBlock.CMEpochNo)
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
