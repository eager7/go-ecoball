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
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/pb"
)

type InvokeInfo struct {
	Method []byte   `json:"method"`
	Param  []string `json:"param"`
}

func NewInvokeContract(from, addr common.AccountName, chainID common.Hash, perm string, method string, param []string, nonce uint64, time int64) (*Transaction, error) {
	invoke := &InvokeInfo{Method: []byte(method), Param: param}
	trans, err := NewTransaction(TxInvoke, from, addr, chainID, perm, invoke, nonce, time)
	if err != nil {
		return nil, err
	}
	return trans, nil
}

func (i *InvokeInfo) Type() uint32 {
	return uint32(TxInvoke)
}

func (i InvokeInfo) GetObject() interface{} {
	return i
}

/**
 *  @brief converts a structure into a sequence of characters
 *  @return []byte - a sequence of characters
 */
func (i *InvokeInfo) Serialize() ([]byte, error) {
	var param []*pb.ParamData
	for _, v := range i.Param {
		p := pb.ParamData{Param: []byte(v)}
		param = append(param, &p)
	}

	p := &pb.InvokeInfo{
		Method: i.Method,
		Param:  param,
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
func (i *InvokeInfo) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New(log, "input data's length is zero")
	}
	var invoke pb.InvokeInfo
	if err := invoke.Unmarshal(data); err != nil {
		return err
	}
	i.Method = common.CopyBytes(invoke.Method)
	for _, v := range invoke.Param {
		p := string(v.Param)
		i.Param = append(i.Param, p)
	}

	return nil
}
func (i *InvokeInfo) Show() {
	fmt.Println(i.JsonString())
}
func (i *InvokeInfo) show() {
	fmt.Println("\t---------Show Invoke Info ----------")
	fmt.Println("\tMethod        :", string(i.Method))
	fmt.Println("\tParam Num     :", len(i.Param))
	for _, v := range i.Param {
		fmt.Println("\tParam         :", v)
	}
	fmt.Println("\t---------------------------")
}

func (i *InvokeInfo) JsonString() string {
	data, _ := json.Marshal(
		struct {
			Method string   `json:"method"`
			Param  []string `json:"param"`
		}{string(i.Method), i.Param})
	return string(data)
}
