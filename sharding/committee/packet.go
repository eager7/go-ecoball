package committee

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/core/types/block"
	sc "github.com/ecoball/go-ecoball/sharding/common"
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

func (c *committee) verifyPacket(csp *sc.NetPacket) {
	log.Debug("verify packet ", csp.BlockType)
	if csp.BlockType == sc.SD_CM_BLOCK {
		c.verifyCmPacket(csp)
	} else if csp.BlockType == sc.SD_FINAL_BLOCK {
		c.verifyFinalPacket(csp)
	} else if csp.BlockType == sc.SD_VIEWCHANGE_BLOCK {
		c.verifyViewChangePacket(csp)
	} else {
		log.Error("wrong block type")
		return
	}
}

func (c *committee) verifyCmPacket(p *sc.NetPacket) {
	var cm block.CMBlock
	err := json.Unmarshal(p.Packet, &cm)
	if err != nil {
		log.Error("cm block unmarshal error ", err)
		return
	}

	last := c.ns.GetLastCMBlock()
	if last != nil {
		if last.Height >= cm.Height {
			log.Debug("old cm packet")
			return
		}
	}

	/*missing_func need verify signature here*/

	var csp sc.CsPacket
	(&csp).Copyhead(p)
	(&csp).Packet = &cm

	c.ppc <- &csp
}

func (c *committee) verifyFinalPacket(p *sc.NetPacket) {
	var final block.FinalBlock
	err := json.Unmarshal(p.Packet, &final)
	if err != nil {
		log.Error("final block unmarshal error ", err)
		return
	}

	last := c.ns.GetLastFinalBlock()
	if last != nil {
		if last.Height >= final.Height {
			log.Debug("old final packet")
			return
		}
	}

	/*missing_func need verify signature here*/

	var csp sc.CsPacket
	csp.Copyhead(p)
	csp.Packet = &final

	c.ppc <- &csp
}

func (c *committee) verifyViewChangePacket(p *sc.NetPacket) {
	var vc block.ViewChangeBlock
	err := json.Unmarshal(p.Packet, &vc)
	if err != nil {
		log.Error("cm block unmarshal error ", err)
		return
	}

	cm := c.ns.GetLastCMBlock()
	if cm != nil {
		if cm.Height > vc.CMEpochNo {
			log.Error("vc block epoch error")
			return
		}
	}

	final := c.ns.GetLastFinalBlock()
	if final != nil {
		if final.Height > vc.FinalBlockHeight {
			log.Error("vc block final block height error")
			return
		}
	}

	last := c.ns.GetLastViewchangeBlock()
	if last != nil {
		if last.Round >= vc.Round {
			log.Error("vc block round error")
			return
		}
	}

	/*missing_func need verify signature here*/

	var csp sc.CsPacket
	csp.Copyhead(p)
	csp.Packet = &vc

	c.ppc <- &csp
}

func (c *committee) dropPacket(packet interface{}) {
	pkt := packet.(*sc.CsPacket)
	log.Debug("drop packet type ", pkt.PacketType)
}
