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
package encoding

import (
	"encoding/binary"
	"io"
)

// EncInt64 encodes an int64 as a slice of 8 bytes.
func EncInt64(i int64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return
}

// DecInt64 decodes a slice of 8 bytes into an int64.
// If len(b) < 8, the slice is padded with zeros.
func DecInt64(b []byte) int64 {
	b2 := make([]byte, 8)
	copy(b2, b)
	return int64(binary.LittleEndian.Uint64(b2))
}

// EncUint64 encodes a uint64 as a slice of 8 bytes.
func EncUint64(i uint64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return
}

// DecUint64 decodes a slice of 8 bytes into a uint64.
// If len(b) < 8, the slice is padded with zeros.
func DecUint64(b []byte) uint64 {
	b2 := make([]byte, 8)
	copy(b2, b)
	return binary.LittleEndian.Uint64(b2)
}

// WriteUint64 writes u to w.
func WriteUint64(w io.Writer, u uint64) error {
	_, err := w.Write(EncUint64(u))
	return err
}

// WriteInt writes i to w.
func WriteInt(w io.Writer, i int) error {
	return WriteUint64(w, uint64(i))
}
