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
	"encoding/binary"
	errIn "errors"
	"fmt"

	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/pb"
)

type ConType uint32

const (
	ConDBFT ConType = 0x01
	CondPos ConType = 0x02
	ConSolo ConType = 0x03
)

func (c ConType) String() string {
	switch c {
	case ConSolo:
		return "ConSolo"
	case ConDBFT:
		return "ConDBFT"
	case CondPos:
		return "CondPos"
	default:
		return "UnKnown"
	}
}

type ConsPayload interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	GetObject() interface{}
	Show()
}

type ConsData struct {
	Type    ConType
	Payload ConsPayload
}

func NewConsensusPayload(Type ConType, payload ConsPayload) *ConsData {
	return &ConsData{Type: Type, Payload: payload}
}

func InitConsensusData(timestamp int64) (*ConsData, error) {
	switch config.ConsensusAlgorithm {
	case "SOLO", "SHARD":
		conType := ConSolo
		conPayload := new(SoloData)
		return NewConsensusPayload(conType, conPayload), nil
	case "DPOS":
		conType := CondPos
		conPayload := GenesisStateInit(timestamp)
		return NewConsensusPayload(conType, conPayload), nil
	default:
		return nil, errors.New("unknown consensus type")
	}
}

func (c *ConsData) proto() (*pb.ConData, error) {
	data, err := c.Payload.Serialize()
	if err != nil {
		return nil, err
	}
	return &pb.ConData{
		Type: uint32(c.Type),
		Data: common.CopyBytes(data),
	}, nil
}

func (c *ConsData) Serialize() ([]byte, error) {
	data, err := c.Payload.Serialize()
	if err != nil {
		return nil, err
	}
	pbCon := pb.ConData{
		Type: uint32(c.Type),
		Data: common.CopyBytes(data),
	}
	b, err := pbCon.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *ConsData) Deserialize(data []byte) error {
	var pbCon pb.ConData
	if err := pbCon.Unmarshal(data); err != nil {
		return err
	}
	c.Type = ConType(pbCon.Type)
	switch c.Type {
	case CondPos:
		c.Payload = new(DPosData)
	case ConDBFT:
		c.Payload = new(DBFTData)
	case ConSolo:
		c.Payload = new(SoloData)
	default:
		return errors.New("unknown consensus type")
	}
	return c.Payload.Deserialize(pbCon.Data)
}

///////////////////////////////////////dPos/////////////////////////////////////////

const (
	Second             = int64(1000)
	BlockInterval      = int64(15000)
	GenerationInterval = GenerationSize * BlockInterval * 10
	GenerationSize     = 4
)

var (
	ErrNotBlockForgTime = errIn.New("current is not time to forge block")
	ErrFoundNilLeader   = errIn.New("found a nil leader")
)

/*
type ConsensusState interface {

	Timestamp() int64
	NextConsensusState(int64) (ConsensusState, error)
	Leader() common.Hash

	Bookkeepers() ([]common.Hash, error)

}*/

type DPosData struct {
	timestamp int64
	leader    common.Hash

	//TODO
	bookkeepers []common.Hash
}

func (d *DPosData) Timestamp() int64 {
	return d.timestamp
}

func (d *DPosData) Leader() common.Hash {
	return d.leader
}

func (d *DPosData) NextConsensusState(passedSecond int64) (*DPosData, error) {
	elapsedSecondInMs := passedSecond * Second
	if elapsedSecondInMs <= 0 || elapsedSecondInMs%BlockInterval != 0 {
		return nil, ErrNotBlockForgTime
	}
	//TODO
	bookkeepers := d.bookkeepers

	consensusState := &DPosData{
		timestamp:   d.timestamp + passedSecond,
		bookkeepers: bookkeepers,
	}

	log.Debug("consensusState, timestamp ", consensusState.timestamp)
	log.Debug(d.timestamp, passedSecond)
	currentInMs := consensusState.timestamp * Second
	offsetInMs := currentInMs % GenerationInterval
	log.Debug("timestamp %", offsetInMs, (offsetInMs*Second)%BlockInterval)
	var err error
	consensusState.leader, err = FindLeader(consensusState.timestamp, bookkeepers)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	return consensusState, nil
}

func (d *DPosData) Bookkeepers() ([]common.Hash, error) {
	return d.bookkeepers, nil
}

func FindLeader(current int64, bookkeepers []common.Hash) (leader common.Hash, err error) {
	currentInMs := current * Second
	offsetInMs := currentInMs % GenerationInterval
	log.Debug("offsetMs = ", offsetInMs)
	if offsetInMs%BlockInterval != 0 {
		log.Debug("In FindLeader, mod not 0")
		return common.NewHash(nil), ErrNotBlockForgTime
	}
	offset := offsetInMs / BlockInterval
	offset %= GenerationSize

	if offset >= 0 && int(offset) < len(bookkeepers) {
		log.Debug("offset = ", offset)
		leader = bookkeepers[offset]
	} else {
		log.Warn("Can't find Leader")
		return common.NewHash(nil), ErrFoundNilLeader
	}
	return leader, nil
}

func GenesisStateInit(timestamp int64) *DPosData {

	//TODO, bookkeepers
	bookkeepers := []common.Hash{}

	addr1 := common.Address{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 1, 2, 3, 4, 5, 6, 7}
	s1 := addr1.ToBase58()

	addr2 := common.Address{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 1, 2, 3, 4, 5, 6, 8}
	s2 := addr2.ToBase58()

	addr3 := common.Address{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 1, 2, 3, 4, 5, 6, 9}
	s3 := addr3.ToBase58()

	addr4 := common.Address{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 1, 2, 3, 4, 5, 6, 6}
	s4 := addr4.ToBase58()

	addresses := []string{}
	addresses = append(addresses, s1)
	addresses = append(addresses, s2)
	addresses = append(addresses, s3)
	addresses = append(addresses, s4)

	for _, v := range addresses {
		hash := common.NewHash(common.AddressFromBase58(v).Bytes())
		bookkeepers = append(bookkeepers, hash)
	}

	//TODO
	data := &DPosData{
		leader:      bookkeepers[0],
		timestamp:   timestamp,
		bookkeepers: bookkeepers,
	}
	return data
}

func (d *DPosData) protoBuf() (*pb.ConsensusState, error) {
	var bookkeepers []*pb.Miner
	for i := 0; i < len(d.bookkeepers); i++ {
		bookkeeper := &pb.Miner{
			Hash: d.bookkeepers[i].Bytes(),
		}
		bookkeepers = append(bookkeepers, bookkeeper)
	}
	consensusState := &pb.ConsensusState{
		Hash:        d.leader.Bytes(),
		Bookkeepers: bookkeepers,
		Timestamp:   d.timestamp,
	}
	return consensusState, nil
}

//TODO
func (d *DPosData) Serialize() ([]byte, error) {
	p, err := d.protoBuf()
	if err != nil {
		return nil, err
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

//TODO
func (d *DPosData) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var state pb.ConsensusState
	if err := state.Unmarshal(data); err != nil {
		return err
	}

	d.timestamp = state.Timestamp
	d.leader = common.NewHash(state.Hash)
	var keepers []common.Hash
	for i := 0; i < len(state.Bookkeepers); i++ {
		bookkeeper := state.Bookkeepers[i]
		keepers = append(keepers, common.NewHash(bookkeeper.Hash))
	}
	d.bookkeepers = keepers
	return nil
}

func (d DPosData) GetObject() interface{} {
	return d
}
func (d *DPosData) Show() {
	//fmt.Println("Proposer:", d.proposer)
}

/////////////////////////////////////////dBft///////////////////////////////////////
type DBFTData struct {
	data uint64
}

func (d *DBFTData) Serialize() ([]byte, error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, d.data)
	return b, nil
}
func (d *DBFTData) Deserialize(data []byte) error {
	d.data = binary.BigEndian.Uint64(data)
	return nil
}
func (d DBFTData) GetObject() interface{} {
	return d
}
func (d *DBFTData) Show() {
	fmt.Println("Data:", d.data)
}

///////////////////////////////////////////Solo/////////////////////////////////////
type SoloData struct{}

func (s *SoloData) Serialize() ([]byte, error) {
	return nil, nil
}
func (s *SoloData) Deserialize(data []byte) error {
	return nil
}
func (s SoloData) GetObject() interface{} {
	return s
}
func (s *SoloData) Show() {
	fmt.Println("Solo Module Data")
}
