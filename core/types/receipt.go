package types

import "github.com/ecoball/go-ecoball/common"

type Receipt struct {
	Hash   common.Hash
	Cpu    float64
	Net    float64
	Result []byte
}

