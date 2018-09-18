package shard

import (
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"log"
)

type shard struct {
	ns *cell.Cell

	msgc chan interface{}
	ppc  <-chan *sc.CsPacket
	pvc  <-chan *sc.NetPacket
}

func MakeShard(ns *cell.Cell) sc.NodeInstance {
	return &shard{ns: ns,
		msgc: make(chan interface{}),
		ppc:  make(chan *sc.CsPacket, sc.DefaultShardMaxMember),
	}
}

func (c *shard) MsgDispatch(msg interface{}) {
	c.msgc <- msg
}

func (c *shard) Start() {
	recvc, err := simulate.Subscribe(c.ns.Self.Port, sc.DefaultShardMaxMember)
	if err != nil {
		log.Panic("simulate error ", err)
		return
	}

	c.pvc = recvc

}
