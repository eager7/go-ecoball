package smartcontract

import (
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/smartcontract/context"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/elog"
)

func ApplyExecOne(ac *context.ApplyContext) (ret []byte, err error){
	//start := time.Now().UnixNano()
	s := ac.St
	tc := ac.Tc
	action := ac.Action
	cpuLimit := tc.CpuLimit
	netLimit := tc.NetLimit
	timeStamp := tc.TimeStamp

	service, err := NewContractService(s, tc.Trx, action, ac, cpuLimit, netLimit, timeStamp)
	if err != nil {
		return nil, err
	}

	ret, err = service.Execute()
	if err != nil {
		return nil, err
	}

	//elapsed := time.Now().UnixNano() - start

	return ret, nil
}

func ApplyExec(ac *context.ApplyContext) (ret []byte, err error){
	ret, err = ApplyExecOne(ac)

	if ac.RecurseDepth > 4 {
		return nil, errors.New(elog.Log, "inline action recurse depth is out of range")
	}

	for _, act := range ac.InlineAction {
		DispatchAction(ac.Tc, &act, ac.RecurseDepth + 1)
	}

	return ret, err
}


func DispatchAction(tc *context.TranscationContext, action *types.Action, recurseDepth int32) (ret []byte, err error){
	apply, _ := context.NewApplyContext(tc.St, tc, action, recurseDepth)
	tc.Trace = append(tc.Trace, *action)
	ret, err = ApplyExec(apply)
	if err != nil {
		return nil, err
	}

	for _, accName := range apply.Accounts {
		tc.AccountDelta[accName] = apply.AccountDelta[accName]
	}
	return ret,nil
}

