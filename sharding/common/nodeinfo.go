package common

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
)

type Worker struct {
	Pubkey  string
	Address string
	Port    string
}

func (a *Worker) Equal(b *Worker) bool {
	return a.Pubkey == b.Pubkey
}

func (a *Worker) EqualNode(b *cs.NodeInfo) bool {
	bkey := string(b.PublicKey)
	return a.Pubkey == bkey
}

func (a *Worker) InitWork(b *cs.NodeInfo) {
	a.Pubkey = string(b.PublicKey)
	a.Address = b.Address
	a.Port = b.Port
}

func (a *Worker) Copy(b *Worker) {
	a.Pubkey = b.Pubkey
	a.Address = b.Address
	a.Port = b.Port
}

type ShardingTopo struct {
	ShardId      uint16
	Pubkey       string
	ShardingInfo [][]Worker
}
