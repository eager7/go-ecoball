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

package ipldecoball

import (
	"fmt"
	"io"
	"github.com/klauspost/reedsolomon"
)

// rsCode is a Reed-Solomon encoder/decoder. It implements the
// ErasureCoder interface.
type rsCode struct {
	enc reedsolomon.Encoder

	numShards  int
	dataShards int
}

// NumPieces returns the number of pieces returned by Encode.
func (rs *rsCode) NumShards() int { return rs.numShards }

// MinPieces return the minimum number of pieces that must be present to
// recover the original data.
func (rs *rsCode) MinShards() int { return rs.dataShards }

// Encode splits data into equal-length pieces, some containing the original
// data and some containing parity data.
func (rs *rsCode) Encode(data []byte) ([][]byte, error) {
	pieces, err := rs.enc.Split(data)
	if err != nil {
		return nil, err
	}
	// err should not be possible if Encode is called on the result of Split,
	// but no harm in checking anyway.
	err = rs.enc.Encode(pieces)
	if err != nil {
		return nil, err
	}
	return pieces, nil
}

// EncodeShards creates the parity shards for an already sharded input.
func (rs *rsCode) EncodeShards(shards [][]byte) ([][]byte, error) {
	// Check that the caller provided the minimum amount of pieces.
	if len(shards) != rs.MinShards() {
		return nil, fmt.Errorf("invalid number of pieces given %v %v", len(shards), rs.MinShards())
	}
	// Add the parity shards.
	for len(shards) < rs.NumShards() {
		shards = append(shards, make([]byte, int(DataShardSize)))
	}
	err := rs.enc.Encode(shards)
	if err != nil {
		return nil, err
	}
	return shards, nil
}

// Recover recovers the original data from Shards and writes it to w.
// shards should be identical to the slice returned by Encode (length and
// order must be preserved), but with missing elements set to nil.
func (rs *rsCode) Recover(shards [][]byte, n uint64, w io.Writer) error {
	err := rs.enc.ReconstructData(shards)
	if err != nil {
		return err
	}
	return rs.enc.Join(w, shards, int(n))
}

// NewRSCode creates a new Reed-Solomon encoder/decoder using the supplied
// parameters.
func NewRSCode(nData, nParity int) (*rsCode, error) {
	enc, err := reedsolomon.New(nData, nParity)
	if err != nil {
		return nil, err
	}
	return &rsCode{
		enc:        enc,
		numShards:  nData + nParity,
		dataShards: nData,
	}, nil
}