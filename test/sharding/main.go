package main

import (
	"github.com/ecoball/go-ecoball/sharding"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"os"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
)

func main() {
	os.RemoveAll("shard")
	simulate.LoadConfig()
	L, err := ledgerimpl.NewLedger("shard", config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey), true)
	errors.CheckErrorPanic(err)

	actor, _ := sharding.NewShardingActor(L)

	topo := actor.SubscribeShardingTopo()

	go func() {
		for {
			<-topo
		}
	}()

	select {}
}
