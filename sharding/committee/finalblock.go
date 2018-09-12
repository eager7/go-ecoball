package committee

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/etime"
	"github.com/ecoball/go-ecoball/core/types/block"
	"github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
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
	return &sc.CsView{EpochNo: b.block.Height}
}

func (b *finalBlockCsi) CacheBlock(packet *sc.CsPacket) *sc.CsView {
	var bk block.FinalBlock
	err := json.Unmarshal(packet.Packet, &bk)
	if err != nil {
		log.Error("final block unmarshal error:%s", err)
		return nil
	}

	b.cache = &bk

	return &sc.CsView{EpochNo: b.cache.CMEpochNo, FinalHeight: b.cache.Height}

}

func (b *finalBlockCsi) MakeCsPacket(round uint16) *sc.CsPacket {
	csp := &sc.CsPacket{Round: round, BlockType: sc.SD_FINAL_BLOCK}

	/*missing_func should fill in signature and bit map*/
	if round == sc.CS_PREPARE_BLOCK {
		b.block.COSign.Round1 = 1
		b.block.COSign.Round2 = 0
	} else if round == sc.CS_PRECOMMIT_BLOCK {
		b.block.COSign.Round2 = 1
	} else if round == sc.CS_COMMIT_BLOCK {
		log.Debug("make final commit block")
	}

	data, err := json.Marshal(csp)
	if err != nil {
		log.Error("final block marshal error:%s", err)
		return nil
	}

	csp.Packet = data

	return csp
}

func (b *finalBlockCsi) GetCsBlock() interface{} {
	return b.block
}

func (b *finalBlockCsi) PrepareRsp(*sc.CsPacket) uint16 {
	return 0
}

func (b *finalBlockCsi) PrecommitRsp(*sc.CsPacket) uint16 {
	return 0
}

func (b *finalBlockCsi) UpdateBlock(*sc.CsPacket) {

}

func (c *committee) productFinalBlock(msg interface{}) {
	etime.StopTime(c.stateTimer)

	last := c.ns.GetLastFinalBlock()
	var height uint64
	if last == nil {
		height = 1
	} else {
		height = last.Height + 1
	}

	final := block.NewFinalBlock(height)

	cms := newFinalBlockCsi(final)

	if c.ns.IsCmLeader() {
		c.cs.StartBlockConsensusLeader(cms)
	} else {
		c.cs.StartBlockConsensusVoter(cms)
	}

	c.stateTimer.Reset(sc.DefaultProductViewChangeBlockTimer * time.Second)
}

func (c *committee) recvConsensusFinalBlock(packet message.EcoBallNetMsg) {

	var final block.FinalBlock

	err := json.Unmarshal(packet.Data(), &final)
	if err != nil {
		log.Error("cm block Unmarshal error:%s", err)
		return
	}

	log.Debug("recv consensus final block height:%d", final.Height)

	simulate.TellBlock(&final)
}

func (c *committee) processWMBStateChange(packet interface{}) {
	if c.ns.IsCmLeader() {
		log.Error("we are leader of commit, drop packet")
		return
	}

	c.productFinalBlock(nil)
	c.cs.ProcessPacket(packet.(message.EcoBallNetMsg))
}
