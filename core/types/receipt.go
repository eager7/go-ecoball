package types

import "github.com/ecoball/go-ecoball/common"

type TransactionReceipt struct {
	Hash   common.Hash
	Cpu    float64
	Net    float64
	Result []byte
}

type BlockReceipt struct {
	BlockCpu float64
	BlockNet float64
}