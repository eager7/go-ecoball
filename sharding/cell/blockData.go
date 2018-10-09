package cell

import "github.com/ecoball/go-ecoball/core/types"

type chainData struct {
	cmBlock         *types.CMBlock
	finalBlock      *types.FinalBlock
	viewchangeBlock *types.ViewChangeBlock
	minorBlock      *types.MinorBlock
	preMinorBlock   *types.MinorBlock
}

func makeChainData() *chainData {
	return &chainData{}
}

func (c *chainData) setCMBlock(cm *types.CMBlock) {
	c.cmBlock = cm

}

func (c *chainData) getCMBlock() *types.CMBlock {
	return c.cmBlock
}

func (c *chainData) setFinalBlock(final *types.FinalBlock) {
	c.finalBlock = final
}

func (c *chainData) getFinalBlock() *types.FinalBlock {
	return c.finalBlock
}

func (c *chainData) setViewchangeBlock(vc *types.ViewChangeBlock) {
	c.viewchangeBlock = vc
}

func (c *chainData) getViewchangeBlock() *types.ViewChangeBlock {
	return c.viewchangeBlock
}

func (c *chainData) setMinorBlock(minor *types.MinorBlock) {
	c.minorBlock = minor
}

func (c *chainData) getMinorBlock() *types.MinorBlock {
	return c.minorBlock
}

func (c *chainData) setPreMinorBlock(minor *types.MinorBlock) {
	c.minorBlock = minor
}

func (c *chainData) getPreMinorBlock() *types.MinorBlock {
	return c.minorBlock
}
