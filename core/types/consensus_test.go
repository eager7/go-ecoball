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
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/types"
	"reflect"
	"testing"
)

func TestDBft(t *testing.T) {
	dposData := &types.DPosData{}
	consensusData := types.ConsData{Type: types.CondPos, Payload: dposData}

	data, err := consensusData.Serialize()
	errors.CheckErrorPanic(err)

	conData := new(types.ConsData)
	errors.CheckErrorPanic(conData.Deserialize(data))

	con := types.ConsData{}
	fmt.Println(reflect.ValueOf(con))
}
