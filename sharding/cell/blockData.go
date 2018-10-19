package cell

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
)

type chainData struct {
	cmBlock         *cs.CMBlock
	finalBlock      *cs.FinalBlock
	viewchangeBlock *cs.ViewChangeBlock
	minorBlock      *cs.MinorBlock
	preMinorBlock   *cs.MinorBlock
}

func makeChainData() *chainData {
	return &chainData{}
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
