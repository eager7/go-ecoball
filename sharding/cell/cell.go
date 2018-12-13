package cell

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	cs "github.com/ecoball/go-ecoball/core/shard"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
)

var (
	log = elog.NewLogger("sharding", elog.DebugLog)
)

type Cell struct {
	NodeType int
	Shardid  uint16 /*only node is shard member*/
	Self     sc.Worker
	cm       *workerSet
	shard    []*sc.Worker
	candidate    []*sc.Worker
	//ss        *shardSet
	//nodes     *workerMap
	//candidate workerSet

	/*last chain data*/
	chain          *chainData
	minorBlockPool *minorBlockSet

	Ledger ledger.Ledger
	Topoc  chan interface{}
}

func MakeCell(l ledger.Ledger) *Cell {
	return &Cell{
		cm:             makeWorkerSet(sc.DefaultCommitteMaxMember),
		chain:          makeChainData(),
		minorBlockPool: makeMinorBlockSet(),
		Ledger:         l,
		Topoc:          make(chan interface{}),
	}
}

func (c *Cell) LoadConfig() {
	self := simulate.GetNodeInfo()
	(&c.Self).Copy(&self)

	cmt := simulate.GetCommittee()

	nodeType := sc.NodeNil
	for _, member := range cmt {
		var worker sc.Worker
		worker.Pubkey = member.Pubkey
		worker.Address = member.Address
		worker.Port = member.Port

		c.addCommitteWorker(&worker)
		if c.Self.Equal(&worker) {
			nodeType = sc.NodeCommittee
		}
	}

	//if nodeType == sc.NodeNil {
		can := simulate.GetCandidate()
		for _, member := range can {
			var worker sc.Worker
			worker.Pubkey = member.Pubkey
			worker.Address = member.Address
			worker.Port = member.Port

			c.addCandidateWorker(&worker)
			if c.Self.Equal(&worker) {
				nodeType = sc.NodeCandidate
			}
		}
	//}

	if nodeType == sc.NodeNil {
		nodeType = sc.NodeShard
	}

	c.NodeType = nodeType
}

func (c *Cell) LoadLastBlock() {
	lastCmBlock, _, err := c.Ledger.GetLastShardBlock(config.ChainHash, cs.HeCmBlock)
	if err != nil || lastCmBlock == nil {
		panic("get cm block error ")
		return
	}

	cm := lastCmBlock.GetObject().(cs.CMBlock)
	c.SyncCmBlockComplete(&cm)

	lastvc, _, err := c.Ledger.GetLastShardBlock(config.ChainHash, cs.HeViewChange)
	if err != nil || lastvc == nil {
		panic("get vc block error ")
		return
	}

	vc := lastvc.GetObject().(cs.ViewChangeBlock)
	c.SaveLastViewchangeBlock(&vc)

	lastFinalBlock, _, err := c.Ledger.GetLastShardBlock(config.ChainHash, cs.HeFinalBlock)
	if err != nil || lastFinalBlock == nil {
		panic("get final block error ")
		return
	}

	final := lastFinalBlock.GetObject().(cs.FinalBlock)
	c.SaveLastFinalBlock(&final)

	if c.NodeType == sc.NodeShard || c.NodeType == sc.NodeCandidate {
		lastMinor, bFinalize, err := c.Ledger.GetLastShardBlock(config.ChainHash, cs.HeMinorBlock)
		if err != nil || lastMinor == nil {
			panic("get minor block error ")
			return
		}

		minor := lastMinor.GetObject().(cs.MinorBlock)

		if !bFinalize {
			last, finalize, err := c.Ledger.GetShardBlockByHash(config.ChainHash, cs.HeMinorBlock, minor.PrevHash, true)
			if err != nil || finalize != true {
				log.Error("get last finalize minor block error", err)
				panic("get last finalize minor block error")
				return
			}

			minor = last.GetObject().(cs.MinorBlock)
		}

		c.SaveLastMinorBlock(&minor)
	}

}

func (c *Cell) SaveLastCMBlock(bk *cs.CMBlock) {
	log.Debug("save cm block epoch ", bk.Height)

	c.chain.setCMBlock(bk)

	worker := &sc.Worker{}
	if len(bk.Candidate.PublicKey) != 0 {
		worker.InitWork(&bk.Candidate)
		c.addCommitteWorker(worker)
	}

	if c.IsCommitteeMember() {
		c.NodeType = sc.NodeCommittee
		c.minorBlockPool.clean()
	} else {
		if c.NodeType == sc.NodeCommittee {
			log.Error("we are not in committee now, restart ")
			panic("we are not in committee now, restart ")
		}

		if bk.Height > 1 {
			c.saveShardsInfoFromCMBlock(bk)
		}
	}

	c.createShardingTopo()

}

func (c *Cell) createShardingTopo() {
	topo := &sc.ShardingTopo{ShardId: c.Shardid, Pubkey: c.Self.Pubkey}

	lastcm := c.GetLastCMBlock()
	if lastcm == nil {
		panic("last cm block is nil")
		return
	}

	total := len(lastcm.Shards) + 2
	topo.ShardingInfo = make([][]sc.Worker, total)

	for _, member := range c.cm.member {
		var worker sc.Worker
		worker = *member
		topo.ShardingInfo[0] = append(topo.ShardingInfo[0], worker)
	}

	for i, shard := range lastcm.Shards {
		for _, member := range shard.Member {
			var worker sc.Worker
			(&worker).InitWork(&member)
			topo.ShardingInfo[i+1] = append(topo.ShardingInfo[i+1], worker)
		}
	}

	for _, member := range c.candidate {
		var worker sc.Worker
		worker = *member
		topo.ShardingInfo[total - 1] = append(topo.ShardingInfo[total - 1], worker)
	}

	log.Debug("send sharding topo to channel ", topo.ShardId, " len ", len(topo.ShardingInfo))
	c.Topoc <- topo
}

func (c *Cell) GetLastCMBlock() *cs.CMBlock {
	return c.chain.getCMBlock()
}

func (c *Cell) SaveLastFinalBlock(bk *cs.FinalBlock) {
	log.Debug("save final block epoch ", bk.EpochNo, " height ", bk.Height)

	cur := c.GetLastFinalBlock()
	if cur != nil {
		if cur.Height >= bk.Height {
			log.Debug("have saved last final block")
			return
		}
	}

	c.chain.setFinalBlock(bk)
	inFinal := false
	for _, minor := range bk.MinorBlocks {
		log.Debug("minor block shard id ", minor.ShardId, " height ", minor.Height)
		if uint32(c.Shardid) == minor.ShardId {
			inFinal = true
			c.chain.saveMinorBlock(minor)
		} else {
			c.chain.setShardHeight(minor.ShardId, minor.Height)
		}
	}

	if !inFinal {
		c.chain.preMinorBlock = nil
	}

	c.minorBlockPool.clean()
}

func (c *Cell) GetLastFinalBlock() *cs.FinalBlock {
	return c.chain.getFinalBlock()
}

func (c *Cell) SaveLastViewchangeBlock(bk *cs.ViewChangeBlock) {
	log.Debug("save view change block epoch ", bk.CMEpochNo, " height ", bk.FinalBlockHeight, " round ", bk.Round)

	cur := c.GetLastViewchangeBlock()
	if cur != nil {
		if cur.Height >= bk.Height {
			log.Debug("have saved last view change block")
			return
		}
	}

	leader := &sc.Worker{}
	leader.InitWork(&bk.Candidate)
	log.Debug("new leader ", leader.Address, " ", leader.Port)

	c.cm.changeLeader(leader)
	c.chain.setViewchangeBlock(bk)
}

func (c *Cell) GetLastViewchangeBlock() *cs.ViewChangeBlock {
	return c.chain.getViewchangeBlock()
}

func (c *Cell) SaveLastMinorBlock(bk *cs.MinorBlock) {
	log.Debug("save last minor block ", bk.Height)
	c.chain.setMinorBlock(bk)
}

func (c *Cell) GetLastMinorBlock() *cs.MinorBlock {
	return c.chain.getMinorBlock()
}

func (c *Cell) SavePreMinorBlock(bk *cs.MinorBlock) {
	c.chain.setPreMinorBlock(bk)
	c.minorBlockPool.saveMinorBlock(bk)
}

func (c *Cell) GetPreMinorBlock() *cs.MinorBlock {
	return c.chain.getPreMinorBlock()
}

func (c *Cell) SyncCmBlockComplete(lastCmblock *cs.CMBlock) {
	curBlock := c.chain.getCMBlock()

	var i uint64
	if curBlock == nil {
		if lastCmblock.Height > sc.DefaultCommitteMaxMember {
			i = lastCmblock.Height - sc.DefaultCommitteMaxMember + 1
		} else {
			i = 1
		}
	} else if curBlock.Height >= lastCmblock.Height {
		log.Debug("cm block is already sync")
		return
	} else if curBlock.Height+sc.DefaultCommitteMaxMember >= lastCmblock.Height {
		i = curBlock.Height + 1
	} else {
		i = lastCmblock.Height - sc.DefaultCommitteMaxMember + 1
	}

	for ; i < lastCmblock.Height; i++ {
		block, _, err := c.Ledger.GetShardBlockByHeight(config.ChainHash, cs.HeCmBlock, i, 0)
		if err != nil {
			log.Error("get block error ", err)
			return
		}

		cm := block.GetObject().(cs.CMBlock)

		var worker sc.Worker
		if len(cm.Candidate.PublicKey) != 0 {
			worker.Pubkey = string(cm.Candidate.PublicKey)
			worker.Address = cm.Candidate.Address
			worker.Port = cm.Candidate.Port

			c.addCommitteWorker(&worker)
		} else {
			log.Error("cm block candidate is nil")
		}
	}

	c.SaveLastCMBlock(lastCmblock)
}

func (c *Cell) SaveMinorBlockToPool(minor *cs.MinorBlock) bool {
	return c.minorBlockPool.saveMinorBlock(minor)
}

func (c *Cell) SyncMinorsBlockToPool(minors []*cs.MinorBlock) {
	c.minorBlockPool.syncMinorBlocks(minors)
}

func (c *Cell) GetMinorBlockHashesFromPool() []common.Hash {
	return c.minorBlockPool.getMinorBlockHashes()
}

func (c *Cell) IsMinorBlockThresholdInPool() bool {
	cm := c.chain.cmBlock
	if cm == nil {
		return true
	}

	if c.minorBlockPool.count() == uint16(len(cm.Shards)*sc.DefaultThresholdOfMinorBlock/100) {
		return true
	} else {
		return false
	}
}

func (c *Cell) IsMinorBlockFullInPool() bool {
	cm := c.chain.cmBlock
	if cm == nil {
		return true
	}

	if c.minorBlockPool.count() == uint16(len(cm.Shards)) {
		return true
	} else {
		return false
	}
}

func (c *Cell) IsLeader() bool {
	if c.NodeType == sc.NodeCommittee {
		return c.cm.isLeader(&c.Self)
	} else if c.NodeType == sc.NodeShard {
		return c.isShardLeader()
	} else {
		return false
	}
}

func (c *Cell) isShardLeader() bool {
	if len(c.shard) == 0 {
		return false
	}

	i := c.CalcShardLeader(len(c.shard))

	if (&c.Self).Equal(c.shard[i]) {
		return true
	} else {
		return false
	}
}

func (c *Cell) IsBackup() bool {
	if c.NodeType == sc.NodeCommittee {
		return c.cm.isBackup(&c.Self)
	} else if c.NodeType == sc.NodeShard {
		return c.isShardBackup()
	} else {
		return false
	}
}

func (c *Cell) isShardBackup() bool {
	if len(c.shard) <= 1 {
		return false
	}

	i := c.CalcShardBackup(len(c.shard))

	if (&c.Self).Equal(c.shard[i]) {
		return true
	} else {
		return false
	}
}

func (c *Cell) SelfIndex() int {
	if c.NodeType == sc.NodeCommittee {
		return c.cm.SelfIndex(&c.Self)
	} else if c.NodeType == sc.NodeShard {
		return c.shardSelfIndex()
	} else {
		return 0
	}
}

func (c *Cell) shardSelfIndex() int {
	for i, work := range c.shard {
		if c.Self.Equal(work) {
			return i
		}
	}

	panic("not shard member")
	return 0
}

func (c *Cell) IsCommitteeMember() bool {
	return c.cm.isMember(&c.Self)
}

func (c *Cell) GetCmWorks() []*sc.Worker {
	return c.cm.member
}

func (c *Cell) GetCmWorksCounter() int {
	return len(c.cm.member)
}

func (c *Cell) GetWorks() []*sc.Worker {
	if c.NodeType == sc.NodeCommittee {
		return c.cm.member
	} else if c.NodeType == sc.NodeShard {
		return c.shard
	} else if c.NodeType == sc.NodeCandidate {
		return c.candidate
	} else {
		return nil
	}
}

func (c *Cell) GetCandidateWorks() []*sc.Worker {
	return c.candidate
}

func (c *Cell) GetWorksCounter() uint32 {
	if c.NodeType == sc.NodeCommittee {
		return uint32(len(c.cm.member))
	} else if c.NodeType == sc.NodeShard {
		return uint32(len(c.shard))
	} else if c.NodeType == sc.NodeCandidate {
		return uint32(len(c.candidate))
	} else {
		return 0
	}
}

func (c *Cell) GetLeader() *sc.Worker {
	if c.NodeType == sc.NodeCommittee {
		return c.cm.member[0]
	} else if c.NodeType == sc.NodeShard {
		i := c.CalcShardLeader(len(c.shard))
		return c.shard[i]
	} else {
		return nil
	}
}

func (c *Cell) GetCommitteeBackup() *sc.Worker {
	if len(c.cm.member) > 1 {
		return c.cm.member[1]
	} else {
		return nil
	}
}

func (c *Cell) addCommitteWorker(worker *sc.Worker) {
	log.Debug("add committee worker key ", worker.Pubkey, " address ", worker.Address, " port ", worker.Port)
	backup := c.GetCommitteeBackup()
	if backup != nil && backup.Equal(worker) {
		c.cm.popLeader()
	} else {
		c.cm.addMember(worker)
	}
}

func (c *Cell) addCandidateWorker(worker *sc.Worker) {
	log.Debug("add candidate worker key ", worker.Pubkey, " address ", worker.Address, " port ", worker.Port)

	c.candidate = append(c.candidate, worker)
}

func (c *Cell) saveShardsInfoFromCMBlock(cmb *cs.CMBlock) {
	c.NodeType = sc.NodeCandidate
	c.shard = c.shard[:0]

	for i, shard := range cmb.Shards {
		for _, member := range shard.Member {
			var worker sc.Worker
			worker.Pubkey = string(member.PublicKey)
			worker.Address = member.Address
			worker.Port = member.Port

			if c.Self.Equal(&worker) {
				c.NodeType = sc.NodeShard
				c.Shardid = uint16(i) + 1
				log.Debug("worker ", worker.Pubkey, " ", worker.Address, " ", worker.Port)
				log.Debug("self ", c.Self.Pubkey, " ", c.Self.Address, " ", c.Self.Port)
				log.Debug("our shardid is ", c.Shardid)
				break
			}
		}

		if c.NodeType != sc.NodeShard {
			continue
		}

		for _, member := range shard.Member {
			var worker sc.Worker
			worker.Pubkey = string(member.PublicKey)
			worker.Address = member.Address
			worker.Port = member.Port
			c.shard = append(c.shard, &worker)
		}

		break
	}

}

func (c *Cell) getShardHeight(shardid uint32) uint64 {
	return c.chain.getShardHeight(shardid)
}

func (c *Cell) CalcShardLeader(size int) uint64 {
	final := c.GetLastFinalBlock()

	height := final.Height

	i := (height % sc.DefaultEpochFinalBlockNumber) % uint64(size)
	log.Debug("current leader i ", i)
	return i
}

func (c *Cell) CalcShardBackup(size int) uint64 {
	final := c.GetLastFinalBlock()
	height := final.Height + 1

	i := (height % sc.DefaultEpochFinalBlockNumber) % uint64(size)
	log.Debug("current backup i ", i)
	return i
}

func (c *Cell) CheckMinorBlockInPool(final *cs.FinalBlock) bool {
	return c.minorBlockPool.checkMinorBlockInPool(final)
}
