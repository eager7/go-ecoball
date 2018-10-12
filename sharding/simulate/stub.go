package simulate

import (
	"github.com/ecoball/go-ecoball/common/event"
	cs "github.com/ecoball/go-ecoball/core/shard"
)

func TellBlock(bl interface{}) {
	if err := event.Send(event.ActorSharding, event.ActorLedger, bl); err != nil {
		log.Fatal(err)
	}
}

func TellMinorBlock(bl interface{}) {
	log.Error("tell ledger minor block")
}

func GetSyncStatus() bool {
	return true
}

func GetMinorBlockPool() []*cs.MinorBlock {
	return nil
}

/*minor block to be packed by committee*/
func GetPreproductionMinorBlock() *cs.MinorBlock {
	return nil
}

func GetCandidateList() (workers []NodeConfig) {
	return
}
