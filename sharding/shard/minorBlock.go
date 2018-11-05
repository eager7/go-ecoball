package shard

import (
	"github.com/ecoball/go-ecoball/common"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message/pb"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

type minorBlockCsi struct {
	bk    *cs.MinorBlock
	cache *cs.MinorBlock
}

func newMinorBlockCsi(block *cs.MinorBlock, sign uint32) *minorBlockCsi {
	block.Step1 = sign
	block.Step2 = sign

	return &minorBlockCsi{bk: block}
}

func (b *minorBlockCsi) GetCsView() *sc.CsView {
	return &sc.CsView{EpochNo: b.bk.CMEpochNo, MinorHeight: b.bk.Height}
}

func (b *minorBlockCsi) CheckBlock(bl interface{}, bLeader bool) bool {
	update := bl.(*cs.MinorBlock)

	if b.bk.CMEpochNo != update.CMEpochNo || b.bk.Height != update.Height {
		log.Error("view error current ", b.bk.CMEpochNo, " ", b.bk.Height, " packet view ", update.CMEpochNo, " ", update.Height)
		return false
	}

	if !sc.Same(b.bk.ProposalPublicKey, update.ProposalPublicKey) {
		log.Error("proposal public key not same")
		return false
	}

	if update.ShardId != b.bk.ShardId {
		log.Error("shardid wrong block ", update.ShardId, " expect ", b.bk.ShardId)
		return false
	}

	if bLeader {
		b.cache = update
	} else {
		b.bk = update
	}

	return true
}

func (b *minorBlockCsi) MakeNetPacket(step uint16) *sc.NetPacket {
	csp := &sc.NetPacket{PacketType: pb.MsgType_APP_MSG_CONSENSUS_PACKET, BlockType: sc.SD_MINOR_BLOCK, Step: step}

	data, err := b.bk.Serialize()
	if err != nil {
		log.Error("minor block Serialize error ", err)
		return nil
	}

	csp.Packet = data

	return csp
}

func (b *minorBlockCsi) GetCsBlock() interface{} {
	return b.bk
}

func (b *minorBlockCsi) PrepareRsp() uint32 {
	log.Debug("prepare receive consign ", b.cache.Step1)

	b.bk.Step1 |= b.cache.Step1

	return b.bk.Step1
}

func (b *minorBlockCsi) PrecommitRsp() uint32 {
	log.Debug("precommit receive consign ", b.cache.Step2)

	b.bk.Step2 |= b.cache.Step2

	return b.bk.Step2
}

func (b *minorBlockCsi) GetCosign() *types.COSign {
	return b.bk.COSign
}

func (b *minorBlockCsi) GetCandidate() *cs.NodeInfo {
	return nil
}

func (s *shard) createMinorBlock() *cs.MinorBlock {
	lastcm := s.ns.GetLastCMBlock()
	if lastcm == nil {
		panic("cm block not exist")
		return nil
	}

	lastMinor := s.ns.GetLastMinorBlock()
	var height uint64
	if lastMinor == nil {
		height = 1
	} else {
		height = lastMinor.Height + 1
	}

	minor := &cs.MinorBlock{
		MinorBlockHeader: cs.MinorBlockHeader{
			ChainID:           common.Hash{},
			Version:           0,
			Height:            0,
			Timestamp:         0,
			PrevHash:          common.Hash{},
			TrxHashRoot:       common.Hash{},
			StateDeltaHash:    common.Hash{},
			CMBlockHash:       common.Hash{},
			ProposalPublicKey: nil,
			ShardId:           0,
			CMEpochNo:         0,
			Receipt:           types.BlockReceipt{},
			COSign:            nil,
		},
		Transactions: nil,
		StateDelta:   nil,
	}

	minor.Height = height
	minor.CMEpochNo = lastcm.Height
	minor.ShardId = uint32(s.ns.Shardid)

	cosign := &types.COSign{}
	cosign.Step1 = 1
	cosign.Step2 = 0

	minor.COSign = cosign

	log.Debug(" create minor block epoch ", minor.CMEpochNo, " height ", minor.Height)

	return minor
}

func (s *shard) productMinorBlock(msg interface{}) {
	if s.ns.IsLeader() {
		lastcm := s.ns.GetLastCMBlock()
		if lastcm == nil {
			panic("cm block not exist")
			return
		}

		lastMinor := s.ns.GetLastMinorBlock()
		var height uint64
		if lastMinor == nil {
			height = 1
		} else {
			height = lastMinor.Height + 1
		}

		simulate.TellLedgerProductMinorBlock(lastcm.Height, height)
	} else {
		minor := s.createMinorBlock()
		sign := s.ns.GetSignBit()
		csi := newMinorBlockCsi(minor, sign)
		s.cs.StartConsensus(csi, sc.DefaultBlockWindow)
	}
}

func (s *shard) processLedgerMinorBlockMsg(p interface{}) {
	minor := p.(*cs.MinorBlock)

	sign := s.ns.GetSignBit()
	csi := newMinorBlockCsi(minor, sign)
	s.cs.StartConsensus(csi, sc.DefaultMinorBlockWindow*time.Millisecond)
}

func (s *shard) reproductMinorBlock(msg interface{}) {
	s.cs.Reset()
	s.productMinorBlock(msg)
}

func (s *shard) commitMinorBlock(bl *cs.MinorBlock) {
	log.Debug("consensus minor block height ", bl.Height)

	simulate.TellBlock(bl)

	s.ns.SavePreMinorBlock(bl)

	s.fsm.Execute(ActWaitBlock, nil)
}

func (s *shard) checkMinorPacket(p interface{}) bool {
	/*check block*/
	csp := p.(*sc.CsPacket)
	if csp.BlockType != sc.SD_MINOR_BLOCK {
		log.Error("it is not minor block, drop it")
		return false
	}

	minor := csp.Packet.(*cs.MinorBlock)
	last := s.ns.GetLastMinorBlock()
	if last != nil {
		if minor.Height <= last.Height {
			log.Error("old minor block, drop it")
			return false
		} else if minor.Height > last.Height+1 {
			log.Debug("last ", last.Height, "recv ", minor.Height, " need sync")
			s.fsm.Execute(ActChainNotSync, nil)
			return false
		}
	}

	return true
}

func (s *shard) processConsensusMinorPacket(p interface{}) {
	log.Debug("process minor consensus packet")

	if !s.checkMinorPacket(p) {
		return
	}

	s.cs.ProcessPacket(p.(*sc.CsPacket))
}
