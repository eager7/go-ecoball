package shard

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/pb"
	"fmt"
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
	HeCmBlock    HeaderType = 1
	HeFinalBlock HeaderType = 2
	HeMinorBlock HeaderType = 3
	HeViewChange HeaderType = 4
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
	//SetSignature(account *account.Account) error
	VerifySignature() (bool, error)
	Hash() common.Hash
	GetChainID() common.Hash
	GetHeight() uint64
}

type BlockInterface interface {
	Payload
	Hash() common.Hash
	GetChainID() common.Hash
	GetHeight() uint64
}

func Serialize(payload Payload) ([]byte, error) {
	if payload == nil {
		return nil, errors.New(log, "the payload is nil")
	}
	data, err := payload.Serialize()
	if err != nil {
		return nil, err
	}
	pbPayload := pb.Payload{
		Type: payload.Type(),
		Data: data,
	}
	data, err = pbPayload.Marshal()
	if err != nil {
		return nil, errors.New(log, fmt.Sprintf("the marshal error:%s", err.Error()))
	}
	return data, nil
}

func BlockDeserialize(data []byte) (BlockInterface, error) {
	if len(data) == 0 {
		return nil, errors.New(log, "input data's length is zero")
	}
	var pbPayload pb.Payload
	if err := pbPayload.Unmarshal(data); err != nil {
		return nil, errors.New(log, err.Error())
	}
	data = pbPayload.Data
	switch HeaderType(pbPayload.Type) {
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
	case HeViewChange:
		block := new(ViewChangeBlock)
		if err := block.Deserialize(data); err != nil {
			return nil, err
		}
		return block, nil
	default:
		return nil, errors.New(log, "unknown header type")
	}
}

func HeaderDeserialize(data []byte) (HeaderInterface, error) {
	if len(data) == 0 {
		return nil, errors.New(log, "input data's length is zero")
	}
	var pbPayload pb.Payload
	if err := pbPayload.Unmarshal(data); err != nil {
		return nil, errors.New(log, err.Error())
	}
	data = pbPayload.Data
	switch HeaderType(pbPayload.Type) {
	case HeCmBlock:
		header := new(CMBlockHeader)
		if err := header.Deserialize(data); err != nil {
			return nil, err
		}
		return header, nil
	case HeMinorBlock:
		header := new(MinorBlockHeader)
		if err := header.Deserialize(data); err != nil {
			return nil, err
		}
		return header, nil
	case HeFinalBlock:
		header := new(FinalBlockHeader)
		if err := header.Deserialize(data); err != nil {
			return nil, err
		}
		return header, nil
	case HeViewChange:
		header := new(ViewChangeBlockHeader)
		if err := header.Deserialize(data); err != nil {
			return nil, err
		}
		return header, nil
	default:
		return nil, errors.New(log, "unknown header type")
	}
}