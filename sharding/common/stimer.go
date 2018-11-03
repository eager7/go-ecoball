package common

import "time"

type Stimer struct {
	T  *time.Timer
	on bool
}

func NewStimer(d time.Duration, start bool) *Stimer {
	s := &Stimer{
		T:  time.NewTimer(d),
		on: start,
	}

	if !start {
		s.T.Stop()
	}

	return s
}

func (s *Stimer) GetStatus() bool {
	return s.on
}

func (s *Stimer) SetStop() {
	s.on = false
}

func (s *Stimer) Stop() {
	if !s.on {
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

	s.on = false
}

func (s *Stimer) Reset(d time.Duration) {
	s.Stop()
	s.T.Reset(d)
	s.on = true
}
