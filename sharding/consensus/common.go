package consensus

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/node"
	"github.com/ecoball/go-ecoball/sharding/simulate"
)

func (c *Consensus) isVoteEnough(counter uint16) bool {
	if c.ns.NodeType == sc.NodeCommittee {
		if counter == c.ns.GetCmWorksCounter() {
			return true
		} else {
			return false
		}
	} else if c.ns.NodeType == sc.NodeShard {
		if counter == c.ns.GetShardWorksCounter() {
			return true
		} else {
			return false
		}
	} else {
		log.Error("wrong node type")
		panic("wrong node type")
	}
}

func (c *Consensus) sendCsPacket() {
	csp := c.instance.MakeCsPacket(c.round)
	data, err := json.Marshal(csp)
	if err != nil {
		log.Error("cm block marshal error:%s", err)
		return
	}

	packet := message.New(message.APP_MSG_CONSENSUS_PACKET, data)
	c.bcastBlock(packet)
}

func (c *Consensus) bcastBlock(packet message.EcoBallNetMsg) {
	var works []*node.Worker
	if c.ns.NodeType == sc.NodeCommittee {
		works = c.ns.GetCmWorks()
	} else if c.ns.NodeType == sc.NodeShard {
		works = c.ns.GetShardWorks()
	} else {
		log.Error("wrong node type")
		return
	}

	for _, work := range works {
		if c.ns.Self.Equal(work) {
			continue
		}

		simulate.Sendto(work.Address, work.Port, packet)
	}
}
