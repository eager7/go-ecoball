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
	case *block.FinalBlock:
		c.recvCommitFinalBlock(bl.(*block.FinalBlock))
	default:
		log.Error("consensus call back wrong packet type ", blockType)
	}
}

func (c *committee) processConsensusPacket(packet netmsg.EcoBallNetMsg) {
	c.fsm.Execute(ActRecvConsensusPacket, packet)
}

func (c *committee) processShardingPacket(packet netmsg.EcoBallNetMsg) {

}

func (c *committee) dropPacket(packet interface{}) {
	pkt := packet.(netmsg.EcoBallNetMsg)
	log.Debug("drop packet type ", pkt.Type())
}
