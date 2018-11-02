package cell

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"fmt"
	"encoding/json"
)

func (c *Cell) VerifyCmPacket(p *sc.NetPacket) *sc.CsPacket {
	cm := new(cs.CMBlock)
	err := cm.Deserialize(p.Packet)
	if err != nil {
		log.Error("cm block Deserialize error ", err)
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
	(&csp).Packet = cm

	return &csp
}

func (c *Cell) VerifyFinalPacket(p *sc.NetPacket) *sc.CsPacket {
	final := new(cs.FinalBlock)
	err := final.Deserialize(p.Packet)
	if err != nil {
		log.Error("final block Deserialize error ", err)
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
		if last.Height >= final.Height {
			log.Debug("wrong final block last ", last.Height, " block ", final.Height)
			return nil
		}
	}

	/*missing_func need verify signature here*/

	var csp sc.CsPacket
	csp.CopyHeader(p)
	csp.Packet = final

	return &csp
}

func (c *Cell) VerifyViewChangePacket(p *sc.NetPacket) *sc.CsPacket {
	vc := new(cs.ViewChangeBlock)
	err := vc.Deserialize(p.Packet)
	if err != nil {
		log.Error("vc block Deserialize error ", err)
		return nil
	}

	last := c.GetLastViewchangeBlock()
	if last == nil {
		panic("vc block not exist")
		return nil
	}

	if last.Height >= vc.Height {
		log.Error("vc block height error last ", last.Height, " recv ", vc.Height)
		return nil
	}

	cm := c.GetLastCMBlock()
	if cm == nil {
		panic("cm block is not exist")
		return nil
	}

	if cm.Height != vc.CMEpochNo {
		log.Error("vc block epoch error last ", cm.Height, " block ", vc.CMEpochNo)
		return nil
	}

	final := c.GetLastFinalBlock()
	if final == nil {
		panic("final block is not exist")
		return nil
	}

	if final.Height != vc.FinalBlockHeight {
		log.Error("vc block height error last ", final.Height, " block ", vc.FinalBlockHeight)
		return nil
	}

	if vc.FinalBlockHeight == last.FinalBlockHeight {
		if last.Round >= vc.Round {
			log.Error("vc block round error last ", last.Round, " block ", vc.Round)
			return nil
		}
	}

	var csp sc.CsPacket
	csp.CopyHeader(p)
	csp.Packet = vc

	return &csp
}

//For test
/*func VerifySyncRequestPacketTest(p *sc.NetPacket) *sc.CsPacket {
	var requestPacket *sc.SyncRequestPacket
	err := json.Unmarshal(p.Packet, &requestPacket)
	if err != nil {
		log.Error("syncResponse unmarshal error ", err)
		return nil
	}
	var csp sc.CsPacket
	csp.CopyHeader(p)
	csp.Packet = &requestPacket
	return &csp
}*/

//Mark Decode
func (c *Cell)VerifySyncRequestPacket(p *sc.NetPacket) *sc.CsPacket {
	var requestPacket *sc.SyncRequestPacket
	err := json.Unmarshal(p.Packet, &requestPacket)
	if err != nil {
		log.Error("syncResponse unmarshal error ", err)
		return nil
	}
	var csp sc.CsPacket
	csp.CopyHeader(p)
	csp.Packet = requestPacket
	return &csp
}

//Mark Decode
func (c *Cell) VerifySyncResponsePacket(p *sc.NetPacket) *sc.CsPacket {
	fmt.Println("syncResponse p = ", p)
	fmt.Println("syncResponse packet = ", p.Packet)
	var syncData *sc.SyncResponseData
	err := json.Unmarshal(p.Packet, &syncData)
	if err != nil {
		log.Error("syncResponse decode error ", err)
		fmt.Println("syncResponse decode error", err)
		return nil
	}
	var csp sc.CsPacket
	csp.CopyHeader(p)
	csp.Packet = syncData
	return &csp
}



func (c *Cell) VerifyMinorPacket(p *sc.NetPacket) *sc.CsPacket {
	minor := new(cs.MinorBlock)
	err := minor.Deserialize(p.Packet)
	if err != nil {
		log.Error("minor block unmarshal error ", err)
		return nil
	}

	/*missing_func need verify signature here*/

	cm := c.GetLastCMBlock()
	if cm == nil {
		log.Error("need product cm block first")
		return nil
	}

	if cm.Height != minor.CMEpochNo {
		log.Error("minor block epoch error ", minor.CMEpochNo, " current epoch ", cm.Height)
		return nil
	}

	if minor.ShardId < 1 || minor.ShardId > sc.DefaultShardMaxMember {
		log.Error("shard id error ", minor.ShardId)
		return nil
	}

	height := c.getShardHeight(minor.ShardId)
	if height >= minor.Height {
		log.Error("minor block height error ", minor.Height, " last ", height)
		return nil
	}

	var csp sc.CsPacket
	csp.CopyHeader(p)
	csp.Packet = minor

	return &csp
}
