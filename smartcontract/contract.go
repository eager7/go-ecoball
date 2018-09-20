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

package smartcontract

import (
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/smartcontract/nativeservice"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/smartcontract/context"
)

var log = elog.NewLogger("contract", elog.DebugLog)

type ContractService interface {
	Execute() ([]byte, error)
}

func NewContractService(s state.InterfaceState, tx *types.Transaction, action *types.Action, context *context.ApplyContext, cpuLimit, netLimit float64, timeStamp int64) (ContractService, error) {
	if s == nil || tx == nil || action == nil {
		return nil, errors.New(log, "the contract service's ledger interface or tx is nil")
	}
	contract, err := s.GetContract(action.ContractAccount)
	if err != nil {
		return nil, err
	}
	invoke, ok := action.Payload.GetObject().(types.InvokeInfo)
	if !ok {
		return nil, errors.New(log, "transaction type error[invoke]")
	}

	log.Debug("NewContractService type: ", contract.TypeVm)

	switch contract.TypeVm {
	case types.VmNative:
		service, err := nativeservice.NewNativeService(s, tx, string(invoke.Method), invoke.Param, cpuLimit, netLimit, timeStamp)
		if err != nil {
			return nil, err
		}
		return service, nil
	case types.VmWasm:
		service, err := wasmservice.NewWasmService(s, action, context, contract, &invoke, timeStamp)
		if err != nil {
			return nil, err
		}
		return service, nil
	default:
		return nil, errors.New(log, "unknown virtual machine")
	}
}
