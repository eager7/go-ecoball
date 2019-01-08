package trie

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/store"
	"sync"
)

type Mpt struct {
	path  string
	trier Trier
	db    TreeDB
	lock  sync.RWMutex
}

func NewMptTrie(path string, root common.Hash) (mpt *Mpt, err error) {
	mpt = &Mpt{path: path}
	levelDB, err := store.NewLevelDBStore(path, 0, 0)
	if err != nil {
		return nil, err
	}
	mpt.db = NewCachingDB(levelDB)
	log.Debug("Open Trie Hash:", path, root.HexString())
	mpt.trier, err = mpt.db.OpenTrie(root)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("open mpt failed:%s", err))
	}
	return mpt, nil
}

func (m *Mpt) Put(key, value []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.trier.TryUpdate(key, value)
}

func (m *Mpt) Get(key []byte) ([]byte, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.trier.TryGet(key)
}

func (m *Mpt) Del(key []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.trier.TryDelete(key)
}

func (m *Mpt) Commit() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	root, err := m.trier.Commit(nil)
	if err != nil {
		return err
	}
	return m.db.TrieDB().Commit(root, false)
}

func (m *Mpt) RollBack(root common.Hash) (err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.trier, err = m.db.OpenTrie(root)
	return err
}

func (m *Mpt) Clone() *Mpt {
	m.lock.Lock()
	defer m.lock.Unlock()
	return &Mpt{path: m.path, trier: m.db.CopyTrie(m.trier), db: m.db, lock: sync.RWMutex{}}
}

func (m *Mpt) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.db.TrieDB().diskDB.Close()
}

func (m *Mpt) Hash() common.Hash {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.trier.Hash()
}

func (m *Mpt) Path() string {
	return m.path
}
