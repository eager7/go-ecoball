package smartcontract

import (
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/smartcontract/context"
	"fmt"
)

func ApplyExecOne(ac *context.ApplyContext) (ret []byte, err error){
	s := ac.St
	tc := ac.Tc
	action := ac.Action
	cpuLimit := ac.CpuLimit
	netLimit := ac.NetLimit
	timeStamp := ac.TimeStamp

	service, err := NewContractService(s, tc.Trx, action, ac, cpuLimit, netLimit, timeStamp)
	if err != nil {
		return nil, err
	}

	ret, err = service.Execute()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func ApplyExec(ac *context.ApplyContext) (ret []byte, err error){
	ret, err = ApplyExecOne(ac)

	for i, act := range ac.InlineAction {
		fmt.Println("inline action ", i)
		DispatchAction(ac.Tc, &act)
	}

	return ret, err
}

//func ApplyExecuteInline(ac *context.ApplyContext, act types.Action) {
//	ac.InlineAction = append(ac.InlineAction, act)
//}


func DispatchAction(tc *context.TranscationContext, action *types.Action) (ret []byte, err error){
	ret, err = TrxExec(tc, action)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func TrxExec(tc *context.TranscationContext, action *types.Action) (ret []byte, err error){
	s := tc.St
	cpuLimit := tc.CpuLimit
	netLimit := tc.NetLimit
	timeStamp := tc.TimeStamp

	apply, _ := context.NewApplyContext(s, tc, action, cpuLimit, netLimit, timeStamp)
	ret, err = ApplyExec(apply)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
