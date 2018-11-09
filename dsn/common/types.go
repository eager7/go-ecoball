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

package common


import (
	"bytes"
	"bufio"
	"encoding/binary"
	"github.com/ecoball/go-ecoball/dsn/host/proof"
	"github.com/ecoball/go-ecoball/dsn/common/crypto"
)

const (
//SegmentSize = 64
	SigSize = 65
	RootAccount = "root"
	FcMethodAn = "reg_store"
	FcMethodProof = "reg_proof"
	FcMethodFile = "reg_file"
	EraDataPiece = 10
)

type FileContract struct {
	PublicKey   []byte
	Cid         string
	LocalPath   string
	FileSize    uint64
	Redundancy  uint8
	Funds       []byte
	StartAt     uint64
	Expiration  uint64
	AccountName string
	PayId       string
}

type HostAncContract struct {
	PublicKey     []byte
	TotalStorage  uint64
	StartAt       uint64
	Collateral    []byte
	MaxCollateral []byte
	AccountName   string
}

type StorageProof struct {
	PublicKey    []byte
	RepoSize     uint64
	Cid          string
	SegmentIndex uint64
	Segment      [proof.SegmentSize]byte
	HashSet      []crypto.Hash
	AtHeight     uint64
	AccountName  string
}

type RscReq struct {
	Cid         string  `json:"cid"`
	Redundency  int     `json:"redundency"`
	IsDir       bool    `json:"dir"`
	Chunk       uint64  `json:"chunk"`
	FileSize    uint64  `json:"filesize"`
}


func int64ToBytes(n int64) []byte {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	binary.Write(writer, binary.BigEndian, &n)
	writer.Flush()
	return buf.Bytes()
}