package committee

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/etime"
	"github.com/ecoball/go-ecoball/core/types/block"
	sc "github.com/ecoball/go-ecoball/sharding/common"
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

func (b *cmBlockCsi) CacheBlock(packet *sc.CsPacket) *sc.CsView {
	var block block.CMBlock
	err := json.Unmarshal(packet.Packet, &block)
	if err != nil {
		log.Error("cm block unmarshal error:%s", err)
		return nil
	}

	b.cache = &block

	return &sc.CsView{EpochNo: b.cache.Height}
}

func (b *cmBlockCsi) MakeCsPacket(round uint16) *sc.CsPacket {
	csp := &sc.CsPacket{BlockType: sc.SD_CM_BLOCK, Round: round}

	/*missing_func should fill in signature and bit map*/
	if round == sc.CS_PREPARE_BLOCK {
		b.block.COSign.Round1 = 1
		b.block.COSign.Round2 = 0
	} else if round == sc.CS_PRECOMMIT_BLOCK {
		b.block.COSign.Round2 = 1
	} else if round == sc.CS_COMMIT_BLOCK {
		log.Debug("make cm commit block")
	}

	data, err := json.Marshal(b.block)
	if err != nil {
		log.Error("cm block marshal error:%s", err)
		return nil
	}

	csp.Packet = data

	return csp
}

func (b *cmBlockCsi) GetCsBlock() interface{} {
	return b.block
}

func (b *cmBlockCsi) PrepareRsp(*sc.CsPacket) uint16 {
	return 0
}

func (b *cmBlockCsi) PrecommitRsp(*sc.CsPacket) uint16 {
	return 0
}

func (b *cmBlockCsi) UpdateBlock(*sc.CsPacket) {

}

func (c *committee) productCommitteeBlock(msg interface{}) {
	etime.StopTime(c.stateTimer)

	last := c.ns.GetLastCMBlock()
	var height uint64
	if last == nil {
		height = 1
	} else {
		height = last.Height + 1
	}

	cm := block.NewCMBlock(height)

	cms := newCmBlockCsi(cm)

	if c.ns.IsCmLeader() {
		c.cs.StartBlockConsensusLeader(cms)
	} else {
		c.cs.StartBlockConsensusVoter(cms)
	}

	c.stateTimer.Reset(sc.DefaultProductCMBlockTimer * time.Second)
}

func (c *committee) recvCommitCmBlock(bl *block.CMBlock) {
	var cm block.CMBlock

	simulate.TellBlock(&cm)
}
