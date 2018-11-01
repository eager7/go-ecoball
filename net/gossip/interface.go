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

// Define the interface for the owner of gossip puller

package gossip

type Receiver interface {
	GetDigests() []string
	ContainItemInDigests(digests []string, item string) bool

	//input map: key is the peer id, value is the digest slice of peer
	//output map: key is the peer id, value is the digest for request
	ShuffelDigests(revDigests map[string][]string, digest []string) map[string][]string

	GetItemData(item string) []byte
	UpdateItemData(dataArray [][]byte) error
}