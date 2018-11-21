package cell

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/sharding/common"
)

type chainData struct {
	cmBlock         *cs.CMBlock
	finalBlock      *cs.FinalBlock
	viewchangeBlock *cs.ViewChangeBlock
	minorBlock      *cs.MinorBlock
	preMinorBlock   *cs.MinorBlock
	shardHeight     []uint64
}

func makeChainData() *chainData {
	return &chainData{shardHeight: make([]uint64, common.DefaultShardMaxMember, common.DefaultShardMaxMember)}
}

func (c *chainData) setCMBlock(cm *cs.CMBlock) {
	c.cmBlock = cm

}

func (c *chainData) getCMBlock() *cs.CMBlock {
	return c.cmBlock
}

func (c *chainData) setFinalBlock(final *cs.FinalBlock) {
	c.finalBlock = final
}

func (c *chainData) getFinalBlock() *cs.FinalBlock {
	return c.finalBlock
}

func (c *chainData) setViewchangeBlock(vc *cs.ViewChangeBlock) {
	c.viewchangeBlock = vc
}

func (c *chainData) getViewchangeBlock() *cs.ViewChangeBlock {
	return c.viewchangeBlock
}

func (c *chainData) saveMinorBlock(header *cs.MinorBlockHeader) {
	if c.preMinorBlock == nil {
		log.Debug("pre minor block not  exist ", header.Height)
		panic("pre minor not exist ")
	} else {
		if c.preMinorBlock.Hash() != header.Hash() ||
			c.preMinorBlock.Height != header.Height {
			log.Debug("pre minor block error ", c.preMinorBlock.Hash(), " ", c.preMinorBlock.Height, "recv ", header.Hash(), " ", header.Height)
			panic("pre minor not exist ")
		} else {
			c.minorBlock = c.preMinorBlock
			c.setShardHeight(c.minorBlock.ShardId, c.minorBlock.Height)
		}

		c.preMinorBlock = nil
	}
}

func (c *chainData) setMinorBlock(minor *cs.MinorBlock) {
	c.minorBlock = minor
}

func (c *chainData) getMinorBlock() *cs.MinorBlock {
	return c.minorBlock
}

func (c *chainData) setPreMinorBlock(minor *cs.MinorBlock) {
	c.preMinorBlock = minor
}

func (c *chainData) getPreMinorBlock() *cs.MinorBlock {
	return c.preMinorBlock
}

func (c *chainData) setShardHeight(shardid uint32, height uint64) {
	if shardid < 1 || shardid > common.DefaultShardMaxMember {
		panic("wrong shard id")
		return
	}

	c.shardHeight[shardid-1] = height
}

func (c *chainData) getShardHeight(shardid uint32) uint64 {
	if shardid < 1 || shardid > common.DefaultShardMaxMember {
		panic("wrong shard id")
		return 0
	}

	return c.shardHeight[shardid-1]
}
