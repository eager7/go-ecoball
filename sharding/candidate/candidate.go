package candidate

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/common/message/mpb"

	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"

	"github.com/ecoball/go-ecoball/sharding/net"
	"time"
)

var (
	log = elog.NewLogger("sharding", elog.DebugLog)
)

const (
	blockSync = iota + 1
	waitBlock
	stateEnd
)

const (
	ActWaitBlock = iota + 1
	ActChainNotSync
	ActRecvShardingPacket
	ActStateTimeout
)

type shard struct {
	ns     *cell.Cell
	fsm    *sc.Fsm
	actorc chan interface{}
	ppc    chan *sc.CsPacket
	pvc    <-chan interface{}

	stateTimer    *sc.Stimer

}

func MakeCandidateShardTest(ns *cell.Cell) *shard {
	instance := MakeCandidateShard(ns)
	return instance.(*shard)
}

func (s *shard) SetNet() {
	net.MakeNet(s.ns)
	s.pvc, _ = net.Np.Subscribe(s.ns.Self.Port, sc.DefaultShardMaxMember)
	s.pvcRoutine()

	go s.sRoutine()
}

func MakeCandidateShard(ns *cell.Cell) sc.NodeInstance {
	s := &shard{ns: ns,
		actorc:        make(chan interface{}),
		ppc:           make(chan *sc.CsPacket, sc.DefaultShardMaxMember),
		stateTimer:    sc.NewStimer(0, false),
	}

	s.fsm = sc.NewFsm(blockSync,
		[]sc.FsmElem{
			{blockSync, ActWaitBlock, nil, nil, nil, waitBlock},
			{blockSync, ActStateTimeout, nil, s.processBlockSyncTimeout, nil, sc.StateNil},
			{waitBlock, ActChainNotSync, nil, s.doBlockSync, nil, blockSync},
			{waitBlock, ActRecvShardingPacket, nil, s.processShardingPacket, nil, sc.StateNil},
		})

	return s
}

func (s *shard) MsgDispatch(msg interface{}) {
	s.actorc <- msg
}

func (s *shard) Start() {
	return
}


func (s *shard) sRoutine() {
	log.Debug("start shard routine")
	s.ns.LoadLastBlock()
	//s.sync.Start()
	go s.setSyncRequest()

	s.stateTimer.Reset(sc.DefaultSyncBlockTimer * time.Second)

	for {
		select {
		case msg := <-s.actorc:
			s.processActorMsg(msg)
		case packet := <-s.ppc:
			s.processPacket(packet)
		case <-s.stateTimer.T.C:
			if s.stateTimer.GetStatus() {
				s.stateTimer.SetStop()
				s.processStateTimeout()
			}
		}
	}
}

func (s *shard) pvcRoutine() {
	for i := 0; i < sc.DefaultShardMaxMember; i++ {
		go func() {
			for {
				msg := <-s.pvc
				packet, err := net.Np.RecvNetMsg(msg)
				if err != nil {
					log.Error("recv net msg error ", err)
				} else {
					s.verifyPacket(packet)
				}
			}
		}()
	}
}

func (s *shard) processActorMsg(msg interface{}) {
	switch msg.(type) {
	case *message.SyncComplete:
		s.processSyncComplete()
	default:
		log.Error("wrong actor message")
	}
}

func (s *shard) processPacket(packet *sc.CsPacket) {
	switch packet.PacketType {
	case mpb.Identify_APP_MSG_SHARDING_PACKET:
		s.recvShardingPacket(packet)
	/*case mpb.Identify_MsgType_APP_MSG_SYNC_REQUEST:
		csp, worker := s.sync.RecvSyncRequestPacket(packet)
		net.Np.SendSyncResponse(csp, worker)
	case mpb.Identify_MsgType_APP_MSG_SYNC_RESPONSE:
		s.sync.RecvSyncResponsePacket(packet)
    */
	default:
		log.Error("wrong packet")
	}
}

func (s *shard) setSyncRequest() {
	log.Debug("miss some blocks, set sync request ")
	//s.sync.SendSyncRequest()
}
