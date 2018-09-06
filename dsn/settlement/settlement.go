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
	"github.com/ipfs/go-ipfs/core"
	"github.com/go-redis/redis"
	"github.com/ecoball/go-ecoball/dsn/renter"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"context"
)

var (
	ipfsNode *core.IpfsNode
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
	rClient  *redis.Client
	msgChan  chan SettleMsg
	ctx      context.Context
}

func NewStorageSettler(ctx context.Context, l ledger.Ledger) *Settler {
	return &Settler{
		ledger: l,
		rClient: InitRedis(DefaultRedisConf()),
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
				err := s.handleHostAnce(v.data)
				if err != nil {
					fmt.Errorf("settle ", err)
				}
			case STYPEPROOF:
				err := s.handleStorageProof(v.data)
				if err != nil {
					fmt.Errorf("settle ", err)
				}
			case STYPEFILECONTRACT:
				err := s.handleFileContract(v.data)
				if err != nil {
					fmt.Errorf("settle ", err)
				}
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Settler) Start() error {
	s.rxLoop()
	return nil
}

func (s *Settler) payToHost(spf host.StorageProof) error {
	//TODO
	return nil
}

// decodeAnnouncement decodes announcement bytes into a host announcement
func (s *Settler) decodeAnnouncement(fullAnnouncement []byte) (contract host.HostAncContract, err error) {
	var announcement host.HostAncContract
	dec := encoding.NewDecoder(bytes.NewReader(fullAnnouncement))
	err = dec.Decode(&announcement)
	if err != nil {
		return announcement, err
	}

	// Read the signature out of the reader
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

func (s *Settler)verifyStorageProof(data []byte) (bool, error) {
	proof, err := s.decodeProof(data)
	if err != nil {
		return false, err
	}
	baseBlockService := ipfsNode.BaseBlocks
	block, err := baseBlockService.Get(&proof.Cid)
	if err != nil {
		return false, err
	}
	blockData := block.RawData()
	rootHash := dproof.MerkleRoot(blockData)
	numberSegment := len(blockData) / dproof.SegmentSize
	ret := dproof.VerifySegment(proof.Segment[:], proof.HashSet, uint64(numberSegment), proof.SegmentIndex, rootHash)
	if ret {
		s.storeReposize(proof)
	}
	return ret, nil
}

func (s *Settler)handleHostAnce(data []byte) error {
	c, err := s.decodeAnnouncement(data)
	if err != nil {
		return err
	}
	return s.storeStoragecap(c)
}

func (s *Settler)handleStorageProof(data []byte) error {
	valid, err := s.verifyStorageProof(data)
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

func (s *Settler) handleFileContract(data []byte) error {
	fc, err := s.decodeFileContract(data)
	if err != nil {
		return err
	}
	s.storeFileContractInfo(fc)
	return nil
}

func SetIpfsNode(node *core.IpfsNode)  {
	ipfsNode = node
}


