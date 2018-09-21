package committee

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/etime"
	"github.com/ecoball/go-ecoball/core/types"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

type cmBlockCsi struct {
	bk    *types.CMBlock
	cache *types.CMBlock
}

func newCmBlockCsi(bk *types.CMBlock) *cmBlockCsi {
	return &cmBlockCsi{bk: bk}
}

func (b *cmBlockCsi) GetCsView() *sc.CsView {
	return &sc.CsView{EpochNo: b.bk.Height}
}

func (b *cmBlockCsi) CheckBlock(bl interface{}, bLeader bool) bool {
	update := bl.(*types.CMBlock)

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

	if update.Height != b.bk.Height {
		log.Error("view error current ", b.bk.Height, " packet view ", update.Height)
		return false
	}

	if !sc.Same(update.LeaderPubKey, b.bk.LeaderPubKey) {
		log.Error("leader public key not same")
		return false
	}

	if bLeader {
		b.cache = update
	} else {
		b.bk = update
	}

	return true
}

func (b *cmBlockCsi) MakeNetPacket(step uint16) *sc.NetPacket {
	csp := &sc.NetPacket{PacketType: netmsg.APP_MSG_CONSENSUS_PACKET, BlockType: sc.SD_CM_BLOCK, Step: step}

	/*missing_func should fill in signature and bit map*/
	if step == consensus.StepPrePare {
		log.Debug("make cm prepare block")
		b.bk.Step1 = 1
	} else if step == consensus.StepPreCommit {
		log.Debug("make cm precommit block")
		b.bk.Step2 = 1
	} else if step == consensus.StepCommit {
		log.Debug("make cm commit block")
	} else {
		log.Fatal("step wrong")
		return nil
	}

	data, err := json.Marshal(b.bk)
	if err != nil {
		log.Error("cm block marshal error ", err)
		return nil
	}

	csp.Packet = data

	return csp
}

func (b *cmBlockCsi) GetCsBlock() interface{} {
	return b.bk
}

func (b *cmBlockCsi) PrepareRsp() uint32 {
	if b.cache.Step1 == 1 {
		b.bk.Step1++
	}

	return b.bk.Step1
}

func (b *cmBlockCsi) PrecommitRsp() uint32 {
	if b.cache.Step2 == 1 {
		b.bk.Step2++
	}

	return b.bk.Step2
}

func (b *cmBlockCsi) GetCandidate() *types.NodeInfo {
	return nil
}

func (c *committee) createCommitteeBlock() *types.CMBlock {
	last := c.ns.GetLastCMBlock()
	var height uint64
	if last == nil {
		height = 1
	} else {
		height = last.Height + 1
	}

	log.Debug("create cm block height ", height)

	cm := &types.CMBlock{}
	cm.Height = height

	cosign := &types.COSign{}
	cosign.Step1 = 1
	cosign.Step2 = 0

	cm.COSign = cosign

	candidate, err := c.ns.Ledger.GetProducerList(config.ChainHash)
	if err == nil && candidate != nil && len(candidate) > 0 {
		panic("missing_func")
		/*missing_func need get account info*/
		//cm.Candidate.PublicKey = []byte(candidate[0].Pubkey)
		//cm.Candidate.Address = candidate[0].Address
		//cm.Candidate.Port = candidate[0].Port
	} else {
		/*missing_func there is no candidate maybe we can select new leader by vrf*/
		backup := c.ns.GetBackup()
		if backup != nil {
			cm.Candidate.PublicKey = []byte(backup.Pubkey)
			cm.Candidate.Address = backup.Address
			cm.Candidate.Port = backup.Port
		}
	}

	return cm

}

func (c *committee) productCommitteeBlock(msg interface{}) {
	etime.StopTime(c.stateTimer)

	cm := c.createCommitteeBlock()

	cms := newCmBlockCsi(cm)

	c.cs.StartConsensus(cms)

	c.stateTimer.Reset(sc.DefaultProductCmBlockTimer * time.Second)
}

func (c *committee) recheckCmPacket(p interface{}) bool {
	/*recheck block*/
	csp := p.(*sc.CsPacket)
	if csp.BlockType != sc.SD_CM_BLOCK {
		log.Error("it is not cm block, drop it")
		return false
	}

	cm := csp.Packet.(*types.CMBlock)
	last := c.ns.GetLastCMBlock()
	if last != nil && cm.Height <= last.Height {
		log.Error("old cm block, drop it")
		return false
	}

	return true
}

func (c *committee) processConsensusCmPacket(p interface{}) {
	log.Debug("process cm consensus packet")

	if !c.recheckCmPacket(p) {
		return
	}

	c.cs.ProcessPacket(p.(*sc.CsPacket))
}

func (c *committee) recvCommitCmBlock(bl *types.CMBlock) {
	log.Debug("recv consensus cm block height ", bl.Height)
	simulate.TellBlock(bl)

	c.ns.SetLastCMBlock(bl)
	c.fsm.Execute(ActWaitMinorBlock, nil)
}
