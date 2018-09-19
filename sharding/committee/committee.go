package committee

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/message"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"time"
)

var (
	log = elog.NewLogger("sdcommittee", elog.DebugLog)
)

const (
	blockSync = iota + 1
	productCommitteBlock
	waitMinorBlock
	productFinalBlock
	productViewChangeBlock
	stateEnd
)

const (
	ActProductCommitteeBlock = iota + 1
	ActWaitMinorBlock
	ActProductFinalBlock
	ActChainNotSync
	ActRecvConsensusPacket
	ActStateTimeout
)

type committee struct {
	ns           *cell.Cell
	fsm          *sc.Fsm
	actorc       chan interface{}
	ppc          chan *sc.CsPacket
	pvc          <-chan *sc.NetPacket
	stateTimer   *time.Timer
	retransTimer *time.Timer
	vccount      uint16

	cs *consensus.Consensus
}

func MakeCommittee(ns *cell.Cell) sc.NodeInstance {
	cm := &committee{
		ns:      ns,
		actorc:  make(chan interface{}),
		ppc:     make(chan *sc.CsPacket, sc.DefaultCommitteMaxMember),
		vccount: 0,
	}

	cm.cs = consensus.MakeConsensus(cm.ns, cm.setRetransTimer, cm.consensusCb)

	cm.fsm = sc.NewFsm(blockSync,
		[]sc.FsmElem{
			{blockSync, ActProductCommitteeBlock, nil, cm.productCommitteeBlock, nil, productCommitteBlock},
			{blockSync, ActWaitMinorBlock, nil, cm.waitMinorBlock, nil, waitMinorBlock},
			{blockSync, ActProductFinalBlock, nil, cm.productFinalBlock, nil, productFinalBlock},
			{blockSync, ActStateTimeout, nil, cm.processBlockSyncTimeout, nil, sc.StateNil},
			{blockSync, ActRecvConsensusPacket, nil, cm.dropPacket, nil, sc.StateNil},

			{productCommitteBlock, ActChainNotSync, nil, cm.doBlockSync, nil, blockSync},
			{productCommitteBlock, ActRecvConsensusPacket, nil, cm.processConsensusCmPacket, nil, sc.StateNil},
			{productCommitteBlock, ActWaitMinorBlock, nil, cm.waitMinorBlock, nil, waitMinorBlock},
			{productCommitteBlock, ActStateTimeout, cm.resetVcCounter, cm.productViewChangeBlock, nil, productViewChangeBlock},

			/*missing_func consensus fail or timeout*/
			/*{ProductCommitteBlock, ActConsensusFail, , },
			{ProductCommitteBlock, ActProductTimeout, , },*/

			{waitMinorBlock, ActChainNotSync, nil, cm.doBlockSync, nil, blockSync},
			{waitMinorBlock, ActProductFinalBlock, nil, cm.productFinalBlock, nil, productFinalBlock},
			{waitMinorBlock, ActStateTimeout, nil, cm.productFinalBlock, nil, productFinalBlock},
			{waitMinorBlock, ActRecvConsensusPacket, cm.processConsensBlockOnWaitStatus, nil, nil, productFinalBlock},

			{productFinalBlock, ActChainNotSync, nil, cm.doBlockSync, nil, blockSync},
			{productFinalBlock, ActWaitMinorBlock, nil, cm.waitMinorBlock, nil, waitMinorBlock},
			{productFinalBlock, ActProductCommitteeBlock, nil, cm.productCommitteeBlock, nil, productCommitteBlock},
			{productFinalBlock, ActRecvConsensusPacket, nil, cm.processConsensusFinalPacket, nil, sc.StateNil},
			{productFinalBlock, ActStateTimeout, cm.resetVcCounter, cm.productViewChangeBlock, nil, productViewChangeBlock},

			/*missing_func consensus fail or timeout*/
			/*{ProductCommitteBlock, ActConsensusFail, , },
			{ProductCommitteBlock, ActProductTimeout, , },*/

			{productViewChangeBlock, ActProductCommitteeBlock, nil, cm.productCommitteeBlock, nil, productCommitteBlock},
			{productViewChangeBlock, ActProductFinalBlock, nil, cm.productFinalBlock, nil, productFinalBlock},
			{productViewChangeBlock, ActStateTimeout, cm.increaseCounter, cm.productViewChangeBlock, nil, productViewChangeBlock},
			{productViewChangeBlock, ActRecvConsensusPacket, nil, cm.processViewchangeConsensusPacket, nil, sc.StateNil},
		})

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
}

func (c *committee) cmRoutine() {
	log.Debug("start committee routine")
	c.stateTimer = time.NewTimer(sc.DefaultSyncBlockTimer * time.Second)
	c.retransTimer = time.NewTimer(sc.DefaultRetransTimer * time.Millisecond)

	for {
		select {
		case msg := <-c.actorc:
			c.processActorMsg(msg)
		case packet := <-c.ppc:
			c.processPacket(packet)
		case <-c.stateTimer.C:
			c.processStateTimeout()
		case <-c.retransTimer.C:
			c.processRetransTimeout()
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
	case message.SyncComplete:
		c.processSyncComplete(msg)
	default:
		log.Error("wrong actor message")
	}
}

func (c *committee) processPacket(packet *sc.CsPacket) {
	switch packet.PacketType {
	case netmsg.APP_MSG_CONSENSUS_PACKET:
		c.processConsensusPacket(packet)
	default:
		log.Error("wrong packet")
	}
}
