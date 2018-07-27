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
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
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

type Payload interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	GetObject() interface{}
	Show()
}

type Transaction struct {
	Version    uint32             `json:"version"`
	Type       TxType             `json:"type"`
	From       common.AccountName `json:"from"`
	Permission string             `json:"permission"`
	Addr       common.AccountName `json:"addr"`
	Nonce      uint64             `json:"nonce"`
	TimeStamp  int64              `json:"timeStamp"`
	Payload    Payload            `json:"payload"`
	Signatures []common.Signature `json:"signatures"`
	Hash       common.Hash        `json:"hash"`
}

func NewTransaction(t TxType, from, addr common.AccountName, perm string, payload Payload, nonce uint64, time int64) (*Transaction, error) {
	if payload == nil {
		return nil, errors.New("the transaction's payload is nil")
	}
	tx := Transaction{
		Version:    VersionTx,
		Type:       t,
		From:       from,
		Permission: perm,
		Addr:       addr,
		Nonce:      nonce,
		TimeStamp:  time,
		Payload:    payload,
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

func (t *Transaction) protoBuf() (*pb.Transaction, error) {
	payload, err := t.Payload.Serialize()
	if err != nil {
		return nil, err
	}
	var sig []*pb.Signature
	for i := 0; i < len(t.Signatures); i++ {
		s := &pb.Signature{PubKey: t.Signatures[i].PubKey, SigData: t.Signatures[i].SigData}
		sig = append(sig, s)
	}
	p := &pb.Transaction{
		Payload: &pb.TxPayload{
			Version:    t.Version,
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
	}
	return p, nil
}

/**
 *  @brief converts a structure into a sequence of characters
 *  @return []byte - a sequence of characters
 */
func (t *Transaction) Serialize() ([]byte, error) {
	p, err := t.protoBuf()
	if err != nil {
		return nil, err
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

/**
 *  @brief converts a sequence of characters into a structure
 *  @param data - a sequence of characters
 */
func (t *Transaction) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	var tx pb.Transaction
	if err := tx.Unmarshal(data); err != nil {
		return err
	}

	t.Version = tx.Payload.Version
	t.Type = TxType(tx.Payload.Type)
	t.From = common.AccountName(tx.Payload.From)
	t.Permission = string(tx.Payload.Permission)
	t.Addr = common.AccountName(tx.Payload.Addr)
	t.Nonce = tx.Payload.Nonce
	t.TimeStamp = tx.Payload.Timestamp
	if t.Payload == nil {
		switch t.Type {
		case TxTransfer:
			t.Payload = new(TransferInfo)
		case TxDeploy:
			t.Payload = new(DeployInfo)
		case TxInvoke:
			t.Payload = new(InvokeInfo)
		default:
			return errors.New("the transaction's payload must not be nil")
		}
	}
	if err := t.Payload.Deserialize(tx.Payload.Payload); err != nil {
		return err
	}
	for i := 0; i < len(tx.Sign); i++ {
		sig := common.Signature{
			PubKey:  common.CopyBytes(tx.Sign[i].PubKey),
			SigData: common.CopyBytes(tx.Sign[i].SigData),
		}
		t.Signatures = append(t.Signatures, sig)
	}
	t.Hash = common.NewHash(tx.Hash)

	return nil
}

func (t *Transaction) Show() {
	fmt.Println("\t---------------Transaction-------------")
	fmt.Println("\tVersion        :", t.Version)
	fmt.Println("\tFrom           :", common.IndexToName(t.From))
	fmt.Println("\tAddr           :", common.IndexToName(t.Addr))
	fmt.Println("\tTime           :", t.TimeStamp)
	fmt.Println("\tHash           :", t.Hash.HexString())
	fmt.Println("\tSig Len        :", len(t.Signatures))
	for i := 0; i < len(t.Signatures); i++ {
		fmt.Println("\tPublicKey      :", common.ToHex(t.Signatures[i].PubKey))
		fmt.Println("\tSigData        :", common.ToHex(t.Signatures[i].SigData))
	}
	t.Payload.Show()
}

func (t *Transaction) JsonString() string {
	data, _ := json.Marshal(t)
	return string(data)
}
