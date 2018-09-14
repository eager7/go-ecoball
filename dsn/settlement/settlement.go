package settlement

import (
	"math/big"
	"fmt"
	"errors"
	dproof "github.com/ecoball/go-ecoball/dsn/proof"
	"bytes"
	"github.com/ecoball/go-ecoball/dsn/host"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/dsn/crypto"
	"github.com/go-redis/redis"
	"github.com/ecoball/go-ecoball/dsn/renter"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"context"
	"github.com/ecoball/go-ecoball/dsn/common"
	"github.com/ecoball/go-ecoball/core/state"
	ecommon "github.com/ecoball/go-ecoball/common"
	//hpb "github.com/ecoball/go-ecoball/dsn/host/pb"
	"github.com/ecoball/go-ecoball/dsn/ipfs/api"
	"io/ioutil"
)

var (
	errProofInvalid = errors.New("Storage proof is invalid")
)

const (
	STYPEANN uint8 = iota
	STYPEPROOF
	STYPEFILECONTRACT
) 

type Currency struct {
	i big.Int
}

type RenterFee struct {
	DownloadSpending Currency
	StorageSpending  Currency
	TotalCost        Currency
}


type DiskResource struct{
	TotalCapacity uint64
	UsedCapacity  uint64
}

type SettleMsg struct {
	MsgType uint8
	data []byte
}

type Settler struct {
	ledger    ledger.Ledger
	rClient   *redis.Client
	msgChan   chan SettleMsg
	ctx       context.Context
}

func NewStorageSettler(ctx context.Context, l ledger.Ledger) *Settler {
	return &Settler{
		ledger: l,
		rClient: common.InitRedis(common.DefaultRedisConf()),
		msgChan: make(chan SettleMsg, 4 * 1024),
		ctx:ctx,
	}
}

func (s *Settler) rxLoop() {
	for {
		select {
		case v := <-s.msgChan:
			switch v.MsgType {
			case STYPEANN:

			case STYPEPROOF:

			case STYPEFILECONTRACT:

			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Settler) Start() error {
	//s.rxLoop()
	return nil
}

func (s *Settler) payToHost(spf host.StorageProof) error {
	//TODO
	return nil
}

// decodeAnnouncement decodes announcement bytes into a host announcement
func (s *Settler) decodeAnnouncement(fullAnnouncement []byte) (host.HostAncContract, error) {
	var announcement host.HostAncContract
	dec := encoding.NewDecoder(bytes.NewReader(fullAnnouncement))
	err := dec.Decode(&announcement)
	if err != nil {
		return announcement, err
	}
	var sig crypto.Signature
	err = dec.Decode(&sig)
	if err != nil {
		return announcement, err
	}
	var pk crypto.PublicKey
	copy(pk[:], announcement.PublicKey)
	annHash := crypto.HashObject(announcement)
	err = crypto.VerifyHash(annHash, pk, sig)
	if err != nil {
		return announcement, err
	}
	return announcement, nil
}

func (s *Settler) storeStoragecap(contract host.HostAncContract) error {
	pk := string(contract.PublicKey)
	dbKey := fmt.Sprintf("host_%s", pk)
	r := s.rClient.HSet(dbKey, "total", contract.TotalStorage)
	r = s.rClient.Expire(dbKey, -1)
	return r.Err()
}

func (s *Settler) decodeProof(proof []byte) (host.StorageProof, error) {
	var sp host.StorageProof
	dec := encoding.NewDecoder(bytes.NewReader(proof))
	err := dec.Decode(&sp)
	if err != nil {
		return sp, err
	}
	// Read the signature out of the reader
	var sig crypto.Signature
	err = dec.Decode(&sig)
	if err != nil {
		return sp, err
	}
	var pk crypto.PublicKey
	copy(pk[:], sp.PublicKey)
	proofHash := crypto.HashObject(sp)
	err = crypto.VerifyHash(proofHash, pk, sig)
	if err != nil {
		return sp, err
	}
	return sp, nil
}

func (s *Settler) decodeFileContract(data []byte) (renter.FileContract, error) {
	var fc renter.FileContract
	dec := encoding.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&fc)
	if err != nil {
		return fc, err
	}
	var sig crypto.Signature
	err = dec.Decode(&sig)
	if err != nil {
		return fc, err
	}
	var pk crypto.PublicKey
	copy(pk[:], fc.PublicKey)
	proofHash := crypto.HashObject(fc)
	err = crypto.VerifyHash(proofHash, pk, sig)
	if err != nil {
		return fc, err
	}
	return fc, nil
}

func (s *Settler)storeReposize(proof host.StorageProof) error {
	pk := string(proof.PublicKey)
	dbKey := fmt.Sprintf("host_%s", pk)
	r := s.rClient.HSet(dbKey, "reposize", proof.RepoSize)
	r = s.rClient.Expire(dbKey, -1)
	return r.Err()
}

func (s *Settler)storeOnlineTime(proof host.StorageProof) error {
	pk := string(proof.PublicKey)
	dbKey := fmt.Sprintf("online_%s", pk)
	r := s.rClient.RPush(dbKey, proof.AtHeight)
	s.rClient.Expire(dbKey, -1)
	return r.Err()
}

func (s *Settler)verifyStorageProof(data []byte, st state.InterfaceState) (bool, error) {
	proof, err := s.decodeProof(data)
	if err != nil {
		return false, err
	}
	block, err := api.IpfsBlockGet(s.ctx, proof.Cid)
	if err != nil {
		return false, err
	}
	blockData, err := ioutil.ReadAll(block)
	if err != nil {
		return false, err
	}
	rootHash := dproof.MerkleRoot(blockData)
	numberSegment := len(blockData) / dproof.SegmentSize
	ret := dproof.VerifySegment(proof.Segment[:], proof.HashSet, uint64(numberSegment), proof.SegmentIndex, rootHash)
	s.storeAccountState(proof, st)
	if ret {
		s.storeReposize(proof)
	}
	return ret, nil
}

func (s *Settler) storeAccountState(data interface{}, st state.InterfaceState) error {
	var err error
	switch data.(type) {
	case *host.HostAncContract:
		sKey := []byte("store_an")
		value := data.(*host.HostAncContract).SeriStateStore()
		err = st.StoreSet(ecommon.NameToIndex(data.(*host.HostAncContract).AccountName), sKey, value)
	case *host.StorageProof:

	case *renter.FileContract:

	default:

	}
	return err
}

func (s *Settler)HandleHostAnce(data []byte, st state.InterfaceState) error {
	c, err := s.decodeAnnouncement(data)
	if err != nil {
		return err
	}
	err = s.storeAccountState(&c,st)
	if err != nil {
		return err
	}
	return s.storeStoragecap(c)
}

func (s *Settler)HandleStorageProof(data []byte, st state.InterfaceState) error {
	valid, err := s.verifyStorageProof(data, st)
	if err != nil {
		return err
	}
	if !valid {
		return errProofInvalid
	}
	return nil
}

func (s *Settler) storeFileContractInfo(fc renter.FileContract) error {
	pk := string(fc.PublicKey)
	hKey := fmt.Sprintf("renter_%s", pk)
	s.rClient.HIncrBy(hKey, "used", int64(fc.FileSize))
	s.rClient.Expire(hKey, -1)
	cKey := fmt.Sprintf("size_%s", fc.Cid)
	s.rClient.HSet(hKey, cKey, fc.FileSize)
	fKey := fmt.Sprintf("files_%s", pk)
	s.rClient.RPush(fKey, fc.Cid)
	s.rClient.Expire(fKey, -1)
	return nil
}

func (s *Settler) HandleFileContract(data []byte, st state.InterfaceState) error {
	fc, err := s.decodeFileContract(data)
	if err != nil {
		return err
	}
	s.storeAccountState(fc, st)
	s.storeFileContractInfo(fc)
	return nil
}


