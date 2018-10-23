package shard

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/etime"
	"github.com/ecoball/go-ecoball/common/message"
	cs "github.com/ecoball/go-ecoball/core/shard"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/net"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

var (
	log = elog.NewLogger("sdshard", elog.DebugLog)
)

const (
	blockSync = iota + 1
	waitBlock
	productMinoBlock
	stateEnd
)

const (
	ActProductMinorBlock = iota + 1
	ActWaitBlock
	ActRecvConsensusPacket
	ActChainNotSync
	ActRecvCommitteePacket
	ActLedgerBlockMsg
	ActStateTimeout
)

type shard struct {
	ns     *cell.Cell
	fsm    *sc.Fsm
	actorc chan interface{}
	ppc    chan *sc.CsPacket
	pvc    <-chan *sc.NetPacket

	stateTimer   *time.Timer
	retransTimer *time.Timer
	cs           *consensus.Consensus
}

func MakeShard(ns *cell.Cell) sc.NodeInstance {
	s := &shard{ns: ns,
		actorc: make(chan interface{}),
		ppc:    make(chan *sc.CsPacket, sc.DefaultShardMaxMember),
	}

	s.cs = consensus.MakeConsensus(s.ns, s.setRetransTimer, s.consensusCb)

	s.fsm = sc.NewFsm(blockSync,
		[]sc.FsmElem{
			{blockSync, ActWaitBlock, nil, nil, nil, waitBlock},
			{blockSync, ActProductMinorBlock, nil, s.productMinorBlock, nil, productMinoBlock},
			{blockSync, ActStateTimeout, nil, s.processBlockSyncTimeout, nil, sc.StateNil},

			{waitBlock, ActProductMinorBlock, nil, s.productMinorBlock, nil, productMinoBlock},
			{waitBlock, ActChainNotSync, nil, nil, nil, blockSync},
			{waitBlock, ActRecvCommitteePacket, nil, s.processCommitteePacket, nil, sc.StateNil},

			{productMinoBlock, ActRecvConsensusPacket, nil, s.processConsensusMinorPacket, nil, sc.StateNil},
			{productMinoBlock, ActWaitBlock, nil, nil, nil, waitBlock},
			{productMinoBlock, ActProductMinorBlock, nil, s.reproductMinorBlock, nil, sc.StateNil},
			{productMinoBlock, ActRecvCommitteePacket, nil, s.processCommitteePacket, nil, sc.StateNil},
			{productMinoBlock, ActLedgerBlockMsg, nil, s.processLedgerMinorBlockMsg, nil, sc.StateNil},
		})

	net.MakeNet(ns)

	return s
}

func (s *shard) MsgDispatch(msg interface{}) {
	s.actorc <- msg
}

func (s *shard) Start() {
	recvc, err := simulate.Subscribe(s.ns.Self.Port, sc.DefaultShardMaxMember)
	if err != nil {
		log.Panic("simulate error ", err)
		return
	}

	s.pvc = recvc
	go s.sRoutine()
	s.pvcRoutine()
}

func (s *shard) sRoutine() {
	log.Debug("start shard routine")
	s.stateTimer = time.NewTimer(sc.DefaultSyncBlockTimer * time.Second)
	s.retransTimer = time.NewTimer(sc.DefaultRetransTimer * time.Millisecond)

	for {
		select {
		case msg := <-s.actorc:
			s.processActorMsg(msg)
		case packet := <-s.ppc:
			s.processPacket(packet)
		case <-s.stateTimer.C:
			s.processStateTimeout()
		case <-s.retransTimer.C:
			s.processRetransTimeout()
		}
	}
}

func (s *shard) pvcRoutine() {
	for i := 0; i < sc.DefaultShardMaxMember; i++ {
		go func() {
			for {
				packet := <-s.pvc
				s.verifyPacket(packet)
			}
		}()
	}
}

func (s *shard) processActorMsg(msg interface{}) {
	switch msg.(type) {
	case *message.SyncComplete:
		s.processSyncComplete()
	case *cs.MinorBlock:
		s.processMinorBlockMsg(msg.(*cs.MinorBlock))
	default:
		log.Error("wrong actor message")
	}
}

func (s *shard) setRetransTimer(bStart bool) {
	etime.StopTime(s.retransTimer)

	if bStart {
		s.retransTimer.Reset(sc.DefaultRetransTimer * time.Second)
	}
}

func (s *shard) processPacket(packet *sc.CsPacket) {
	switch packet.PacketType {
	case netmsg.APP_MSG_CONSENSUS_PACKET:
		s.recvConsensusPacket(packet)
	case netmsg.APP_MSG_SHARDING_PACKET:
		s.recvCommitteePacket(packet)
	default:
		log.Error("wrong packet")
	}
}
