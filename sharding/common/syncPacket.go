package common

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
	"fmt"
)

//SyncType
const (
	ShardInternal = iota
	ShardToShard
	CommiteeInternal
	CommiteeToShard
	ShardToCommitee
)

//MessageType
const (
	SyncRequest = iota
	SyncResponse
)

type SyncPacket struct {
	SyncType  uint16
	MessageType uint8
	Packet interface{}
}

//toHeight = -1, means to newest height
type SyncRequestPacket struct {
	BlockType int8
	FromHeight int64
	ToHeight int64
	Worker *WorkerId
	ShardID uint32
}


type SyncResponsePacket struct {
	Len uint
	BlockType cs.HeaderType
	Blocks []cs.Payload
	LastHeight uint64
	ShardID uint32
	Compelte bool
}



type SyncResponseData struct {
	BlockType cs.HeaderType
	Len uint
	Data [][]byte
	LastHeight uint64
	ShardID uint32
	Compelte bool
}


func (p *SyncResponsePacket)Encode(blockType cs.HeaderType, shardID uint32) *SyncResponseData {
	var data [][]byte
	p.BlockType = blockType
	p.Len = uint(len(p.Blocks))
	fmt.Println("Encoding", p.Len)
	for i := uint(0); i < p.Len; i++ {
		fmt.Println("Encoding block ", p.Blocks[i] )
		blockData, err := p.Blocks[i].Serialize()

		if err != nil {
			log.Error("SyncResponseData Serialize")
			fmt.Println("SyncResponseData Serialize error")
		}
		data = append(data, blockData)
	}
	syncData := &SyncResponseData {
		p.BlockType,
		p.Len,
		data,
		p.LastHeight,
		shardID,
		p.Compelte,
	}
	return syncData
}

type DataSync interface {

	SendSyncRequest()
}




