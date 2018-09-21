package committee

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/etime"
	"github.com/ecoball/go-ecoball/core/types"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

type vcBlockCsi struct {
	bk    *types.ViewChangeBlock
	cache *types.ViewChangeBlock
}

func newVcBlockCsi(block *types.ViewChangeBlock) *vcBlockCsi {
	return &vcBlockCsi{bk: block}
}

func (b *vcBlockCsi) GetCsView() *sc.CsView {
	return &sc.CsView{EpochNo: b.bk.CMEpochNo, FinalHeight: b.bk.FinalBlockHeight, Round: b.bk.Round}
}

func (b *vcBlockCsi) CheckBlock(bl interface{}, bLeader bool) bool {
	update := bl.(*types.ViewChangeBlock)

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
	csp := &sc.NetPacket{PacketType: netmsg.APP_MSG_CONSENSUS_PACKET, BlockType: sc.SD_VIEWCHANGE_BLOCK, Step: step}

	/*missing_func should fill in signature and bit map*/
	if step == consensus.StepPrePare {
		log.Debug("make prepare vc block")
		b.bk.Step1 = 1
	} else if step == consensus.StepPreCommit {
		log.Debug("make precommit vc block")
		b.bk.Step2 = 1
	} else if step == consensus.StepCommit {
		log.Debug("make commit vc block")
	} else {
		log.Fatal("step wrong")
		return nil
	}

	data, err := json.Marshal(b.bk)
	if err != nil {
		log.Error("vc block marshal error ", err)
		return nil
	}

	csp.Packet = data

	return csp
}

func (b *vcBlockCsi) GetCsBlock() interface{} {
	return b.bk
}

func (b *vcBlockCsi) PrepareRsp() uint32 {
	if b.cache.Step1 == 1 {
		b.bk.Step1++
	}

	return b.bk.Step1
}

func (b *vcBlockCsi) PrecommitRsp() uint32 {
	if b.cache.Step2 == 1 {
		b.bk.Step2++
	}

	return b.bk.Step2
}

func (b *vcBlockCsi) GetCandidate() *types.NodeInfo {
	return &b.bk.Candidate
}

func (c *committee) createVcBlock() (*types.ViewChangeBlock, bool) {
	lastcm := c.ns.GetLastCMBlock()
	lastfinal := c.ns.GetLastFinalBlock()
	lastvc := c.ns.GetLastViewchangeBlock()

	var round uint16
	var epoch uint64
	var height uint64

	if lastcm != nil {
		epoch = lastcm.Height
	}

	if lastfinal != nil {
		height = lastfinal.Height
	}

	if lastvc == nil {
		round = 1
	} else {
		if lastfinal != nil && lastfinal.Height > lastvc.FinalBlockHeight {
			round = 1
		} else {
			round = lastvc.Round + 1
		}
	}

	works := c.ns.GetWorks()
	length := len(works)
	if length == 0 {
		panic("works is empty")
		return nil, false
	} else if length == 1 {
		return nil, false
	}

	log.Debug("create vc block epoch ", epoch, " height ", height, " round ", round)
	vc := &types.ViewChangeBlock{}
	vc.CMEpochNo = epoch
	vc.FinalBlockHeight = height
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

	log.Debug("candidate address ", candi.Address, " port ", candi.Port)

	if c.ns.Self.Equal(candi) {
		return vc, true
	} else {
		return vc, false
	}
}

func (c *committee) productViewChangeBlock(msg interface{}) {
	etime.StopTime(c.stateTimer)

	vc, bCandi := c.createVcBlock()
	if vc == nil {
		return
	}

	vci := newVcBlockCsi(vc)

	c.cs.StartVcConsensus(vci, bCandi)

	c.stateTimer.Reset(time.Duration(sc.DefaultProductViewChangeBlockTimer*(c.vccount+1)) * time.Second)
}

func (c *committee) recheckVcPacket(p interface{}) bool {
	/*recheck block*/
	csp := p.(*sc.CsPacket)
	if csp.BlockType != sc.SD_VIEWCHANGE_BLOCK {
		log.Error("it is not vc block, drop it")
		return false
	}

	return true
}

func (c *committee) processViewchangeConsensusPacket(p interface{}) {
	log.Debug("process view change consensus block")

	if !c.recheckVcPacket(p) {
		return
	}

	c.cs.ProcessPacket(p.(*sc.CsPacket))
}

func (c *committee) recvCommitViewchangeBlock(bl *types.ViewChangeBlock) {
	log.Debug("recv consensus view change block epoch ", bl.CMEpochNo, " height ", bl.FinalBlockHeight, " round  ", bl.Round)
	simulate.TellBlock(bl)

	c.ns.SetLastViewchangeBlock(bl)
	c.resetVcCounter(nil)

	lastcm := c.ns.GetLastCMBlock()
	if lastcm.Height > bl.CMEpochNo {
		c.fsm.Execute(ActProductFinalBlock, nil)
	} else {
		if bl.FinalBlockHeight%sc.DefaultEpochFinalBlockNumber == 0 {
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
