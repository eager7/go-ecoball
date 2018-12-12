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

package types

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/pb"
)

//type COSign struct {
//	//Step1         TBLS_SIG
//	BitValidateor1 [32]uint8
//	//Step2         TBLS_SIG
//	BitValidateor2 [32]uint8
//}

type COSign struct {
	TPubKey []byte
	Step1   uint32
	Sign1   [][]byte
	Step2   uint32
	Sign2   [][]byte
}

func (c *COSign) Proto() *pb.COSign {
	p := pb.COSign{
		TPubKey: common.CopyBytes(c.TPubKey),
		Step1:   c.Step1,
		Sign1:   nil,
		Step2:   c.Step2,
		Sign2:   nil,
	}
	p.Sign1 = append(p.Sign1, c.Sign1...)
	p.Sign2 = append(p.Sign2, c.Sign2...)
	return &p
}
