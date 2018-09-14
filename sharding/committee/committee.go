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
	ns          *cell.Cell
	fsm         *sc.Fsm
	actorMsgc   chan interface{}
	packetRecvc <-chan netmsg.EcoBallNetMsg
	stateTimer  *time.Timer

	cs *consensus.Consensus
}

func MakeCommittee(ns *cell.Cell) sc.NodeInstance {
	cm := &committee{
		ns:        ns,
		actorMsgc: make(chan interface{}),
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
			{productCommitteBlock, ActRecvConsensusPacket, cm.processCmConsensusPacket, sc.StateNil},
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
			{productFinalBlock, ActRecvConsensusPacket, cm.processFinalConsensusPacket, sc.StateNil},
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
	c.actorMsgc <- msg
}

func (c *committee) Start() {
	recvc, err := simulate.Subscribe(c.ns.Self.Port)
	if err != nil {
		log.Panic("simulate error ", err)
		return
	}

	c.packetRecvc = recvc
	go c.cmRoutine()

	c.stateTimer = time.NewTimer(sc.DefaultSyncBlockTimer * time.Second)
}

func (c *committee) cmRoutine() {
	log.Debug("start committee routine")

	for {
		select {
		case msg := <-c.actorMsgc:
			c.processActorMsg(msg)
		case packet := <-c.packetRecvc:
			c.processPacket(packet)
		case <-c.stateTimer.C:
			c.processStateTimeout()
		}
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

func (c *committee) processPacket(packet netmsg.EcoBallNetMsg) {
	switch packet.Type() {
	case netmsg.APP_MSG_CONSENSUS_PACKET:
		c.processConsensusPacket(packet)
	case netmsg.APP_MSG_SHARDING_PACKET:
		c.processShardingPacket(packet)
	default:
		log.Error("wrong packet")
	}
}
