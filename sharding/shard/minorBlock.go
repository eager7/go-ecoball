package shard

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/simulate"
)

type minorBlockCsi struct {
	bk    *types.MinorBlock
	cache *types.MinorBlock
}

func newMinorBlockCsi(block *types.MinorBlock) *minorBlockCsi {
	return &minorBlockCsi{bk: block}
}

func (b *minorBlockCsi) GetCsView() *sc.CsView {
	return &sc.CsView{EpochNo: b.bk.CMEpochNo, MinorHeight: b.bk.Height}
}

func (b *minorBlockCsi) CheckBlock(bl interface{}, bLeader bool) bool {
	update := bl.(*types.MinorBlock)

	if b.bk.CMEpochNo != update.CMEpochNo || b.bk.Height != update.Height {
		log.Error("view error current ", b.bk.CMEpochNo, " ", b.bk.Height, " packet view ", update.CMEpochNo, " ", update.Height)
		return false
	}

	if !sc.Same(b.bk.ProposalPublicKey, update.ProposalPublicKey) {
		log.Error("proposal public key not same")
		return false
	}

	if update.ShardId != b.bk.ShardId {
		log.Error("candidate address not same")
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
	csp := &sc.NetPacket{PacketType: netmsg.APP_MSG_CONSENSUS_PACKET, BlockType: sc.SD_MINOR_BLOCK, Step: step}

	/*missing_func should fill in signature and bit map*/
	if step == consensus.StepPrePare {
		log.Debug("make prepare minor block")
		b.bk.Step1 = 1
	} else if step == consensus.StepPreCommit {
		log.Debug("make precommit minor block")
		b.bk.Step2 = 1
	} else if step == consensus.StepCommit {
		log.Debug("make commit minor block")
	} else {
		log.Fatal("step wrong")
		return nil
	}

	data, err := json.Marshal(b.bk)
	if err != nil {
		log.Error("minor block marshal error ", err)
		return nil
	}

	csp.Packet = data

	return csp
}

func (b *minorBlockCsi) GetCsBlock() interface{} {
	return b.bk
}

func (b *minorBlockCsi) PrepareRsp() uint32 {
	if b.cache.Step1 == 1 {
		b.bk.Step1++
	}

	return b.bk.Step1
}

func (b *minorBlockCsi) PrecommitRsp() uint32 {
	if b.cache.Step2 == 1 {
		b.bk.Step2++
	}

	return b.bk.Step2
}

func (b *minorBlockCsi) GetCandidate() *types.NodeInfo {
	return nil
}

func (s *shard) createMinorBlock() *types.MinorBlock {
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

	minor := &types.MinorBlock{
		MinorBlockHeader: types.MinorBlockHeader{
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

	return minor
}

func (s *shard) productMinorBlock(msg interface{}) {
	minor := s.createMinorBlock()

	csi := newMinorBlockCsi(minor)

	s.cs.StartConsensus(csi)
}

func (s *shard) recvCommitMinorBlock(bl *types.MinorBlock) {
	log.Debug("recv consensus minor block height ", bl.Height)

	simulate.TellMinorBlock(bl)

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

	minor := csp.Packet.(*types.MinorBlock)
	last := s.ns.GetLastMinorBlock()
	if last != nil && minor.Height <= last.Height {
		log.Error("old minor block, drop it")
		return false
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
