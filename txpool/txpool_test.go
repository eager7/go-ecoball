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
package txpool_test

import (
	"testing"
	"time"

	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/test/example"
	"github.com/ecoball/go-ecoball/txpool"
)

func TestTxPool(t *testing.T) {
	ledger := example.Ledger("/tmp/txPool")
	_, err := txpool.Start(ledger)
	errors.CheckErrorPanic(err)

	tx := example.TestTransfer()

	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, tx))
	time.Sleep(time.Duration(1) * time.Second)
}
