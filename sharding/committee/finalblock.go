package committee

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/etime"
	"github.com/ecoball/go-ecoball/core/types"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

type finalBlockCsi struct {
	bk    *types.FinalBlock
	cache *types.FinalBlock
}

func newFinalBlockCsi(bk *types.FinalBlock) *finalBlockCsi {
	return &finalBlockCsi{bk: bk}
}

func (b *finalBlockCsi) GetCsView() *sc.CsView {
	return &sc.CsView{EpochNo: b.bk.EpochNo, FinalHeight: b.bk.Height}
}

func (b *finalBlockCsi) CheckBlock(bl interface{}, bLeader bool) bool {
	update := bl.(*types.FinalBlock)
	if b.bk.Height != update.Height || b.bk.EpochNo != update.EpochNo {
		log.Error("view error current ", b.bk.EpochNo, " ", b.bk.Height, " packet view ", update.EpochNo, " ", update.Height)
		return false
	}

	if !sc.Same(b.bk.ProposalPubKey, update.ProposalPubKey) {
		log.Error("proposal not same")
		return false
	}

	if bLeader {
		b.cache = update
	} else {
		b.bk = update
	}
	return true
}

func (b *finalBlockCsi) MakeNetPacket(step uint16) *sc.NetPacket {
	csp := &sc.NetPacket{PacketType: netmsg.APP_MSG_CONSENSUS_PACKET, BlockType: sc.SD_FINAL_BLOCK, Step: step}

	/*missing_func should fill in signature and bit map*/
	if step == consensus.StepPrePare {
		log.Debug("make final prepare block")
		b.bk.Step1 = 1
	} else if step == consensus.StepPreCommit {
		log.Debug("make final precommit block")
		b.bk.Step2 = 1
	} else if step == consensus.StepCommit {
		log.Debug("make final commit block")
	} else {
		log.Fatal("step wrong")
		return nil
	}

	data, err := json.Marshal(b.bk)
	if err != nil {
		log.Error("final block marshal error ", err)
		return nil
	}

	csp.Packet = data

	return csp
}

func (b *finalBlockCsi) GetCsBlock() interface{} {
	return b.bk
}

func (b *finalBlockCsi) PrepareRsp() uint32 {
	if b.cache.Step1 == 1 {
		b.bk.Step1++
	}

	return b.bk.Step1
}

func (b *finalBlockCsi) PrecommitRsp() uint32 {
	if b.cache.Step2 == 1 {
		b.bk.Step2++
	}

	return b.bk.Step2
}

func (b *finalBlockCsi) GetCandidate() *types.NodeInfo {
	return nil
}

func (c *committee) createFinalBlock() *types.FinalBlock {

	lastcm := c.ns.GetLastCMBlock()
	if lastcm == nil {
		panic("cm block not exist")
		return nil
	}

	lastfinal := c.ns.GetLastFinalBlock()
	var height uint64
	if lastfinal == nil {
		height = 1
	} else {
		height = lastfinal.Height + 1
	}

	final := &types.FinalBlock{
		FinalBlockHeader: types.FinalBlockHeader{
			ChainID:            common.Hash{},
			Version:            0,
			Height:             0,
			Timestamp:          0,
			TrxCount:           0,
			PrevHash:           common.Hash{},
			ProposalPubKey:     nil,
			EpochNo:            0,
			CMBlockHash:        common.Hash{},
			TrxRootHash:        common.Hash{},
			StateDeltaRootHash: common.Hash{},
			MinorBlocksHash:    common.Hash{},
			StateHashRoot:      common.Hash{},
			COSign:             nil,
		},
		MinorBlocks: nil,
	}
	final.Height = height
	final.EpochNo = lastcm.Height

	cosign := &types.COSign{}
	cosign.Step1 = 1
	cosign.Step2 = 0

	final.COSign = cosign

	log.Debug("create final block epoch ", lastcm.Height, " height ", height)

	return final

}

func (c *committee) productFinalBlock(msg interface{}) {
	log.Debug("product final block")
	etime.StopTime(c.stateTimer)

	final := c.createFinalBlock()
	if final == nil {
		return
	}

	csi := newFinalBlockCsi(final)

	c.cs.StartConsensus(csi)

	c.stateTimer.Reset(sc.DefaultProductFinalBlockTimer * time.Second)
}

func (c *committee) checkFinalPacket(p interface{}) bool {

	/*recheck block*/
	csp := p.(*sc.CsPacket)
	if csp.BlockType != sc.SD_FINAL_BLOCK {
		log.Error("it is not final block, drop it")
		return false
	}

	final := csp.Packet.(*types.FinalBlock)
	last := c.ns.GetLastFinalBlock()
	if last != nil && final.Height <= last.Height {
		log.Error("old final block, drop it")
		return false
	}

	return true

}

func (c *committee) processConsensusFinalPacket(p interface{}) {
	log.Debug("process final consensus block")

	if !c.checkFinalPacket(p) {
		return
	}

	c.cs.ProcessPacket(p.(*sc.CsPacket))
}

func (c *committee) processConsensBlockOnWaitStatus(p interface{}) bool {
	log.Debug("process final consensus packet on waiting status")
	if c.ns.IsLeader() {
		log.Error("we are leader of committee, drop packet")
		return false
	}

	if !c.checkFinalPacket(p) {
		return false
	}

	c.productFinalBlock(nil)

	return c.cs.ProcessPacket(p.(*sc.CsPacket))
}

func (c *committee) commitFinalBlock(bl *types.FinalBlock) {
	log.Debug("recv consensus final block height ", bl.Height)
	simulate.TellBlock(bl)

	c.ns.SaveLastFinalBlock(bl)
	if bl.Height%sc.DefaultEpochFinalBlockNumber == 0 {
		c.fsm.Execute(ActProductCommitteeBlock, nil)
	} else {
		c.fsm.Execute(ActWaitMinorBlock, nil)
	}

}
