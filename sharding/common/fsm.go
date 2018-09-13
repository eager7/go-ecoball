package common

import "github.com/ecoball/go-ecoball/common/elog"

var (
	log = elog.NewLogger("sdcommon", elog.DebugLog)
)

const (
	StateNil = iota
)

type callFunc func(msg interface{})

type FsmElem struct {
	State     int
	Action    int
	Call      callFunc
	Nextstate int
}

type Fsm struct {
	state int
	elems []FsmElem
}

func NewFsm(state int, elems []FsmElem) *Fsm {
	fsm := &Fsm{state: state}

	fsm.elems = elems
	return fsm
}

func (f *Fsm) Execute(action int, msg interface{}) {
	log.Debug("state ", f.state, " action ", action)
	for _, elem := range f.elems {
		if f.state == elem.State &&
			action == elem.Action {
			if elem.Call != nil {
				elem.Call(msg)
			}

			if elem.Nextstate != StateNil {
				f.state = elem.Nextstate
				log.Debug("new state ", f.state)
			}
			return
		}
	}

	log.Panic("wrong fsm action ", action, " state ", f.state)
	panic("wrong fsm")
}

func (f *Fsm) getState() int {
	return f.state
}
