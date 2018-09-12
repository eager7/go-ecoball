package committee

import (
	"github.com/ecoball/go-ecoball/common/etime"
	"github.com/ecoball/go-ecoball/core/types/block"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

func (c *committee) productViewChangeBlock(msg interface{}) {
	etime.StopTime(c.stateTimer)

	c.stateTimer.Reset(sc.DefaultProductViewChangeBlockTimer * time.Second)
}

func (c *committee) processViewchangeConsensusPacket(packet interface{}) {
	if c.ns.IsCmLeader() {
		if !c.cs.IsCsRunning() {
			panic("consensus is not running")
			return
		}
	} else {
		if !c.cs.IsCsRunning() {
			c.productViewChangeBlock(nil)
		}
	}

	c.cs.ProcessPacket(packet.(netmsg.EcoBallNetMsg))
}

func (c *committee) recvCommitViewchangeBlock(bl *block.ViewChangeBlock) {
	log.Debug("recv consensus view chaneg block epoch ", bl.CMEpochNo, " round  ", bl.Round)
	simulate.TellBlock(bl)
}
