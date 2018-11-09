package datasync

import (
	"testing"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/cell"
	"fmt"
	sh "github.com/ecoball/go-ecoball/core/shard"
)

func TestSync_SendSyncRequest(t *testing.T) {
	//var worker *sc.WorkerId
	worker := &sc.WorkerId{
		"111",
		"127.0.0.1",
		"8888",
	}
	fmt.Println("First")
	requestPacket := MakeSyncRequestPacket(1, 10, -1,  worker, 0)
	fmt.Println("Second")
	cell := &cell.Cell{}
	sync := MakeSync(cell)
	csp := cell.VerifySyncRequestPacket(requestPacket)
	fmt.Println("Third")

	//s := shard.MakeShardTest(cell)
	//s.Start()
	fmt.Println("4th")
	requestPacket1 := csp.Packet.(*sc.SyncRequestPacket)
	fmt.Println(requestPacket1.Worker, requestPacket1.ToHeight, requestPacket1.FromHeight)
	fmt.Println("5th")
	responseNetPacket := sync.DealSyncRequestHelperTest(requestPacket1)
	fmt.Println("6th")

	responseCsp := cell.VerifySyncResponsePacket(responseNetPacket)
	fmt.Println("7th")
	data := responseCsp.Packet.(*sc.SyncResponseData)
	fmt.Println("8th")

	fmt.Println("SyncResponseData = ", data)
	responsePacket := sync.SyncResponseDecode(data)
	blocks := responsePacket.Blocks
	for _, payload := range blocks {
		block := payload.(*sh.MinorBlock)
		fmt.Println(block.Height, block.Version)
	}
	fmt.Println("9th")
}