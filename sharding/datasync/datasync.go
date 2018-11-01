package datasync

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/sharding/net"
	"github.com/ecoball/go-ecoball/sharding/cell"
	"github.com/ecoball/go-ecoball/net/message/pb"
)

var (
	log = elog.NewLogger("sync", elog.DebugLog)
)

type DataSync interface {

	SendSyncRequest(fromHeight int64)
}


type sync struct {
	syncType int
	cell *cell.Cell
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


func (sync *sync)SendSyncRequest(blockType int8, fromHeight int64)  {
	worker := &sc.WorkerId{
		sync.cell.Self.Pubkey,
		sync.cell.Self.Address,
		sync.cell.Self.Port,
	}
	csp := MakeSyncRequestPacket(blockType, fromHeight, -1, worker)

	net.Np.SendSyncMessage(csp)
}

func (sync *sync)SendSyncRequestTo(blockType int8, fromHeight int64, toHeight int64)  {
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
func (sync *sync)processShardInternalSync(packet *sc.SyncPacket)  {
	switch packet.MessageType {
	case sc.SyncRequest:
		//TODO,response with block data
	case sc.SyncResponse:
		//TODO, add into chain(ledger)
	}
}

func (sync *sync)processSyncPacket(packet *sc.SyncPacket) {

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




