package node

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

type Node struct {
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

func MakeNode() *Node {
	return &Node{
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

func (n *Node) readConfigFile(filename string) *config {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Info("read config file error")
		return nil
	}

	str := string(bytes)

	var c config
	if err := json.Unmarshal([]byte(str), &c); err != nil {
		log.Info("json unmarshal error")
		return nil
	}

	return &c
}

func (n *Node) LoadConfig() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)

	c := n.readConfigFile(dir + "\\config.json")
	if c == nil {
		return
	}

	n.Self.Pubkey = c.Pubkey
	n.Self.Address = c.Address
	n.Self.Port = c.Port

	nodeType := sc.NodeNil
	for _, member := range c.Committee {
		var worker Worker
		worker.Pubkey = member.Pubkey
		worker.Address = member.Address
		worker.Port = member.Port

		n.addCommitteWorker(&worker)
		if n.Self.Equal(&worker) {
			nodeType = sc.NodeCommittee
		}
	}

	if nodeType == sc.NodeNil {
		nodeType = sc.NodeCandidate
	}

	n.NodeType = nodeType
}

func (n *Node) SetLastCMBlock(cmb *block.CMBlock) {
	n.chain.setCMBlock(cmb)

	var worker Worker
	worker.Pubkey = string(cmb.Candidate.PublicKey)
	worker.Address = cmb.Candidate.Address
	worker.Port = cmb.Candidate.Port

	n.addCommitteWorker(&worker)
	if n.NodeType == sc.NodeShard {
		n.saveShardsInfoFromCMBlock(cmb)
	}

	n.minorBlockPool.resize(len(cmb.Shards))

}

func (n *Node) GetLastCMBlock() *block.CMBlock {
	return n.chain.getCMBlock()
}

func (n *Node) GetLastFinalBlock() *block.FinalBlock {
	return n.chain.getFinalBlock()
}

func (n *Node) SetLastFinalBlock(block *block.FinalBlock) {
	n.chain.setFinalBlock(block)
	n.minorBlockPool.clean()
}

func (n *Node) SyncCMBlockComplete(lastCMblock *block.CMBlock) {
	curBlock := n.chain.getCMBlock()

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

		n.addCommitteWorker(&worker)
	}

	n.SetLastCMBlock(lastCMblock)
}

func (n *Node) SetMinorBlockToPool(minor *block.MinorBlock) {
	n.minorBlockPool.setMinorBlock(minor)
}

func (n *Node) SyncMinorsBlockToPool(minors []*block.MinorBlock) {
	n.minorBlockPool.syncMinorBlocks(minors)
}

func (n *Node) GetMinorBlockFromPool() *minorBlockSet {
	return n.minorBlockPool
}

func (n *Node) GetMinorBlockPoolCount() uint16 {
	return n.minorBlockPool.count()
}

func (n *Node) IsCmLeader() bool {
	return n.cm.isLeader(&n.Self)
}

func (n *Node) IsCmCandidateLeader() bool {
	/*should do vrf by cmblock*/
	return n.cm.isCandidateLeader(&n.Self)
}

func (n *Node) GetCmWorks() []*Worker {
	return n.cm.member
}

func (n *Node) GetCmWorksCounter() uint16 {
	return uint16(len(n.cm.member))
}

func (n *Node) GetShardWorks() []*Worker {
	return n.shard
}

func (n *Node) GetShardWorksCounter() uint16 {
	return uint16(len(n.shard))
}

func (n *Node) addCommitteWorker(worker *Worker) {
	n.cm.addMember(worker)
}

func (n *Node) saveShardsInfoFromCMBlock(cmb *block.CMBlock) {
	n.NodeType = sc.NodeCandidate
	n.shard = n.shard[:0]

	for i, shard := range cmb.Shards {
		for _, member := range shard.Member {
			var worker Worker
			worker.Pubkey = string(member.PublicKey)
			worker.Address = member.Address
			worker.Port = member.Port

			if n.Self.Equal(&worker) {
				n.NodeType = sc.NodeShard
				n.Shardid = uint16(i + 1)
				break
			}
		}

		if n.NodeType != sc.NodeShard {
			continue
		}

		for _, member := range shard.Member {
			var worker Worker
			worker.Pubkey = string(member.PublicKey)
			worker.Address = member.Address
			worker.Port = member.Port
			n.shard = append(n.shard, &worker)
		}

		break
	}

}
