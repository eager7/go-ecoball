package simulate

import (
	"github.com/ecoball/go-ecoball/common"
	cc "github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/types"
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

func TellBlock(bl cs.BlockInterface) {
	log.Debug("tell ledger block type ", bl.Type(), " height ", bl.GetHeight())
	if err := event.Send(event.ActorSharding, event.ActorLedger, bl); err != nil {
		log.Fatal(err)
	}
}

func TellLedgerProductFinalBlock(epoch uint64, height uint64, hashes []common.Hash) {
	log.Debug("tell ledger product final block hashes ", len(hashes))
	if len(hashes) > 0 {
		log.Debug(hashes[0])
	}

	pb := message.ProducerBlock{
		ChainID: cc.ChainHash,
		Height:  height,
		Type:    cs.HeFinalBlock,
	}

	pb.Hashes = append(pb.Hashes, hashes...)

	if err := event.Send(event.ActorSharding, event.ActorLedger, pb); err != nil {
		log.Fatal(err)
	}
}

func TellLedgerProductMinorBlock(epoch uint64, height uint64) {
	log.Debug("tell ledger product minor block")

	pb := message.ProducerBlock{
		ChainID: cc.ChainHash,
		Height:  height,
		Type:    cs.HeMinorBlock,
	}

	if err := event.Send(event.ActorSharding, event.ActorLedger, pb); err != nil {
		log.Fatal(err)
	}
}

func CheckFinalBlock(f *cs.FinalBlock) bool {
	log.Error("ledger check final block")
	return true
}

//func TellLedgerProductMinorBlock(epoch uint64, height uint64, shardid uint32) {
//	log.Error("tell ledger product minor block")
//
//	minor := createMinorBlock(epoch, height, shardid)
//	if minor == nil {
//		return
//	}
//
//	if err := event.Send(event.ActorLedger, event.ActorSharding, minor); err != nil {
//		log.Fatal(err)
//	}
//}
//
//func CheckMinorBlock(b *cs.MinorBlock) bool {
//	log.Error("ledger check minor block")
//	return true
//}

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

func GetCandidateList() (workers []sc.Worker) {
	workers = make([]sc.Worker, 0, 0)
	return
}

func createFinalBlock(epoch uint64, height uint64) *cs.FinalBlock {

	final := &cs.FinalBlock{
		FinalBlockHeader: cs.FinalBlockHeader{
			ChainID:            common.Hash{},
			Version:            0,
			Height:             0,
			Timestamp:          0,
			TrxCount:           0,
			PrevHash:           common.Hash{},
			ProposalPubKey:     nil,
			EpochNo:            0,
			CMBlockHash:        common.Hash{},
			TrxRootHash:        common.Hash{},
			StateDeltaRootHash: common.Hash{},
			MinorBlocksHash:    common.Hash{},
			StateHashRoot:      common.Hash{},
			COSign:             nil,
		},
		MinorBlocks: nil,
	}
	final.Height = height
	final.EpochNo = epoch

	cosign := &types.COSign{}
	cosign.Step1 = 1
	cosign.Step2 = 0

	final.COSign = cosign

	log.Debug("create final block epoch ", epoch, " height ", height)

	return final

}

func createMinorBlock(epoch uint64, height uint64, sharid uint32) *cs.MinorBlock {

	minor := &cs.MinorBlock{
		MinorBlockHeader: cs.MinorBlockHeader{
			ChainID:           common.Hash{},
			Version:           0,
			Height:            0,
			Timestamp:         0,
			PrevHash:          common.Hash{},
			TrxHashRoot:       common.Hash{},
			StateDeltaHash:    common.Hash{},
			CMBlockHash:       common.Hash{},
			ProposalPublicKey: nil,
			ShardId:           0,
			CMEpochNo:         0,
			Receipt:           types.BlockReceipt{},
			COSign:            nil,
		},
		Transactions: nil,
		StateDelta:   nil,
	}

	minor.Height = height
	minor.CMEpochNo = epoch
	minor.ShardId = sharid

	cosign := &types.COSign{}
	cosign.Step1 = 1
	cosign.Step2 = 0

	minor.COSign = cosign

	log.Debug(" create minor block epoch ", minor.CMEpochNo, " height ", minor.Height)

	return minor
}
