package committee

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

type vcBlockCsi struct {
	bk    *cs.ViewChangeBlock
	cache *cs.ViewChangeBlock
}

func newVcBlockCsi(block *cs.ViewChangeBlock, sign uint32) *vcBlockCsi {
	block.Step1 = sign
	block.Step2 = sign

	return &vcBlockCsi{bk: block}
}

func (b *vcBlockCsi) GetCsView() *sc.CsView {
	return &sc.CsView{EpochNo: b.bk.CMEpochNo, FinalHeight: b.bk.FinalBlockHeight, Round: b.bk.Round}
}

func (b *vcBlockCsi) CheckBlock(bl interface{}, bLeader bool) bool {
	update := bl.(*cs.ViewChangeBlock)

	if b.bk.CMEpochNo != update.CMEpochNo || b.bk.FinalBlockHeight != update.FinalBlockHeight || b.bk.Round != update.Round {
		log.Error("view error current ", b.bk.CMEpochNo, " ", b.bk.FinalBlockHeight, " ", b.bk.Round, " packet view ", update.CMEpochNo, " ", update.FinalBlockHeight, " ", update.Round)
		return false
	}

	if !sc.Same(b.bk.Candidate.PublicKey, update.Candidate.PublicKey) {
		log.Error("candidate public key not same")
		return false
	}

	if update.Candidate.Address != b.bk.Candidate.Address {
		log.Error("candidate address not same")
		return false
	}

	if update.Candidate.Port != b.bk.Candidate.Port {
		log.Error("candidate port not same")
		return false
	}

	if bLeader {
		b.cache = update
	} else {
		b.bk = update
	}

	return true
}

func (b *vcBlockCsi) MakeNetPacket(step uint16) *sc.NetPacket {
	csp := &sc.NetPacket{PacketType: mpb.Identify_APP_MSG_CONSENSUS_PACKET, BlockType: sc.SD_VIEWCHANGE_BLOCK, Step: step}

	data, err := b.bk.Serialize()
	if err != nil {
		log.Error("vc block Serialize error ", err)
		return nil
	}

	csp.Packet = data

	return csp
}

func (b *vcBlockCsi) GetCsBlock() interface{} {
	return b.bk
}

func (b *vcBlockCsi) PrepareRsp() uint32 {
	log.Debug("prepare receive consign ", b.cache.Step1)

	b.bk.Step1 |= b.cache.Step1
	return b.bk.Step1
}

func (b *vcBlockCsi) PrecommitRsp() uint32 {
	log.Debug("precommit receive consign ", b.cache.Step2)
	b.bk.Step2 |= b.cache.Step2

	return b.bk.Step2
}

func (b *vcBlockCsi) GetCosign() *types.COSign {
	return b.bk.COSign
}

func (b *vcBlockCsi) GetCandidate() *cs.NodeInfo {
	return &b.bk.Candidate
}

func (c *committee) createVcBlock() (*cs.ViewChangeBlock, bool) {
	lastcm := c.ns.GetLastCMBlock()
	lastfinal := c.ns.GetLastFinalBlock()
	lastvc := c.ns.GetLastViewchangeBlock()

	var round uint16
	var epoch uint64
	var fheight uint64
	var height uint64

	if lastcm != nil {
		epoch = lastcm.Height
	}

	if lastfinal != nil {
		fheight = lastfinal.Height
	}

	if lastvc == nil {
		height = 1
		round = 1
	} else {
		if lastfinal != nil && lastfinal.Height > lastvc.FinalBlockHeight {
			round = 1
		} else {
			round = lastvc.Round + 1
		}

		height = lastvc.Height + 1
	}

	works := c.ns.GetWorks()
	length := len(works)
	if length == 0 {
		panic("works is empty")
		return nil, false
	} else if length == 1 {
		return nil, false
	}

	log.Debug("create vc block epoch ", epoch, " fheight ", fheight, " round ", round, " height ", height)
	vc := &cs.ViewChangeBlock{
		ViewChangeBlockHeader: cs.ViewChangeBlockHeader{
			CMEpochNo:        0,
			FinalBlockHeight: 0,
			Round:            0,
			Candidate:        cs.NodeInfo{},
			Timestamp:        0,
			COSign:           nil,
		},
	}

	vc.Height = height
	vc.CMEpochNo = epoch
	vc.FinalBlockHeight = fheight
	vc.Round = round

	works = works[1:]
	length--
	i := c.vccount % uint16(length)
	candi := works[i]

	vc.Candidate.PublicKey = []byte(candi.Pubkey)
	vc.Candidate.Address = candi.Address
	vc.Candidate.Port = candi.Port

	cosign := &types.COSign{}
	cosign.Step1 = 1
	cosign.Step2 = 0

	vc.COSign = cosign

	log.Debug(" create view change block epoch ", vc.CMEpochNo, " height ", vc.FinalBlockHeight, " round ", vc.Round, " candidate address ", candi.Address, " port ", candi.Port)

	if c.ns.Self.Equal(candi) {
		return vc, true
	} else {
		return vc, false
	}

}

func (c *committee) productViewChangeBlock(msg interface{}) {
	c.stateTimer.Stop()

	vc, bCandidate := c.createVcBlock()
	if vc == nil {
		return
	}

	sign := c.ns.GetSignBit()
	vci := newVcBlockCsi(vc, sign)

	c.cs.StartVcConsensus(vci, sc.DefaultViewchangeBlockWindow*time.Millisecond, bCandidate)

	c.stateTimer.Reset(time.Duration(sc.DefaultProductViewChangeBlockTimer*(c.vccount+1)*(c.vccount+1)) * time.Second)
}

func (c *committee) checkVcPacket(p interface{}) bool {
	/*check block*/
	csp := p.(*sc.CsPacket)
	if csp.BlockType != sc.SD_VIEWCHANGE_BLOCK {
		log.Error("it is not vc block, drop it")
		return false
	}

	vc := csp.Packet.(*cs.ViewChangeBlock)
	last := c.ns.GetLastViewchangeBlock()
	lastfinal := c.ns.GetLastFinalBlock()
	if last == nil || lastfinal == nil {
		panic("last block is nil")
		return false
	}

	if last.Height >= vc.Height {
		log.Debug("old vc packet height ", vc.Height, " last height ", last.Height)
		return false
	} else if vc.Height > last.Height+1 {
		log.Debug("vc packet height ", vc.Height, " last height ", last.Height, " need sync")
		c.fsm.Execute(ActChainNotSync, nil)
	}

	if vc.FinalBlockHeight < lastfinal.Height {
		log.Debug("wrong vc packet final height ", vc.FinalBlockHeight, " last final height", lastfinal.Height)
		return false
	} else if vc.FinalBlockHeight == lastfinal.Height {
		if last.FinalBlockHeight == vc.FinalBlockHeight {
			if last.Round >= vc.Round {
				log.Debug("old vc packet vc round ", vc.Round, " last round ", last.Round)
				return false
			} else if vc.Round > last.Round+1 {
				log.Debug("vc round ", vc.Round, " last round ", last.Round, " need sync")
				c.fsm.Execute(ActChainNotSync, nil)
				return false
			}
		} else {
			if vc.Round > 1 {
				log.Debug("vc round ", vc.Round, " need sync")
				c.fsm.Execute(ActChainNotSync, nil)
				return false
			} else if vc.Round < 1 {
				log.Debug("wrong round ", vc.Round)
				return false
			}
		}
	} else {
		log.Debug("last final ", lastfinal.Height, " recv view change final height ", vc.FinalBlockHeight, " need sync")
		c.fsm.Execute(ActChainNotSync, nil)
		return false
	}
	return true
}

func (c *committee) processViewchangeConsensusPacket(p interface{}) {
	log.Debug("process view change consensus block")

	if !c.checkVcPacket(p) {
		return
	}

	c.cs.ProcessPacket(p.(*sc.CsPacket))
}

func (c *committee) commitViewchangeBlock(vc *cs.ViewChangeBlock) {
	log.Debug("recv consensus view change block epoch ", vc.CMEpochNo, " height ", vc.FinalBlockHeight, " round  ", vc.Round)

	lastcm := c.ns.GetLastCMBlock()
	if lastcm == nil {
		panic("cm block is nil")
		return
	}

	lastfinal := c.ns.GetLastFinalBlock()
	if lastfinal == nil {
		panic("final block is nil")
		return
	}

	if lastcm.Height != vc.CMEpochNo {

		return
	}

	if lastfinal.Height != vc.FinalBlockHeight {
		return
	}


	simulate.TellBlock(vc)

	c.ns.SaveLastViewchangeBlock(vc)
	c.resetVcCounter(nil)

	if lastcm.Height == 1 {
		c.fsm.Execute(ActProductCommitteeBlock, nil)
	} else if lastcm.Height > lastfinal.EpochNo {

		c.fsm.Execute(ActProductFinalBlock, nil)
	} else {
		if lastfinal.Height%sc.DefaultEpochFinalBlockNumber == 0 {
			c.fsm.Execute(ActProductCommitteeBlock, nil)
		} else {
			c.fsm.Execute(ActProductFinalBlock, nil)
		}
	}

}

func (c *committee) resetVcCounter(p interface{}) bool {
	c.vccount = 0
	return true
}

func (c *committee) increaseCounter(p interface{}) bool {
	c.vccount++
	return true
}
