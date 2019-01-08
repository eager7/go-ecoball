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

package types_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/test/example"
	"math/big"
	"testing"
)

func TestTransfer(t *testing.T) {
	tx := example.TestTransfer()
	receipt := &types.TrxReceipt{
		From:   common.NameToIndex("root"),
		Addr:   common.NameToIndex("root"),
		Token:  "",
		Amount: new(big.Int).SetUint64(0),
		Cpu:    10,
		Net:    20,
		Result: []byte("result"),
	}
	tx.Receipt = receipt
	result, err := tx.VerifySignature()
	errors.CheckErrorPanic(err)
	errors.CheckEqualPanic(result)

	transferData, err := tx.Serialize()
	errors.CheckErrorPanic(err)

	tx2 := &types.Transaction{}
	errors.CheckErrorPanic(tx2.Deserialize(transferData))

	elog.Log.Debug(tx.String())
	elog.Log.Info(tx2.String())
	errors.CheckEqualPanic(tx.String() == tx2.String())
}

func TestDeploy(t *testing.T) {
	deploy := example.TestDeploy([]byte("test"))
	data, err := deploy.Serialize()
	errors.CheckErrorPanic(err)

	dep := &types.Transaction{Payload: new(types.DeployInfo)}
	errors.CheckErrorPanic(dep.Deserialize(data))
	elog.Log.Debug(deploy.String())
	elog.Log.Info(dep.String())
	errors.CheckEqualPanic(dep.String() == deploy.String())
}

func TestInvoke(t *testing.T) {
	i := example.TestInvoke("main")
	data, err := i.Serialize()
	errors.CheckErrorPanic(err)

	i2 := new(types.Transaction)
	errors.CheckErrorPanic(i2.Deserialize(data))

	errors.CheckEqualPanic(i.String() == i2.String())
}

func TestSize(t *testing.T) {
	tx := example.TestTransfer()
	data, _ := tx.Serialize()
	fmt.Println(len(data))
}
