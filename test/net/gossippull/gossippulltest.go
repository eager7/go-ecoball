package gossippull
import (
	"github.com/ecoball/go-ecoball/net/gossip"
	"strconv"
	"time"
	"github.com/ecoball/go-ecoball/net/gossip/protos"
	"context"
)

func StartBlockPuller(ctx context.Context) {
	cfg := gossip.PullConfig{
		ChainId:  1,
		PullPeersCount: 3,
		PullInterval: time.Duration(5)*time.Second,
		MsgType: pb.PullMsgType_BLOCK_MSG,
	}
	receiver := NewPullReceiver()

	gossip.NewPullMediator(ctx, cfg, receiver)
}


func NewPullReceiver() *PullReceiver {
	pr := &PullReceiver{
		data: make([][]byte, 0),
	}
	for i:=0; i<3; i++ {
		pr.data = append(pr.data, []byte(strconv.Itoa(int(i))+"abcd"))
	}

	return pr
}

type PullReceiver struct {
	gossip.Receiver
	data [][]byte
}

func (pr *PullReceiver) GetDigests() []string {
	digest := []string{strconv.Itoa(len(pr.data)-1)}
	return digest
}

func (pr *PullReceiver) ContainItemInDigests(digests []string, item string) bool {
	// only one digest in this example
	sd, _ := strconv.ParseInt(digests[0], 10, 32)
	id, _ := strconv.ParseInt(item, 10, 32)
	if sd >= id {
		return true
	}

	return false
}

func (pr *PullReceiver) ShuffelDigests(revDigests map[string][]string, digest []string) map[string][]string {
	result := make(map[string][]string)
	for key, v := range revDigests {
		// only one digest in this example
		sd, _ := strconv.ParseInt(v[0], 10, 32)
		id, _ := strconv.ParseInt(digest[0], 10, 32)
		for id=id+1; id <= sd; id++ {
			result[key] = append(result[key], strconv.Itoa(int(id)))
		}
		break
	}
	return result
}

func (pr *PullReceiver) GetItemData(item string) []byte {
	id, _ := strconv.ParseInt(item, 10, 32)
	if int(id) >= len(pr.data) {
		return nil
	}
	return pr.data[id]

}

func (pr *PullReceiver) UpdateItemData(dataArray [][]byte) error {
	for _, data := range dataArray {
		pr.data = append(pr.data, data)
		//pr.data[len(pr.data)] = data
	}

	return nil
}
