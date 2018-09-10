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

package config_test

import (
	"fmt"
	"testing"

	"github.com/ecoball/go-ecoball/common/config"
	_ "github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common"
)

func TestConfig(t *testing.T) {
	fmt.Println(config.HttpLocalPort)
	fmt.Println(config.ConsensusAlgorithm)
	fmt.Println(config.PeerList)
	fmt.Println(config.StartNode)

	fmt.Println("worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString())
	fmt.Println("worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString())
	fmt.Println("worker3", common.AddressFromPubKey(config.Worker3.PublicKey).HexString())

	fmt.Println("chain Id:", config.ChainHash.HexString())
}
