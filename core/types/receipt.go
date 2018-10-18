package types

import (
	"github.com/ecoball/go-ecoball/common"
	"math/big"
)

type AccountReceipt struct {
	Balance *big.Int
}

type TransactionReceipt struct {
	From   AccountReceipt
	To     AccountReceipt
	Hash   common.Hash
	Cpu    float64
	Net    float64
	Account [2][]byte
	Result []byte
}

type BlockReceipt struct {
	BlockCpu float64
	BlockNet float64
}
