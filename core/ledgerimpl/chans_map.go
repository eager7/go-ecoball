package ledgerimpl

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/transaction"
	"sync"
)

type ChainsMap struct {
	Chains map[common.Hash]*transaction.ChainTx
	lock   sync.RWMutex
}

func (c *ChainsMap) Initialize() ChainsMap {
	c.Chains = make(map[common.Hash]*transaction.ChainTx)
	return *c
}

func (c *ChainsMap) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.Chains)
}

func (c *ChainsMap) Add(key common.Hash, value *transaction.ChainTx) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Chains[key] = value
}

func (c *ChainsMap) Get(key common.Hash) *transaction.ChainTx {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if k, ok := c.Chains[key]; ok {
		return k
	} else {
		return nil
	}
}

func (c *ChainsMap) Contains(key common.Hash) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if _, ok := c.Chains[key]; ok {
		return true
	} else {
		return false
	}
}

func (c *ChainsMap) Del(key common.Hash) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.Chains[key]; ok {
		delete(c.Chains, key)
	}
}

func (c *ChainsMap) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for k := range c.Chains {
		delete(c.Chains, k)
	}
}

func (c *ChainsMap) Clone() ChainsMap {
	n := ChainsMap{}
	n.Initialize()
	for k, v := range c.Chains {
		n.Chains[k] = v
	}
	return n
}

func (c *ChainsMap) Iterator() <-chan *transaction.ChainTx {
	channel := make(chan *transaction.ChainTx)
	go func() {
		c.lock.RLock()
		defer c.lock.RUnlock()
		for _, v := range c.Chains {
			channel <- v
		}
		close(channel)
	}()
	return channel
}
