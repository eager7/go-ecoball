package mutex

import "sync"

type Mutex struct {
	m sync.RWMutex
}

func (m *Mutex) Lock() {
	m.m.Lock()
}

func (m *Mutex) Unlock() {
	m.m.Unlock()
}

func (m *Mutex) RLock() {
	m.m.RLock()
}

func (m *Mutex) RUnlock() {
	m.m.RUnlock()
}