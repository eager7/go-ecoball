package cell

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"io/ioutil"
)

var (
	log = elog.NewLogger("sdcell", elog.DebugLog)
)

type Cell struct {
	NodeType int
	Shardid  uint16 /*only node is shard member*/
	Self     Worker
	cm       *workerSet
	shard    []*Worker
	//ss        *shardSet
	//nodes     *workerMap
	//candidate workerSet

	/*last chain data*/
	chain          *chainData
	minorBlockPool *minorBlockSet

	Ledger ledger.Ledger
}

func MakeCell(l ledger.Ledger) *Cell {
	return &Cell{
		cm:             makeWorkerSet(sc.DefaultCommitteMaxMember),
		chain:          makeChainData(),
		minorBlockPool: makeMinorBlockSet(),
		Ledger:         l,
	}
}

type NodeConfig struct {
	Pubkey  string
	Address string
	Port    string
}

type sconfig struct {
	Pubkey    string
	Address   string
	Port      string
	Committee []NodeConfig
	Shard     []NodeConfig
}

func (c *Cell) readConfigFile(filename string) *sconfig {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Info("read config file error")
		return nil
	}

	str := string(bytes)

	var cfg sconfig
	if err := json.Unmarshal([]byte(str), &cfg); err != nil {
		log.Info("json unmarshal error")
		return nil
	}

	return &cfg
}

func (c *Cell) LoadConfig() {
	cfg := c.readConfigFile("config.json")

	c.Self.Pubkey = cfg.Pubkey
	c.Self.Address = cfg.Address
	c.Self.Port = cfg.Port

	nodeType := sc.NodeNil
	for _, member := range cfg.Committee {
		var worker Worker
		worker.Pubkey = member.Pubkey
		worker.Address = member.Address
		worker.Port = member.Port

		c.addCommitteWorker(&worker)
		if c.Self.Equal(&worker) {
			nodeType = sc.NodeCommittee
		}
	}

	if nodeType == sc.NodeNil {
		nodeType = sc.NodeCandidate
	}

	c.NodeType = nodeType
}

func (c *Cell) SaveLastCMBlock(bk *types.CMBlock) {
	c.chain.setCMBlock(bk)

	worker := &Worker{}
	if len(bk.Candidate.PublicKey) != 0 {
		worker.InitWork(&bk.Candidate)
		c.addCommitteWorker(worker)
	}

	if c.NodeType == sc.NodeShard {
		c.saveShardsInfoFromCMBlock(bk)
	}

	c.minorBlockPool.resize(len(bk.Shards))
}

func (c *Cell) GetLastCMBlock() *types.CMBlock {
	return c.chain.getCMBlock()
}

func (c *Cell) SaveLastFinalBlock(bk *types.FinalBlock) {
	c.chain.setFinalBlock(bk)
	c.minorBlockPool.clean()
}

func (c *Cell) GetLastFinalBlock() *types.FinalBlock {
	return c.chain.getFinalBlock()
}

func (c *Cell) SaveLastViewchangeBlock(bk *types.ViewChangeBlock) {
	leader := &Worker{}
	leader.InitWork(&bk.Candidate)

	c.cm.resetNewLeader(leader)
	c.chain.setViewchangeBlock(bk)
}

func (c *Cell) GetLastViewchangeBlock() *types.ViewChangeBlock {
	return c.chain.getViewchangeBlock()
}

func (c *Cell) SaveLastMinorBlock(bk *types.MinorBlock) {
	c.chain.setMinorBlock(bk)
}

func (c *Cell) GetLastMinorBlock() *types.MinorBlock {
	return c.chain.getMinorBlock()
}

func (c *Cell) SavePreMinorBlock(bk *types.MinorBlock) {
	c.chain.setPreMinorBlock(bk)
}

func (c *Cell) GetPreMinorBlock() *types.MinorBlock {
	return c.chain.getPreMinorBlock()
}

func (c *Cell) SyncCmBlockComplete(lastCmblock *types.CMBlock) {
	curBlock := c.chain.getCMBlock()

	var i uint64
	if curBlock == nil {
		i = 1
	} else if curBlock.Height >= lastCmblock.Height {
		log.Debug("cm block is already sync")
		return
	} else if curBlock.Height+sc.DefaultCommitteMaxMember >= lastCmblock.Height {
		i = curBlock.Height + 1
	} else {
		i = lastCmblock.Height - sc.DefaultCommitteMaxMember + 1
	}

	for ; i < lastCmblock.Height; i++ {
		block, err := c.Ledger.GetShardBlockByHeight(config.ChainHash, types.HeCmBlock, i)
		if err != nil {
			log.Error("get block error ", err)
			return
		}

		cm := block.GetObject().(types.CMBlock)

		var worker Worker
		worker.Pubkey = string(cm.Candidate.PublicKey)
		worker.Address = cm.Candidate.Address
		worker.Port = cm.Candidate.Port

		c.addCommitteWorker(&worker)
	}

	c.SaveLastCMBlock(lastCmblock)
}

func (c *Cell) SaveMinorBlockToPool(minor *types.MinorBlock) {
	c.minorBlockPool.saveMinorBlock(minor)
}

func (c *Cell) SyncMinorsBlockToPool(minors []*types.MinorBlock) {
	c.minorBlockPool.syncMinorBlocks(minors)
}

func (c *Cell) GetMinorBlockFromPool() *minorBlockSet {
	return c.minorBlockPool
}

func (c *Cell) IsMinorBlockEnoughInPool() bool {
	cm := c.chain.cmBlock
	if cm == nil {
		return true
	}

	if c.minorBlockPool.count() >= uint16(len(cm.Shards)*sc.DefaultThresholdOfMinorBlock/100) {
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

	if (&c.Self).Equal(c.shard[0]) {
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

	if (&c.Self).Equal(c.shard[1]) {
		return true
	} else {
		return false
	}
}

func (c *Cell) GetCmWorks() []*Worker {
	return c.cm.member
}

func (c *Cell) GetWorks() []*Worker {
	if c.NodeType == sc.NodeCommittee {
		return c.cm.member
	} else if c.NodeType == sc.NodeShard {
		return c.shard
	} else {
		return nil
	}
}

func (c *Cell) GetWorksCounter() uint32 {
	if c.NodeType == sc.NodeCommittee {
		return uint32(len(c.cm.member))
	} else if c.NodeType == sc.NodeShard {
		return uint32(len(c.shard))
	} else {
		return 0
	}
}

func (c *Cell) GetLeader() *Worker {
	if c.NodeType == sc.NodeCommittee {
		return c.cm.member[0]
	} else if c.NodeType == sc.NodeShard {
		return c.shard[0]
	} else {
		return nil
	}
}

func (c *Cell) GetBackup() *Worker {
	if c.NodeType == sc.NodeCommittee {
		if len(c.cm.member) > 1 {
			return c.cm.member[1]
		} else {
			return nil
		}
	} else if c.NodeType == sc.NodeShard {
		if len(c.shard) > 1 {
			return c.shard[0]
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func (c *Cell) addCommitteWorker(worker *Worker) {
	log.Debug("add commit worker key ", worker.Pubkey, " address ", worker.Address, " port ", worker.Port)
	backup := c.GetBackup()
	if backup != nil && backup.Equal(worker) {
		c.cm.popLeader()
	} else {
		c.cm.addMember(worker)
	}
}

func (c *Cell) saveShardsInfoFromCMBlock(cmb *types.CMBlock) {
	c.NodeType = sc.NodeCandidate
	c.shard = c.shard[:0]

	for i, shard := range cmb.Shards {
		for _, member := range shard.Member {
			var worker Worker
			worker.Pubkey = string(member.PublicKey)
			worker.Address = member.Address
			worker.Port = member.Port

			if c.Self.Equal(&worker) {
				c.NodeType = sc.NodeShard
				c.Shardid = uint16(i + 1)
				break
			}
		}

		if c.NodeType != sc.NodeShard {
			continue
		}

		for _, member := range shard.Member {
			var worker Worker
			worker.Pubkey = string(member.PublicKey)
			worker.Address = member.Address
			worker.Port = member.Port
			c.shard = append(c.shard, &worker)
		}

		break
	}

}
