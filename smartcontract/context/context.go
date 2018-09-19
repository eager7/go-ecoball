package context

import (
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
)

type ApplyContext struct {
	Tc 				*TranscationContext
	Action			*types.Action
	InlineAction	[]types.Action
	St 				*state.State
	RecurseDepth	int32
}

func NewApplyContext(s *state.State, tc *TranscationContext, action	*types.Action, recurseDepth int32) (*ApplyContext, error){
	context := &ApplyContext{
		Tc:				tc,
		Action:			action,
		InlineAction:	nil,
		St:				s,
		RecurseDepth:	recurseDepth,
	}

	return context, nil
}


type TranscationContext struct {
	Trx 		*types.Transaction
	St 			*state.State
	TimeStamp 	int64
	CpuLimit 	float64
	NetLimit 	float64
}

func NewTranscationContext(s *state.State, tx *types.Transaction, cpuLimit, netLimit float64, timeStamp int64) (*TranscationContext, error){
	trxContext := &TranscationContext{
		Trx:		tx,
		St:			s,
		TimeStamp:	timeStamp,
		CpuLimit:	cpuLimit,
		NetLimit:	netLimit,
	}

	return trxContext, nil
}