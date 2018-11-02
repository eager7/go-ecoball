package cell

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	cs "github.com/ecoball/go-ecoball/core/shard"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"fmt"

	"github.com/gin-gonic/gin/json"
	"time"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message/pb"
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

	if nodeType == sc.NodeNil {
		nodeType = sc.NodeShard
	}

	c.NodeType = nodeType
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
		c.saveShardsInfoFromCMBlock(bk)
	}

	c.createShardingTopo()
}

func (c *Cell) createShardingTopo() {
	topo := &sc.ShardingTopo{ShardId: c.Shardid}

	lastcm := c.GetLastCMBlock()
	if lastcm == nil {
		panic("last cm block is nil")
		return
	}

	total := len(lastcm.Shards) + 1

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

	log.Debug("send sharding topo to channel ", topo.ShardId, " len ", len(topo.ShardingInfo))
	c.Topoc <- topo
}

func (c *Cell) GetLastCMBlock() *cs.CMBlock {
	return c.chain.getCMBlock()
}

func (c *Cell) SaveLastFinalBlock(bk *cs.FinalBlock) {
	log.Debug("save final block epoch ", bk.EpochNo, " height ", bk.Height)

	c.chain.setFinalBlock(bk)

	for _, minor := range bk.MinorBlocks {
		log.Debug("minor block shard id ", minor.ShardId, " height ", minor.Height)
		c.chain.setShardHeight(minor.ShardId, minor.Height)
		if uint32(c.Shardid) == minor.ShardId {
			c.chain.saveMinorBlock(minor)
		}
	}

	c.minorBlockPool.clean()
}

func (c *Cell) GetLastFinalBlock() *cs.FinalBlock {
	return c.chain.getFinalBlock()
}

func (c *Cell) SaveLastViewchangeBlock(bk *cs.ViewChangeBlock) {
	log.Debug("save view change block epoch ", bk.CMEpochNo, " height ", bk.FinalBlockHeight, " round ", bk.Round)
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
	c.chain.setMinorBlock(bk)
}

func (c *Cell) GetLastMinorBlock() *cs.MinorBlock {
	return c.chain.getMinorBlock()
}

func (c *Cell) SavePreMinorBlock(bk *cs.MinorBlock) {
	c.chain.setPreMinorBlock(bk)
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
		block, err := c.Ledger.GetShardBlockByHeight(config.ChainHash, cs.HeCmBlock, i)
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

func (c *Cell) IsCommitteeMember() bool {
	return c.cm.isMember(&c.Self)
}

func (c *Cell) GetCmWorks() []*sc.Worker {
	return c.cm.member
}

func (c *Cell) GetWorks() []*sc.Worker {
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

func (c *Cell) GetLeader() *sc.Worker {
	if c.NodeType == sc.NodeCommittee {
		return c.cm.member[0]
	} else if c.NodeType == sc.NodeShard {
		return c.shard[0]
	} else {
		return nil
	}
}

func (c *Cell) GetBackup() *sc.Worker {
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

func (c *Cell) addCommitteWorker(worker *sc.Worker) {
	log.Debug("add commit worker key ", worker.Pubkey, " address ", worker.Address, " port ", worker.Port)
	backup := c.GetBackup()
	if backup != nil && backup.Equal(worker) {
		c.cm.popLeader()
	} else {
		c.cm.addMember(worker)
	}
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


func (s *Cell) SyncResponseDecode(syncData *sc.SyncResponseData) (*sc.SyncResponsePacket)   {

	blockType := syncData.BlockType
	len := syncData.Len
	data := syncData.Data
	lastHeight := syncData.LastHeight

	fmt.Println("len = ", len)
	fmt.Println("data = ", data)

	var list []cs.Payload
	for i := 0; i < int(len); i++ {
		blockInterface, err := cs.BlockDeserialize(data[i], cs.HeaderType(blockType))
		if err != nil {
			log.Error("minor block deserialize err")
			return nil
		}
		list = append(list,blockInterface)
	}
	csp := &sc.SyncResponsePacket{
		uint(len),
		blockType,
		list,
		lastHeight,
		syncData.Compelte,
	}

	return csp
}

//TODO, make sure TellBlock will be all right
func (s *Cell) dealSyncResponse(response *sc.SyncResponsePacket) {
	blocks := response.Blocks
	for _, block := range blocks {
		simulate.TellBlock(block.(cs.BlockInterface))
	}
}

func (s *Cell) DealSyncRequestHelper(request *sc.SyncRequestPacket) (*sc.NetPacket)  {
	from := request.FromHeight
	to := request.ToHeight
	blockType := cs.HeaderType(request.BlockType)

	var response sc.SyncResponsePacket

	fmt.Println("from = ", from)
	lastBlock, err := s.Ledger.GetLastShardBlock(config.ChainHash, blockType)
	if err != nil {
		log.Error("GetLastShardBlock error", err)
		return nil
	}
	response.LastHeight = uint64(to)
	if to < 0 {
		to = int64(lastBlock.GetHeight())
	}
	if to > from + 10 {
		to = from + 10
		response.Compelte = false
	} else {
		response.Compelte = true
	}



	fmt.Println("to = ", to)


	for i := from; i <= to; i++ {
		blockInterface, err := s.Ledger.GetShardBlockByHeight(config.ChainHash, blockType, uint64(i))
		if err == nil {
			minorBlock := blockInterface.GetObject().(cs.Payload)
			response.Blocks = append(response.Blocks, minorBlock)
		}
	}

	data := response.Encode(uint8(blockType))

	csp := &sc.NetPacket{
		PacketType: pb.MsgType_APP_MSG_SYNC_RESPONSE,
		BlockType: sc.SD_SYNC,
	}
	jsonData,err := json.Marshal(data)
	if err != nil {
		log.Error("GetLastShardBlock error", err)
		return nil
	}
	csp.Packet = jsonData

	return csp
}

//TODO, Restrict max block counts
func (s *Cell) dealSyncRequest(request *sc.SyncRequestPacket) (*sc.NetPacket, *sc.WorkerId) {

	worker := request.Worker
	csp := s.DealSyncRequestHelper(request)

	return csp, worker
}

func (s *Cell)  RecvSyncRequestPacket(packet *sc.CsPacket) (*sc.NetPacket, *sc.WorkerId){
	requestPacket := packet.Packet.(*sc.SyncRequestPacket)
	return s.dealSyncRequest(requestPacket)
}

func (s *Cell)  RecvSyncResponsePacket(packet *sc.CsPacket){
	data := packet.Packet.(sc.SyncResponseData)

	p := s.SyncResponseDecode(&data)
	s.dealSyncResponse(p)
	if p.Compelte {
		simulate.SyncComplete()
	} else {
		log.Info("Data sync not complete")

	}
}

func (s *Cell) DealSyncRequestHelperTest(request *sc.SyncRequestPacket) (*sc.NetPacket)  {
	from := request.FromHeight
	to := request.ToHeight

	fmt.Println("from = ", from)
	if to < 0 {
		to = 20
	}
	fmt.Println("to = ", to)

	var response sc.SyncResponsePacket
	for i := from; i <= to; i++ {

		header := cs.MinorBlockHeader {
			Version: 213,
			Height: 21392,
			Timestamp:    time.Now().UnixNano(),

			COSign:       nil,



		}
		cosign := &types.COSign{}
		cosign.Step1 = 1
		cosign.Step2 = 0

		header.COSign = cosign

		minorBlock := cs.MinorBlock {
			MinorBlockHeader: header,
			Transactions: nil  ,
			StateDelta: nil ,
		}
		response.Blocks = append(response.Blocks, &minorBlock)

	}

	data := response.Encode(uint8(cs.HeMinorBlock))

	csp := &sc.NetPacket{
		PacketType: pb.MsgType_APP_MSG_SYNC_RESPONSE,
		BlockType: sc.SD_SYNC,
	}
	jsonData,err := json.Marshal(data)
	if err != nil {
		log.Error("GetLastShardBlock error", err)
		return nil
	}
	csp.Packet = jsonData

	return csp
}


