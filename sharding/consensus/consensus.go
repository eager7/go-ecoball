package consensus

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

var (
	log = elog.NewLogger("sdconsensus", elog.DebugLog)
)

const (
	StepPrePare = iota + 1
	StepPreCommit
	StepCommit
	StepNIL
)

type Consensus struct {
	ns   *cell.Cell
	step uint16
	view *sc.CsView

	instance sc.ConsensusInstance
	rcb      retransCb
	ccb      csCompleteCb
}

type csCompleteCb func(bl interface{})
type retransCb func(bStart bool)

func MakeConsensus(ns *cell.Cell, rcb retransCb, ccb csCompleteCb) *Consensus {
	return &Consensus{
		step: StepNIL,
		ns:   ns,
		rcb:  rcb,
		ccb:  ccb,
	}
}

func (c *Consensus) StartConsensus(instance sc.ConsensusInstance) {
	if c.ns.IsLeader() {
		c.startBlockConsensusLeader(instance)
	} else {
		c.startBlockConsensusVoter(instance)
	}
}

func (c *Consensus) StartVcConsensus(instance sc.ConsensusInstance, bCandi bool) {
	if bCandi {
		c.startBlockConsensusLeader(instance)
	} else {
		c.startBlockConsensusVoter(instance)
	}
}

func (c *Consensus) ProcessPacket(csp *sc.CsPacket) bool {
	candidate := c.instance.GetCandidate()
	if candidate != nil {
		if c.ns.Self.EqualNode(candidate) {
			if !c.instance.CheckBlock(csp.Packet, true) {
				log.Error("check packet error")
				return false
			}
			c.processPacketByLeader(csp)
		} else {
			if !c.instance.CheckBlock(csp.Packet, false) {
				log.Error("check packet error")
				return false
			}

			c.processPacketByVoter(csp)
		}
	} else {
		if c.ns.IsLeader() {
			if !c.instance.CheckBlock(csp.Packet, true) {
				log.Error("check packet error")
				return false
			}
			c.processPacketByLeader(csp)
		} else {
			if !c.instance.CheckBlock(csp.Packet, false) {
				log.Error("check packet error")
				return false
			}

			c.processPacketByVoter(csp)
		}
	}
	return true
}

func (c *Consensus) ProcessRetransPacket() {
	if c.instance == nil {
		return
	}

	if c.ns.GetWorksCounter() == 1 {
		c.sendCommit()
	} else {
		log.Debug("resend packet  step ", c.step)
		packet := c.instance.MakeNetPacket(c.step)
		c.sendCsPacket(packet)
	}
}
