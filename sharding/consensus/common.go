package consensus

func (c *Consensus) isVoteEnough(counter uint16) bool {
	if counter == c.ns.GetWorksCounter() {
		return true
	} else {
		return false
	}

}

func (c *Consensus) sendCsPacket() {
	csp := c.instance.MakeCsPacket(c.step)

	c.BroadcastBlock(csp)
}

func (c *Consensus) sendCsRspPacket() {
	csp := c.instance.MakeCsPacket(c.step)

	c.sendToLeader(csp)
}

func (c *Consensus) reset() {
	c.step = StepNIL
	c.instance = nil
	c.view = nil
}

func (c *Consensus) csComplete() {
	bl := c.instance.GetCsBlock()
	c.reset()
	c.completeCb(bl)
}
