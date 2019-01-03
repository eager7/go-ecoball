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

package geneses

import (
	"errors"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"math/big"
)

func PresetContract(s *state.State, timeStamp int64, addr common.Address) error {
	if s == nil {
		return errors.New("state is nil")
	}
	root := common.NameToIndex("root")
	fmt.Println("preset insert a root account:", addr.HexString())
	if root, err := s.AddAccount(root, addr, timeStamp); err != nil {
		return err
	} else {
		root.SetContract(types.VmNative, []byte("system contract"), nil, nil)
	}

	dsn := common.NameToIndex("dsn")
	fmt.Println("preset insert dsn account:", addr.HexString())
	if _, err := s.AddAccount(dsn, addr, timeStamp); err != nil {
		return err
	}

	abaToken := common.NameToIndex("abatoken")
	fmt.Println("preset insert a token account:", addr.HexString())
	if _, err := s.AddAccount(abaToken, addr, timeStamp); err != nil {
		return err
	}

	s.CreateToken(state.AbaToken, new(big.Int).SetUint64(state.AbaTotal), abaToken, root)

	s.IssueToken(root, new(big.Int).SetUint64(90000), state.AbaToken)

	fmt.Println("set root account's resource to [cpu:100, net:100]")
	if err := s.SetResourceLimits(root, root, 10000, 10000, config.BlockCpuLimit, config.BlockNetLimit); err != nil {
		fmt.Println(err)
		return err
	}

	tester := common.NameToIndex("tester")
	addr = common.AddressFromPubKey(config.Worker1.PublicKey)
	fmt.Println("preset insert a tester account:", addr.HexString())
	if tester, err := s.AddAccount(tester, addr, timeStamp); err != nil {
		return err
	} else {
		tester.SetContract(types.VmNative, []byte("system contract"), nil, nil)
	}

	if err := s.AccountAddBalance(tester, state.AbaToken, new(big.Int).SetUint64(50000)); err != nil {
		return err
	}

	fmt.Println("set root account's resource to [cpu:10000, net:10000]")
	if err := s.SetResourceLimits(tester, tester, 10000, 10000, config.BlockCpuLimit, config.BlockNetLimit); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
