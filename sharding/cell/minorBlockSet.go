package cell

import (
	"github.com/ecoball/go-ecoball/common"
	cs "github.com/ecoball/go-ecoball/core/shard"
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

type minorBlockSet struct {
	blocks []*cs.MinorBlock
}

func makeMinorBlockSet() *minorBlockSet {
	return &minorBlockSet{blocks: make([]*cs.MinorBlock, sc.DefaultShardMaxMember, sc.DefaultShardMaxMember)}
}

func (m *minorBlockSet) clean() {
	for i := 0; i < len(m.blocks); i++ {
		m.blocks[i] = nil
	}
}

func (m *minorBlockSet) saveMinorBlock(minor *cs.MinorBlock) bool {
	shardid := minor.ShardId
	if int(shardid) > len(m.blocks) || shardid < 1 {
		log.Error("save minorBlock error shardid ", shardid)
		return false
	}

	if m.blocks[shardid-1] != nil {
		log.Debug("minorBlock already exist")
		return false
	}

	m.blocks[shardid-1] = minor
	return true
}

func (m *minorBlockSet) syncMinorBlocks(minors []*cs.MinorBlock) {
	if len(m.blocks) < len(minors) {
		panic("sync minor block length error")
		log.Panic("sync minor block error len ", len(m.blocks), "  sync blocks len ", len(minors))
		return
	}

	for i := 0; i < len(m.blocks); i++ {
		m.blocks[i] = minors[i]
	}
}

func (m *minorBlockSet) getMinorBlockHashes() (hashes []common.Hash) {
	for _, minor := range m.blocks {
		if minor != nil {
			hashes = append(hashes, minor.Hash())
		}
	}
	return
}

func (m *minorBlockSet) getMinorBlock(shardid uint16) *cs.MinorBlock {
	if int(shardid) > len(m.blocks) || shardid < 1 {
		log.Error("get minorBlock error shardid ", shardid)
		return nil
	}

	return m.blocks[shardid-1]
}

func (m *minorBlockSet) count() uint16 {
	var length uint16
	for i := 0; i < len(m.blocks); i++ {
		if m.blocks[i] != nil {
			length++
		}
	}

	return length
}
