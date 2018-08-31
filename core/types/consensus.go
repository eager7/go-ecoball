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
	errIn "errors"
	"encoding/binary"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/pb"
	"sort"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common/errors"
)

type ConType uint32

const (
	ConDBFT ConType = 0x01
	CondPos ConType = 0x02
	ConSolo ConType = 0x03
	ConABFT ConType = 0x04
)

func (c ConType) String() string {
	switch c {
	case ConSolo:
		return "ConSolo"
	case ConDBFT:
		return "ConDBFT"
	case CondPos:
		return "CondPos"
	case ConABFT:
		return "ConABFT"
	default:
		return "UnKnown"
	}
}

type ConsensusPayload interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	GetObject() interface{}
	Show()
}

type ConsensusData struct {
	Type    ConType
	Payload ConsensusPayload
}

func NewConsensusPayload(Type ConType, payload ConsensusPayload) *ConsensusData {
	return &ConsensusData{Type: Type, Payload: payload}
}

func InitConsensusData(timestamp int64) (*ConsensusData, error) {

	switch config.ConsensusAlgorithm {
	case "SOLO":
		conType := ConSolo
		conPayload := new(SoloData)
		return NewConsensusPayload(conType, conPayload), nil
	case "DPOS":
		conType := CondPos
		conPayload := GenesisStateInit(timestamp)
		return NewConsensusPayload(conType, conPayload), nil
	case "ABABFT":
		conType := ConABFT
		conPayload := GenesisABABFTInit(timestamp)
		return NewConsensusPayload(conType, conPayload), nil
		//TODO
	default:
		return nil, errors.New(log, "unknown consensus type")
	}
}

func (c *ConsensusData) ProtoBuf() (*pb.ConsensusData, error) {
	data, err := c.Payload.Serialize()
	if err != nil {
		return nil, err
	}
	return &pb.ConsensusData{
		Type: uint32(c.Type),
		Data: common.CopyBytes(data),
	}, nil
}

func (c *ConsensusData) Serialize() ([]byte, error) {
	data, err := c.Payload.Serialize()
	if err != nil {
		return nil, err
	}
	pbCon := pb.ConsensusData{
		Type: uint32(c.Type),
		Data: common.CopyBytes(data),
	}
	b, err := pbCon.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *ConsensusData) Deserialize(data []byte) error {
	var pbCon pb.ConsensusData
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
	case ConABFT:
		c.Payload = new(AbaBftData)
	default:
		return errors.New(log, "unknown consensus type")
	}
	return c.Payload.Deserialize(pbCon.Data)
}

///////////////////////////////////////dPos/////////////////////////////////////////

const (
	Second             = int64(1000)
	BlockInterval      = int64(15000)
	GenerationInterval = GenerationSize * BlockInterval * 10
	GenerationSize     = 4
	ConsensusThreshold = GenerationSize*2/3 + 1
	MaxProduceDuration = int64(5250)
	MinProduceDuration = int64(2250)
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

func (ds *DPosData) Timestamp() int64 {
	return ds.timestamp
}

func (ds *DPosData) Leader() common.Hash {
	return ds.leader
}

func (ds *DPosData) NextConsensusState(passedSecond int64) (*DPosData, error) {
	elapsedSecondInMs := passedSecond * Second
	if elapsedSecondInMs <= 0 || elapsedSecondInMs%BlockInterval != 0 {
		return nil, ErrNotBlockForgTime
	}
	//TODO
	bookkeepers := ds.bookkeepers

	consensusState := &DPosData{
		timestamp:   ds.timestamp + passedSecond,
		bookkeepers: bookkeepers,
	}

	log.Debug("consensusState, timestamp ", consensusState.timestamp)
	log.Debug(ds.timestamp, passedSecond)
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

func (dposData *DPosData) Bookkeepers() ([]common.Hash, error) {
	return dposData.bookkeepers, nil
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

func (data *DPosData) protoBuf() (*pb.ConsensusState, error) {
	var bookkeepers []*pb.Miner
	for i := 0; i < len(data.bookkeepers); i++ {
		bookkeeper := &pb.Miner{
			Hash: data.bookkeepers[i].Bytes(),
		}
		bookkeepers = append(bookkeepers, bookkeeper)
	}
	consensusState := &pb.ConsensusState{
		data.leader.Bytes(),
		bookkeepers,
		data.timestamp,
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
		return errors.New(log, "input data's length is zero")
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

///////////////////////////////////////////aBft/////////////////////////////////////
type AbaBftData struct {
	NumberRound        uint32
	PreBlockSignatures []common.Signature
}

func (a *AbaBftData) Serialize() ([]byte, error) {
	var sig []*pb.Signature
	for i := 0; i < len(a.PreBlockSignatures); i++ {
		s := &pb.Signature{PubKey: a.PreBlockSignatures[i].PubKey, SigData: a.PreBlockSignatures[i].SigData}
		sig = append(sig, s)
	}
	pbData := pb.AbaBftData{
		NumberRound: a.NumberRound,
		Sign:        sig,
	}
	data, err := pbData.Marshal()
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (a *AbaBftData) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New(log, fmt.Sprintf("AbaBftData is nil:%v", data))
	}
	var pbData pb.AbaBftData
	if err := pbData.Unmarshal(data); err != nil {
		return err
	}
	a.NumberRound = pbData.NumberRound
	for i := 0; i < len(pbData.Sign); i++ {
		sig := common.Signature{
			PubKey:  common.CopyBytes(pbData.Sign[i].PubKey),
			SigData: common.CopyBytes(pbData.Sign[i].SigData),
		}
		a.PreBlockSignatures = append(a.PreBlockSignatures, sig)
	}
	return nil
}
func (a AbaBftData) GetObject() interface{} {
	return a
}
func (a *AbaBftData) Show() {
	fmt.Println("\t-----------AbaBft------------")
	fmt.Println("\tNumberRound    :", a.NumberRound)
	fmt.Println("\tSig Len        :", len(a.PreBlockSignatures))
	for i := 0; i < len(a.PreBlockSignatures); i++ {
		fmt.Println("\tPublicKey      :", common.ToHex(a.PreBlockSignatures[i].PubKey))
		fmt.Println("\tSigData        :", common.ToHex(a.PreBlockSignatures[i].SigData))
	}
}

func GenesisABABFTInit(timestamp int64)  *AbaBftData{
	// array the peers list
	/*
	Num_peers_t := len(ababft.Peers_list)
	var Peers_list_t []string
	for i := 0; i < Num_peers_t; i++ {
		Peers_list_t[i] = string(ababft.Peers_list[i].PublicKey)
	}
	// sort the peers as list
	sort.Strings(Peers_list_t)
	for i := 0; i < Num_peers_t; i++ {
		ababft.Peers_list[i].PublicKey = []byte(Peers_list_t[i])
		ababft.Peers_list[i].Index = i
	}
	log.Debug("generate the geneses")
	var sigs []common.Signature
	for i := 0; i < Num_peers_t; i++ {
		sigs = append(sigs,common.Signature{ababft.Peers_list[i].PublicKey, []byte("hello,ababft")})
	}
	*/
	var Num_peers_t int
	Num_peers_t = 3
	var Peers_list_t []string
	Peers_list_t = append(Peers_list_t,string(config.Worker1.PublicKey))
	Peers_list_t = append(Peers_list_t,string(config.Worker2.PublicKey))
	Peers_list_t = append(Peers_list_t,string(config.Worker3.PublicKey))

	sort.Strings(Peers_list_t)
	var Peers_list []account.Account
	for i := 0; i < Num_peers_t; i++ {
		var peer account.Account
		peer.PublicKey = []byte(Peers_list_t[i])
		Peers_list = append(Peers_list,peer)
	}
	var sigs []common.Signature
	for i := 0; i < Num_peers_t; i++ {
		sigs = append(sigs,common.Signature{Peers_list[i].PublicKey, []byte("hello,ababft")})
	}
	abaData := AbaBftData{0,sigs}
	return &abaData
}
