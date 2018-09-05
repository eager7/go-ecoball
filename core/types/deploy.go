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

type DeployInfo struct {
	TypeVm   VmType `json:"typeVm"`
	Describe []byte `json:"describe"`
	Code     []byte `json:"code"`
	Abi      []byte `json:"abi"`
}

func NewDeployContract(from, addr common.AccountName, chainID common.Hash, perm string, vm VmType, des string, code []byte, abi []byte, nonce uint64, time int64) (*Transaction, error) {
	deploy := &DeployInfo{
		TypeVm:   vm,
		Describe: []byte(des),
		Code:     code,
		Abi:	  abi,
	}
	trans, err := NewTransaction(TxDeploy, from, addr, chainID, perm, deploy, nonce, time)
	if err != nil {
		return nil, err
	}
	return trans, nil
}

/**
 *  @brief converts a structure into a sequence of characters
 *  @return []byte - a sequence of characters
 */
func (d *DeployInfo) Serialize() ([]byte, error) {
	p := &pb.DeployInfo{
		TypeVm:   uint32(d.TypeVm),
		Describe: d.Describe,
		Code:     d.Code,
		Abi:      d.Abi,
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
func (d *DeployInfo) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New(log, "input data's length is zero")
	}
	var deploy pb.DeployInfo
	if err := deploy.Unmarshal(data); err != nil {
		return err
	}
	d.TypeVm = VmType(deploy.TypeVm)
	d.Describe = common.CopyBytes(deploy.Describe)
	d.Code = common.CopyBytes(deploy.Code)
	d.Abi = common.CopyBytes(deploy.Abi)

	return nil
}

func (d *DeployInfo) Type() uint32 {
	return uint32(TxDeploy)
}

func (d DeployInfo) GetObject() interface{} {
	return d
}

func (d *DeployInfo) Show() {
	fmt.Println(d.JsonString())
}

func (d *DeployInfo) show() {
	fmt.Println("\t---------Show Deploy Info ----------")
	fmt.Println("\tTypeVm        :", d.TypeVm)
	fmt.Println("\tDescribe      :", string(d.Describe))
}

func (d *DeployInfo) JsonString() string {
	data, _ := json.Marshal(d)
	return string(data)
}
