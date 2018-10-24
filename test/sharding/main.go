package main

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/sharding"
	"github.com/ecoball/go-ecoball/sharding/cell"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"github.com/ecoball/go-ecoball/txpool"
	"os"
)

func main() {
	simulate.LoadConfig()

	os.RemoveAll("shard")
	L, err := ledgerimpl.NewLedger("shard", config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey), true)
	errors.CheckErrorPanic(err)

	_, err = txpool.Start(L)
	if err != nil {
		panic("txpool error")
	}

	actor, _ := sharding.NewShardingActor(L)

	topo := actor.SubscribeShardingTopo()

	go func() {
		for {
			t := <-topo
			var st *cell.ShardingTopo
			st = t.(*cell.ShardingTopo)
			for _, cm := range st.ShardingInfo {
				print(len(cm))
			}
		}
	}()

	select {}
}
