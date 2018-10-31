package cell

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

func (c *Cell) GetSignBit() uint32 {
	works := c.GetWorks()
	for i, work := range works {
		if c.Self.Equal(work) {
			return 1 << uint32(i)
		}
	}

	return 0
}

func (c *Cell) getSignCount(sign uint32) uint32 {
	var counter uint32
	var i uint32
	for i = 0; i < c.GetWorksCounter(); i++ {
		mask := 1 << uint32(i)
		if sign&uint32(mask) > 0 {
			counter++
		}
	}

	return counter
}

func (c *Cell) IsVoteEnough(sign uint32) bool {
	counter := c.getSignCount(sign)

	if counter >= c.GetWorksCounter()*sc.DefaultThresholdOfConsensus/1000+1 {
		return true
	} else {
		return false
	}
}

func (c *Cell) IsVoteOnThreshold(sign uint32) bool {
	counter := c.getSignCount(sign)

	if counter == c.GetWorksCounter()*sc.DefaultThresholdOfConsensus/1000+1 {
		return true
	} else {
		return false
	}
}

func (c *Cell) IsVoteFull(sign uint32) bool {
	counter := c.getSignCount(sign)

	if counter == c.GetWorksCounter() {
		return true
	} else {
		return false
	}
}
