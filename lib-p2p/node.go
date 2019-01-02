// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball.
//
// The go-ecoball is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.

package p2p

import (
	"context"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/lib-p2p/net"
)

var (
	log         = elog.NewLogger("net", elog.DebugLog)
	NodeNetWork *net.Instance
)

func InitNetWork(ctx context.Context) *net.Instance {
	var err error
	NodeNetWork, err = net.NewInstance(ctx, config.SwarmConfig.PrivateKey, config.SwarmConfig.ListenAddress[0])
	if err != nil {
		log.Error(err)
	}
	if err := NewNetActor(&netActor{instance: NodeNetWork, ctx: ctx}); err != nil {
		log.Panic(err)
	}
	return nil
}
