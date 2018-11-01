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

// Implement the gossip pull algo

package gossip

import (
	"time"
	"sync"
	"context"
	"sync/atomic"
	"github.com/ecoball/go-ecoball/common/elog"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"	
)

const (
	GspPullWaitforDigestTime  = time.Duration(1) * time.Second
	GspPullWaitforReqackTime  = time.Duration(2) * time.Second
)

const (
	PullEngine_State_Init           = 0
	PullEngine_State_AcceptDigest   = 1
	PullEngine_State_AcceptResponse = 2
)

var (
	log = elog.NewLogger("gossip", elog.DebugLog)
)

type GspPullAdapter interface {
	Receiver //who will own the gossiper

	SelectRemotePeers() []peer.ID

	Hello(id peer.ID) error
	SendDigest(id peer.ID, digest interface{}) error
	SendRequest(id peer.ID, request interface{}) error
	SendResponse(id peer.ID, response interface{}) error
}

type GspPullEngine struct {
	GspPullAdapter

	sentHello       map[string] bool
	sentDigest      map[string][]string
	sentRequest     map[string][]string
	receivedDigests	map[string][]string
	lock            sync.Mutex

	digest          []string
	runningState    int32

	stopCh          chan struct{}
	stopOnce        sync.Once
}

func newGspPullEngine(adpt GspPullAdapter) *GspPullEngine {
	pe := &GspPullEngine{
		GspPullAdapter:  adpt,
		sentHello:       make(map[string] bool),
		sentDigest:      make(map[string][]string),
		sentRequest:     make(map[string][]string),
		receivedDigests: make(map[string][]string),
		digest:          make([]string, 0),
		runningState:    PullEngine_State_Init,
		stopCh:          make(chan struct{}),
	}

	return pe
}

func (pe *GspPullEngine) start(ctx context.Context, interval time.Duration) {
	pe.startPull()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-pe.stopCh:
				return
			case <-time.After(interval):
				pe.startPull()
			}
		}
	}()
}

func (pe *GspPullEngine) Stop() {
	stopFunc := func() {
		close(pe.stopCh)
	}
	pe.stopOnce.Do(stopFunc)
}

func (pe *GspPullEngine) startPull() {
	log.Debug("gossip engine start")
	pe.lock.Lock()
	defer pe.lock.Unlock()

	pe.digest =  pe.GetDigests()
	peers := pe.SelectRemotePeers()
	pe.acceptDigest()
	for _, p := range peers {
		pe.Hello(p)
		pe.sentHello[string(p)] = true
	}

	time.AfterFunc(GspPullWaitforDigestTime, pe.processReceivedDigests)
}

func (pe *GspPullEngine) OnDigest(m *GspPullDigest) {
	log.Debug("gossip engine receive digest", m.Digests)
	if !pe.isInAcceptDigestSate() || !pe.sentHello[string(m.SenderId)] {
		return
	}
	pe.lock.Lock()
	defer pe.lock.Unlock()
	pe.sentHello[string(m.SenderId)] = false
	pe.receivedDigests[string(m.SenderId)] = m.Digests
}

func (pe *GspPullEngine) processReceivedDigests() {
	log.Debug("gossip engine process input digest")
	pe.acceptResponse()

	pe.lock.Lock()
	defer pe.lock.Unlock()

	shufferedDigests := pe.ShuffelDigests(pe.receivedDigests, pe.digest)

	for p, digest := range shufferedDigests {
		pe.SendRequest(peer.ID(p), digest)
		pe.sentRequest[p] = digest
	}

	time.AfterFunc(GspPullWaitforReqackTime, pe.endPull)}

func (pe *GspPullEngine) OnResponse(m *GspPullReqAck) {
	log.Debug("gossip engine receive response")
	if !pe.isInAcceptResponseSate() || pe.sentRequest[string(m.Responser)] == nil {
		log.Error("receive a response before sending request message")
		return
	}

	var dataArray [][]byte
	for _, env := range m.Payload {
		dataArray = append(dataArray, env.Data)
	}

	pe.UpdateItemData(dataArray)
}

func (pe *GspPullEngine) endPull() {
	pe.lock.Lock()
	defer pe.lock.Unlock()

	pe.resetState()
	pe.sentHello = make(map[string]bool)
	pe.sentRequest = make(map[string][]string)
	log.Debug("gossip engine exit")
}

func (pe *GspPullEngine) OnHello(m *GspPullHello) {
	log.Debug("gossip engine receive a hello message")

	pe.SendDigest(m.SenderId, pe.digest)
	pe.sentDigest[string(m.SenderId)] = pe.digest
}

func (pe *GspPullEngine) OnRequest(m *GspPullRequest) {
	log.Debug("gossip engine receive a request message", m.ReqItems)

	if pe.sentDigest[string(m.Asker)] == nil {
		log.Error("receive a request before sending digest message", string(m.Asker))
		return
	}

	pe.lock.Lock()
	defer pe.lock.Unlock()

	var sendItems []string
	for _, item := range m.ReqItems {
		if !pe.requestExistInMyDigests(item) {
			continue
		}
		sendItems = append(sendItems, item)
	}

	pe.SendResponse(m.Asker, sendItems)
}

func (pe *GspPullEngine) requestExistInMyDigests(item string) bool {
	if pe.ContainItemInDigests(pe.digest, item) {
		return true
	}
	return false
}

func (pe *GspPullEngine) resetState() {
	atomic.StoreInt32(&(pe.runningState), PullEngine_State_Init)
}

func (pe *GspPullEngine) acceptDigest() {
	atomic.StoreInt32(&(pe.runningState), PullEngine_State_AcceptDigest)
}

func (pe *GspPullEngine) isInAcceptDigestSate() bool {
	return atomic.LoadInt32(&(pe.runningState)) == PullEngine_State_AcceptDigest
}

func (pe *GspPullEngine) acceptResponse() {
	atomic.StoreInt32(&(pe.runningState), PullEngine_State_AcceptResponse)
}

func (pe *GspPullEngine) isInAcceptResponseSate() bool {
	return atomic.LoadInt32(&(pe.runningState)) == PullEngine_State_AcceptResponse
}
