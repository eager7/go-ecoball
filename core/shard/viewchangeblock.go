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
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/core/types"
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

	Hashes common.Hash
	*types.COSign
}

func (h *ViewChangeBlockHeader) ComputeHash() error {
	data, err := h.unSignatureData()
	if err != nil {
		return err
	}
	h.Hashes, err = common.DoubleHash(data)
	if err != nil {
		return err
	}
	return nil
}

func (h *ViewChangeBlockHeader) VerifySignature() (bool, error) {
	/*for _, v := range h.Signatures {
		b, err := secp256k1.Verify(h.Hash.Bytes(), v.SigData, v.PubKey)
		if err != nil || b != true {
			return false, err
		}
	}*/
	return true, nil
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
		Hash:   h.Hashes.Bytes(),
		COSign: h.COSign.Proto(),
	}, nil
}

func (h *ViewChangeBlockHeader) unSignatureData() ([]byte, error) {
	pbHeader, err := h.proto()
	if err != nil {
		return nil, err
	}
	pbHeader.Hash = nil
	pbHeader.COSign.Sign1 = nil
	pbHeader.COSign.Sign2 = nil
	pbHeader.COSign.Step1 = 0
	pbHeader.COSign.Step2 = 0
	data, err := pbHeader.Marshal()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
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
		return nil, errors.New(fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
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
	h.Hashes = common.NewHash(pbHeader.Hash)
	h.COSign = &types.COSign{
		TPubKey: pbHeader.COSign.TPubKey,
		Step1:   pbHeader.COSign.Step1,
		Sign1:   nil,
		Step2:   pbHeader.COSign.Step2,
		Sign2:   nil,
	}
	h.COSign.Sign1 = append(h.COSign.Sign1, pbHeader.COSign.Sign1...)
	h.COSign.Sign2 = append(h.COSign.Sign2, pbHeader.COSign.Sign2...)
	return nil
}

func (h *ViewChangeBlockHeader) JsonString() string {
	data, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return "hash:" + h.Hashes.HexString() + string(data)
}

func (h *ViewChangeBlockHeader) Type() uint32 {
	return uint32(HeViewChange)
}

func (h *ViewChangeBlockHeader) Hash() common.Hash {
	return h.Hashes
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
		return errors.New("input data's length is zero")
	}
	var pbBlock pb.ViewChangeBlock
	if err := pbBlock.Unmarshal(data); err != nil {
		return errors.New(err.Error())
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
	return "hash:" + b.Hashes.HexString() + string(data)
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
	sigData, err := account.Sign(b.Hashes.Bytes())
	if err != nil {
		return err
	}
	sig := common.Signature{}
	sig.SigData = common.CopyBytes(sigData)
	sig.PubKey = common.CopyBytes(account.PublicKey)
	//t.Signatures = append(t.Signatures, sig)
	return nil
}
