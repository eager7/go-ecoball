package mobsync

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/gogo/protobuf/sortkeys"
	"sync"
)

type BlockMap struct {
	Blocks map[uint64]*types.Block
	lock   sync.RWMutex
}

func (b *BlockMap) Initialize() *BlockMap {
	b.Blocks = make(map[uint64]*types.Block)
	return b
}

func (b *BlockMap) Add(height uint64, block *types.Block) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if _, ok := b.Blocks[height]; ok {
		return
	}
	b.Blocks[height] = block
}

func (b *BlockMap) Del(height uint64) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if _, ok := b.Blocks[height]; ok {
		delete(b.Blocks, height)
	}
}

func (b *BlockMap) Get(height uint64) *types.Block {
	b.lock.RLock()
	defer b.lock.RUnlock()
	if info, ok := b.Blocks[height]; ok {
		return info
	}
	return nil
}

func (b *BlockMap) Iterator() <-chan *types.Block {
	channel := make(chan *types.Block)
	go func() {
		b.lock.RLock()
		defer b.lock.RUnlock()
		for _, v := range b.Blocks {
			channel <- v
		}
		close(channel)
	}()
	return channel
}

func (b *BlockMap) IteratorByHeight(chainId common.Hash) <-chan *types.Block {
	b.lock.RLock()
	var keys []uint64
	for h := range b.Blocks {
		keys = append(keys, h)
	}
	b.lock.RUnlock()
	sortkeys.Uint64s(keys)

	channel := make(chan *types.Block)
	go func() {
		for _, key := range keys {
			channel <- b.Get(key)
		}
		close(channel)
	}()
	return channel
}

type ChainMap struct {
	Chains map[common.Hash]BlockMap
	lock   sync.RWMutex
}

func (c *ChainMap) Initialize() *ChainMap {
	c.Chains = make(map[common.Hash]BlockMap)
	return c
}

func (c *ChainMap) Add(chainId common.Hash, block *types.Block) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if chain, ok := c.Chains[chainId]; ok {
		chain.Add(block.Height, block)
		return
	}
	chain := new(BlockMap).Initialize()
	chain.Add(block.Height, block)
	c.Chains[chainId] = *chain

}

func (c *ChainMap) Del(chainId common.Hash, block *types.Block) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if chain, ok := c.Chains[chainId]; ok {
		chain.Del(block.Height)
	}
}

func (c *ChainMap) Get(chainId common.Hash) *BlockMap {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if chain, ok := c.Chains[chainId]; ok {
		return &chain
	}
	return nil
}

func (c *ChainMap) Iterator(chainId common.Hash) <-chan *types.Block {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if chain, ok := c.Chains[chainId]; ok {
		return chain.Iterator()
	}
	return nil
}
