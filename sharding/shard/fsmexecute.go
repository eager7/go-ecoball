package shard

import (
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/sharding/simulate"
)

func (s *shard) processSyncComplete() {
	lastCmBlock, err := s.ns.Ledger.GetLastShardBlock(config.ChainHash, types.HeCmBlock)
	if err != nil || lastCmBlock == nil {
		s.fsm.Execute(ActWaitBlock, nil)
		return
	}

	cm := lastCmBlock.GetObject().(*types.CMBlock)
	s.ns.SyncCmBlockComplete(cm)

	/* missing_func vc block */

	lastFinalBlock, err := s.ns.Ledger.GetLastShardBlock(config.ChainHash, types.HeFinalBlock)
	if err != nil || lastFinalBlock == nil {
		s.fsm.Execute(ActProductMinorBlock, nil)
		return
	}

	final := lastFinalBlock.GetObject().(*types.FinalBlock)
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
