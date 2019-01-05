package mobsync

import (
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/types"
	"testing"
)

func TestBlockMap_Iterator(t *testing.T) {
	mm := new(ChainMap).Initialize()
	mm.Add(config.ChainHash, &types.Block{Header: &types.Header{Height: 10}})
	mm.Add(config.ChainHash, &types.Block{Header: &types.Header{Height: 7}})
	mm.Add(config.ChainHash, &types.Block{Header: &types.Header{Height: 4}})
	mm.Add(config.ChainHash, &types.Block{Header: &types.Header{Height: 9}})
	mm.Add(config.ChainHash, &types.Block{Header: &types.Header{Height: 5}})
	if b := mm.Get(config.ChainHash); b != nil {
		for block := range b.IteratorByHeight(config.ChainHash) {
			log.Debug(block.String())
		}
	}
}


