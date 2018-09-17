package types

import "github.com/ecoball/go-ecoball/common"

type Payload interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	GetObject() interface{}
	Type() uint32
	JsonString() string
}

type HeaderType uint32

const (
	HeMinorBlock HeaderType = 1
	HeCmBlock    HeaderType = 2
	HeFinalBlock HeaderType = 3
)

func (h HeaderType) String() string {
	switch h {
	case HeCmBlock:
		return "HeCmBlock Type"
	case HeMinorBlock:
		return "HeMinorBlock Type"
	case HeFinalBlock:
		return "HeFinalBlock Type"
	default:
		return "unknown type"
	}
}

type HeaderInterface interface {
	Payload
	Hash() common.Hash
	GetHeight() uint64
}

type BlockInterface interface {
	HeaderInterface
}