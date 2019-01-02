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
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/pb"
	"math/big"
)

type TransferInfo struct {
	Token string   `json:"token"`
	Value *big.Int `json:"value"`
}

func NewTransferInfo(token string, v *big.Int) *TransferInfo {
	t := new(TransferInfo)
	t.Token = token
	t.Value = new(big.Int).Set(v)
	return t
}

func NewTransfer(from, to common.AccountName, chainID common.Hash, perm string, value *big.Int, nonce uint64, time int64) (*Transaction, error) {
	payload := NewTransferInfo("ABA", value)
	return NewTransaction(TxTransfer, from, to, chainID, perm, payload, nonce, time)
}

/**
 *  @brief converts a structure into a sequence of characters
 *  @return []byte - a sequence of characters
 */
func (t *TransferInfo) Serialize() ([]byte, error) {

	data, err := t.Value.GobEncode()
	if err != nil {
		return nil, err
	}
	pbTransfer := pb.Transfer{
		Token: t.Token,
		Value: data,
	}
	b, err := pbTransfer.Marshal()
	if err != nil {
		return nil, errors.New("marshal failed")
	}
	return b, nil
}

/**
 *  @brief converts a sequence of characters into a structure
 *  @param data - a sequence of characters
 */
func (t *TransferInfo) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("data len is 0")
	}
	var pbTransfer pb.Transfer
	if err := pbTransfer.Unmarshal(data); err != nil {
		return errors.New("unMarshal failed")
	}
	t.Token = pbTransfer.Token
	t.Value = new(big.Int)
	return t.Value.GobDecode(pbTransfer.Value)
}

func (t *TransferInfo) GetInstance() interface{} {
	return t
}

func (t *TransferInfo) Identify() mpb.Identify {
	return mpb.Identify_APP_MSG_TRANSACTION_TRANSFER
}

func (t *TransferInfo) String() string {
	data, _ := json.Marshal(t)
	return string(data)
}
