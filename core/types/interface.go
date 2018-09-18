package types

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
)

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

func BlockDeserialize(data []byte, typ HeaderType) (BlockInterface, error) {
	switch typ {
	case HeCmBlock:
		block := new(CMBlock)
		if err := block.Deserialize(data); err != nil {
			return nil, err
		}
		return block, nil
	case HeMinorBlock:
		block := new(MinorBlock)
		if err := block.Deserialize(data); err != nil {
			return nil, err
		}
		return block, nil
	case HeFinalBlock:
		block := new(FinalBlock)
		if err := block.Deserialize(data); err != nil {
			return nil, err
		}
		return block, nil
	default:
		return nil, errors.New(log, "unknown header type")
	}
}
