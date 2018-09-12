package types

import "github.com/ecoball/go-ecoball/common"

type FinalBlockHeader struct {
	ProposalPubKey     []byte
	CMEpochNo          uint64
	CMBlockHash        common.Hash
	TrxCount           uint32
	TrxRootHash        common.Hash
	StateDeltaRootHash common.Hash
	MinorBlocksHash    common.Hash
}

type FinalBlock struct {
	*FinalBlockHeader
	MinorBlocks []MinorBlockHeader
}

func (b *Block) SetFinalBlockData() {

}