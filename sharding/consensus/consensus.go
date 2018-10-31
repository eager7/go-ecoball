package consensus

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"time"
)

var (
	log = elog.NewLogger("sharding", elog.DebugLog)
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
	rcb      timerdCb
	fcb      timerCb
	ccb      csCompleteCb
}

type csCompleteCb func(bl interface{})
type timerdCb func(bStart bool, d time.Duration)
type timerCb func(bStart bool)

func MakeConsensus(ns *cell.Cell, rcb timerdCb, fcb timerCb, ccb csCompleteCb) *Consensus {
	return &Consensus{
		step: StepNIL,
		ns:   ns,
		rcb:  rcb,
		fcb:  fcb,
		ccb:  ccb,
	}
}

func (c *Consensus) StartConsensus(instance sc.ConsensusInstance, d time.Duration) {
	if c.ns.IsLeader() {
		c.startBlockConsensusLeader(instance, d)
	} else {
		c.startBlockConsensusVoter(instance)
	}
}

func (c *Consensus) StartVcConsensus(instance sc.ConsensusInstance, d time.Duration, bCandi bool) {
	if bCandi {
		c.startBlockConsensusLeader(instance, d)
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

	log.Debug("process restrans packet")

	if c.ns.GetWorksCounter() == 1 {
		c.sendCommit()
	} else {
		log.Debug("resend packet  step ", c.step)
		packet := c.instance.MakeNetPacket(c.step)
		c.sendCsPacket(packet)

		c.rcb(true, sc.DefaultRetransTimer*time.Second)
	}
}

func (c *Consensus) ProcessFullVoteTimeout() {
	if c.instance == nil {
		return
	}

	log.Debug("full vote timer out")

	cosig := c.instance.GetCosign()
	if c.step == StepPrePare {
		if c.ns.IsVoteEnough(cosig.Step1) {
			c.sendPreCommit()
		} else {
			panic("step prepare wrong vote counter")
		}
	} else if c.step == StepPreCommit {
		if c.ns.IsVoteEnough(cosig.Step2) {
			c.sendCommit()
		} else {
			panic("step precommit wrong vote counter")
		}
	}
}
