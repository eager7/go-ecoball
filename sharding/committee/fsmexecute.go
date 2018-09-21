package committee

import (
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/etime"
	"github.com/ecoball/go-ecoball/core/types"
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
	lastCmBlock, err := c.ns.Ledger.GetLastShardBlock(config.ChainHash, types.HeCmBlock)
	if err != nil || lastCmBlock == nil {
		c.fsm.Execute(ActProductCommitteeBlock, msg)
		return
	}

	cm := lastCmBlock.GetObject().(*types.CMBlock)
	c.ns.SyncCmBlockComplete(cm)

	lastFinalBlock, err := c.ns.Ledger.GetLastShardBlock(config.ChainHash, types.HeFinalBlock)
	if err != nil || lastFinalBlock == nil {
		c.fsm.Execute(ActWaitMinorBlock, msg)
		return
	}

	final := lastFinalBlock.GetObject().(*types.FinalBlock)
	c.ns.SetLastFinalBlock(final)

	if cm.Height > final.EpochNo {
		c.fsm.Execute(ActWaitMinorBlock, msg)
		return
	} else if cm.Height < final.EpochNo {
		panic("wrong sync status")
		log.Panic("wrong sync status, cm block height ", cm.Height, " final block number ", final.EpochNo)
		return
	}

	if final.Height%sc.DefaultEpochFinalBlockNumber == 0 {
		c.fsm.Execute(ActProductCommitteeBlock, msg)
		return
	}

	blocks := simulate.GetMinorBlockPool()
	c.ns.SyncMinorsBlockToPool(blocks)

	/*haven't collect enough shard's minor block, the wait time will be longer than default configure when we enter
	  WaitMinorBlock status, maybe we can recalculate the left time by check the minor block's timestamps */
	if c.ns.GetMinorBlockPoolCount() < uint16(len(cm.Shards)*sc.DefaultThresholdOfMinorBlock/100) {
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
