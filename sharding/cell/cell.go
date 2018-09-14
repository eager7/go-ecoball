package cell

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/types/block"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	log = elog.NewLogger("sdnode", elog.DebugLog)
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
}

func MakeCell() *Cell {
	return &Cell{
		cm:             makeWorkerSet(sc.DefaultCommitteMaxMember),
		chain:          makeChainData(),
		minorBlockPool: makeMinorBlockSet(),
	}
}

type NodeConfig struct {
	Pubkey  string
	Address string
	Port    string
}

type config struct {
	Pubkey    string
	Address   string
	Port      string
	Committee []NodeConfig
	Shard     []NodeConfig
}

func (c *Cell) readConfigFile(filename string) *config {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Info("read config file error")
		return nil
	}

	str := string(bytes)

	var cfg config
	if err := json.Unmarshal([]byte(str), &cfg); err != nil {
		log.Info("json unmarshal error")
		return nil
	}

	return &cfg
}

func (c *Cell) LoadConfig() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)

	cfg := c.readConfigFile(dir + "\\config.json")
	if cfg == nil {
		return
	}

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

func (c *Cell) SetLastCMBlock(cmb *block.CMBlock) {
	c.chain.setCMBlock(cmb)

	var worker Worker
	if len(cmb.Candidate.PublicKey) != 0 {
		worker.Pubkey = string(cmb.Candidate.PublicKey)
		worker.Address = cmb.Candidate.Address
		worker.Port = cmb.Candidate.Port
		c.addCommitteWorker(&worker)
	}

	if c.NodeType == sc.NodeShard {
		c.saveShardsInfoFromCMBlock(cmb)
	}

	c.minorBlockPool.resize(len(cmb.Shards))
}

func (c *Cell) GetLastCMBlock() *block.CMBlock {
	return c.chain.getCMBlock()
}

func (c *Cell) GetLastFinalBlock() *block.FinalBlock {
	return c.chain.getFinalBlock()
}

func (c *Cell) SetLastFinalBlock(block *block.FinalBlock) {
	c.chain.setFinalBlock(block)
	c.minorBlockPool.clean()
}

func (c *Cell) SyncCMBlockComplete(lastCMblock *block.CMBlock) {
	curBlock := c.chain.getCMBlock()

	var i uint64
	if curBlock == nil {
		i = 1
	} else if curBlock.Height >= lastCMblock.Height {
		log.Error("sync cm block error")
		return
	} else if curBlock.Height+sc.DefaultCommitteMaxMember >= lastCMblock.Height {
		i = curBlock.Height + 1
	} else {
		i = lastCMblock.Height - sc.DefaultCommitteMaxMember + 1
	}

	for ; i < lastCMblock.Height; i++ {
		cmb := simulate.GetCMBlockByNumber(i)
		var worker Worker
		worker.Pubkey = string(cmb.Candidate.PublicKey)
		worker.Address = cmb.Candidate.Address
		worker.Port = cmb.Candidate.Port

		c.addCommitteWorker(&worker)
	}

	c.SetLastCMBlock(lastCMblock)
}

func (c *Cell) SetMinorBlockToPool(minor *block.MinorBlock) {
	c.minorBlockPool.setMinorBlock(minor)
}

func (c *Cell) SyncMinorsBlockToPool(minors []*block.MinorBlock) {
	c.minorBlockPool.syncMinorBlocks(minors)
}

func (c *Cell) GetMinorBlockFromPool() *minorBlockSet {
	return c.minorBlockPool
}

func (c *Cell) GetMinorBlockPoolCount() uint16 {
	return c.minorBlockPool.count()
}

func (c *Cell) IsCmLeader() bool {
	return c.cm.isLeader(&c.Self)
}

func (c *Cell) IsCmCandidateLeader() bool {
	/*should do vrf by cmblock*/
	return c.cm.isCandidateLeader(&c.Self)
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

func (c *Cell) GetWorksCounter() uint16 {
	if c.NodeType == sc.NodeCommittee {
		return uint16(len(c.cm.member))
	} else if c.NodeType == sc.NodeShard {
		return uint16(len(c.shard))
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

func (c *Cell) addCommitteWorker(worker *Worker) {
	log.Debug("add commit worker key ", worker.Pubkey, " address ", worker.Address, " port ", worker.Port)
	c.cm.addMember(worker)
}

func (c *Cell) saveShardsInfoFromCMBlock(cmb *block.CMBlock) {
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
