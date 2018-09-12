package node

import (
	"github.com/ecoball/go-ecoball/core/types/block"
)

type minorBlockSet struct {
	blocks []*block.MinorBlock
}

func makeMinorBlockSet() *minorBlockSet {
	return &minorBlockSet{blocks: make([]*block.MinorBlock, 0, 10)}
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

func (m *minorBlockSet) setMinorBlock(minor *block.MinorBlock) {
	shardid := minor.ShardId
	if int(shardid) > len(m.blocks) || shardid < 1 {
		log.Error("set minorBlock error shardid %d", shardid)
		return
	}

	m.blocks[shardid-1] = minor
}

func (m *minorBlockSet) syncMinorBlocks(minors []*block.MinorBlock) {
	if len(m.blocks) != len(minors) {
		panic("sync minor block length error")
		log.Panic("sync minor block error len %d sync blocks len%d", len(m.blocks), len(minors))
		return
	}

	for i := 0; i < len(m.blocks); i++ {
		m.blocks[i] = minors[i]
	}
}

func (m *minorBlockSet) getMinorBlock(shardid uint16) *block.MinorBlock {
	if int(shardid) > len(m.blocks) || shardid < 1 {
		log.Error("get minorBlock error shardid %d", shardid)
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
