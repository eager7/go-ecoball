package main

import (
	"github.com/ecoball/go-ecoball/sharding"
	"github.com/ecoball/go-ecoball/test/example"
)

func main() {
	L := example.Ledger("sharding")

	sharding.NewShardingActor(L)

	select {}
}
