package shard

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

func (c *shard) doBlockSync(msg interface{}) {
	c.setSyncRequest()
	c.stateTimer.Reset(sc.DefaultSyncBlockTimer * time.Second)
}

func (s *shard) processStateTimeout() {
	log.Debug("state time out")

	s.fsm.Execute(ActStateTimeout, nil)
}

func (s *shard) processBlockSyncTimeout(msg interface{}) {
	log.Debug("processBlockSyncTimeout")
	if simulate.GetSyncStatus() {
		log.Error("complete sync , maybe lose message")
		//s.processSyncComplete()
	} else {
		log.Debug("didn't complete sync , wait next time")
		s.stateTimer.Reset(sc.DefaultSyncBlockTimer * time.Second)
	}
}


func (s *shard) processSyncComplete() {
	log.Debug("recv sync complete")
	/*
	lastCmBlock, _, err := s.ns.Ledger.GetLastShardBlock(config.ChainHash, mpb.Identify_APP_MSG_CM_BLOCK)
	if err != nil || lastCmBlock == nil {
		panic("get cm block error ")
		return
	}

	cm := lastCmBlock.GetInstance().(*cs.CMBlock)
	s.ns.SyncCmBlockComplete(cm)

	lastvc, _, err := s.ns.Ledger.GetLastShardBlock(config.ChainHash, mpb.Identify_APP_MSG_VC_BLOCK)
	if err != nil || lastvc == nil {
		panic("get vc block error ")
		return
	}

	vc := lastvc.GetInstance().(*cs.ViewChangeBlock)
	s.ns.SaveLastViewchangeBlock(vc)

	lastFinalBlock, _, err := s.ns.Ledger.GetLastShardBlock(config.ChainHash, mpb.Identify_APP_MSG_FINAL_BLOCK)
	if err != nil || lastFinalBlock == nil {
		panic("get final block error ")
		return
	}

	final := lastFinalBlock.GetInstance().(*cs.FinalBlock)
	s.ns.SaveLastFinalBlock(final)

	lastMinor, bFinalize, err := s.ns.Ledger.GetLastShardBlock(config.ChainHash, mpb.Identify_APP_MSG_MINOR_BLOCK)
	if err != nil || lastMinor == nil {
		panic("get minor block error ")
		return
	}

	minor := lastMinor.GetInstance().(*cs.MinorBlock)

	if !bFinalize {
		last, finalize, err := s.ns.Ledger.GetShardBlockByHash(config.ChainHash, mpb.Identify_APP_MSG_MINOR_BLOCK, minor.PrevHash, true)
		if err != nil || finalize != true {
			log.Error("get last finalize minor block error", err)
			panic("get last finalize minor block error")
			return
		}

		minor = last.GetInstance().(*cs.MinorBlock)

	}

	s.ns.SaveLastMinorBlock(minor)

	if cm.Height == 1 && final.Height == 1 {
		s.fsm.Execute(ActWaitBlock, nil)
		return
	}
	*/
	preMinor := simulate.GetPreproductionMinorBlock()
	if preMinor != nil {
		s.fsm.Execute(ActWaitBlock, nil)
		return
	} else {
		s.fsm.Execute(ActProductMinorBlock, nil)
		return
	}
}

func (s *shard) processMinorBlockMsg(minor *cs.MinorBlock) {
	s.fsm.Execute(ActLedgerBlockMsg, minor)
}
