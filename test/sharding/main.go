package main

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/sharding"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"github.com/ecoball/go-ecoball/txpool"
	"os"
	"github.com/ecoball/go-ecoball/test/example"
)

func main() {
	simulate.LoadConfig("/tmp/sharding.json")

	os.RemoveAll("shard")
	L, err := ledgerimpl.NewLedger("shard", config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey), true)
	errors.CheckErrorPanic(err)

	_, err = txpool.Start(L)
	if err != nil {
		panic("txpool error")
	}

	actor, _ := sharding.NewShardingActor(L)

	topo := actor.SubscribeShardingTopo()
	actor.SetNet(nil)


	go example.TransferExample()
	//go simulate.SyncComplete()

	go func() {
		for {
			t := <-topo
			var st *sc.ShardingTopo
			st = t.(*sc.ShardingTopo)
			for _, cm := range st.ShardingInfo {
				print(len(cm))
			}
		}
	}()

	select {}
}
