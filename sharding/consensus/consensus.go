package consensus

import (
	"encoding/json"
	"github.com/ecoball/go-ecoball/common/elog"
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/node"
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
	ns    *node.Node
	round uint16
	view  *sc.CsView

	instance sc.ConsensusInstance

	completeCb csCompleteCb
}

type csCompleteCb func(bl interface{})

func MakeConsensus(ns *node.Node, cb csCompleteCb) *Consensus {
	return &Consensus{
		round:      RoundNIL,
		ns:         ns,
		completeCb: cb,
	}
}

func (c *Consensus) ProcessPacket(packet netmsg.EcoBallNetMsg) {
	var csp sc.CsPacket
	err := json.Unmarshal(packet.Data(), &csp)
	if err != nil {
		log.Error("net packet unmarshal error:%s", err)
		return
	}

	view := c.instance.CacheBlock(&csp)
	if !c.view.Equal(view) {
		log.Error("view error current:%d %d %d %, recv: %d %d %d", c.view.EpochNo, c.view.FinalHeight, c.view.MinorHeight,
			view.EpochNo, view.FinalHeight, view.MinorHeight)
		return
	}

}
