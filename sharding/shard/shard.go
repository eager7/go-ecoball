package shard

import (
	netmsg "github.com/ecoball/go-ecoball/net/message"
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"log"
)

type shard struct {
	ns *cell.Cell

	msgc        chan interface{}
	packetRecvc <-chan netmsg.EcoBallNetMsg
}

func MakeShard(ns *cell.Cell) sc.NodeInstance {
	return &shard{ns: ns,
		msgc: make(chan interface{}),
	}
}

func (c *shard) MsgDispatch(msg interface{}) {
	c.msgc <- msg
}

func (c *shard) Start() {
	recvc, err := simulate.Subscribe(c.ns.Self.Port)
	if err != nil {
		log.Panic("simulate error ", err)
		return
	}

	c.packetRecvc = recvc

}
