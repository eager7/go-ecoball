package state

import (
	"github.com/ecoball/go-ecoball/common"
	"sync"
)

type Param struct {
	Key   string
	Value uint64
}

type ParamsMap struct {
	Params map[string]Param
	lock   sync.RWMutex
}

func (p *ParamsMap) Initialize() {
	p.Params = make(map[string]Param)
}

func (p *ParamsMap) Len() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return len(p.Params)
}

func (p *ParamsMap) Add(key string, value uint64) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Params[key] = Param{Key: key, Value: value}
}

func (p *ParamsMap) Get(key string) *Param {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if k, ok := p.Params[key]; ok {
		return &k
	} else {
		return nil
	}
}

func (p *ParamsMap) Contains(key string) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if _, ok := p.Params[key]; ok {
		return true
	} else {
		return false
	}
}

func (p *ParamsMap) Del(key string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.Params[key]; ok {
		delete(p.Params, key)
	}
}

func (p *ParamsMap) Purge() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for k := range p.Params {
		delete(p.Params, k)
	}
}

func (p *ParamsMap) Clone() ParamsMap {
	n := ParamsMap{}
	n.Initialize()
	for k, v := range p.Params {
		n.Params[k] = v
	}
	return n
}

func (p *ParamsMap) Iterator() <-chan Param {
	channel := make(chan Param)
	go func() {
		p.lock.RLock()
		defer p.lock.RUnlock()
		for _, v := range p.Params {
			channel <- v
		}
		close(channel)
	}()
	return channel
}

/**********************************************************************************************************************/
type ProducersMap struct {
	Producers map[common.AccountName]Producer
	lock      sync.RWMutex
}

func (p *ProducersMap) Initialize() {
	p.Producers = make(map[common.AccountName]Producer)
}

func (p *ProducersMap) Len() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return len(p.Producers)
}

func (p *ProducersMap) Add(key common.AccountName, value uint64) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Producers[key] = Producer{Index: key, Amount: value}
}

func (p *ProducersMap) Get(key common.AccountName) *Producer {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if k, ok := p.Producers[key]; ok {
		return &k
	} else {
		return nil
	}
}

func (p *ProducersMap) Contains(key common.AccountName) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if _, ok := p.Producers[key]; ok {
		return true
	} else {
		return false
	}
}

func (p *ProducersMap) Del(key common.AccountName) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.Producers[key]; ok {
		delete(p.Producers, key)
	}
}

func (p *ProducersMap) Purge() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for k := range p.Producers {
		delete(p.Producers, k)
	}
}

func (p *ProducersMap) Clone() ProducersMap {
	n := ProducersMap{}
	n.Initialize()
	for k, v := range p.Producers {
		n.Producers[k] = v
	}
	return n
}

func (p *ProducersMap) Iterator() <-chan Producer {
	channel := make(chan Producer)
	go func() {
		p.lock.RLock()
		defer p.lock.RUnlock()
		for _, v := range p.Producers {
			channel <- v
		}
		close(channel)
	}()
	return channel
}

/**********************************************************************************************************************/
type TokensMap struct {
	Tokens map[string]TokenInfo
	lock   sync.RWMutex
}

func (t *TokensMap) Initialize() {
	t.Tokens = make(map[string]TokenInfo)
}

func (t *TokensMap) Len() int {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return len(t.Tokens)
}

func (t *TokensMap) Add(key string, value TokenInfo) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.Tokens[key] = value
}

func (t *TokensMap) Get(key string) *TokenInfo {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if k, ok := t.Tokens[key]; ok {
		return &k
	} else {
		return nil
	}
}

func (t *TokensMap) Contains(key string) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if _, ok := t.Tokens[key]; ok {
		return true
	} else {
		return false
	}
}

func (t *TokensMap) Del(key string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if _, ok := t.Tokens[key]; ok {
		delete(t.Tokens, key)
	}
}

func (t *TokensMap) Purge() {
	t.lock.Lock()
	defer t.lock.Unlock()
	for k := range t.Tokens {
		delete(t.Tokens, k)
	}
}

func (t *TokensMap) Clone() TokensMap {
	n := TokensMap{}
	n.Initialize()
	for k, v := range t.Tokens {
		n.Tokens[k] = v
	}
	return n
}

func (t *TokensMap) Iterator() <-chan TokenInfo {
	channel := make(chan TokenInfo)
	go func() {
		t.lock.RLock()
		defer t.lock.RUnlock()
		for _, v := range t.Tokens {
			channel <- v
		}
		close(channel)
	}()
	return channel
}

/**********************************************************************************************************************/
type Chain struct {
	Hash    common.Hash
	TxHash  common.Hash
	Address common.Address
	Index   common.AccountName
}

type ChainsMap struct {
	Chains map[common.Hash]Chain
	lock   sync.RWMutex
}

func (c *ChainsMap) Initialize() {
	c.Chains = make(map[common.Hash]Chain)
}

func (c *ChainsMap) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.Chains)
}

func (c *ChainsMap) Add(key common.Hash, value Chain) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Chains[key] = value
}

func (c *ChainsMap) Get(key common.Hash) *Chain {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if k, ok := c.Chains[key]; ok {
		return &k
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

func (c *ChainsMap) Iterator() <-chan Chain {
	channel := make(chan Chain)
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
