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

package request

import (
	innerCommon "github.com/ecoball/go-ecoball/common"
)

type AccountName struct {
	Name      string
	ChainHash innerCommon.Hash
}

type TokenName struct {
	Name      string
	ChainHash innerCommon.Hash
}

type BlockHeight struct {
	Height    uint64
	ChainHash innerCommon.Hash
}

type TransactionHash struct {
	Hash      innerCommon.Hash
	ChainHash innerCommon.Hash
}

type PermissionPublicKeys struct {
	Name       string
	Permission string
	ChainHash  innerCommon.Hash
}
