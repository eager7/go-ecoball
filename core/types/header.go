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

package types

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
)

const VersionHeader = 1

type Header struct {
	Version    uint32
	ChainID    common.Hash
	TimeStamp  int64
	Height     uint64
	ConsData   ConsData
	PrevHash   common.Hash
	MerkleHash common.Hash
	StateHash  common.Hash
	Bloom      bloom.Bloom

	Receipt    BlockReceipt
	Signatures []common.Signature
	Hash       common.Hash
}

var log = elog.NewLogger("LedgerImpl", elog.DebugLog)

func (h *Header) ComputeHash() error {
	data, err := h.unSignatureData()
	if err != nil {
		return err
	}
	h.Hash, err = common.DoubleHash(data)
	if err != nil {
		return err
	}
	return nil
}

func (h *Header) SetSignature(account *account.Account) error {
	sigData, err := account.Sign(h.Hash.Bytes())
	if err != nil {
		return err
	}
	sig := common.Signature{}
	sig.SigData = common.CopyBytes(sigData)
	sig.PubKey = common.CopyBytes(account.PublicKey)
	h.Signatures = append(h.Signatures, sig)
	return nil
}

func (h *Header) VerifySignature() (bool, error) {
	for _, v := range h.Signatures {
		b, err := secp256k1.Verify(h.Hash.Bytes(), v.SigData, v.PubKey)
		if err != nil || b != true {
			return false, err
		}
	}
	return true, nil
}

/**
** Used to compute hash
 */
func (h *Header) unSignatureData() ([]byte, error) {
	p, err := h.proto()
	if err != nil {
		return nil, err
	}
	p.Hash = nil
	p.Signature = nil
	p.Receipt = nil
	data, err := p.Marshal()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
	}
	return data, nil
}

/**
 *  @brief converts a structure into a sequence of characters
 *  @return []byte - a sequence of characters
 */
func (h *Header) Serialize() ([]byte, error) {
	p, err := h.proto()
	if err != nil {
		return nil, err
	}
	data, err := p.Marshal()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
	}
	return data, nil
}

/**
 *  @brief converts a sequence of characters into a structure
 *  @param data - a sequence of characters
 */
func (h *Header) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var pbHeader pb.Header
	if err := pbHeader.Unmarshal(data); err != nil {
		return err
	}

	h.Version = pbHeader.Version
	h.ChainID = common.NewHash(pbHeader.ChainID)
	h.TimeStamp = pbHeader.Timestamp
	h.Height = pbHeader.Height
	h.PrevHash = common.NewHash(pbHeader.PrevHash)
	h.MerkleHash = common.NewHash(pbHeader.MerkleHash)
	for i := 0; i < len(pbHeader.Signature); i++ {
		sig := common.Signature{
			PubKey:  common.CopyBytes(pbHeader.Signature[i].PubKey),
			SigData: common.CopyBytes(pbHeader.Signature[i].SigData),
		}
		h.Signatures = append(h.Signatures, sig)
	}
	h.StateHash = common.NewHash(pbHeader.StateHash)
	h.Hash = common.NewHash(pbHeader.Hash)
	h.Receipt = BlockReceipt{BlockNet: pbHeader.Receipt.BlockNet, BlockCpu: pbHeader.Receipt.BlockCpu}
	h.Bloom = bloom.NewBloom(pbHeader.Bloom)

	dataCon, err := pbHeader.ConsensusData.Marshal()
	if err != nil {
		return err
	}
	if err := h.ConsData.Deserialize(dataCon); err != nil {
		return err
	}

	return nil
}

func (h *Header) String() string {
	data, err := json.Marshal(
		struct {
			ChainID    string
			Version    uint32
			TimeStamp  int64
			Height     uint64
			ConsData   ConsData
			PrevHash   string
			MerkleHash string
			StateHash  string
			bloom      bloom.Bloom
			Signatures []common.Signature
			Hash       string
		}{
			ChainID:    h.ChainID.HexString(),
			Version:    h.Version,
			TimeStamp:  h.TimeStamp,
			Height:     h.Height,
			ConsData:   h.ConsData,
			PrevHash:   h.PrevHash.HexString(),
			MerkleHash: h.MerkleHash.HexString(),
			StateHash:  h.StateHash.HexString(),
			Signatures: h.Signatures,
			Hash:       h.Hash.HexString(),
		})
	if err != nil {
		log.Error(err)
		return ""
	}
	return string(data)
}

func (h *Header) proto() (*pb.Header, error) {
	var sig []*pb.Signature
	for i := 0; i < len(h.Signatures); i++ {
		s := &pb.Signature{PubKey: h.Signatures[i].PubKey, SigData: h.Signatures[i].SigData}
		sig = append(sig, s)
	}
	pbCon, err := h.ConsData.proto()
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &pb.Header{
		Version:       h.Version,
		ChainID:       h.ChainID.Bytes(),
		Timestamp:     h.TimeStamp,
		Height:        h.Height,
		ConsensusData: pbCon,
		PrevHash:      h.PrevHash.Bytes(),
		MerkleHash:    h.MerkleHash.Bytes(),
		StateHash:     h.StateHash.Bytes(),
		Bloom:         h.Bloom.Bytes(),
		Hash:          h.Hash.Bytes(),
		Signature:     sig,
		Receipt:       &pb.BlockReceipt{BlockCpu: h.Receipt.BlockCpu, BlockNet: h.Receipt.BlockNet},
	}, nil
}

func (h *Header) Identify() mpb.Identify {
	return mpb.Identify_APP_MSG_HEADER
}

func (h *Header) GetInstance() interface{} {
	return h
}
