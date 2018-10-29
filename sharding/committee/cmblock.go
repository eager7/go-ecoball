package committee

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/etime"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/types"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

type cmBlockCsi struct {
	bk    *cs.CMBlock
	cache *cs.CMBlock
}

func newCmBlockCsi(bk *cs.CMBlock) *cmBlockCsi {
	return &cmBlockCsi{bk: bk}
}

func (b *cmBlockCsi) GetCsView() *sc.CsView {
	return &sc.CsView{EpochNo: b.bk.Height}
}

func (b *cmBlockCsi) CheckBlock(bl interface{}, bLeader bool) bool {
	update := bl.(*cs.CMBlock)

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

	data, err := b.bk.Serialize()
	if err != nil {
		log.Error("cm block Serialize error ", err)
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

func (b *cmBlockCsi) GetCosign() *types.COSign {
	return b.bk.COSign
}

func (b *cmBlockCsi) GetCandidate() *cs.NodeInfo {
	return nil
}

func (c *committee) reshardWorker(height uint64) (candidate *cs.NodeInfo, shards []cs.Shard) {
	/*missing_func need get deposit account info*/
	//candidate, err := c.ns.Ledger.GetProducerList(config.ChainHash)
	cw := simulate.GetCandidateList()
	if len(cw) > 0 {
		var can cs.NodeInfo
		can.PublicKey = []byte(cw[0].Pubkey)
		can.Address = cw[0].Address
		can.Port = cw[0].Port

		candidate = &can
	} else {
		var can cs.NodeInfo
		back := c.ns.GetBackup()
		if back != nil {
			can.PublicKey = []byte(back.Pubkey)
			can.Address = back.Address
			can.Port = back.Port

			candidate = &can
		} else {
			candidate = nil
		}
	}

	ss := simulate.GetShards()

	var shard cs.Shard
	for i, member := range ss {
		var worker cs.NodeInfo
		worker.PublicKey = []byte(member.Pubkey)
		worker.Address = member.Address
		worker.Port = member.Port

		shard.Member = append(shard.Member, worker)
		if (i+1)%5 == 0 {
			shards = append(shards, shard)
			shard.Member = make([]cs.NodeInfo, 0, 5)
		}
	}

	return
}

func (c *committee) createCommitteeBlock() *cs.CMBlock {
	last := c.ns.GetLastCMBlock()
	var height uint64
	if last == nil {
		panic("last cm block is nil")
		return nil
	}

	height = last.Height + 1

	header := cs.CMBlockHeader{
		ChainID:      config.ChainHash,
		Version:      last.Version,
		Height:       height,
		Timestamp:    time.Now().UnixNano(),
		PrevHash:     last.Hash(),
		LeaderPubKey: nil,
		Nonce:        0,
		Candidate:    cs.NodeInfo{},
		ShardsHash:   common.Hash{},
		COSign:       nil,
	}

	cosign := &types.COSign{}
	cosign.Step1 = 1
	cosign.Step2 = 0

	header.COSign = cosign

	candidate, shards := c.reshardWorker(height)
	if candidate != nil {
		header.Candidate.PublicKey = candidate.PublicKey
		header.Candidate.Address = candidate.Address
		header.Candidate.Port = candidate.Port
	}

	cmb, err := cs.NewCmBlock(header, shards)
	if err != nil {
		log.Error("new cm block err ", err)
		return nil
	}

	log.Debug("create cm block height ", cmb.Height)

	return cmb

}

func (c *committee) productCommitteeBlock(msg interface{}) {
	etime.StopTime(c.stateTimer)

	cm := c.createCommitteeBlock()
	if cm == nil {
		return
	}

	cms := newCmBlockCsi(cm)

	c.cs.StartConsensus(cms, sc.DefaultCmBlockWindow*time.Millisecond)

	c.stateTimer.Reset(sc.DefaultProductCmBlockTimer * time.Second)
}

func (c *committee) checkCmPacket(p interface{}) bool {
	/*check block*/
	csp := p.(*sc.CsPacket)
	if csp.BlockType != sc.SD_CM_BLOCK {
		log.Error("it is not cm block, drop it")
		return false
	}

	cm := csp.Packet.(*cs.CMBlock)
	last := c.ns.GetLastCMBlock()
	if last != nil && cm.Height <= last.Height {
		log.Error("old cm block, drop it")
		return false
	}

	return true
}

func (c *committee) processConsensusCmPacket(p interface{}) {
	log.Debug("process cm consensus packet")

	if !c.checkCmPacket(p) {
		return
	}

	c.cs.ProcessPacket(p.(*sc.CsPacket))
}

func (c *committee) commitCmBlock(bl *cs.CMBlock) {
	log.Debug("recv consensus cm block height ", bl.Height)
	simulate.TellBlock(bl)

	c.ns.SaveLastCMBlock(bl)
	c.fsm.Execute(ActCollectMinorBlock, nil)
}
