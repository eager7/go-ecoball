package committee

import (
	"github.com/ecoball/go-ecoball/common/config"
	cs "github.com/ecoball/go-ecoball/core/shard"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
	"github.com/ecoball/go-ecoball/common/message/mpb"
)

func (c *committee) processStateTimeout() {
	log.Debug("state time out")
	c.setRetransTimer(false, 0)
	c.setFullVoteTimer(false)
	c.fsm.Execute(ActStateTimeout, nil)
}

func (c *committee) recvConsensusPacket(packet *sc.CsPacket) {
	c.fsm.Execute(ActRecvConsensusPacket, packet)
}

func (c *committee) recvShardPacket(packet *sc.CsPacket) {
	c.fsm.Execute(ActRecvShardPacket, packet)
}

func (c *committee) processSyncComplete(msg interface{}) {
	log.Debug("recv sync complete")

	lastCmBlock, _, err := c.ns.Ledger.GetLastShardBlock(config.ChainHash, mpb.Identify_APP_MSG_CM_BLOCK)
	if err != nil || lastCmBlock == nil {
		panic("get cm block error ")
		return
	}

	cm := lastCmBlock.GetInstance().(*cs.CMBlock)
	c.ns.SyncCmBlockComplete(cm)

	lastvc, _, err := c.ns.Ledger.GetLastShardBlock(config.ChainHash, mpb.Identify_APP_MSG_VC_BLOCK)
	if err != nil || lastvc == nil {
		panic("get vc block error ")
		return
	}

	vc := lastvc.GetInstance().(*cs.ViewChangeBlock)
	c.ns.SaveLastViewchangeBlock(vc)

	lastFinalBlock, _, err := c.ns.Ledger.GetLastShardBlock(config.ChainHash, mpb.Identify_APP_MSG_FINAL_BLOCK)
	if err != nil || lastFinalBlock == nil {
		panic("get final block error ")
		return
	}

	final := lastFinalBlock.GetInstance().(*cs.FinalBlock)
	c.ns.SaveLastFinalBlock(final)

	if cm.Height == 1 && final.Height == 1 {
		c.fsm.Execute(ActProductCommitteeBlock, msg)
		return
	}

	if cm.Height > final.EpochNo {
		c.fsm.Execute(ActCollectMinorBlock, msg)
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
	if c.ns.IsLeader() {
		c.fsm.Execute(ActProductFinalBlock, msg)
		return
	} else {
		if c.ns.IsMinorBlockThresholdInPool() {
			c.fsm.Execute(ActProductFinalBlock, msg)
			return
		} else {
			c.fsm.Execute(ActCollectMinorBlock, msg)
			return
		}
	}
}

func (c *committee) doBlockSync(msg interface{}) {
	c.setSyncRequest()
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

func (c *committee) collectMinorBlock(msg interface{}) {
	c.stateTimer.Reset(sc.DefaultWaitMinorBlockTimer * time.Second)
}
