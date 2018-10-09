package cell

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/core/types"
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

func (c *Cell) VerifyCmPacket(p *sc.NetPacket) *sc.CsPacket {
	var cm types.CMBlock
	err := json.Unmarshal(p.Packet, &cm)
	if err != nil {
		log.Error("cm block unmarshal error ", err)
		return nil
	}

	last := c.GetLastCMBlock()
	if last != nil {
		if last.Height >= cm.Height {
			log.Debug("old cm packet")
			return nil
		}
	}

	/*missing_func need verify signature here*/

	var csp sc.CsPacket
	(&csp).Copyhead(p)
	(&csp).Packet = &cm

	return &csp
}

func (c *Cell) VerifyFinalPacket(p *sc.NetPacket) *sc.CsPacket {
	var final types.FinalBlock
	err := json.Unmarshal(p.Packet, &final)
	if err != nil {
		log.Error("final block unmarshal error ", err)
		return nil
	}

	last := c.GetLastFinalBlock()
	if last != nil {
		if last.Height >= final.Height {
			log.Debug("old final packet")
			return nil
		}
	}

	/*missing_func need verify signature here*/

	var csp sc.CsPacket
	csp.Copyhead(p)
	csp.Packet = &final

	return &csp
}

func (c *Cell) VerifyViewChangePacket(p *sc.NetPacket) *sc.CsPacket {
	var vc types.ViewChangeBlock
	err := json.Unmarshal(p.Packet, &vc)
	if err != nil {
		log.Error("vc block unmarshal error ", err)
		return nil
	}

	cm := c.GetLastCMBlock()
	if cm != nil {
		if cm.Height > vc.CMEpochNo {
			log.Error("vc block epoch error")
			return nil
		}
	}

	final := c.GetLastFinalBlock()
	if final != nil {
		if final.Height > vc.FinalBlockHeight {
			log.Error("vc block final block height error")
			return nil
		}
	}

	last := c.GetLastViewchangeBlock()
	if last != nil {
		if last.Round >= vc.Round {
			log.Error("vc block round error")
			return nil
		}
	}

	/*missing_func need verify signature here*/

	var csp sc.CsPacket
	csp.Copyhead(p)
	csp.Packet = &vc

	return &csp
}

func (c *Cell) VerifyMinorPacket(p *sc.NetPacket) *sc.CsPacket {
	var minor types.MinorBlock
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

	if cm != nil {
		if cm.Height != minor.CMEpochNo {
			log.Error("minor block epoch error ", minor.CMEpochNo, " current epoch ", cm.Height)
			return nil
		}
	}

	/*missing_func need verify signature here*/
	var csp sc.CsPacket
	csp.Copyhead(p)
	csp.Packet = &minor

	return &csp
}
