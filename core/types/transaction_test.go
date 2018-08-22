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
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/test/example"
	"testing"
	"github.com/ecoball/go-ecoball/common/elog"
)

func TestTransfer(t *testing.T) {
	tx := example.TestTransfer()
	result, err := tx.VerifySignature()
	errors.CheckErrorPanic(err)
	errors.CheckEqualPanic(result)

	transferData, err := tx.Serialize()
	errors.CheckErrorPanic(err)

	tx2 := &types.Transaction{}
	errors.CheckErrorPanic(tx2.Deserialize(transferData))

	elog.Log.Debug(tx.JsonString())
	elog.Log.Info(tx2.JsonString())
	errors.CheckEqualPanic(tx.JsonString() == tx2.JsonString())
}

func TestDeploy(t *testing.T) {
	deploy := example.TestDeploy([]byte("test"))
	data, err := deploy.Serialize()
	errors.CheckErrorPanic(err)

	dep := &types.Transaction{Payload: new(types.DeployInfo)}
	errors.CheckErrorPanic(dep.Deserialize(data))

	errors.CheckEqualPanic(dep.JsonString() == deploy.JsonString())
}

func TestInvoke(t *testing.T) {
	i := example.TestInvoke("main")
	data, err := i.Serialize()
	errors.CheckErrorPanic(err)

	i2 := new(types.Transaction)
	errors.CheckErrorPanic(i2.Deserialize(data))

	errors.CheckEqualPanic(i.JsonString() == i2.JsonString())
}
