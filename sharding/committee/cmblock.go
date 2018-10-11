package committee

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common"
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

func (c *committee) reshardWorker() (candidate *types.NodeInfo, shards []types.Shard) {
	/*missing_func need get deposit account info*/
	//candidate, err := c.ns.Ledger.GetProducerList(config.ChainHash)

	var can types.NodeInfo
	backup := c.ns.GetBackup()
	if backup != nil {
		can.PublicKey = []byte(backup.Pubkey)
		can.Address = backup.Address
		can.Port = backup.Port

		candidate = &can
	} else {
		candidate = nil
	}

	ss := simulate.GetShards()
	var shard types.Shard
	for _, member := range ss {
		var worker types.NodeInfo
		worker.PublicKey = []byte(member.Pubkey)
		worker.Address = member.Address
		worker.Port = member.Port

		shard.Member = append(shard.Member, worker)
	}

	if len(ss) > 0 {
		shards = append(shards, shard)
	}

	return
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

	cosign := &types.COSign{}
	cosign.Step1 = 1
	cosign.Step2 = 0

	header := types.CMBlockHeader{
		ChainID:      common.Hash{},
		Version:      0,
		Height:       0,
		Timestamp:    0,
		PrevHash:     common.Hash{},
		LeaderPubKey: nil,
		Nonce:        0,
		Candidate:    types.NodeInfo{},
		ShardsHash:   common.Hash{},
		COSign:       nil,
	}
	header.Height = height
	header.COSign = cosign

	candidate, shards := c.reshardWorker()
	if candidate != nil {
		header.Candidate.PublicKey = candidate.PublicKey
		header.Candidate.Address = candidate.Address
		header.Candidate.Port = candidate.Port
	}

	cmb := &types.CMBlock{
		CMBlockHeader: header,
		Shards:        make([]types.Shard, len(shards)),
	}

	copy(cmb.Shards, shards)

	//for i, shard := range shards {
	//	cmb.Shards[i].Id = shard.Id
	//	cmb.Shards[i].Member = make([]types.NodeInfo, len(shard.Member))
	//	for j, node := range shard.Member {
	//		cmb.Shards[i].Member[j].PublicKey = node.PublicKey
	//		cmb.Shards[i].Member[j].Port = node.Port
	//		cmb.Shards[i].Member[j].Address = node.Address
	//	}
	//}

	return cmb

}

func (c *committee) productCommitteeBlock(msg interface{}) {
	etime.StopTime(c.stateTimer)

	cm := c.createCommitteeBlock()

	cms := newCmBlockCsi(cm)

	c.cs.StartConsensus(cms)

	c.stateTimer.Reset(sc.DefaultProductCmBlockTimer * time.Second)
}

func (c *committee) checkCmPacket(p interface{}) bool {
	/*check block*/
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

	if !c.checkCmPacket(p) {
		return
	}

	c.cs.ProcessPacket(p.(*sc.CsPacket))
}

func (c *committee) recvCommitCmBlock(bl *types.CMBlock) {
	log.Debug("recv consensus cm block height ", bl.Height)
	simulate.TellBlock(bl)

	c.ns.SaveLastCMBlock(bl)
	c.fsm.Execute(ActWaitMinorBlock, nil)
}
