package state_test

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/state"
	"math/big"
	"os"
	"testing"
	"time"
	"github.com/ecoball/go-ecoball/common/config"
)

func TestStateObject(t *testing.T) {
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	indexAcc := common.NameToIndex("pct")
	os.RemoveAll("/tmp/state_object/")
	acc, _ := state.NewAccount("/tmp/state_object", indexAcc, addr, time.Now().UnixNano())
	//add balance
	errors.CheckErrorPanic(acc.AddBalance(state.AbaToken, new(big.Int).SetUint64(1000)))
	//add perm
	perm := state.NewPermission(state.Active, state.Owner, 2, []state.KeyFactor{}, []state.AccFactor{{Actor: common.NameToIndex("worker1"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker2"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker3"), Weight: 1, Permission: "active"}})
	acc.AddPermission(perm)
	//set cpu net
	acc.AddResourceLimits(true, 100, 100, 200, 200, config.BlockCpuLimit, config.BlockNetLimit)

	data, err := acc.Serialize()
	errors.CheckErrorPanic(err)

	acc2 := new(state.Account)
	errors.CheckErrorPanic(acc2.Deserialize(data))
	acc2.Show()
	errors.CheckEqualPanic(acc.JsonString(false) == acc2.JsonString(false))
}

func TestResourceRecover(t *testing.T) {
	os.RemoveAll("/tmp/state_object_recover/")
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	indexAcc := common.NameToIndex("pct")
	acc, err := state.NewAccount("/tmp/acc", indexAcc, addr, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(acc.AddBalance(state.AbaToken, new(big.Int).SetUint64(1000)))
	acc.AddResourceLimits(true, 100, 100, 100, 100, config.BlockCpuLimit, config.BlockNetLimit)
	elog.Log.Debug(common.JsonString(acc.Resource, false))

	acc.SubResourceLimits(1.0, 1.0, 100, 100, config.BlockCpuLimit, config.BlockNetLimit)
	available := acc.Net.Available
	elog.Log.Debug(common.JsonString(acc.Resource, false))

	time.Sleep(time.Microsecond * 100)
	ti := time.Now().UnixNano()
	errors.CheckErrorPanic(acc.RecoverResources(100, 100, ti, config.BlockCpuLimit, config.BlockNetLimit))
	elog.Log.Debug(common.JsonString(acc.Resource, false))
	if acc.Net.Available < available {
		elog.Log.Error(acc.Net.Available, available)
		t.Fatal("recover failed")
	}

	data, err := acc.Serialize()
	errors.CheckErrorPanic(err)
	accNew := new(state.Account)
	errors.CheckErrorPanic(accNew.Deserialize(data))
	errors.CheckEqualPanic(acc.JsonString(false) == accNew.JsonString(false))

	errors.CheckErrorPanic(acc.SubResourceLimits(1, 1, 100, 100, config.BlockCpuLimit, config.BlockNetLimit))
}
