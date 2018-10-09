package simulate

import (
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/types"
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

func GetMinorBlockPool() []*types.MinorBlock {
	return nil
}

/*minor block to be packed by committee*/
func GetPreproductionMinorBlock() *types.MinorBlock {
	return nil
}
