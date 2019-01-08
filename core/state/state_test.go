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

package state_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/state"
	"math/big"
	"os"
	"testing"
	"time"
)

func TestStateNew(t *testing.T) {
	acc := common.NameToIndex("root")
	_ = os.RemoveAll("/tmp/state/")
	s, err := state.NewState("/tmp/state", common.Hash{})
	errors.CheckErrorPanic(err)
	fmt.Println("Trie Root:", s.GetHashRoot().HexString())

	addr := common.AddressFromPubKey(config.Root.PublicKey)
	_, err = s.AddAccount(acc, addr, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	_, _ = s.CreateToken(state.AbaToken, new(big.Int).SetUint64(state.AbaTotal), acc, acc)
	errors.CheckErrorPanic(s.AccountAddBalance(acc, state.AbaToken, new(big.Int).SetUint64(90000)))

	balance, err := s.AccountGetBalance(acc, state.AbaToken)
	errors.CheckErrorPanic(err)
	fmt.Println("Value From:", balance)

	value := new(big.Int).SetUint64(100)
	if err := s.AccountAddBalance(acc, state.AbaToken, value); err != nil {
		fmt.Println("Update Error:", err)
	}

	fmt.Println("Hash Root:", s.GetHashRoot().HexString())
	_ = s.CommitToDB()
	balance, err = s.AccountGetBalance(acc, state.AbaToken)
	errors.CheckErrorPanic(err)
	fmt.Println("Value:", balance)
}

func TestStateRoot(t *testing.T) {
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	acc := common.NameToIndex("pct")
	token := state.AbaToken
	_ = os.RemoveAll("/tmp/state_root/")
	s, err := state.NewState("/tmp/state_root", common.Hash{})
	errors.CheckErrorPanic(err)

	_, err = s.AddAccount(acc, addr, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	_, err = s.CreateToken(state.AbaToken, new(big.Int).SetUint64(state.AbaTotal), acc, acc)
	errors.CheckErrorPanic(err)

	errors.CheckErrorPanic(s.AccountAddBalance(acc, token, new(big.Int).SetInt64(100)))

	value, err := s.AccountGetBalance(acc, token)
	errors.CheckErrorPanic(err)
	fmt.Println("value:", value)
	fmt.Println("root:", s.GetHashRoot().HexString())
	errors.CheckEqualPanic(value.Uint64() == 100)

	errors.CheckErrorPanic(s.AccountAddBalance(acc, token, new(big.Int).SetInt64(150)))

	value, err = s.AccountGetBalance(acc, token)
	errors.CheckErrorPanic(err)

	fmt.Println("value:", value)
	fmt.Println("root:", s.GetHashRoot().HexString())
	errors.CheckEqualPanic(value.Uint64() == 250)
	_ = s.CommitToDB()
}

func TestStateDBCopy(t *testing.T) {
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	acc := common.NameToIndex("pct")
	_ = os.RemoveAll("/tmp/state_copy/")
	s, err := state.NewState("/tmp/state_copy", common.HexToHash(""))
	errors.CheckErrorPanic(err)
	if _, err := s.AddAccount(acc, addr, time.Now().UnixNano()); err != nil {
		t.Fatal(err)
	}
	_, _ = s.CreateToken(state.AbaToken, new(big.Int).SetUint64(state.AbaTotal), acc, acc)
	errors.CheckErrorPanic(s.AccountAddBalance(acc, state.AbaToken, new(big.Int).SetInt64(100)))
	errors.CheckErrorPanic(s.SetResourceLimits(acc, acc, 10, 10, config.BlockCpuLimit, config.BlockNetLimit))
	_ = s.CommitToDB()
	value, err := s.AccountGetBalance(acc, state.AbaToken)
	errors.CheckErrorPanic(err)
	elog.Log.Info(value)
	errors.CheckEqualPanic(value.Uint64() == 80)

	copyS, err := s.StateCopy()
	errors.CheckErrorPanic(err)
	origin, err := copyS.GetAccountByName(acc)
	errors.CheckErrorPanic(err)
	elog.Log.Info(s.Accounts.Get(acc).String())
	elog.Log.Warn(origin.String())
	errors.CheckEqualPanic(s.Accounts.Get(acc).String() == origin.String())

	_ = copyS.AccountAddBalance(acc, state.AbaToken, new(big.Int).SetUint64(300))
	balance, err := copyS.AccountGetBalance(acc, state.AbaToken)
	errors.CheckErrorPanic(err)
	elog.Log.Info(balance)
	errors.CheckEqualPanic(balance.Uint64() == 380)

	errors.CheckErrorPanic(copyS.SubResources(acc, 1, 1, config.BlockCpuLimit, config.BlockNetLimit))
	elog.Log.Debug(copyS.RequireResources(acc, config.BlockCpuLimit, config.BlockNetLimit, time.Now().UnixNano()))
}

func TestStateDBReset(t *testing.T) {
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	acc := common.NameToIndex("pct")
	_ = os.RemoveAll("/tmp/state_copy/")
	s, err := state.NewState("/tmp/state_copy", common.HexToHash(""))
	errors.CheckErrorPanic(err)
	tm, err := time.Parse("02/01/2006 15:04:05 PM", "21/02/1990 00:00:00 AM")
	errors.CheckErrorPanic(err)
	timeStamp := tm.UnixNano()
	_, err = s.AddAccount(acc, addr, timeStamp)
	errors.CheckErrorPanic(err)

	_, _ = s.CreateToken(state.AbaToken, new(big.Int).SetUint64(state.AbaTotal), acc, acc)
	errors.CheckErrorPanic(s.AccountAddBalance(acc, state.AbaToken, new(big.Int).SetInt64(100)))
	_ = s.CommitToDB()

	checkBalance(100, acc, s)

	prevHash := s.GetHashRoot()
	elog.Log.Info(prevHash.HexString())

	errors.CheckErrorPanic(s.AccountAddBalance(acc, state.AbaToken, new(big.Int).SetInt64(100)))
	_ = s.CommitToDB()

	checkBalance(200, acc, s)

	errors.CheckErrorPanic(s.Reset(prevHash))

	checkBalance(100, acc, s)
}

func checkBalance(value uint64, index common.AccountName, s *state.State) {
	balance, err := s.AccountGetBalance(index, state.AbaToken)
	errors.CheckErrorPanic(err)
	elog.Log.Info(balance)
	errors.CheckEqualPanic(balance.Uint64() == value)
}

func TestState_Store(t *testing.T) {
	acc := common.NameToIndex("pct")
	_ = os.RemoveAll("/tmp/state_store/")
	s, err := state.NewState("/tmp/state_store", common.HexToHash(""))
	errors.CheckErrorPanic(err)
	_, err = s.AddAccount(acc, common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330")), time.Now().UnixNano())
	errors.CheckErrorPanic(err)

	errors.CheckErrorPanic(s.StoreSet(acc, []byte("key"), []byte("value")))
	value, err := s.StoreGet(acc, []byte("key"))
	errors.CheckErrorPanic(err)
	if string(value) != "value" {
		t.Fatal("must be value:", value)
	}
}
