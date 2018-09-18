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

type cmBlockCsi struct {
	block *block.CMBlock
	cache *block.CMBlock
}

func newCmBlockCsi(block *block.CMBlock) *cmBlockCsi {
	return &cmBlockCsi{block: block}
}

func (b *cmBlockCsi) GetCsView() *sc.CsView {
	return &sc.CsView{EpochNo: b.block.Height}
}

func (b *cmBlockCsi) CacheBlock(bl interface{}) *sc.CsView {
	b.cache = bl.(*block.CMBlock)

	return &sc.CsView{EpochNo: b.cache.Height}
}

func (b *cmBlockCsi) MakeCsPacket(step uint16) *sc.CsPacket {
	csp := &sc.CsPacket{PacketType: netmsg.APP_MSG_CONSENSUS_PACKET, BlockType: sc.SD_CM_BLOCK, Step: step}

	/*missing_func should fill in signature and bit map*/
	if step == consensus.StepPrePare {
		log.Debug("make cm prepare block")
		b.block.Step1 = 1
	} else if step == consensus.StepPreCommit {
		log.Debug("make cm precommit block")
		b.block.Step2 = 1
	} else if step == consensus.StepCommit {
		log.Debug("make cm commit block")
	} else {
		log.Fatal("step wrong")
		return nil
	}

	data, err := json.Marshal(b.block)
	if err != nil {
		log.Error("cm block marshal error ", err)
		return nil
	}

	csp.Packet = data

	return csp
}

func (b *cmBlockCsi) GetCsBlock() interface{} {
	return b.block
}

func (b *cmBlockCsi) PrepareRsp() uint16 {
	if b.cache.Step1 == 1 {
		b.block.Step1++
	}

	return b.block.Step1
}

func (b *cmBlockCsi) PrecommitRsp() uint16 {
	if b.cache.Step2 == 1 {
		b.block.Step2++
	}

	return b.block.Step2
}

func (b *cmBlockCsi) UpdateBlock(*sc.CsPacket) {
	b.block = b.cache
	b.cache = nil
}

func (c *committee) createCommitteeBlock() *block.CMBlock {
	last := c.ns.GetLastCMBlock()
	var height uint64
	if last == nil {
		height = 1
	} else {
		height = last.Height + 1
	}

	log.Debug("create cm block height ", height)

	cm := block.NewCMBlock(height)

	return cm

}

func (c *committee) productCommitteeBlock(msg interface{}) {
	etime.StopTime(c.stateTimer)

	cm := c.createCommitteeBlock()

	cms := newCmBlockCsi(cm)

	c.cs.StartConsensus(cms)

	c.stateTimer.Reset(sc.DefaultProductCmBlockTimer * time.Second)
}

func (c *committee) processConsensusCmPacket(p interface{}) {
	log.Debug("process cm consensus packet")

	c.cs.ProcessPacket(p.(*sc.CsPacket))
}

func (c *committee) recvCommitCmBlock(bl *block.CMBlock) {
	log.Debug("recv consensus cm block height ", bl.Height)
	simulate.TellBlock(bl)

	c.ns.SetLastCMBlock(bl)
	c.fsm.Execute(ActWaitMinorBlock, nil)
}
