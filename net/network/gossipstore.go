// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball.
//
// The go-ecoball is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.

// Implement a gossip message store

package network

import (
	"context"
	"github.com/ecoball/go-ecoball/net/message"
	"sync"
	"time"
)

func NewMsgStore(ctx context.Context, ttl time.Duration) MsgStore {
	store := &MsgStoreImpl{
		messages: make(map[uint64]*gmsg, 0),
		ttl:      ttl,
		stopCh:   make(chan struct{}),
	}

	go store.expireRoting(ctx)
	return store
}

type gmsg struct {
	data    interface{}
	created time.Time
	expired bool
}

type MsgStore interface {
	Add(msg interface{}) bool
	CheckValid(msg interface{}) bool
	Stop()
}

type MsgStoreImpl struct {
	messages map[uint64]*gmsg
	ttl      time.Duration
	lock     sync.RWMutex

	stopCh   chan struct{}
	stopOnce sync.Once
}

func (msi *MsgStoreImpl) Add(msg interface{}) bool {
	m, ok := msg.(message.EcoBallNetMsg)
	if !ok || !msi.CheckValid(msg) {
		return false
	}

	msi.lock.Lock()
	defer msi.lock.Unlock()

	msi.messages[m.Nonce()] = &gmsg{data: msg, created: time.Now()}

	return true
}

func (msi *MsgStoreImpl) CheckValid(msg interface{}) bool {
	msi.lock.Lock()
	defer msi.lock.Unlock()

	m, ok := msg.(message.EcoBallNetMsg)
	if !ok {
		return false
	}

	if msi.messages[m.Nonce()] != nil {
		return false
	}

	return true
}

func (msi *MsgStoreImpl) Stop() {
	stopFunc := func() {
		close(msi.stopCh)
	}
	msi.stopOnce.Do(stopFunc)
}

func (msi *MsgStoreImpl) expireRoting(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-msi.stopCh:
			return
		case <-time.After(msi.expirePollInterval()):
			if msi.hasExpiredMsg() {
				msi.expireMsg()
			}
		}
	}
}

func (msi *MsgStoreImpl) hasExpiredMsg() bool {
	expired := func(m *gmsg) bool {
		if !m.expired && time.Since(m.created) > msi.ttl {
			return true
		} else if time.Since(m.created) > (msi.ttl * 2) {
			return true
		}
		return false
	}
	msi.lock.Lock()
	defer msi.lock.Unlock()
	for _, v := range msi.messages {
		if expired(v) {
			return true
		}
	}

	return false
}

func (msi *MsgStoreImpl) expireMsg() bool {
	msi.lock.Lock()
	defer msi.lock.Unlock()

	for k, v := range msi.messages {
		if !v.expired && time.Since(v.created) > msi.ttl {
			v.expired = true
		} else if time.Since(v.created) > (2 * msi.ttl) {
			delete(msi.messages, k)
		}
	}

	return false
}

func (msi *MsgStoreImpl) expirePollInterval() time.Duration {
	inv := msi.ttl / 100 * 5

	if inv == 0 {
		inv = 5
	}

	return inv
}
