package simulate

import (
	"github.com/ecoball/go-ecoball/core/types/block"
)

func SendSyncCompleteMsg() {

}

func GetCMBlockByNumber(heigh uint64) *block.CMBlock {
	return nil
}

func GetLastCMBlock() *block.CMBlock {
	return nil
}

func GetFinalBlockByNumber() *block.FinalBlock {
	return nil
}

func GetLastFinalBlock() *block.FinalBlock {
	return nil
}

func GetMinorBlockByNumber() *block.FinalBlock {
	return nil
}

func GetLastMinorBlock() *block.FinalBlock {
	return nil
}

func GetMinorBlockPool() []*block.MinorBlock {
	return nil
}

func GetProducerList() []*block.NodeInfo {
	return nil
}

func GetSyncStatus() bool {
	return true
}

func GetCandidate() []NodeConfig {
	if !configLoad {
		LoadConfig()
	}

	return candidate
}

func TellBlock(bl interface{}) {

}
