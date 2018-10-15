package shard

import (
	"github.com/ecoball/go-ecoball/common/config"
	cs "github.com/ecoball/go-ecoball/core/shard"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

func (s *shard) processStateTimeout() {
	s.fsm.Execute(ActStateTimeout, nil)
}

func (s *shard) processBlockSyncTimeout(msg interface{}) {
	log.Debug("processBlockSyncTimeout")
	if simulate.GetSyncStatus() {
		log.Error("complete sync , maybe lose message")
		s.processSyncComplete()
	} else {
		log.Debug("didn't complete sync , wait next time")
		s.stateTimer.Reset(sc.DefaultSyncBlockTimer * time.Second)
	}
}

func (s *shard) processSyncComplete() {
	lastCmBlock, err := s.ns.Ledger.GetLastShardBlock(config.ChainHash, cs.HeCmBlock)
	if err != nil || lastCmBlock == nil {
		s.fsm.Execute(ActWaitBlock, nil)
		return
	}

	cm := lastCmBlock.GetObject().(*cs.CMBlock)
	s.ns.SyncCmBlockComplete(cm)

	/* missing_func vc block */

	lastFinalBlock, err := s.ns.Ledger.GetLastShardBlock(config.ChainHash, cs.HeFinalBlock)
	if err != nil || lastFinalBlock == nil {
		s.fsm.Execute(ActProductMinorBlock, nil)
		return
	}

	final := lastFinalBlock.GetObject().(*cs.FinalBlock)
	s.ns.SaveLastFinalBlock(final)

	//lastMinor, err := s.ns.Ledger.GetLastShardBlock(config.ChainHash, types.HeMinorBlock)
	//minor := lastMinor.GetObject().(*types.MinorBlock)
	//s.ns.SaveLastMinorBlock(minor)

	minor := simulate.GetPreproductionMinorBlock()
	if minor != nil {
		s.fsm.Execute(ActWaitBlock, nil)
		return
	} else {
		s.fsm.Execute(ActProductMinorBlock, nil)
		return
	}
}
