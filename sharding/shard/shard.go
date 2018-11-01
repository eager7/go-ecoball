package shard

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/message"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/net"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

var (
	log = elog.NewLogger("sharding", elog.DebugLog)
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
	ActRecvShardingPacket
	ActLedgerBlockMsg
	ActStateTimeout
)

type shard struct {
	ns     *cell.Cell
	fsm    *sc.Fsm
	actorc chan interface{}
	ppc    chan *sc.CsPacket
	pvc    <-chan *sc.NetPacket

	stateTimer    *sc.Stimer
	retransTimer  *sc.Stimer
	fullVoteTimer *sc.Stimer
	cs            *consensus.Consensus
}

func MakeShardTest(ns *cell.Cell) *shard {
	instance := MakeShard(ns)
	return instance.(*shard)
}

func MakeShard(ns *cell.Cell) sc.NodeInstance {
	s := &shard{ns: ns,
		actorc:        make(chan interface{}),
		ppc:           make(chan *sc.CsPacket, sc.DefaultShardMaxMember),
		stateTimer:    sc.NewStimer(0, false),
		retransTimer:  sc.NewStimer(0, false),
		fullVoteTimer: sc.NewStimer(0, false),
	}

	s.cs = consensus.MakeConsensus(s.ns, s.setRetransTimer, s.setFullVoeTimer, s.consensusCb)

	s.fsm = sc.NewFsm(blockSync,
		[]sc.FsmElem{
			{blockSync, ActWaitBlock, nil, nil, nil, waitBlock},
			{blockSync, ActProductMinorBlock, nil, s.productMinorBlock, nil, productMinoBlock},
			{blockSync, ActStateTimeout, nil, s.processBlockSyncTimeout, nil, sc.StateNil},

			{waitBlock, ActProductMinorBlock, nil, s.productMinorBlock, nil, productMinoBlock},
			{waitBlock, ActChainNotSync, nil, nil, nil, blockSync},
			{waitBlock, ActRecvShardingPacket, nil, s.processShardingPacket, nil, sc.StateNil},

			{productMinoBlock, ActRecvConsensusPacket, nil, s.processConsensusMinorPacket, nil, sc.StateNil},
			{productMinoBlock, ActWaitBlock, nil, nil, nil, waitBlock},
			{productMinoBlock, ActProductMinorBlock, nil, s.reproductMinorBlock, nil, sc.StateNil},
			{productMinoBlock, ActRecvShardingPacket, nil, s.processShardingPacket, nil, sc.StateNil},
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
	s.stateTimer.Reset(sc.DefaultSyncBlockTimer * time.Second)

	for {
		select {
		case msg := <-s.actorc:
			s.processActorMsg(msg)
		case packet := <-s.ppc:
			s.processPacket(packet)
		case <-s.stateTimer.T.C:
			if s.stateTimer.On {
				s.processStateTimeout()
			}
		case <-s.retransTimer.T.C:
			if s.retransTimer.On {
				s.processRetransTimeout()
			}
		case <-s.fullVoteTimer.T.C:
			if s.fullVoteTimer.On {
				s.processFullVoteTimeout()
			}
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

func (s *shard) setRetransTimer(bStart bool, d time.Duration) {
	log.Debug("set restrans timer ", bStart)

	if bStart {
		s.retransTimer.Reset(d)
	} else {
		s.retransTimer.Stop()
	}
}

func (s *shard) processPacket(packet *sc.CsPacket) {
	switch packet.PacketType {
	case pb.MsgType_APP_MSG_CONSENSUS_PACKET:
		s.recvConsensusPacket(packet)
	case pb.MsgType_APP_MSG_SHARDING_PACKET:
		s.recvShardingPacket(packet)
	case pb.MsgType_APP_MSG_SYNC_REQUEST:
		csp, worker := s.ns.RecvSyncRequestPacket(packet)
		net.Np.SendSyncResponse(csp, worker)
	case pb.MsgType_APP_MSG_SYNC_RESPONSE:
		s.ns.RecvSyncResponsePacket(packet)

	default:
		log.Error("wrong packet")
	}
}


func (s *shard) processFullVoteTimeout() {
	s.cs.ProcessFullVoteTimeout()
}

func (s *shard) setFullVoeTimer(bStart bool) {
	log.Debug("set full vote timer ", bStart)

	if bStart {
		s.fullVoteTimer.Reset(sc.DefaultFullVoteTimer * time.Second)
	} else {
		s.fullVoteTimer.Stop()
	}
}

