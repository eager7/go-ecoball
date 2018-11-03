// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package shard

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/common/errors"
	"fmt"
	"encoding/json"
	"github.com/ecoball/go-ecoball/account"
)

type ViewChangeBlockHeader struct {
	ChainID   common.Hash
	Version   uint32
	Height    uint64
	Timestamp int64
	PrevHash  common.Hash

	CMEpochNo        uint64
	FinalBlockHeight uint64
	Round            uint16
	Candidate        NodeInfo

	hash common.Hash
	*types.COSign
}


func (h *ViewChangeBlockHeader) ComputeHash() error {
	data, err := h.unSignatureData()
	if err != nil {
		return err
	}
	h.hash, err = common.DoubleHash(data)
	if err != nil {
		return err
	}
	return nil
}

func (h *ViewChangeBlockHeader) proto() (*pb.ViewChangeBlockHeader, error) {
	return &pb.ViewChangeBlockHeader{
		ChainID:          h.ChainID.Bytes(),
		Version:          h.Version,
		Height:           h.Height,
		Timestamp:        h.Timestamp,
		PrevHash:         h.PrevHash.Bytes(),
		CMEpochNo:        h.CMEpochNo,
		FinalBlockHeight: h.FinalBlockHeight,
		Round:            uint32(h.Round),
		Candidate: &pb.NodeInfo{
			PublicKey: h.Candidate.PublicKey,
			Address:   h.Candidate.Address,
			Port:      h.Candidate.Port,
		},
		Hash: h.hash.Bytes(),
		COSign: &pb.COSign{
			Step1: h.COSign.Step1,
			Step2: h.COSign.Step2,
		},
	}, nil
}

func (h *ViewChangeBlockHeader) unSignatureData() ([]byte, error) {
	pbHeader, err := h.proto()
	if err != nil {
		return nil, err
	}
	pbHeader.Hash = nil
	pbHeader.COSign = nil
	data, err := pbHeader.Marshal()
	if err != nil {
		return nil, errors.New(log, fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
	}
	return data, nil
}

func (h *ViewChangeBlockHeader) Serialize() ([]byte, error) {
	pbHeader, err := h.proto()
	if err != nil {
		return nil, err
	}
	data, err := pbHeader.Marshal()
	if err != nil {
		return nil, errors.New(log, fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
	}
	return data, nil
}

func (h *ViewChangeBlockHeader) Deserialize(data []byte) error {
	var pbHeader pb.ViewChangeBlockHeader
	if err := pbHeader.Unmarshal(data); err != nil {
		return err
	}
	h.ChainID = common.NewHash(pbHeader.ChainID)
	h.Version = pbHeader.Version
	h.Height = pbHeader.Height
	h.Timestamp = pbHeader.Timestamp
	h.PrevHash = common.NewHash(pbHeader.PrevHash)
	h.CMEpochNo = pbHeader.CMEpochNo
	h.FinalBlockHeight = pbHeader.FinalBlockHeight
	h.Round = uint16(pbHeader.Round)
	h.Candidate = NodeInfo{
		PublicKey: common.CopyBytes(pbHeader.Candidate.PublicKey),
		Address:   pbHeader.Candidate.Address,
		Port:      pbHeader.Candidate.Port,
	}
	h.hash = common.NewHash(pbHeader.Hash)
	h.COSign = &types.COSign{
		Step1: pbHeader.COSign.Step1,
		Step2: pbHeader.COSign.Step2,
	}

	return nil
}

func (h *ViewChangeBlockHeader) JsonString() string {
	data, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(data)
}

func (h *ViewChangeBlockHeader) Type() uint32 {
	return uint32(HeViewChange)
}

func (h *ViewChangeBlockHeader) Hash() common.Hash {
	return h.hash
}

func (h *ViewChangeBlockHeader) GetHeight() uint64 {
	return h.Height
}

func (h *ViewChangeBlockHeader) GetChainID() common.Hash {
	return h.ChainID
}

func (h ViewChangeBlockHeader) GetObject() interface{} {
	return h
}

type ViewChangeBlock struct {
	ViewChangeBlockHeader
}

func (b *ViewChangeBlock) proto() (*pb.ViewChangeBlock, error) {
	pbHeader, err := b.ViewChangeBlockHeader.proto()
	if err != nil {
		return nil, err
	}
	pbBlock := pb.ViewChangeBlock{
		Header: pbHeader,
	}

	return &pbBlock, nil
}

func (b *ViewChangeBlock) Serialize() ([]byte, error) {
	p, err := b.proto()
	if err != nil {
		return nil, err
	}
	data, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *ViewChangeBlock) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New(log, "input data's length is zero")
	}
	var pbBlock pb.ViewChangeBlock
	if err := pbBlock.Unmarshal(data); err != nil {
		return err
	}
	dataHeader, err := pbBlock.Header.Marshal()
	if err != nil {
		return err
	}

	err = b.ViewChangeBlockHeader.Deserialize(dataHeader)
	if err != nil {
		return err
	}


	return nil
}

func (b *ViewChangeBlock) JsonString() string {
	data, err := json.Marshal(b)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(data)
}

func (b ViewChangeBlock) GetObject() interface{} {
	return b
}

func NewVCBlock(header ViewChangeBlockHeader) (*ViewChangeBlock, error) {
	if err := header.ComputeHash(); err != nil {
		return nil, err
	}
	return &ViewChangeBlock{
		ViewChangeBlockHeader: header,
	}, nil
}

func (b *ViewChangeBlock) SetSignature(account *account.Account) error {
	sigData, err := account.Sign(b.hash.Bytes())
	if err != nil {
		return err
	}
	sig := common.Signature{}
	sig.SigData = common.CopyBytes(sigData)
	sig.PubKey = common.CopyBytes(account.PublicKey)
	//t.Signatures = append(t.Signatures, sig)
	return nil
}