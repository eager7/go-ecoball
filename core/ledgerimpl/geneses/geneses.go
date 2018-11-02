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

/*
func GenesisBlockInit(ledger ledger.Ledger, timeStamp int64) (*types.Block, error) {
	if ledger == nil {
		return nil, errors.New("ledger is nil")
	}

	//TODO start
	SecondInMs := int64(1000)
	BlockIntervalInMs := int64(15000)
	timeStamp = int64((timeStamp*SecondInMs-SecondInMs)/BlockIntervalInMs) * BlockIntervalInMs
	timeStamp = timeStamp / SecondInMs
	//TODO end

	hash := common.NewHash([]byte("EcoBall Geneses Block"))
	conData := types.GenesesBlockInitConsensusData(timeStamp)
	txs, err := PresetContract(ledger, timeStamp)
	if err != nil {
		return nil, err
	}


	hashState := ledger.StateDB().GetHashRoot()
	header, err := types.NewHeader(types.VersionHeader, 1, hash, hash, hashState, *conData, bloom.Bloom{}, timeStamp)
	if err != nil {
		return nil, err
	}
	block := types.Block{Header: header, CountTxs: uint32(len(txs)), Transactions: txs}

	if err := block.SetSignature(&config.Root); err != nil {
		return nil, err
	}
	return &block, nil
}*/

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

	s.CreateToken(state.AbaToken, new(big.Int).SetUint64(state.AbaTotal), root, root)

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

	s.CreateToken(state.AbaToken, new(big.Int).SetUint64(state.AbaTotal), root, root)
	s.IssueToken(root, new(big.Int).SetUint64(90000), state.AbaToken)

	fmt.Println("set root account's resource to [cpu:10000, net:10000]")
	if err := s.SetResourceLimits(root, root, 10000, 10000, config.BlockCpuLimit, config.BlockNetLimit); err != nil {
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
