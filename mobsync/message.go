package mobsync

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/core/types"
)

type BlockRequest struct {
	ChainId     common.Hash
	BlockHeight uint64
}

func (b *BlockRequest) Identify() mpb.Identify {
	return mpb.Identify_APP_MSG_BLOCK_REQUEST
}
func (b *BlockRequest) String() string {
	return fmt.Sprintf("chain hash:%s, height:%d", b.ChainId.String(), b.BlockHeight)
}
func (b *BlockRequest) GetInstance() interface{} {
	return b
}
func (b *BlockRequest) Serialize() ([]byte, error) {
	proto := pb.BlockRequest{
		ChainId:     b.ChainId.Bytes(),
		BlockHeight: b.BlockHeight,
	}

	data, err := proto.Marshal()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("serialize block request message error:%s", err.Error()))
	}
	return data, nil
}
func (b *BlockRequest) Deserialize(data []byte) error {
	proto := &pb.BlockRequest{}
	if err := proto.Unmarshal(data); err != nil {
		return errors.New(fmt.Sprintf("deserialize block request message error:%s", err.Error()))
	}
	b.ChainId = common.NewHash(proto.ChainId)
	b.BlockHeight = proto.BlockHeight
	return nil
}

type BlockResponse struct {
	ChainId common.Hash
	Blocks  []*types.Block
}

func (b *BlockResponse) Identify() mpb.Identify {
	return mpb.Identify_APP_MSG_BLOCK_RESPONSE
}
func (b *BlockResponse) String() string {
	return fmt.Sprintf("chain hash:%s, block number:%d", b.ChainId.String(), len(b.Blocks))
}
func (b *BlockResponse) GetInstance() interface{} {
	return b
}
func (b *BlockResponse) Serialize() ([]byte, error) {
	var pbBlocks []*pb.Block
	for _, block := range b.Blocks {
		if pbBlock, err := block.Proto(); err != nil {
			return nil, err
		} else {
			pbBlocks = append(pbBlocks, pbBlock)
		}
	}
	proto := pb.BlockResponse{
		ChainId: b.ChainId.Bytes(),
		Blocks:  pbBlocks,
	}
	data, err := proto.Marshal()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("serialize block response message error:%s", err.Error()))
	}
	return data, nil
}
func (b *BlockResponse) Deserialize(data []byte) error {
	proto := &pb.BlockResponse{}
	if err := proto.Unmarshal(data); err != nil {
		return errors.New(fmt.Sprintf("deserialize block response message error:%s", err.Error()))
	}
	b.ChainId = common.NewHash(proto.ChainId)
	for _, pbBlock := range proto.Blocks {
		if data, err := pbBlock.Marshal(); err != nil {
			return errors.New(fmt.Sprintf("marshal pb block error:%s", err.Error()))
		} else {
			block := new(types.Block)
			if err := block.Deserialize(data); err != nil {
				return err
			}
			b.Blocks = append(b.Blocks, block)
		}
	}
	return nil
}
