package committee

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/etime"
	"github.com/ecoball/go-ecoball/core/types/block"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

type finalBlockCsi struct {
	block *block.FinalBlock
	cache *block.FinalBlock
}

func newFinalBlockCsi(bk *block.FinalBlock) *finalBlockCsi {
	return &finalBlockCsi{block: bk}
}

func (b *finalBlockCsi) GetCsView() *sc.CsView {
	return &sc.CsView{EpochNo: b.block.CMEpochNo, FinalHeight: b.block.Height}
}

func (b *finalBlockCsi) CacheBlock(bl interface{}) *sc.CsView {
	b.cache = bl.(*block.FinalBlock)

	return &sc.CsView{EpochNo: b.cache.CMEpochNo, FinalHeight: b.cache.Height}
}

func (b *finalBlockCsi) MakeCsPacket(step uint16) *sc.CsPacket {
	csp := &sc.CsPacket{PacketType: netmsg.APP_MSG_CONSENSUS_PACKET, BlockType: sc.SD_FINAL_BLOCK, Step: step}

	/*missing_func should fill in signature and bit map*/
	if step == consensus.StepPrePare {
		log.Debug("make final prepare block")
		b.block.Step1 = 1
	} else if step == consensus.StepPreCommit {
		log.Debug("make final precommit block")
		b.block.Step2 = 1
	} else if step == consensus.StepCommit {
		log.Debug("make final commit block")
	} else {
		log.Fatal("step wrong")
		return nil
	}

	data, err := json.Marshal(b.block)
	if err != nil {
		log.Error("final block marshal error ", err)
		return nil
	}

	csp.Packet = data

	return csp
}

func (b *finalBlockCsi) GetCsBlock() interface{} {
	return b.block
}

func (b *finalBlockCsi) PrepareRsp() uint16 {
	if b.cache.Step1 == 1 {
		b.block.Step1++
	}

	return b.block.Step1
}

func (b *finalBlockCsi) PrecommitRsp() uint16 {
	if b.cache.Step2 == 1 {
		b.block.Step2++
	}

	return b.block.Step2
}

func (b *finalBlockCsi) UpdateBlock(*sc.CsPacket) {
	b.block = b.cache
	b.cache = nil
}

func (c *committee) createFinalBlock() *block.FinalBlock {

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

	log.Debug("create final block epoch ", lastcm.Height, " height ", height)
	final := block.NewFinalBlock(lastcm.Height, height)

	return final

}

func (c *committee) productFinalBlock(msg interface{}) {
	log.Debug("product final block")
	etime.StopTime(c.stateTimer)

	final := c.createFinalBlock()
	if final == nil {
		return
	}

	cms := newFinalBlockCsi(final)

	c.cs.StartConsensus(cms)

	c.stateTimer.Reset(sc.DefaultProductFinalBlockTimer * time.Second)
}

func (c *committee) processConsensusFinalPacket(p interface{}) {
	log.Debug("process final consensus block")

	c.cs.ProcessPacket(p.(*sc.CsPacket))
}

func (c *committee) processWMBStateChange(p interface{}) {
	log.Debug("process final consensus packet on waiting status")

	if c.ns.IsCmLeader() {
		log.Error("we are leader of commit, drop packet")
		return
	}

	c.productFinalBlock(nil)
	c.cs.ProcessPacket(p.(*sc.CsPacket))
}

func (c *committee) recvCommitFinalBlock(bl *block.FinalBlock) {
	log.Debug("recv consensus final block height ", bl.Height)
	simulate.TellBlock(bl)

	c.ns.SetLastFinalBlock(bl)
	if bl.Height%sc.DefaultEpochFinalBlockNumber == 0 {
		c.fsm.Execute(ActProductCommitteeBlock, nil)
	} else {
		c.fsm.Execute(ActWaitMinorBlock, nil)
	}

}
