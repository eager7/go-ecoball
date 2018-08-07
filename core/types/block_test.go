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
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/types"
	"testing"
	"time"
	"github.com/ecoball/go-ecoball/common/errors"
)

func TestHeader(t *testing.T) {
	conData := types.ConsensusData{Type: types.ConSolo, Payload: &types.SoloData{}}
	h, err := types.NewHeader(types.VersionHeader, 10, common.Hash{}, common.Hash{}, common.Hash{}, conData, bloom.Bloom{}, time.Now().Unix())
	errors.CheckErrorPanic(err)
	acc, err := account.NewAccount(0)
	errors.CheckErrorPanic(err)

	errors.CheckErrorPanic(h.SetSignature(&acc))

	data, err := h.Serialize()
	errors.CheckErrorPanic(err)

	h2 := new(types.Header)
	errors.CheckErrorPanic(h2.Deserialize(data))

	errors.CheckEqualPanic(h.JsonString() == h2.JsonString())
	h2.Show()
}

func TestBlockCreate(t *testing.T) {
	g, err := types.GenesesBlockInit()
	errors.CheckErrorPanic(err)

	acc, err := account.NewAccount(0)
	errors.CheckErrorPanic(err)

	errors.CheckErrorPanic(g.SetSignature(&acc))

	data, err := g.Serialize()
	errors.CheckErrorPanic(err)

	block := new(types.Block)
	errors.CheckErrorPanic(block.Deserialize(data))

	errors.CheckEqualPanic(g.JsonString() == block.JsonString())
}
