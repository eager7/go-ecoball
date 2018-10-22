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
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
	"math/big"
)

const VersionTx = 1

type TxType uint32

const (
	TxDeploy   TxType = 0x01
	TxInvoke   TxType = 0x02
	TxTransfer TxType = 0x03
)

type VmType uint32

const (
	VmWasm   VmType = 0x01
	VmNative VmType = 0x02
)

type Transaction struct {
	Version    uint32             `json:"version"`
	ChainID    common.Hash        `json:"chain_id"`
	Type       TxType             `json:"type"`
	From       common.AccountName `json:"from"`
	Permission string             `json:"permission"`
	Addr       common.AccountName `json:"addr"`
	Nonce      uint64             `json:"nonce"`
	TimeStamp  int64              `json:"timeStamp"`
	Payload    Payload            `json:"payload"`
	Signatures []common.Signature `json:"signatures"`
	Hash       common.Hash        `json:"hash"`
	Receipt    TransactionReceipt
}

func NewTransaction(t TxType, from, addr common.AccountName, chainID common.Hash, perm string, payload Payload, nonce uint64, time int64) (*Transaction, error) {
	if payload == nil {
		return nil, errors.New(log, "the transaction's payload is nil")
	}
	tx := Transaction{
		Version:    VersionTx,
		ChainID:    chainID,
		Type:       t,
		From:       from,
		Permission: perm,
		Addr:       addr,
		Nonce:      nonce,
		TimeStamp:  time,
		Payload:    payload,
		Signatures: nil,
		Hash:       common.Hash{},
		Receipt: TransactionReceipt{
			From:      from,
			To:        addr,
			TokenName: "",
			Amount:    new(big.Int).SetUint64(0),
			Hash:      common.Hash{},
			Cpu:       0,
			Net:       0,
			NewToken:  nil,
			Accounts:  make(map[int][]byte, 1),
			Producer:  0,
			Result:    nil,
		},
	}
	if tx.Permission == "" {
		tx.Permission = "active"
	}
	data, err := tx.unSignatureData()
	if err != nil {
		return nil, err
	}
	tx.Hash, err = common.DoubleHash(data)
	if err != nil {
		return nil, err
	}
	tx.Receipt.Hash = tx.Hash
	return &tx, nil
}

func (t *Transaction) SetSignature(account *account.Account) error {
	sigData, err := account.Sign(t.Hash.Bytes())
	if err != nil {
		return err
	}
	sig := common.Signature{}
	sig.SigData = common.CopyBytes(sigData)
	sig.PubKey = common.CopyBytes(account.PublicKey)
	t.Signatures = append(t.Signatures, sig)
	return nil
}

func (t *Transaction) VerifySignature() (bool, error) {
	for _, v := range t.Signatures {
		result, err := secp256k1.Verify(t.Hash.Bytes(), v.SigData, v.PubKey)
		if err != nil || result != true {
			return false, err
		}
	}
	return true, nil
}

func (t *Transaction) unSignatureData() ([]byte, error) {
	payload, err := t.Payload.Serialize()
	if err != nil {
		return nil, err
	}
	p := &pb.TxPayload{
		Version:    t.Version,
		ChainID:    t.ChainID.Bytes(),
		From:       uint64(t.From),
		Permission: []byte(t.Permission),
		Addr:       uint64(t.Addr),
		Payload:    payload,
		Nonce:      t.Nonce,
		Timestamp:  t.TimeStamp,
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (t *Transaction) ProtoBuf() (*pb.Transaction, error) {
	payload, err := t.Payload.Serialize()
	if err != nil {
		return nil, err
	}
	var sig []*pb.Signature
	for i := 0; i < len(t.Signatures); i++ {
		s := &pb.Signature{PubKey: t.Signatures[i].PubKey, SigData: t.Signatures[i].SigData}
		sig = append(sig, s)
	}
	receipt, err := t.Receipt.Serialize()
	if err != nil {
		return nil, err
	}
	p := &pb.Transaction{
		Payload: &pb.TxPayload{
			Version:    t.Version,
			ChainID:    t.ChainID.Bytes(),
			Type:       uint32(t.Type),
			From:       uint64(t.From),
			Permission: []byte(t.Permission),
			Addr:       uint64(t.Addr),
			Payload:    payload,
			Nonce:      t.Nonce,
			Timestamp:  t.TimeStamp,
		},
		Sign: sig,
		Hash: t.Hash.Bytes(),
		Receipt: receipt,
	}
	return p, nil
}

/**
 *  @brief converts a structure into a sequence of characters
 *  @return []byte - a sequence of characters
 */
func (t *Transaction) Serialize() ([]byte, error) {
	p, err := t.ProtoBuf()
	if err != nil {
		return nil, err
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, errors.New(log, fmt.Sprintf("Marshal error:%s", err.Error()))
	}
	return b, nil
}

/**
 *  @brief converts a sequence of characters into a structure
 *  @param data - a sequence of characters
 */
func (t *Transaction) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New(log, "input data's length is zero")
	}

	var txPb pb.Transaction
	if err := txPb.Unmarshal(data); err != nil {
		return errors.New(log, fmt.Sprintf("data len: %d, unMarshal error:%s", len(data), err.Error()))
	}

	t.Version = txPb.Payload.Version
	t.ChainID = common.NewHash(txPb.Payload.ChainID)
	t.Type = TxType(txPb.Payload.Type)
	t.From = common.AccountName(txPb.Payload.From)
	t.Permission = string(txPb.Payload.Permission)
	t.Addr = common.AccountName(txPb.Payload.Addr)
	t.Nonce = txPb.Payload.Nonce
	t.TimeStamp = txPb.Payload.Timestamp


	if err := t.Receipt.Deserialize(txPb.Receipt); err != nil {
		return err
	}
	if t.Payload == nil {
		switch t.Type {
		case TxTransfer:
			t.Payload = new(TransferInfo)
		case TxDeploy:
			t.Payload = new(DeployInfo)
		case TxInvoke:
			t.Payload = new(InvokeInfo)
		default:
			return errors.New(log, "the transaction's payload must not be nil")
		}
	}
	if err := t.Payload.Deserialize(txPb.Payload.Payload); err != nil {
		return err
	}
	for i := 0; i < len(txPb.Sign); i++ {
		sig := common.Signature{
			PubKey:  common.CopyBytes(txPb.Sign[i].PubKey),
			SigData: common.CopyBytes(txPb.Sign[i].SigData),
		}
		t.Signatures = append(t.Signatures, sig)
	}
	t.Hash = common.NewHash(txPb.Hash)

	return nil
}

func (t *Transaction) JsonString() string {
	/*data, _ := json.Marshal(struct {
		Version    uint32 `json:"version"`
		ChainID    string
		Type       string             `json:"type"`
		From       string             `json:"from"`
		Permission string             `json:"permission"`
		Addr       string             `json:"addr"`
		Nonce      uint64             `json:"nonce"`
		TimeStamp  int64              `json:"timeStamp"`
		Payload    string             `json:"payload"`
		Signatures []common.Signature `json:"signatures"`
		Hash       string             `json:"hash"`
		Receipt    TransactionReceipt `json:"receipt"`
	}{Version: t.Version, ChainID: t.ChainID.HexString(), Type: t.Type.String(), From: t.From.String(),
		Permission: t.Permission, Addr: t.Addr.String(), Nonce: t.Nonce,
		TimeStamp: t.TimeStamp, Payload: t.Payload.JsonString(), Signatures: t.Signatures,
		Hash: t.Hash.HexString(), Receipt: t.Receipt})*/
	data, _ := json.Marshal(t)
	return string(data)
}

func (t TxType) String() string {
	switch t {
	case TxDeploy:
		return "deploy transaction"
	case TxInvoke:
		return "invoke transaction"
	case TxTransfer:
		return "transfer transaction"
	default:
		return "unknown type"
	}
}
