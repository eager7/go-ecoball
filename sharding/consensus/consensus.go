package consensus

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/elog"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

var (
	log = elog.NewLogger("sdconsensus", elog.DebugLog)
)

const (
	RoundPrePare = iota + 1
	RoundPreCommit
	RoundCommit
	RoundNIL
)

type Consensus struct {
	ns   *cell.Cell
	step uint16
	view *sc.CsView

	instance sc.ConsensusInstance

	completeCb csCompleteCb
}

type csCompleteCb func(bl interface{})

func MakeConsensus(ns *cell.Cell, cb csCompleteCb) *Consensus {
	return &Consensus{
		step:       RoundNIL,
		ns:         ns,
		completeCb: cb,
	}
}

func (c *Consensus) StartConsensus(instance sc.ConsensusInstance) {
	if c.ns.IsCmLeader() {
		c.startBlockConsensusLeader(instance)
	} else {
		c.startBlockConsensusVoter(instance)
	}
}

func (c *Consensus) ProcessPacket(packet netmsg.EcoBallNetMsg) {
	var csp sc.CsPacket
	err := json.Unmarshal(packet.Data(), &csp)
	if err != nil {
		log.Error("net packet unmarshal error ", err)
		return
	}

	view := c.instance.CacheBlock(&csp)
	if view == nil {
		log.Error("cache packet error")
		return
	}

	if !c.view.Equal(view) {
		log.Error("view error current ", c.view.EpochNo, " ", c.view.FinalHeight, " ", c.view.MinorHeight, " packet view ",
			view.EpochNo, " ", view.FinalHeight, " ", view.MinorHeight)
		return
	}

	if c.ns.IsCmLeader() {
		c.processPacketByLeader(&csp)
	} else {
		c.processPacketByVoter(&csp)
	}
}

func (c *Consensus) IsCsRunning() bool {
	if c.instance == nil && c.view == nil && c.step == RoundNIL {
		return false
	} else if c.instance != nil && c.view != nil && c.step != RoundNIL {
		return true
	} else {
		panic("consensus wrong status")
		return false
	}
}
