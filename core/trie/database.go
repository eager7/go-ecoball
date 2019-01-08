// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package trie

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/hashicorp/golang-lru"
	"sync"
)

var MaxTrieCacheGen = uint16(120)

const (
	maxPastTries      = 12
	codeSizeCacheSize = 100000
)

type TreeDB interface {
	OpenTrie(root common.Hash) (Trier, error)
	OpenStorageTrie(addrHash, root common.Hash) (Trier, error)
	CopyTrie(Trier) Trier
	ContractCode(addrHash, codeHash common.Hash) ([]byte, error)
	ContractCodeSize(addrHash, codeHash common.Hash) (int, error)
	TrieDB() *Database
}

type Trier interface {
	TryGet(key []byte) ([]byte, error)
	TryUpdate(key, value []byte) error
	TryDelete(key []byte) error
	Commit(onLeaf LeafCallback) (common.Hash, error)
	Hash() common.Hash
	NodeIterator(startKey []byte) NodeIterator
	GetKey([]byte) []byte // TODO(fjl): remove this when SecureTrie is removed
	Prove(key []byte, fromLevel uint, proofDb store.Putter) error
}

type cachedTrie struct {
	*SecureTrie
	db *cachingDB
}

type cachingDB struct {
	db            *Database
	mu            sync.Mutex
	pastTries     []*SecureTrie
	codeSizeCache *lru.Cache
}

func NewCachingDB(db store.Database) TreeDB {
	csc, _ := lru.New(codeSizeCacheSize)
	return &cachingDB{
		db:            NewDatabase(db),
		codeSizeCache: csc,
	}
}

func (db *cachingDB) OpenTrie(root common.Hash) (Trier, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := len(db.pastTries) - 1; i >= 0; i-- {
		if db.pastTries[i].Hash() == root {
			return cachedTrie{db.pastTries[i].Copy(), db}, nil
		}
	}
	tr, err := NewSecure(root, db.db, MaxTrieCacheGen)
	if err != nil {
		return nil, err
	}
	return cachedTrie{tr, db}, nil
}

func (db *cachingDB) pushTrie(t *SecureTrie) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if len(db.pastTries) >= maxPastTries {
		copy(db.pastTries, db.pastTries[1:])
		db.pastTries[len(db.pastTries)-1] = t
	} else {
		db.pastTries = append(db.pastTries, t)
	}
}

func (db *cachingDB) OpenStorageTrie(addrHash, root common.Hash) (Trier, error) {
	return NewSecure(root, db.db, 0)
}

func (db *cachingDB) CopyTrie(t Trier) Trier {
	switch t := t.(type) {
	case cachedTrie:
		return cachedTrie{t.SecureTrie.Copy(), db}
	case *SecureTrie:
		return t.Copy()
	default:
		panic(fmt.Errorf("unknown trie type %T", t))
	}
}

func (db *cachingDB) ContractCode(addrHash, codeHash common.Hash) ([]byte, error) {
	code, err := db.db.Node(codeHash)
	if err == nil {
		db.codeSizeCache.Add(codeHash, len(code))
	}
	return code, err
}

// ContractCodeSize retrieves a particular contracts code's size.
func (db *cachingDB) ContractCodeSize(addrHash, codeHash common.Hash) (int, error) {
	if cached, ok := db.codeSizeCache.Get(codeHash); ok {
		return cached.(int), nil
	}
	code, err := db.ContractCode(addrHash, codeHash)
	return len(code), err
}

func (db *cachingDB) TrieDB() *Database {
	return db.db
}

func (m cachedTrie) Commit(onLeaf LeafCallback) (common.Hash, error) {
	root, err := m.SecureTrie.Commit(onLeaf)
	if err == nil {
		m.db.pushTrie(m.SecureTrie)
	}
	return root, err
}

func (m cachedTrie) Prove(key []byte, fromLevel uint, proofDb store.Putter) error {
	return m.SecureTrie.Prove(key, fromLevel, proofDb)
}
