package datasync

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/sharding/net"
	"github.com/ecoball/go-ecoball/sharding/cell"
	"github.com/ecoball/go-ecoball/net/message/pb"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/common/config"
	"fmt"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common"
	"reflect"
)

var (
	log = elog.NewLogger("sync", elog.DebugLog)
)

type BlocksCache struct {
	needHeight map[cs.HeaderType]uint64
	needHeightMinor map[uint32]uint64
	blocks map[cs.HeaderType][]cs.BlockInterface
	minorBlocks map[uint32][]cs.BlockInterface
	finalBlockComplete bool
	complete bool
}

type Sync struct {
	cell *cell.Cell
	cache BlocksCache
}

func MakeSync(c *cell.Cell) *Sync {
	return &Sync{
		cell: c,
		cache: BlocksCache {
			needHeight: make(map[cs.HeaderType]uint64, 0),
			needHeightMinor: make(map[uint32]uint64, 0),
			blocks: make(map[cs.HeaderType][]cs.BlockInterface, 0),
			minorBlocks: make(map[uint32][]cs.BlockInterface, 0),
			finalBlockComplete: false,
			complete:false,
		},
	}
}

func MakeSyncRequestPacket(blockType int8, fromHeight int64, to int64, worker *sc.WorkerId, shardID uint32) (*sc.NetPacket) {
	csp := &sc.NetPacket{
		PacketType: pb.MsgType_APP_MSG_SYNC_REQUEST,
		BlockType: sc.SD_SYNC,
		Step: 0,
	}
	request := &sc.SyncRequestPacket{
		BlockType: blockType,
		FromHeight: fromHeight,
		ToHeight: to,
		Worker: worker,
		ShardID: shardID,
	}

	data, err := json.Marshal(request)
	if err != nil {
		log.Error("vc block marshal error ", err)
		return nil
	}
	csp.Packet = data

	return csp
}

//Request order is important
func (sync *Sync)SendSyncRequest()  {
	log.Debug("SendSyncRequest, node type = ", sync.cell.NodeType)

	//Special case treatment
	if sync.cell.NodeType == sc.NodeShard {
		log.Debug("Node is ", sc.NodeShard)
		lastBlock, err := sync.cell.Ledger.GetLastShardBlock(config.ChainHash, cs.HeMinorBlock)
		if err != nil {
			log.Error("GetLastShardBlock ", err)
		}
		log.Debug("SendSyncRequest, get Height = ", lastBlock.GetHeight())
		if lastBlock.GetHeight() == 1 {
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

func (sync *Sync)SendSyncRequestWithType(blockType cs.HeaderType) {
	if blockType != cs.HeMinorBlock {

		var height int64 = -1

		blocks := sync.cache.blocks[blockType]
		if len(blocks) > 0 {
			l := len(blocks)
			height = int64(blocks[l-1].GetHeight() + 1)
		}

		if height < 0 {
			lastBlock, err := sync.cell.Ledger.GetLastShardBlock(config.ChainHash, blockType)
			if err != nil {
				log.Error("get last block faield", err)
				return
			}
			height = int64(lastBlock.GetHeight() + 1)
			sync.cache.needHeight[blockType] = uint64(height)
		}
		//For non-MinorBlock, shardID isn't important
		sync.SendSyncRequestWithHeightType(int8(blockType), int64(height), 0)
	} else {

		finalBlocks := sync.cache.blocks[cs.HeFinalBlock]
		var markBlock cs.FinalBlock
		if len(finalBlocks) > 0 {
			markBlock = finalBlocks[0].GetObject().(cs.FinalBlock)
		} else {
			//TODO, problem: no final block situation
			lastBlock, err := sync.cell.Ledger.GetLastShardBlock(config.ChainHash, cs.HeFinalBlock)
			if err != nil {
				log.Error("GetLastShardBlock, ", err)
				return
			}
			markBlock = lastBlock.GetObject().(cs.FinalBlock)
		}
		//TODO, deal with non-Fix sharding
		for _, minorBlock := range markBlock.MinorBlocks {
			height := minorBlock.Height
			shardID := minorBlock.ShardId
			list := sync.cache.minorBlocks[shardID]
			len := len(list)
			if len > 0 {
				tHeight := list[len-1].GetHeight()
				if tHeight > height {
					height = tHeight
				}
			}
			height = height + 1
			sync.SendSyncRequestWithHeightType(int8(blockType), int64(height), shardID)
		}
	}
}


func (sync *Sync)SendSyncRequestWithHeightType(blockType int8, fromHeight int64, shardID uint32)  {
	worker := &sc.WorkerId{
		sync.cell.Self.Pubkey,
		sync.cell.Self.Address,
		sync.cell.Self.Port,
	}
	csp := MakeSyncRequestPacket(blockType, fromHeight, -1, worker, shardID)

	net.Np.SendSyncMessage(csp)
}

func (sync *Sync)SendSyncRequestTo(blockType int8, fromHeight int64, toHeight int64, shardID uint32)  {
	worker := &sc.WorkerId{
		sync.cell.Self.Pubkey,
		sync.cell.Self.Address,
		sync.cell.Self.Port,
	}
	csp := MakeSyncRequestPacket(blockType, fromHeight, toHeight, worker, shardID)

	net.Np.SendSyncMessage(csp)
}

/*func (sync *Sync)dealSyncRequest() {

}*/

func (s *Sync) SyncResponseDecode(syncData *sc.SyncResponseData) (*sc.SyncResponsePacket)   {

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
		syncData.ShardID,
		syncData.Compelte,
	}

	return csp
}

//TODO, make sure TellBlock will be all right
func (s *Sync) tellLedgerSyncComplete() {
	/*blocks := response.Blocks
	for _, block := range blocks {
		simulate.TellBlock(block.(cs.BlockInterface))
	}*/
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
				simulate.TellBlock(curMinorBlock)
			}
		}
	}
}

func (s *Sync) DealSyncRequestHelper(request *sc.SyncRequestPacket) (*sc.NetPacket)  {
	from := request.FromHeight
	to := request.ToHeight
	blockType := cs.HeaderType(request.BlockType)
	shardID := request.ShardID
	log.Debug("type = ", request.BlockType)
	log.Debug("from = ", from, " to = ", to, " blockType = ", blockType)

	var response sc.SyncResponsePacket

	fmt.Println("from = ", from)
	lastBlock, err := s.cell.Ledger.GetLastShardBlock(config.ChainHash, blockType)
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

	for i := from; i <= to; i++ {
		blockInterface, err := s.cell.Ledger.GetShardBlockByHeight(config.ChainHash, blockType, uint64(i), shardID)
		log.Debug("Block Type = ", blockType, " index = ", i)
		if err == nil {
			//TODO, need to simplize, and refactor
			log.Debug("block type = ", reflect.TypeOf(blockInterface.GetObject()))
			o := blockInterface.GetObject()
			block := s.getConcreteBlockObject(o)
			payload := block.(cs.Payload)
			response.Blocks = append(response.Blocks, payload)
			/*o1 := blockInterface.GetObject()

			typeStr := reflect.TypeOf(o1).String()
			if strings.Contains(typeStr, "CMBlock") {
				o := o1.(cs.CMBlock)
				payload := interface{}(&o).(cs.Payload)
				response.Blocks = append(response.Blocks, payload)
			} else if strings.Contains(typeStr, "MinorBlock") {
				o := o1.(cs.MinorBlock)
				payload := interface{}(&o).(cs.Payload)
				response.Blocks = append(response.Blocks, payload)
			} else if strings.Contains(typeStr, "FinalBlock") {
				o := o1.(cs.FinalBlock)
				payload := interface{}(&o).(cs.Payload)
				response.Blocks = append(response.Blocks, payload)
			} else if strings.Contains(typeStr, "ViewChangeBlock") {
				o := o1.(cs.ViewChangeBlock)
				payload := interface{}(&o).(cs.Payload)
				response.Blocks = append(response.Blocks, payload)
			} else {
				log.Error("wrong block type, ", typeStr)
			}*/
		}
	}

	data := response.Encode(blockType, shardID)

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
func (s *Sync) dealSyncRequest(request *sc.SyncRequestPacket) (*sc.NetPacket, *sc.WorkerId) {

	worker := request.Worker
	csp := s.DealSyncRequestHelper(request)

	return csp, worker
}

func (s *Sync)  RecvSyncRequestPacket(packet *sc.CsPacket) (*sc.NetPacket, *sc.WorkerId){
	requestPacket := packet.Packet.(*sc.SyncRequestPacket)
	return s.dealSyncRequest(requestPacket)
}

func (s *Sync) CheckSyncCompleteForMinorBlock(toHeight uint64, blocks *[]cs.BlockInterface, syncResponse *sc.SyncResponsePacket) bool {
	len := len(*blocks)
	if len > 0 && (*blocks)[len-1].GetHeight() >= toHeight {
		return true
	} else {
		return false
	}
}

func (s *Sync) CheckSyncCompleteForCMBlock(hash *common.Hash, blocks *[]cs.BlockInterface) bool {
	list := *blocks
	len := len(list)
	for i := len-1; i >= 0; i-- {
		h := list[i].Hash()
		if h.Equals(hash) {
			return  true
		}
	}
	return false
 }

func (s *Sync) CheckSyncComplete(syncResponse *sc.SyncResponsePacket) bool  {
	log.Debug("CheckSyncComplete blockType = ",
		syncResponse.BlockType, " complete = ", syncResponse.Compelte)
	log.Debug("Block type = ", uint8(syncResponse.BlockType))
	var lastFinalBlock cs.FinalBlock
	if syncResponse.BlockType == cs.HeFinalBlock && syncResponse.Compelte {
		log.Debug("CheckSyncComplete, set complete")
		s.cache.finalBlockComplete = true
		blocks := s.cache.blocks[cs.HeFinalBlock]
		len := len(blocks)
		log.Debug("final block cache len = ",len)
		if len == 0 {
			s.cache.finalBlockComplete = true
			return true
		} else {
			lastFinalBlock = blocks[len-1].GetObject().(cs.FinalBlock)
		}
	}
	if s.cache.finalBlockComplete {

		for _, minorBlock := range lastFinalBlock.MinorBlocks {
			blocks := s.cache.minorBlocks[minorBlock.ShardId]
			complete := s.CheckSyncCompleteForMinorBlock(minorBlock.GetHeight(), &blocks,  syncResponse)
			if !complete {
				return false
			}
		}
		//TODO, check cm block
		blocks := s.cache.blocks[cs.HeCmBlock]
		if !s.CheckSyncCompleteForCMBlock(&lastFinalBlock.CMBlockHash, &blocks) {
			return false
		}


		return true



	} else {
		return false
	}
}

func (s *Sync) getConcreteBlockObject(o interface{}) interface{}  {
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
	list := *p
	blockType := syncResponse.BlockType
	for _, payload := range syncResponse.Blocks {
		o := payload.GetObject()
		var block cs.BlockInterface
		block = s.getConcreteBlockObject(o).(cs.BlockInterface)

		len := len(list)
		var needH uint64
		if len > 0 {
			needH = list[len-1].GetHeight()+1
		} else {
			needH = needHeight//s.cache.needHeight[blockType]
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
	} else {
		shardID := syncResponse.ShardID
		list := s.cache.minorBlocks[shardID]
		s.FillSyncDataInCacheHelper(&list, s.cache.needHeightMinor[shardID], syncResponse)
	}


}

func (s *Sync)  RecvSyncResponsePacket(packet *sc.CsPacket){
	data := packet.Packet.(*sc.SyncResponseData)

	p := s.SyncResponseDecode(data)
	s.FillSyncDataInCache(p)

	if s.CheckSyncComplete(p) {
		s.tellLedgerSyncComplete()
		simulate.SyncComplete()
		log.Info("invoke SyncComplete")
	} else {
		log.Info("Data sync not complete")
		s.SendSyncRequest()
	}
}

func (s *Sync) DealSyncRequestHelperTest(request *sc.SyncRequestPacket) (*sc.NetPacket)  {
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

	data := response.Encode(cs.HeMinorBlock, 0 )

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




