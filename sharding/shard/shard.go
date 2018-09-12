package shard

import (
	netmsg "github.com/ecoball/go-ecoball/net/message"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/node"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"log"
)

type shard struct {
	ns *node.Node

	msgc        chan interface{}
	packetRecvc <-chan netmsg.EcoBallNetMsg
}

func MakeShard(ns *node.Node) sc.NodeInstance {
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
		log.Panic("simulate error %s", err)
		return
	}

	c.packetRecvc = recvc

}
