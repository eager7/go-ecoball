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
	StateNil = iota + 1
	BlockSync
	Consensus
	ProductCommitteBlock
	WaitMinorBlock
	ProductFinalBlock
	ProductViewChangeBlock
	StateEnd
)

const (
	ActProductCommitteeBlock = iota + 1
	ActWaitMinorBlock
	ActProductFinalBlock

	ActChainNotSync
	ActRecvConsensusPacket
	ActConsensusSuccess
	ActConsensusFail

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

	cm.fsm = sc.NewFsm(BlockSync,
		[]sc.FsmElem{
			{BlockSync, ActProductCommitteeBlock, cm.productCommitteeBlock, ProductCommitteBlock},
			{BlockSync, ActWaitMinorBlock, cm.waitMinorBlock, WaitMinorBlock},
			{BlockSync, ActProductFinalBlock, cm.productFinalBlock, ProductFinalBlock},
			{BlockSync, ActStateTimeout, cm.processBlockSyncTimeout, StateNil},
			{BlockSync, ActRecvConsensusPacket, cm.dropPacket, BlockSync},

			{ProductCommitteBlock, ActChainNotSync, cm.doBlockSync, BlockSync},
			{ProductCommitteBlock, ActConsensusSuccess, cm.waitMinorBlock, WaitMinorBlock},
			{ProductCommitteBlock, ActStateTimeout, cm.productViewChangeBlock, ProductViewChangeBlock},
			{ProductCommitteBlock, ActRecvConsensusPacket, cm.processCmConsensusPacket, ProductCommitteBlock},
			/*missing_func consensus fail or timeout*/
			/*{ProductCommitteBlock, ActConsensusFail, , },
			{ProductCommitteBlock, ActProductTimeout, , },*/

			{WaitMinorBlock, ActChainNotSync, cm.doBlockSync, BlockSync},
			{WaitMinorBlock, ActProductFinalBlock, cm.productFinalBlock, ProductFinalBlock},
			{WaitMinorBlock, ActStateTimeout, cm.productFinalBlock, ProductFinalBlock},
			{WaitMinorBlock, ActRecvConsensusPacket, cm.processWMBStateChange, ProductFinalBlock},

			{ProductFinalBlock, ActChainNotSync, cm.doBlockSync, BlockSync},
			{ProductFinalBlock, ActWaitMinorBlock, cm.waitMinorBlock, WaitMinorBlock},
			{ProductFinalBlock, ActProductCommitteeBlock, cm.productCommitteeBlock, ProductCommitteBlock},
			{ProductFinalBlock, ActStateTimeout, cm.productViewChangeBlock, ProductViewChangeBlock},
			{ProductFinalBlock, ActRecvConsensusPacket, cm.processFinalConsensusPacket, ProductFinalBlock},

			/*missing_func consensus fail or timeout*/
			/*{ProductCommitteBlock, ActConsensusFail, , },
			{ProductCommitteBlock, ActProductTimeout, , },*/

			{ProductViewChangeBlock, ActProductCommitteeBlock, cm.productCommitteeBlock, ProductCommitteBlock},
			{ProductViewChangeBlock, ActProductFinalBlock, cm.productFinalBlock, ProductFinalBlock},
			{ProductViewChangeBlock, ActStateTimeout, cm.productViewChangeBlock, ProductViewChangeBlock},
			{ProductViewChangeBlock, ActRecvConsensusPacket, cm.processViewchangeConsensusPacket, ProductViewChangeBlock},
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

func (c *committee) processStateTimeout() {
	c.fsm.Execute(ActStateTimeout, nil)
}
