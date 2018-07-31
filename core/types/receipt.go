package types

import "github.com/ecoball/go-ecoball/common"

type Receipt struct {
	Hash   common.Hash
	Cpu    float32
	Net    float32
	Result []byte
}

