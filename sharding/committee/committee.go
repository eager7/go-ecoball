package committee

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/message"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/datasync"
	"github.com/ecoball/go-ecoball/sharding/net"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

var (
	log = elog.NewLogger("sharding", elog.DebugLog)
)

const (
	blockSync = iota + 1
	productCommitteBlock
	collectMinorBlock
	productFinalBlock
	productViewChangeBlock
	stateEnd
)

const (
	ActProductCommitteeBlock = iota + 1
	ActCollectMinorBlock
	ActProductFinalBlock
	ActChainNotSync
	ActRecvConsensusPacket
	ActRecvShardPacket
	ActLedgerBlockMsg
	ActStateTimeout
)

type committee struct {
	ns            *cell.Cell
	fsm           *sc.Fsm
	actorc        chan interface{}
	ppc           chan *sc.CsPacket
	pvc           <-chan *sc.NetPacket
	stateTimer    *sc.Stimer
	retransTimer  *sc.Stimer
	fullVoteTimer *sc.Stimer
	vccount       uint16
	cs            *consensus.Consensus

	sync *datasync.Sync
}

func MakeCommittee(ns *cell.Cell) sc.NodeInstance {
	cm := &committee{
		ns:            ns,
		actorc:        make(chan interface{}),
		ppc:           make(chan *sc.CsPacket, sc.DefaultCommitteMaxMember),
		vccount:       0,
		stateTimer:    sc.NewStimer(0, false),
		retransTimer:  sc.NewStimer(0, false),
		fullVoteTimer: sc.NewStimer(0, false),
		sync:          datasync.MakeSync(ns),
	}

	cm.cs = consensus.MakeConsensus(cm.ns, cm.setRetransTimer, cm.setFullVoeTimer, cm.consensusCb)

	cm.fsm = sc.NewFsm(blockSync,
		[]sc.FsmElem{
			{blockSync, ActProductCommitteeBlock, nil, cm.productCommitteeBlock, nil, productCommitteBlock},
			{blockSync, ActCollectMinorBlock, nil, cm.collectMinorBlock, nil, collectMinorBlock},
			{blockSync, ActProductFinalBlock, nil, cm.productFinalBlock, nil, productFinalBlock},
			{blockSync, ActStateTimeout, nil, cm.processBlockSyncTimeout, nil, sc.StateNil},

			{productCommitteBlock, ActChainNotSync, nil, cm.doBlockSync, nil, blockSync},
			{productCommitteBlock, ActRecvConsensusPacket, nil, cm.processConsensusCmPacket, nil, sc.StateNil},
			{productCommitteBlock, ActCollectMinorBlock, nil, cm.collectMinorBlock, nil, collectMinorBlock},
			{productCommitteBlock, ActStateTimeout, cm.resetVcCounter, cm.productViewChangeBlock, nil, productViewChangeBlock},

			{collectMinorBlock, ActChainNotSync, nil, cm.doBlockSync, nil, blockSync},
			{collectMinorBlock, ActProductFinalBlock, nil, cm.productFinalBlock, nil, productFinalBlock},
			{collectMinorBlock, ActStateTimeout, nil, cm.productFinalBlock, nil, productFinalBlock},
			{collectMinorBlock, ActRecvConsensusPacket, cm.processConsensBlockOnWaitStatus, nil, cm.afterProcessConsensBlockOnWaitStatus, productFinalBlock},
			{collectMinorBlock, ActRecvShardPacket, nil, cm.processShardBlockOnWaitStatus, nil, sc.StateNil},

			{productFinalBlock, ActChainNotSync, nil, cm.doBlockSync, nil, blockSync},
			{productFinalBlock, ActCollectMinorBlock, nil, cm.collectMinorBlock, nil, collectMinorBlock},
			{productFinalBlock, ActProductCommitteeBlock, nil, cm.productCommitteeBlock, nil, productCommitteBlock},
			{productFinalBlock, ActRecvConsensusPacket, nil, cm.processConsensusFinalPacket, nil, sc.StateNil},
			{productFinalBlock, ActStateTimeout, cm.resetVcCounter, cm.productViewChangeBlock, nil, productViewChangeBlock},
			{productFinalBlock, ActLedgerBlockMsg, nil, cm.processLedgerFinalBlockMsg, nil, sc.StateNil},

			{productViewChangeBlock, ActProductCommitteeBlock, nil, cm.productCommitteeBlock, nil, productCommitteBlock},
			{productViewChangeBlock, ActProductFinalBlock, nil, cm.productFinalBlock, nil, productFinalBlock},
			{productViewChangeBlock, ActStateTimeout, cm.increaseCounter, cm.productViewChangeBlock, nil, productViewChangeBlock},
			{productViewChangeBlock, ActRecvConsensusPacket, nil, cm.processViewchangeConsensusPacket, nil, sc.StateNil},
			{productViewChangeBlock, ActChainNotSync, nil, cm.doBlockSync, nil, blockSync},
		})

	net.MakeNet(ns)

	return cm
}

func (c *committee) MsgDispatch(msg interface{}) {
	c.actorc <- msg
}

func (c *committee) Start() {
	recvc, err := simulate.Subscribe(c.ns.Self.Port, sc.DefaultCommitteMaxMember)
	if err != nil {
		log.Panic("simulate error ", err)
		return
	}

	c.pvc = recvc
	go c.cmRoutine()
	c.pvcRoutine()

	c.setSyncRequest()
}

func (c *committee) cmRoutine() {
	log.Debug("start committee routine")
	c.stateTimer.Reset(sc.DefaultSyncBlockTimer * time.Second)

	for {
		select {
		case msg := <-c.actorc:
			c.processActorMsg(msg)
		case packet := <-c.ppc:
			c.processPacket(packet)
		case <-c.stateTimer.T.C:
			if c.stateTimer.On {
				c.processStateTimeout()
			}
		case <-c.retransTimer.T.C:
			if c.retransTimer.On {
				c.processRetransTimeout()
			}
		case <-c.fullVoteTimer.T.C:
			if c.fullVoteTimer.On {
				c.processFullVoteTimeout()
			}
		}
	}
}

func (c *committee) pvcRoutine() {
	for i := 0; i < sc.DefaultCommitteMaxMember; i++ {
		go func() {
			for {
				packet := <-c.pvc
				c.verifyPacket(packet)
			}
		}()
	}
}

func (c *committee) processActorMsg(msg interface{}) {
	switch msg.(type) {
	case *message.SyncComplete:
		c.processSyncComplete(msg)
	case *cs.FinalBlock:
		c.processFinalBlockMsg(msg.(*cs.FinalBlock))
	default:
		log.Error("wrong actor message")
	}
}

func (c *committee) processPacket(packet *sc.CsPacket) {
	switch packet.PacketType {
	case pb.MsgType_APP_MSG_CONSENSUS_PACKET:
		c.recvConsensusPacket(packet)
	case pb.MsgType_APP_MSG_SHARDING_PACKET:
		c.recvShardPacket(packet)
	default:
		log.Error("wrong packet")
	}
}

func (c *committee) processFullVoteTimeout() {
	c.cs.ProcessFullVoteTimeout()
}

func (c *committee) setFullVoeTimer(bStart bool) {
	log.Debug("set full vote timer ", bStart)

	if bStart {
		c.fullVoteTimer.Reset(sc.DefaultFullVoteTimer * time.Second)
	} else {
		c.fullVoteTimer.Stop()
	}
}

func (c *committee) setSyncRequest() {
	log.Debug("set sync request ")
}
