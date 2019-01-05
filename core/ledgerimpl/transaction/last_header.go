package transaction

import (
	"github.com/ecoball/go-ecoball/core/types"
	"sync"
)

type LastHeader struct {
	Header *types.Header
	lock   sync.RWMutex
}

func (l *LastHeader) Set(header *types.Header) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.Header = header
}

func (l *LastHeader) Get() *types.Header {
	l.lock.RLock()
	defer l.lock.RUnlock()
	if l.Header == nil {
		return nil
	}
	if h, err := l.Header.Clone(); err != nil {
		return l.Header
	} else {
		return h
	}
}
