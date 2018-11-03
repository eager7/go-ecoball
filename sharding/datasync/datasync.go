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
)

var (
	log = elog.NewLogger("sync", elog.DebugLog)
)



type Sync struct {
	syncType int
	cell *cell.Cell
}

func MakeSync(c *cell.Cell) *Sync {
	return &Sync{cell: c}
}

func MakeSyncRequestPacket(blockType int8, fromHeight int64, to int64, worker *sc.WorkerId) (*sc.NetPacket) {
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
	sync.SendSyncRequestWithType(cs.HeCmBlock)
	sync.SendSyncRequestWithType(cs.HeViewChange)
	sync.SendSyncRequestWithType(cs.HeFinalBlock)
	sync.SendSyncRequestWithType(cs.HeMinorBlock)
}

func (sync *Sync)SendSyncRequestWithType(blockType cs.HeaderType) {
	lastBlock, err := sync.cell.Ledger.GetLastShardBlock(config.ChainHash, cs.HeCmBlock)
	if err != nil {
		log.Error("get last block faield", err)
		return
	}
	height := lastBlock.GetHeight() + 1
	sync.SendSyncRequestWithHeightType(int8(blockType), int64(height))
}


func (sync *Sync)SendSyncRequestWithHeightType(blockType int8, fromHeight int64)  {
	worker := &sc.WorkerId{
		sync.cell.Self.Pubkey,
		sync.cell.Self.Address,
		sync.cell.Self.Port,
	}
	csp := MakeSyncRequestPacket(blockType, fromHeight, -1, worker)

	net.Np.SendSyncMessage(csp)
}

func (sync *Sync)SendSyncRequestTo(blockType int8, fromHeight int64, toHeight int64)  {
	worker := &sc.WorkerId{
		sync.cell.Self.Pubkey,
		sync.cell.Self.Address,
		sync.cell.Self.Port,
	}
	csp := MakeSyncRequestPacket(blockType, fromHeight, toHeight, worker)

	net.Np.SendSyncMessage(csp)
}

/*func (sync *Sync)dealSyncRequest() {

}*/

//TODO, now only treat shardInternal and commiteeInternal
func (sync *Sync)processShardInternalSync(packet *sc.SyncPacket)  {
	switch packet.MessageType {
	case sc.SyncRequest:
		//TODO,response with block data
	case sc.SyncResponse:
		//TODO, add into chain(ledger)
	}
}

func (sync *Sync)processSyncPacket(packet *sc.SyncPacket) {

	switch packet.SyncType {
	case sc.ShardInternal:
		sync.processShardInternalSync(packet)
	case sc.CommiteeInternal:

	case sc.ShardToShard:

	case sc.ShardToCommitee:

	case sc.CommiteeToShard:

	default:
		log.Error("wrong packet")
	}
}





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
		syncData.Compelte,
	}

	return csp
}

//TODO, make sure TellBlock will be all right
func (s *Sync) dealSyncResponse(response *sc.SyncResponsePacket) {
	blocks := response.Blocks
	for _, block := range blocks {
		simulate.TellBlock(block.(cs.BlockInterface))
	}
}

func (s *Sync) DealSyncRequestHelper(request *sc.SyncRequestPacket) (*sc.NetPacket)  {
	from := request.FromHeight
	to := request.ToHeight
	blockType := cs.HeaderType(request.BlockType)
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



	fmt.Println("to = ", to)


	for i := from; i <= to; i++ {
		blockInterface, err := s.cell.Ledger.GetShardBlockByHeight(config.ChainHash, blockType, uint64(i))
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
func (s *Sync) dealSyncRequest(request *sc.SyncRequestPacket) (*sc.NetPacket, *sc.WorkerId) {

	worker := request.Worker
	csp := s.DealSyncRequestHelper(request)

	return csp, worker
}

func (s *Sync)  RecvSyncRequestPacket(packet *sc.CsPacket) (*sc.NetPacket, *sc.WorkerId){
	requestPacket := packet.Packet.(*sc.SyncRequestPacket)
	return s.dealSyncRequest(requestPacket)
}

func (s *Sync)  RecvSyncResponsePacket(packet *sc.CsPacket){
	data := packet.Packet.(*sc.SyncResponseData)

	p := s.SyncResponseDecode(data)
	s.dealSyncResponse(p)
	if p.Compelte {
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




