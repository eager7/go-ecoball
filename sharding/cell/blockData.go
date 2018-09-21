package cell

import "github.com/ecoball/go-ecoball/core/types"

type chainData struct {
	cmBlock         *types.CMBlock
	finalBlock      *types.FinalBlock
	viewchangeBlock *types.ViewChangeBlock
	//minorBlocks *minorBlockSet
}

func makeChainData() *chainData {
	return &chainData{}
}

func (c *chainData) setCMBlock(cm *types.CMBlock) {
	c.cmBlock = cm

}

func (c *chainData) setFinalBlock(final *types.FinalBlock) {
	c.finalBlock = final
}

func (c *chainData) setViewchangeBlock(vc *types.ViewChangeBlock) {
	c.viewchangeBlock = vc
}

func (c *chainData) getCMBlock() *types.CMBlock {
	return c.cmBlock
}

func (c *chainData) getFinalBlock() *types.FinalBlock {
	return c.finalBlock
}

func (c *chainData) getViewchangeBlock() *types.ViewChangeBlock {
	return c.viewchangeBlock
}
