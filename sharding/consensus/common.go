package consensus

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/net/message"
)

func (c *Consensus) isVoteEnough(counter uint16) bool {
	if counter == c.ns.GetWorksCounter() {
		return true
	} else {
		return false
	}

}

func (c *Consensus) sendCsPacket() {
	csp := c.instance.MakeCsPacket(c.step)
	data, err := json.Marshal(csp)
	if err != nil {
		log.Error("cm block marshal error ", err)
		return
	}

	packet := message.New(message.APP_MSG_CONSENSUS_PACKET, data)

	c.BroadcastBlock(packet)
}

func (c *Consensus) sendCsRspPacket() {
	csp := c.instance.MakeCsPacket(c.step)
	data, err := json.Marshal(csp)
	if err != nil {
		log.Error("cm block marshal error ", err)
		return
	}

	packet := message.New(message.APP_MSG_CONSENSUS_PACKET, data)

	c.sendToLeader(packet)
}

func (c *Consensus) csComplete() {
	bl := c.instance.GetCsBlock()
	c.completeCb(bl)

	c.reset()
}

func (c *Consensus) reset() {
	c.step = RoundNIL
	c.instance = nil
	c.view = nil
}
