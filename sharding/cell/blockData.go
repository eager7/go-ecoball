package cell

import "github.com/ecoball/go-ecoball/core/types/block"

type chainData struct {
	cmBlock         *block.CMBlock
	finalBlock      *block.FinalBlock
	viewchangeBlock *block.ViewChangeBlock
	//minorBlocks *minorBlockSet
}

func makeChainData() *chainData {
	return &chainData{}
}

func (c *chainData) setCMBlock(cm *block.CMBlock) {
	c.cmBlock = cm

}

func (c *chainData) setFinalBlock(final *block.FinalBlock) {
	c.finalBlock = final
}

func (c *chainData) setViewchangeBlock(vc *block.ViewChangeBlock) {
	c.viewchangeBlock = vc
}

func (c *chainData) getCMBlock() *block.CMBlock {
	return c.cmBlock
}

func (c *chainData) getFinalBlock() *block.FinalBlock {
	return c.finalBlock
}

func (c *chainData) getViewchangeBlock() *block.ViewChangeBlock {
	return c.viewchangeBlock
}
