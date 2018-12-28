package datasync

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/net"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"reflect"
	"sync"
	"time"
)

var (
	log = elog.NewLogger("sync", elog.DebugLog)
)

/**
TODO,maybe change cache to map:height->block base, will be easy to code?and more efficient
*/
type BlocksCache struct {
	needHeight         map[cs.HeaderType]uint64
	needHeightMinor    map[uint32]uint64
	blocks             map[cs.HeaderType][]cs.BlockInterface
	minorBlocks        map[uint32][]cs.BlockInterface
	finalBlockComplete bool
	complete           bool
}

type Sync struct {
	cell          *cell.Cell
	cache         BlocksCache
	synchronizing bool
	lock          sync.Mutex
	receiveCh     chan *sc.CsPacket
	sendCh        chan int
	retryTimer    *sc.Stimer
}

func MakeSync(c *cell.Cell) *Sync {
	return &Sync{
		cell: c,
		cache: BlocksCache{
			needHeight:         make(map[cs.HeaderType]uint64, 0),
			needHeightMinor:    make(map[uint32]uint64, 0),
			blocks:             make(map[cs.HeaderType][]cs.BlockInterface, 0),
			minorBlocks:        make(map[uint32][]cs.BlockInterface, 0),
			finalBlockComplete: false,
			complete:           false,
		},
		synchronizing: false,
		receiveCh:     make(chan *sc.CsPacket, 10),
		sendCh:        make(chan int, 2),
		retryTimer:    sc.NewStimer(0, false),
	}
}

func (sync *Sync) Start() {
	go sync.working()
}

func (sync *Sync) working() {
	for {
		select {
		case packet := <-sync.receiveCh:
			log.Debug("Receive Sync Packet")
			sync.RecvSyncResponsePacketHelper(packet)
		case <-sync.sendCh:
			log.Debug("Send Request")
			sync.SendSyncRequestHelper()
		case <-sync.retryTimer.T.C:
			log.Debug("Retry sync")
			sync.SendSyncRequest()
		}
	}
}

func MakeSyncRequestPacket(blockType cs.HeaderType, fromHeight int64, to int64, worker *sc.WorkerId, shardID uint32) *sc.NetPacket {
	csp := &sc.NetPacket{
		ChainId:    0,
		PacketType: pb.MsgType_APP_MSG_SYNC_REQUEST,
		BlockType:  sc.SD_SYNC,
		Step:       0,
		Packet:     nil,
	}
	request := &sc.SyncRequestPacket{
		BlockType:  blockType,
		FromHeight: fromHeight,
		ToHeight:   to,
		Worker:     worker,
		ShardID:    shardID,
	}

	data, err := json.Marshal(request)
	if err != nil {
		log.Error("vc block marshal error ", err)
		return nil
	}
	csp.Packet = data

	return csp
}

func (sync *Sync) SendSyncRequest() {
	sync.sendCh <- 1
}

//Request order is important
func (sync *Sync) SendSyncRequestHelper() {

	sync.retryTimer.Reset(3 * time.Second)

	//special case: commitee worker length = 1
	works := sync.cell.GetWorks()
	if len(works) == 1 {
		log.Debug("Commitee worker len = 1, don't send sync request")
		log.Debug("retryTimer stop 1")
		sync.retryTimer.Stop()
		simulate.SyncComplete()
		return
	}

	/*if sync.cell.NodeType == sc.NodeCommittee {
		time.Sleep(3 * time.Second)
	}*/

	sync.lock.Lock()
	sync.synchronizing = true
	sync.lock.Unlock()

	log.Debug("SendSyncRequest, node type = ", sync.cell.NodeType)

	//TODO, Special case treatment, not working when start a new shard node from scratch
	if sync.cell.NodeType == sc.NodeShard || sync.cell.NodeType == sc.NodeCandidate {
		log.Debug("Node is ", sync.cell.NodeType)
		//lastBlock, _, err := sync.cell.Ledger.GetLastShardBlock(config.ChainHash, cs.HeMinorBlock)
		//if err != nil {
		//	log.Error("GetLastShardBlock ", err)
		//}
		lastBlock := sync.cell.GetLastCMBlock()
		if lastBlock == nil {
			panic("last cm block not exist")
			return
		}
		log.Debug("SendSyncRequest, get Height = ", lastBlock.GetHeight())
		if lastBlock.GetHeight() == 1 {
			log.Debug("retryTimer stop 2")
			sync.retryTimer.Stop()
			simulate.SyncComplete()
			log.Info("invoke SyncComplete cause Height = 1")
			return
		}
	} else {
		log.Debug("Node isn't NodeShard")
	}

	sync.SendSyncRequestWithType(cs.HeCmBlock)
	sync.SendSyncRequestWithType(cs.HeViewChange)
	sync.SendSyncRequestWithType(cs.HeFinalBlock)
	sync.SendSyncRequestWithType(cs.HeMinorBlock)
}

func (sync *Sync) SendSyncRequestWithType(blockType cs.HeaderType) {
	if blockType != cs.HeMinorBlock {

		var height int64 = -1

		blocks := sync.cache.blocks[blockType]
		if len(blocks) > 0 {
			l := len(blocks)
			height = int64(blocks[l-1].GetHeight() + 1)
		}

		if height < 0 {
			lastBlock, _, err := sync.cell.Ledger.GetLastShardBlock(config.ChainHash, blockType)
			if err != nil {
				log.Error("get last block faield", err)
				return
			}
			height = int64(lastBlock.GetHeight() + 1)
			sync.cache.needHeight[blockType] = uint64(height)
		}
		//For non-MinorBlock, shardID isn't important
		sync.SendSyncRequestWithHeightType(blockType, int64(height), 0)
	} else {

		finalBlocks := sync.cache.blocks[cs.HeFinalBlock]
		var markBlock cs.FinalBlock
		var offset uint64

		if len(finalBlocks) > 0 {
			markBlock = finalBlocks[0].GetObject().(cs.FinalBlock)
			offset = 0
		} else {
			//TODO, problem: no final block situation
			lastBlock, _, err := sync.cell.Ledger.GetLastShardBlock(config.ChainHash, cs.HeFinalBlock)
			if err != nil {
				log.Error("GetLastShardBlock, ", err)
				return
			}
			markBlock = lastBlock.GetObject().(cs.FinalBlock)

			offset = 1
		}
		//TODO, deal with non-Fix sharding. Or final block missing minor block?
		for _, minorBlock := range markBlock.MinorBlocks {
			height := minorBlock.Height + offset
			shardID := minorBlock.ShardId
			list := sync.cache.minorBlocks[shardID]
			len := len(list)
			if len > 0 {
				tHeight := list[len-1].GetHeight() + 1
				if tHeight > height {
					height = tHeight
				}
			}
			sync.cache.needHeightMinor[shardID] = height
			sync.SendSyncRequestWithHeightType(blockType, int64(height), shardID)
		}
	}
}

func (sync *Sync) SendSyncRequestWithHeightType(blockType cs.HeaderType, fromHeight int64, shardID uint32) {
	log.Debug("SendSyncRequestWithHeightType, blockType = ",
		blockType, " from height = ", fromHeight, " shardID = ", shardID)
	worker := &sc.WorkerId{
		sync.cell.Self.Pubkey,
		sync.cell.Self.Address,
		sync.cell.Self.Port,
	}
	csp := MakeSyncRequestPacket(blockType, fromHeight, -1, worker, shardID)

	net.Np.SendSyncMessage(csp)
}

func (sync *Sync) SendSyncRequestTo(blockType cs.HeaderType, fromHeight int64, toHeight int64, shardID uint32) {
	worker := &sc.WorkerId{
		sync.cell.Self.Pubkey,
		sync.cell.Self.Address,
		sync.cell.Self.Port,
	}
	csp := MakeSyncRequestPacket(blockType, fromHeight, toHeight, worker, shardID)

	net.Np.SendSyncMessage(csp)
}

func (s *Sync) SyncResponseDecode(syncData *sc.SyncResponseData) *sc.SyncResponsePacket {

	blockType := syncData.BlockType
	len := syncData.Len
	data := syncData.Data
	lastHeight := syncData.LastHeight

	//fmt.Println("len = ", len)
	//fmt.Println("data = ", data)

	var list []cs.Payload
	for i := 0; i < int(len); i++ {

		blockInterface, err := cs.BlockDeserialize(data[i])
		if err != nil {
			log.Error("minor block deserialize err")
			return nil
		}
		list = append(list, blockInterface)
	}
	csp := &sc.SyncResponsePacket{
		uint(len),
		blockType,
		list,
		lastHeight,
		syncData.ShardID,
		syncData.Compelte,
	}

	return csp
}

func (s *Sync) clearCache() {
	s.cache = BlocksCache{
		needHeight:         make(map[cs.HeaderType]uint64, 0),
		needHeightMinor:    make(map[uint32]uint64, 0),
		blocks:             make(map[cs.HeaderType][]cs.BlockInterface, 0),
		minorBlocks:        make(map[uint32][]cs.BlockInterface, 0),
		finalBlockComplete: false,
		complete:           false,
	}
}

//TODO, make sure TellBlock will be all right
func (s *Sync) tellLedgerSyncComplete() {
	log.Debug("tellLedgerSyncComplete in")
	ledger := s.cell.Ledger

	for _, block := range s.cache.blocks[cs.HeCmBlock] {
		log.Debug("SaveShardBlock cm, height = ",
			block.GetHeight())
		ledger.SaveShardBlock(config.ChainHash, block)
	}

	finalBlocks := s.cache.blocks[cs.HeFinalBlock]
	l := len(finalBlocks)
	if l > 0 {
		lastFinalBlock := finalBlocks[l-1].GetObject().(cs.FinalBlock)
		for _, minorBlock := range lastFinalBlock.MinorBlocks {
			shardID := minorBlock.ShardId
			minorBlocks := s.cache.minorBlocks[shardID]
			ll := len(minorBlocks)
			limit := minorBlock.Height
			for i := 0; i < ll; i++ {
				curMinorBlock := minorBlocks[i]
				if curMinorBlock.GetHeight() > limit {
					break
				}
				log.Debug("SaveShardBlock minor, height = ",
					curMinorBlock.GetHeight(), " shardID = ", shardID)
				ledger.SaveShardBlock(config.ChainHash, curMinorBlock)
			}
		}
		for _, finalBlock := range finalBlocks {
			log.Debug("SaveShardBlock final, height = ",
				finalBlock.GetHeight())
			ledger.SaveShardBlock(config.ChainHash, finalBlock)
		}
	}

	for _, block := range s.cache.blocks[cs.HeViewChange] {
		log.Debug("SaveShardBlock change view, height = ",
			block.GetHeight())
		ledger.SaveShardBlock(config.ChainHash, block)
	}
}

func (s *Sync) DealSyncRequestHelper(request *sc.SyncRequestPacket) *sc.NetPacket {
	from := request.FromHeight
	to := request.ToHeight
	blockType := cs.HeaderType(request.BlockType)
	shardID := request.ShardID
	log.Debug("type = ", request.BlockType)
	log.Debug("from = ", from, " to = ", to, " blockType = ", blockType)

	var response sc.SyncResponsePacket

	fmt.Println("from = ", from)

	if blockType == cs.HeMinorBlock {
		blockType = cs.HeFinalBlock
	}
	lastBlock, _, err := s.cell.Ledger.GetLastShardBlock(config.ChainHash, blockType)
	blockType = cs.HeaderType(request.BlockType)

	if err != nil {
		log.Error("GetLastShardBlock error", err)
		return nil
	}

	response.LastHeight = uint64(to)
	if to < 0 {
		to = int64(lastBlock.GetHeight())
	}
	if to > from+10 {
		to = from + 10
		response.Compelte = false
	} else {
		response.Compelte = true
	}

	log.Debug("from = ", from, " to = ", to, " blockType = ", blockType)

	for i := from; i <= to; i++ {
		blockInterface, _, err := s.cell.Ledger.GetShardBlockByHeight(config.ChainHash, blockType, uint64(i), shardID)
		log.Debug("Block Type = ", blockType, " index = ", i)
		if err == nil {
			//TODO, need to simplize, and refactor
			log.Debug("block type = ", reflect.TypeOf(blockInterface.GetObject()))
			o := blockInterface.GetObject()
			block := s.getConcreteBlockObject(o)
			payload := block.(cs.Payload)
			response.Blocks = append(response.Blocks, payload)
		}
	}

	data := response.Encode(blockType, shardID)

	csp := &sc.NetPacket{
		PacketType: pb.MsgType_APP_MSG_SYNC_RESPONSE,
		BlockType:  sc.SD_SYNC,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("GetLastShardBlock error", err)
		return nil
	}
	csp.Packet = jsonData

	return csp
}

//TODO, Restrict max block counts
func (s *Sync) dealSyncRequest(request *sc.SyncRequestPacket) (*sc.NetPacket, *sc.WorkerId) {

	worker := request.Worker
	csp := s.DealSyncRequestHelper(request)

	return csp, worker
}

func (s *Sync) RecvSyncRequestPacket(packet *sc.CsPacket) (*sc.NetPacket, *sc.WorkerId) {
	requestPacket := packet.Packet.(*sc.SyncRequestPacket)
	return s.dealSyncRequest(requestPacket)
}

func (s *Sync) CheckSyncCompleteForMinorBlock(minBlock *cs.MinorBlockHeader, blocks *[]cs.BlockInterface, syncResponse *sc.SyncResponsePacket) bool {
	len := len(*blocks)

	if len > 0 {
		lastMinFinal := (*blocks)[len-1]

		log.Debug("lastMin height = ", lastMinFinal.GetHeight(), " minBlock.Height = ", minBlock.Height)
		if lastMinFinal.GetHeight() >= minBlock.GetHeight() {
			return true
		} else {
			return false
		}
	} else {
		ledger := s.cell.Ledger
		_, _, err := ledger.GetShardBlockByHash(config.ChainHash, cs.HeMinorBlock, minBlock.Hash(), true)
		//log.Debug("minorBlock hash = ", minBlock.Hash(), " err = ", nil, " o.hash = ", o.Hash())
		//log.Debug("minorBlock height = ", minBlock.Height, " err = ", nil, " o.height = ", o.GetHeight())
		if err != nil {
			log.Debug("GetShardBlockByHash", err)
			return false
		} else {
			return true
		}

	}
}

func (s *Sync) CheckSyncCompleteForCMBlock(epoch uint64, blocks *[]cs.BlockInterface) bool {
	list := *blocks
	len := len(list)
	var lastCMBlock cs.CMBlock
	if len > 0 {
		lastCMBlock = (*blocks)[len-1].GetObject().(cs.CMBlock)
	} else {
		o, _, err := s.cell.Ledger.GetLastShardBlock(config.ChainHash, cs.HeCmBlock)
		if err != nil {
			log.Error("GetLastShardBlock error")
		} else {
			lastCMBlock = o.GetObject().(cs.CMBlock)
		}
	}
	log.Debug("final epoch = ", epoch, " lastCMBlock.Height = ", lastCMBlock.Height)
	if epoch <= lastCMBlock.Height {
		return true
	}
	return false
}

//TODO, probably missing some corner case?
func (s *Sync) CheckSyncComplete(syncResponse *sc.SyncResponsePacket) bool {
	log.Debug("CheckSyncComplete blockType = ",
		syncResponse.BlockType, " complete = ", syncResponse.Compelte)
	log.Debug("Block type = ", uint8(syncResponse.BlockType))

	var lastFinalBlock cs.FinalBlock
	hasFinalBlock := false
	if syncResponse.BlockType == cs.HeFinalBlock && syncResponse.Compelte {
		log.Debug("CheckSyncComplete, set complete")
		s.cache.finalBlockComplete = true
		blocks := s.cache.blocks[cs.HeFinalBlock]
		len := len(blocks)
		log.Debug("final block cache len = ", len)
		if len > 0 {
			lastFinalBlock = blocks[len-1].GetObject().(cs.FinalBlock)
			hasFinalBlock = true
		}
	}
	if s.cache.finalBlockComplete {

		if hasFinalBlock {
			log.Debug("check complete, step 1")
			for _, minorBlock := range lastFinalBlock.MinorBlocks {
				log.Debug("MinorBlock hash = ", minorBlock.Hash())
				blocks := s.cache.minorBlocks[minorBlock.ShardId]
				complete := s.CheckSyncCompleteForMinorBlock(minorBlock, &blocks, syncResponse)
				if !complete {
					return false
				}
			}
			log.Debug("check complete, step 2")
			//TODO, check cm block
			blocks := s.cache.blocks[cs.HeCmBlock]

			if !s.CheckSyncCompleteForCMBlock(lastFinalBlock.EpochNo, &blocks) {
				return false
			}
			log.Debug("check complete, step 3")
		} else {

			/*
				Case commitee start
			*/
			lastFinalBlock, _, err := s.cell.Ledger.GetLastShardBlock(config.ChainHash, cs.HeFinalBlock)
			if err != nil {
				log.Error("GetLastShardBlock", err)
				return true
			}
			//TODO, check cm block
			if lastFinalBlock.GetHeight() == 1 {
				return true
			}

			lenFinalBlock := len(s.cache.blocks[cs.HeCmBlock])
			lenCMBlock := len(s.cache.blocks[cs.HeFinalBlock])
			lenChangeViewBlock := len(s.cache.blocks[cs.HeViewChange])
			if lenFinalBlock == 0 && lenCMBlock == 0 && lenChangeViewBlock == 0 {
				log.Debug("Empty Cache")
				return false
			}
		}

		return true

	} else {
		return false
	}
}

func (s *Sync) getConcreteBlockObject(o interface{}) interface{} {
	var block interface{}
	switch t := o.(type) {
	case cs.CMBlock:
		o1 := o.(cs.CMBlock)
		block = interface{}(&o1)
	case cs.MinorBlock:
		o1 := o.(cs.MinorBlock)
		block = interface{}(&o1)
	case cs.FinalBlock:
		o1 := o.(cs.FinalBlock)
		block = interface{}(&o1)
	case cs.ViewChangeBlock:
		o1 := o.(cs.ViewChangeBlock)
		block = interface{}(&o1)
	default:
		log.Error("Wrong type ", t)
	}
	return block
}

func (s *Sync) FillSyncDataInCacheHelper(p *[]cs.BlockInterface, needHeight uint64, syncResponse *sc.SyncResponsePacket) {
	log.Debug("FillSyncDataInCacheHelper, needHeight = ", needHeight, " block type = ", syncResponse.BlockType)
	log.Debug("Fill Block size = ", len(syncResponse.Blocks), " cache size = ", len(*p))
	blockType := syncResponse.BlockType
	for _, payload := range syncResponse.Blocks {
		o := payload.GetObject()
		var block cs.BlockInterface
		block = s.getConcreteBlockObject(o).(cs.BlockInterface)

		list := *p
		len := len(list)
		var needH uint64
		if len > 0 {
			needH = list[len-1].GetHeight() + 1
		} else {
			needH = needHeight //s.cache.needHeight[blockType]
		}
		if block.GetHeight() != needH {
			log.Debug("Not need height, needHeight = ",
				needH, " currentHeight = ", block.GetHeight())
		} else {
			log.Debug("Append block in cache, type = ",
				blockType, " height = ", block.GetHeight())
			*p = append(*p, block)
		}
	}
}

func (s *Sync) FillSyncDataInCache(syncResponse *sc.SyncResponsePacket) {

	blockType := syncResponse.BlockType

	if blockType != cs.HeMinorBlock {
		list := s.cache.blocks[blockType]
		s.FillSyncDataInCacheHelper(&list, s.cache.needHeight[blockType], syncResponse)
		//TODO, not efficent
		s.cache.blocks[blockType] = list
	} else {
		shardID := syncResponse.ShardID
		list := s.cache.minorBlocks[shardID]
		s.FillSyncDataInCacheHelper(&list, s.cache.needHeightMinor[shardID], syncResponse)
		//TODO, not efficent
		s.cache.minorBlocks[shardID] = list
	}

}

func (s *Sync) RecvSyncResponsePacket(packet *sc.CsPacket) {
	s.receiveCh <- packet
}

func (s *Sync) RecvSyncResponsePacketHelper(packet *sc.CsPacket) {

	data := packet.Packet.(*sc.SyncResponseData)

	p := s.SyncResponseDecode(data)
	log.Debug("RecvSyncResponsePacket, blockType = ", p.BlockType)

	s.FillSyncDataInCache(p)

	if s.CheckSyncComplete(p) {
		s.lock.Lock()

		if s.synchronizing {
			s.tellLedgerSyncComplete()
			s.clearCache()
			log.Debug("retryTimer stop 0")
			s.retryTimer.Stop()
			simulate.SyncComplete()
			s.synchronizing = false
			log.Info("invoke SyncComplete")
		}

		s.lock.Unlock()
	} else {
		log.Info("Data sync not complete")
		//s.SendSyncRequest()
	}
}

func (s *Sync) DealSyncRequestHelperTest(request *sc.SyncRequestPacket) *sc.NetPacket {
	from := request.FromHeight
	to := request.ToHeight

	fmt.Println("from = ", from)
	if to < 0 {
		to = 20
	}
	fmt.Println("to = ", to)

	var response sc.SyncResponsePacket
	for i := from; i <= to; i++ {

		header := cs.MinorBlockHeader{
			Version:   213,
			Height:    21392,
			Timestamp: time.Now().UnixNano(),

			COSign: nil,
		}
		cosign := &types.COSign{}
		cosign.Step1 = 1
		cosign.Step2 = 0

		header.COSign = cosign

		minorBlock := cs.MinorBlock{
			MinorBlockHeader: header,
			Transactions:     nil,
			StateDelta:       nil,
		}
		response.Blocks = append(response.Blocks, &minorBlock)

	}

	data := response.Encode(cs.HeMinorBlock, 0)

	csp := &sc.NetPacket{
		PacketType: pb.MsgType_APP_MSG_SYNC_RESPONSE,
		BlockType:  sc.SD_SYNC,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("GetLastShardBlock error", err)
		return nil
	}
	csp.Packet = jsonData

	return csp
}
