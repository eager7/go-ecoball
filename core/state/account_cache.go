package state

import (
	"fmt"
	. "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/hashicorp/golang-lru"
	"sync"
)

type AccountCache struct {
	lock         sync.RWMutex
	AccountCache *lru.Cache
}

func (a *AccountCache) Initialize() error {
	csc, err := lru.New(10000)
	if err != nil {
		return errors.New(fmt.Sprintf("New Lru error:%s", err.Error()))
	}
	a.AccountCache = csc
	return nil
}

func (a *AccountCache) Add(acc *Account) {
	a.AccountCache.Add(acc.Index, acc)
}

func (a *AccountCache) Get(index AccountName) *Account {
	if value, ok := a.AccountCache.Get(index); ok {
		return value.(*Account)
	}
	return nil
}

func (a *AccountCache) Contains(index AccountName) bool {
	return a.AccountCache.Contains(index)
}

func (a *AccountCache) Purge() {
	a.AccountCache.Purge()
}
