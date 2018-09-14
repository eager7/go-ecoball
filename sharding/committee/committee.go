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
	ns         *cell.Cell
	fsm        *sc.Fsm
	actorc     chan interface{}
	ppc        chan *sc.CsPacket
	pvc        <-chan *sc.NetPacket
	stateTimer *time.Timer

	cs *consensus.Consensus
}

func MakeCommittee(ns *cell.Cell) sc.NodeInstance {
	cm := &committee{
		ns:     ns,
		actorc: make(chan interface{}),
		ppc:    make(chan *sc.CsPacket, sc.DefaultCommitteMaxMember),
	}

	cm.cs = consensus.MakeConsensus(cm.ns, cm.consensusCb)

	cm.fsm = sc.NewFsm(blockSync,
		[]sc.FsmElem{
			{blockSync, ActProductCommitteeBlock, cm.productCommitteeBlock, productCommitteBlock},
			{blockSync, ActWaitMinorBlock, cm.waitMinorBlock, waitMinorBlock},
			{blockSync, ActProductFinalBlock, cm.productFinalBlock, productFinalBlock},
			{blockSync, ActStateTimeout, cm.processBlockSyncTimeout, sc.StateNil},
			{blockSync, ActRecvConsensusPacket, cm.dropPacket, sc.StateNil},

			{productCommitteBlock, ActChainNotSync, cm.doBlockSync, blockSync},
			{productCommitteBlock, ActRecvConsensusPacket, cm.processConsensusCmPacket, sc.StateNil},
			{productCommitteBlock, ActWaitMinorBlock, cm.waitMinorBlock, waitMinorBlock},
			{productCommitteBlock, ActStateTimeout, cm.productViewChangeBlock, productViewChangeBlock},

			/*missing_func consensus fail or timeout*/
			/*{ProductCommitteBlock, ActConsensusFail, , },
			{ProductCommitteBlock, ActProductTimeout, , },*/

			{waitMinorBlock, ActChainNotSync, cm.doBlockSync, blockSync},
			{waitMinorBlock, ActProductFinalBlock, cm.productFinalBlock, productFinalBlock},
			{waitMinorBlock, ActStateTimeout, cm.productFinalBlock, productFinalBlock},
			{waitMinorBlock, ActRecvConsensusPacket, cm.processWMBStateChange, productFinalBlock},

			{productFinalBlock, ActChainNotSync, cm.doBlockSync, blockSync},
			{productFinalBlock, ActWaitMinorBlock, cm.waitMinorBlock, waitMinorBlock},
			{productFinalBlock, ActProductCommitteeBlock, cm.productCommitteeBlock, productCommitteBlock},
			{productFinalBlock, ActRecvConsensusPacket, cm.processConsensusFinalPacket, sc.StateNil},
			{productFinalBlock, ActStateTimeout, cm.productViewChangeBlock, productViewChangeBlock},

			/*missing_func consensus fail or timeout*/
			/*{ProductCommitteBlock, ActConsensusFail, , },
			{ProductCommitteBlock, ActProductTimeout, , },*/

			{productViewChangeBlock, ActProductCommitteeBlock, cm.productCommitteeBlock, productCommitteBlock},
			{productViewChangeBlock, ActProductFinalBlock, cm.productFinalBlock, productFinalBlock},
			{productViewChangeBlock, ActStateTimeout, cm.productViewChangeBlock, productViewChangeBlock},
			{productViewChangeBlock, ActRecvConsensusPacket, cm.processViewchangeConsensusPacket, sc.StateNil},
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

	for {
		select {
		case msg := <-c.actorc:
			c.processActorMsg(msg)
		case packet := <-c.ppc:
			c.processPacket(packet)
		case <-c.stateTimer.C:
			c.processStateTimeout()
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
