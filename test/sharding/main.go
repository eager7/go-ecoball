package main

import (
	"github.com/ecoball/go-ecoball/sharding"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"github.com/ecoball/go-ecoball/test/example"
)

func main() {
	L := example.Ledger("sharding")

	simulate.LoadConfig()

	actor, _ := sharding.NewShardingActor(L)

	topo := actor.SubscribeShardingTopo()

	go func() {
		for {
			<-topo
		}
	}()

	select {}
}
