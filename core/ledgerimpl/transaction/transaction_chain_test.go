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

package transaction_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/transaction"
	"github.com/ecoball/go-ecoball/core/types"
	"testing"
)

func TestBlockAdd(t *testing.T) {
	c, err := transaction.NewTransactionChain("/tmp/quaker/Tx", nil, false)
	errors.CheckErrorPanic(err)

	re, err := c.BlockStore.SearchAll()
	errors.CheckErrorPanic(err)
	for k, v := range re {
		hash := common.NewHash([]byte(k))
		block := new(types.Block)
		block.Deserialize([]byte(v))
		fmt.Println(hash.HexString())
		fmt.Println(block)
	}
}

func TestNewMinorBlock(t *testing.T) {

}