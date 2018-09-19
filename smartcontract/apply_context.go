package smartcontract

import (
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/smartcontract/context"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/elog"
)

func ApplyExecOne(ac *context.ApplyContext) (err error){
	s := ac.St
	tc := ac.Tc
	action := ac.Action
	cpuLimit := tc.CpuLimit
	netLimit := tc.NetLimit
	timeStamp := tc.TimeStamp

	service, err := NewContractService(s, tc.Trx, action, ac, cpuLimit, netLimit, timeStamp)
	if err != nil {
		return err
	}

	_, err = service.Execute()
	if err != nil {
		return err
	}

	return nil
}

func ApplyExec(ac *context.ApplyContext) (err error){
	err = ApplyExecOne(ac)

	if ac.RecurseDepth > 4 {
		return errors.New(elog.Log, "inline action recurse depth is out of range")
	}

	for _, act := range ac.InlineAction {
		DispatchAction(ac.Tc, &act, ac.RecurseDepth + 1)
	}

	return err
}


func DispatchAction(tc *context.TranscationContext, action *types.Action, recurseDepth int32) (err error){
	apply, _ := context.NewApplyContext(tc.St, tc, action, recurseDepth)
	err = ApplyExec(apply)
	if err != nil {
		return err
	}

	return nil
}
