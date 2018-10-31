package common

import "time"

type Stimer struct {
	T  *time.Timer
	On bool
}

func NewStimer(d time.Duration, start bool) *Stimer {
	s := &Stimer{
		T:  time.NewTimer(d),
		On: start,
	}

	if !start {
		s.T.Stop()
	}

	return s
}

func (s *Stimer) Stop() {
	if !s.On {
		return
	}

	if !s.T.Stop() {
		log.Debug("timer stop faild")
		select {
		case <-s.T.C:
			log.Debug("select timer c")
		default:
		}
	} else {
		log.Debug("timer stop success")
	}

	s.On = false
}

func (s *Stimer) Reset(d time.Duration) {
	s.Stop()
	s.T.Reset(d)
	s.On = true
}
