package cell

import (
	"encoding/json"
	cs "github.com/ecoball/go-ecoball/core/shard"
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

func (c *Cell) VerifyCmPacket(p *sc.NetPacket) *sc.CsPacket {
	var cm cs.CMBlock
	err := json.Unmarshal(p.Packet, &cm)
	if err != nil {
		log.Error("cm block unmarshal error ", err)
		return nil
	}

	last := c.GetLastCMBlock()
	if last != nil {
		if last.Height+1 != cm.Height {
			log.Debug("old cm block last ", last.Height, " block ", cm.Height)
			return nil
		}
	}

	/*missing_func need verify signature here*/

	var csp sc.CsPacket
	(&csp).CopyHeader(p)
	(&csp).Packet = &cm

	return &csp
}

func (c *Cell) VerifyFinalPacket(p *sc.NetPacket) *sc.CsPacket {
	var final cs.FinalBlock
	err := json.Unmarshal(p.Packet, &final)
	if err != nil {
		log.Error("final block unmarshal error ", err)
		return nil
	}

	cm := c.GetLastCMBlock()
	if cm == nil {
		log.Error("cm block not exist")
		return nil
	}

	if final.EpochNo != cm.Height {
		log.Error("block epoch error ", cm.Height, " block ", final.EpochNo)
		return nil
	}

	last := c.GetLastFinalBlock()
	if last != nil {
		if last.Height+1 != final.Height {
			log.Debug("wrong final block last ", last.Height, " block ", final.Height)
			return nil
		}
	}

	/*missing_func need verify signature here*/

	var csp sc.CsPacket
	csp.CopyHeader(p)
	csp.Packet = &final

	return &csp
}

func (c *Cell) VerifyViewChangePacket(p *sc.NetPacket) *sc.CsPacket {
	var vc cs.ViewChangeBlock
	err := json.Unmarshal(p.Packet, &vc)
	if err != nil {
		log.Error("vc block unmarshal error ", err)
		return nil
	}

	cm := c.GetLastCMBlock()
	if cm != nil {
		if cm.Height != vc.CMEpochNo {
			log.Error("vc block epoch error last ", cm.Height, " block ", vc.CMEpochNo)
			return nil
		}
	}

	final := c.GetLastFinalBlock()
	if final != nil {
		if final.Height != vc.FinalBlockHeight {
			log.Error("vc block height error last ", final.Height, " block ", vc.FinalBlockHeight)
			return nil
		}
	}

	last := c.GetLastViewchangeBlock()
	if last == nil {
		if vc.Round != 1 {
			log.Error("vc block round error ", vc.Round, "should be 1")
			return nil
		}
	} else {
		if final != nil && final.Height > vc.FinalBlockHeight {
			if vc.Round != 1 {
				log.Error("vc block round error ", vc.Round, "should be 1")
				return nil
			}
		} else {
			if last.Round+1 != vc.Round {
				log.Error("vc block round error last ", last.Round, " block ", vc.Round)
				return nil
			}
		}
	}

	/*missing_func need verify signature here*/

	var csp sc.CsPacket
	csp.CopyHeader(p)
	csp.Packet = &vc

	return &csp
}

func (c *Cell) VerifyMinorPacket(p *sc.NetPacket) *sc.CsPacket {
	var minor cs.MinorBlock
	err := json.Unmarshal(p.Packet, &minor)
	if err != nil {
		log.Error("minor block unmarshal error ", err)
		return nil
	}

	cm := c.GetLastCMBlock()
	if cm == nil {
		log.Error("need product cm block first")
		return nil
	}

	if cm.Height != minor.CMEpochNo {
		log.Error("minor block epoch error ", minor.CMEpochNo, " current epoch ", cm.Height)
		return nil
	}

	/*missing_func need verify signature here*/
	var csp sc.CsPacket
	csp.CopyHeader(p)
	csp.Packet = &minor

	return &csp
}
