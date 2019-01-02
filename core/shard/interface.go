package shard

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/core/types"
)

type HeaderInterface interface {
	types.EcoMessage
	VerifySignature() (bool, error)
	Hash() common.Hash
	GetChainID() common.Hash
	GetHeight() uint64
}

type BlockInterface interface {
	HeaderInterface
}

func Serialize(payload types.EcoMessage) ([]byte, error) {
	if payload == nil {
		return nil, errors.New("the payload is nil")
	}
	data, err := payload.Serialize()
	if err != nil {
		return nil, err
	}
	pbPayload := pb.Payload{
		Type: uint32(payload.Identify()),
		Data: data,
	}
	data, err = pbPayload.Marshal()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("the marshal error:%s", err.Error()))
	}
	return data, nil
}

func BlockDeserialize(data []byte) (BlockInterface, error) {
	if len(data) == 0 {
		return nil, errors.New("input data's length is zero")
	}
	var pbPayload pb.Payload
	if err := pbPayload.Unmarshal(data); err != nil {
		return nil, errors.New(err.Error())
	}
	data = pbPayload.Data
	switch mpb.Identify(pbPayload.Type) {
	case mpb.Identify_APP_MSG_CM_BLOCK:
		block := new(CMBlock)
		if err := block.Deserialize(data); err != nil {
			return nil, err
		}
		return block, nil
	case mpb.Identify_APP_MSG_MINOR_BLOCK:
		block := new(MinorBlock)
		if err := block.Deserialize(data); err != nil {
			return nil, err
		}
		return block, nil
	case mpb.Identify_APP_MSG_FINAL_BLOCK:
		block := new(FinalBlock)
		if err := block.Deserialize(data); err != nil {
			return nil, err
		}
		return block, nil
	case mpb.Identify_APP_MSG_VC_BLOCK:
		block := new(ViewChangeBlock)
		if err := block.Deserialize(data); err != nil {
			return nil, err
		}
		return block, nil
	default:
		return nil, errors.New("unknown header type")
	}
}

func HeaderDeserialize(data []byte) (HeaderInterface, error) {
	if len(data) == 0 {
		return nil, errors.New("input data's length is zero")
	}
	var pbPayload pb.Payload
	if err := pbPayload.Unmarshal(data); err != nil {
		return nil, errors.New(err.Error())
	}
	data = pbPayload.Data
	switch mpb.Identify(pbPayload.Type) {
	case mpb.Identify_APP_MSG_CM_BLOCK:
		header := new(CMBlockHeader)
		if err := header.Deserialize(data); err != nil {
			return nil, err
		}
		return header, nil
	case mpb.Identify_APP_MSG_MINOR_BLOCK:
		header := new(MinorBlockHeader)
		if err := header.Deserialize(data); err != nil {
			return nil, err
		}
		return header, nil
	case mpb.Identify_APP_MSG_FINAL_BLOCK:
		header := new(FinalBlockHeader)
		if err := header.Deserialize(data); err != nil {
			return nil, err
		}
		return header, nil
	case mpb.Identify_APP_MSG_VC_BLOCK:
		header := new(ViewChangeBlockHeader)
		if err := header.Deserialize(data); err != nil {
			return nil, err
		}
		return header, nil
	default:
		return nil, errors.New("unknown header type")
	}
}
