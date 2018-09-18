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
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/types"
	"reflect"
	"testing"
)

func TestDBft(t *testing.T) {
	dposData := &types.DPosData{}
	consensusData := types.ConsensusData{Type: types.CondPos, Payload: dposData}

	data, err := consensusData.Serialize()
	errors.CheckErrorPanic(err)

	conData := new(types.ConsensusData)
	errors.CheckErrorPanic(conData.Deserialize(data))

	con := types.ConsensusData{}
	fmt.Println(reflect.ValueOf(con))
}

func TestAbaBft(t *testing.T) {
	sig1 := common.Signature{PubKey: []byte("1234"), SigData: []byte("5678")}
	sig2 := common.Signature{PubKey: []byte("4321"), SigData: []byte("8765")}
	var sigPer []common.Signature
	sigPer = append(sigPer, sig1)
	sigPer = append(sigPer, sig2)
	abaData := types.AbaBftData{NumberRound: 5, PreBlockSignatures: sigPer}

	conData := types.NewConsensusPayload(types.ConABFT, &abaData)

	data, err := conData.Serialize()
	errors.CheckErrorPanic(err)

	conDataDeserialize := new(types.ConsensusData)
	errors.CheckErrorPanic(conDataDeserialize.Deserialize(data))

	conDataObj, ok := conDataDeserialize.Payload.GetObject().(types.AbaBftData)
	if !ok {
		t.Fatal("type error")
	}
	if conDataObj.NumberRound != 5 {
		t.Fatal("NumberRound mismatch")
	}
	if len(conDataObj.PreBlockSignatures) != 2 {
		t.Fatal("PreBlockSignatures mismatch")
	}
}
