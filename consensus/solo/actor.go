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

package solo

import (
	"reflect"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/ecoball/go-ecoball/common/event"
)

type soloActor struct {
	solo *Solo
	pid  *actor.PID
}

func NewSoloActor(l *soloActor) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor {
		return l
	})
	pid, err := actor.SpawnNamed(props, "soloActor")
	if err != nil {
		return nil, err
	}
	event.RegisterActor(event.ActorConsensus, pid)

	return pid, nil
}

func (l *soloActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Stop:
		l.pid.Stop()
		l.solo.stop = true
	case *actor.Restarting:

	default:
		log.Warn("unknown type message:", msg, "type", reflect.TypeOf(msg))
	}
}
