package renter

import (
	"math/big"
	"github.com/ecoball/go-ecoball/common"
)

type AddFileReq struct {
	FileSize      uint64
	AccountName   string
	Redundancy    uint8
	Allowance     big.Int
	Collateral    big.Int
	MaxCollateral big.Int
	ChainId       common.Hash
}

type AddFileRsp struct {
	Result string
}
