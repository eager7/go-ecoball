package committee

import (
	netmsg "github.com/ecoball/go-ecoball/net/message"
	//sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/core/types/block"
)

func (c *committee) consensusCb(bl interface{}) {
	switch blockType := bl.(type) {
	case *block.CMBlock:
		c.recvCommitCmBlock(bl.(*block.CMBlock))
	default:
		log.Error("consensus call back wrong packet type %d", blockType)
	}
}

func (c *committee) processConsensusPacket(packet netmsg.EcoBallNetMsg) {
	c.fsm.Execute(ActRecvConsensusPacket, packet)
}

func (c *committee) processNetConsensusPacket(packet interface{}) {
	c.cs.ProcessPacket(packet.(netmsg.EcoBallNetMsg))
}

func (c *committee) processShardingPacket(packet netmsg.EcoBallNetMsg) {

}

func (c *committee) dropPacket(packet interface{}) {
	pkt := packet.(netmsg.EcoBallNetMsg)
	log.Debug("drop packet type %d", pkt.Type())
}
