package cell

import (
	"github.com/ecoball/go-ecoball/core/types"
)

type minorBlockSet struct {
	blocks []*types.MinorBlock
}

func makeMinorBlockSet() *minorBlockSet {
	return &minorBlockSet{blocks: make([]*types.MinorBlock, 0, 10)}
}

func (m *minorBlockSet) resize(size int) {
	m.blocks = m.blocks[:size]
	m.clean()
}

func (m *minorBlockSet) clean() {
	for i := 0; i < len(m.blocks); i++ {
		m.blocks[i] = nil
	}
}

func (m *minorBlockSet) setMinorBlock(minor *types.MinorBlock) {
	shardid := minor.ShardId
	if int(shardid) > len(m.blocks) || shardid < 1 {
		log.Error("set minorBlock error shardid ", shardid)
		return
	}

	m.blocks[shardid-1] = minor
}

func (m *minorBlockSet) syncMinorBlocks(minors []*types.MinorBlock) {
	if len(m.blocks) != len(minors) {
		panic("sync minor block length error")
		log.Panic("sync minor block error len ", len(m.blocks), "  sync blocks len ", len(minors))
		return
	}

	for i := 0; i < len(m.blocks); i++ {
		m.blocks[i] = minors[i]
	}
}

func (m *minorBlockSet) getMinorBlock(shardid uint16) *types.MinorBlock {
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
