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
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common/config"
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

	//// set root control token account
	//perm := state.Permission{Keys: make(map[string]state.KeyFactor, 1), Accounts: make(map[string]state.AccFactor, 1)}
	//perm.Accounts["root"] = state.AccFactor{Actor: common.NameToIndex("root"), Weight: 1, Permission: "active"}
	//s.AddPermission(abaToken, perm)

	//saving := common.NameToIndex("saving")
	//savingAddr := common.AddressFromPubKey(config.Saving.PublicKey)
	//fmt.Println("preset insert a bpay account:", savingAddr.HexString())
	//if root, err := s.AddAccount(saving, savingAddr, timeStamp); err != nil {
	//	return err
	//} else {
	//	root.SetContract(types.VmNative, []byte("system contract"), nil, nil)
	//}
	//
	//bpay := common.NameToIndex("bpay")
	//bpayAddr := common.AddressFromPubKey(config.Bpay.PublicKey)
	//fmt.Println("preset insert a bpay account:", bpayAddr.HexString())
	//if root, err := s.AddAccount(bpay, bpayAddr, timeStamp); err != nil {
	//	return err
	//} else {
	//	root.SetContract(types.VmNative, []byte("system contract"), nil, nil)
	//}

	s.CreateToken(state.AbaToken, new(big.Int).SetUint64(state.AbaTotal), abaToken, root)

	//if err := s.AccountAddBalance(root, state.AbaToken, new(big.Int).SetUint64(90000)); err != nil {
	//	return err
	//}

	s.IssueToken(root, new(big.Int).SetUint64(90000), state.AbaToken)

	fmt.Println("set root account's resource to [cpu:100, net:100]")
	if err := s.SetResourceLimits(root, root, 10000, 10000, config.BlockCpuLimit, config.BlockNetLimit); err != nil {
		fmt.Println(err)
		return err
	}

	/*
		delegate := common.NameToIndex("delegate")
		if _, err := s.AddAccount(delegate, common.AddressFromPubKey(config.Delegate.PublicKey), timeStamp); err != nil {
			return err
		}
		if err := s.AccountAddBalance(delegate, state.AbaToken, new(big.Int).SetUint64(10000)); err != nil {
			return err
		}
		fmt.Println("set root account's resource to [cpu:100, net:100]")
		if err := s.SetResourceLimits(delegate, delegate, 1000, 1000, types.BlockCpuLimit, types.BlockNetLimit); err != nil {
			fmt.Println(err)
			return err
		}
	*/
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

func PresetShardContract(s *state.State, timeStamp int64, addr common.Address) error {
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

	//// set root control token account
	//perm := state.Permission{Keys: make(map[string]state.KeyFactor, 1), Accounts: make(map[string]state.AccFactor, 1)}
	//perm.Accounts["root"] = state.AccFactor{Actor: common.NameToIndex("root"), Weight: 1, Permission: "active"}
	//s.AddPermission(abaToken, perm)

	s.CreateToken(state.AbaToken, new(big.Int).SetUint64(state.AbaTotal), abaToken, root)

	s.IssueToken(root, new(big.Int).SetUint64(900000), state.AbaToken)

	fmt.Println("set root account's resource to [cpu:10000, net:10000]")
	if err := s.SetResourceLimits(root, root, 20000, 50000, config.BlockCpuLimit, config.BlockNetLimit); err != nil {
		fmt.Println(err)
		return err
	}

	worker := common.NameToIndex("testeru")
	addr = common.AddressFromPubKey(config.Worker.PublicKey)
	fmt.Println("preset insert a tester account:", addr.HexString())
	if _, err := s.AddAccount(worker, addr, timeStamp); err != nil {
		return err
	}

	if err := s.AccountAddBalance(worker, state.AbaToken, new(big.Int).SetUint64(50000)); err != nil {
		return err
	}

	fmt.Println("set root account's resource to [cpu:10000, net:10000]")
	if err := s.SetResourceLimits(worker, worker, 10000, 10000, config.BlockCpuLimit, config.BlockNetLimit); err != nil {
		fmt.Println(err)
		return err
	}

	worker1 := common.NameToIndex("testerh")
	addr = common.AddressFromPubKey(config.Worker1.PublicKey)
	fmt.Println("preset insert a tester account:", addr.HexString())
	if _, err := s.AddAccount(worker1, addr, timeStamp); err != nil {
		return err
	}

	if err := s.AccountAddBalance(worker1, state.AbaToken, new(big.Int).SetUint64(50000)); err != nil {
		return err
	}

	fmt.Println("set root account's resource to [cpu:10000, net:10000]")
	if err := s.SetResourceLimits(worker1, worker1, 10000, 10000, config.BlockCpuLimit, config.BlockNetLimit); err != nil {
		fmt.Println(err)
		return err
	}

	worker2 := common.NameToIndex("testerl")
	addr = common.AddressFromPubKey(config.Worker2.PublicKey)
	fmt.Println("preset insert a tester account:", addr.HexString())
	if _, err := s.AddAccount(worker2, addr, timeStamp); err != nil {
		return err
	}

	if err := s.AccountAddBalance(worker2, state.AbaToken, new(big.Int).SetUint64(50000)); err != nil {
		return err
	}

	fmt.Println("set root account's resource to [cpu:10000, net:10000]")
	if err := s.SetResourceLimits(worker2, worker2, 10000, 10000, config.BlockCpuLimit, config.BlockNetLimit); err != nil {
		fmt.Println(err)
		return err
	}

	worker3 := common.NameToIndex("testerp")
	addr = common.AddressFromPubKey(config.Worker3.PublicKey)
	fmt.Println("preset insert a tester account:", addr.HexString())
	if _, err := s.AddAccount(worker3, addr, timeStamp); err != nil {
		return err
	}

	if err := s.AccountAddBalance(worker3, state.AbaToken, new(big.Int).SetUint64(50000)); err != nil {
		return err
	}

	fmt.Println("set root account's resource to [cpu:10000, net:10000]")
	if err := s.SetResourceLimits(worker3, worker3, 10000, 10000, config.BlockCpuLimit, config.BlockNetLimit); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
