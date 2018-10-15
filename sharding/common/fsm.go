package common

import "github.com/ecoball/go-ecoball/common/elog"

var (
	log = elog.NewLogger("sdcommon", elog.DebugLog)
)

const (
	StateNil = iota
)

type PreCallFunc func(msg interface{}) bool
type ActCallFunc func(msg interface{})
type AfterCallFunc func(msg interface{})

type FsmElem struct {
	State     int
	Action    int
	PreAct    PreCallFunc
	Act       ActCallFunc
	AfterAct  AfterCallFunc
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

			if elem.PreAct != nil {
				if !elem.PreAct(msg) {
					return
				}
			}

			if elem.Act != nil {
				elem.Act(msg)
			}

			if elem.Nextstate != StateNil {
				f.state = elem.Nextstate
				log.Debug("new state ", f.state)
			}

			if elem.AfterAct != nil {
				elem.AfterAct(msg)
			}

			return
		}
	}

	log.Error("wrong fsm state ", f.state, "action  ", action)
}

func (f *Fsm) getState() int {
	return f.state
}
