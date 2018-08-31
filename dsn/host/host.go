package host

import (
	"time"
	"errors"
	"math/big"
	"bytes"
	"github.com/ecoball/go-ecoball/dsn/crypto"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/types"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/store"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	"fmt"
	"bufio"
	"encoding/binary"
	"github.com/ipfs/go-ipfs/core/corerepo"
	"context"
	"github.com/ipfs/go-ipfs/core"
	"math/rand"
	//chunker "gx/ipfs/QmVDjhUMtkRskBFAVNwyXuLSKbeAya7JKPnzAxMKDaK4x4/go-ipfs-chunker"
	//dcommon "github.com/ecoball/go-ecoball/dsn/common"
	dproof "github.com/ecoball/go-ecoball/dsn/proof"
)

var (
	dbPath string = "/tmp/store/leveldb"
	contractDesc string = "storage host"
	errGetBlockSyncState = errors.New("failed to get block sync state")
	errCreateAnnouncement = errors.New("failed to create announcement")
	errCreateStorageProof = errors.New("failed to create announcement")
	ipfsNode *core.IpfsNode
)

type HostAncContract struct {
//	PublicKey    crypto.PublicKey
	PublicKey    []byte
	TotalStorage uint64
//	StartAt      proof.BlockHeight
	StartAt      uint64
//	WindowSize   proof.BlockHeight
	Collateral    big.Int
	MaxCollateral big.Int
//	RevisionNumber uint64
//	Version        string
}

type StorageHostConf struct {
	TotalStorage  uint64
	Collateral    big.Int
	MaxCollateral big.Int
	AccountName   string
}

type StorageProof struct {
	PublicKey     []byte
	RepoSize      uint64
	Cid           cid.Cid
	SegmentIndex  uint64
	Segment       [dproof.SegmentSize]byte
	HashSet       []crypto.Hash
	AtHeight      uint64
}


type StorageHost interface {
	Announce() error
	TotalStorage() uint64
	ProvideStorageProof() error
}

type storageHost struct {
	isBlockSynced     bool
	announced         bool
	announceConfirmed bool
	totalStorage      uint64
	account           account.Account
	collateral    	  big.Int
	maxCollateral     big.Int
	ledger            ledger.Ledger
	chainId           common.Hash
	accountName       string
	db                store.Storage

	ctx               context.Context
}

func NewStorageHost(l ledger.Ledger, acc account.Account, conf StorageHostConf) StorageHost {
	return &storageHost{
		account:       acc,
		totalStorage:  conf.TotalStorage,
		collateral:    conf.Collateral,
		maxCollateral: conf.MaxCollateral,
		ledger:        l,
		chainId:       config.ChainHash,
		accountName:   conf.AccountName,
		ctx:           context.Background(),
	}
}

func SetIpfsNode(node *core.IpfsNode)  {
	ipfsNode = node
}

func (h *storageHost) Start() error {
	db, err := store.NewBlockStore(dbPath)
	if err != nil {
		return err
	}
	h.db = db
	return nil
}

func (h *storageHost) getBlockSyncState(chainId common.Hash) bool {
	go func() {
		timerChan := time.NewTicker(10 * time.Second).C
		//var syncState bool
		for {
			select {
			case timerChan:
				//TODO get current block synced state
			}
		}
	}()
	return false
}

// Announce creates a storage host announcement transaction
func (h *storageHost) Announce() error {
	syncState := h.getBlockSyncState(h.chainId)
	if !syncState {
		return errGetBlockSyncState
	}
	announcement, err := h.createAnnouncement()
	if err != nil {
		return errCreateAnnouncement
	}
	timeNow := time.Now().Unix()
	transaction, err := types.NewInvokeContract(innerCommon.NameToIndex(h.accountName),
		innerCommon.NameToIndex("root"), h.chainId,
		"owner", "reg_store", []string{string(announcement)}, 0, timeNow)
	if err != nil {
		return err
	}
	err = event.Send(event.ActorNil, event.ActorTxPool, transaction)
	if err != nil {
		return err
	}
	return nil
}

func (h *storageHost) TotalStorage() uint64 {
	return h.totalStorage
}

// createAnnouncement will take a storage host announcement and encode it, returning the
// exact []byte that should be added to the arbitrary data of a transaction
func (h *storageHost) createAnnouncement() (signedAnnounce []byte, err error) {
	curBlockHeight := h.ledger.GetCurrentHeight(h.chainId)
	annBytes := encoding.Marshal(HostAncContract{
		PublicKey:    h.account.PublicKey,
		TotalStorage: h.totalStorage,
		StartAt:      curBlockHeight,
		Collateral:   h.collateral,
		MaxCollateral:h.maxCollateral,
	})

	// Create a signature for the announcement
	annHash := crypto.HashBytes(annBytes)
	var sk crypto.SecretKey
	copy(sk[:], h.account.PrivateKey)
	sig := crypto.SignHash(annHash, sk)
	return append(annBytes, sig[:]...), nil
}

// decodeAnnouncement decodes announcement bytes into a host announcement
func (h *storageHost) DecodeAnnouncement(fullAnnouncement []byte) (contract HostAncContract, err error) {
	var announcement HostAncContract
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

func (h *storageHost) storeStoragecap(contract HostAncContract) error {
	pk := string(contract.PublicKey)
	dbKey := fmt.Sprintf("storage_total_%s", pk)
	value := int64ToBytes(int64(contract.TotalStorage))
	return h.db.Put([]byte(dbKey), value)
}

func int64ToBytes(n int64) []byte {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	binary.Write(writer, binary.BigEndian, &n)
	writer.Flush()
	return buf.Bytes()
}

func (h *storageHost)createStorageProof() ([]byte, error) {
	repoStat, err := corerepo.RepoStat(h.ctx, ipfsNode)
	if err != nil {
		return nil, err
	}
	var proof StorageProof
	proof.PublicKey = h.account.PublicKey
	proof.RepoSize = repoStat.RepoSize
	baseBlockService := ipfsNode.BaseBlocks
	allCids, err := baseBlockService.AllKeysChan(h.ctx)
	j := rand.Intn(int(repoStat.NumObjects))
	cnt := 0
	var proofCid *cid.Cid
	for cid := range allCids {
		if cnt == j {
			proofCid = cid
		}
		cnt++
	}
	block, err := baseBlockService.Get(proofCid)
	if err != nil {
		return nil, err
	}
	proof.Cid = *proofCid
	blockData := block.RawData()
	dataSize := len(blockData)
	numberSegment := dataSize / dproof.SegmentSize
	segmentIndex := rand.Intn(int(numberSegment))
	base, cachedHashSet := dproof.MerkleProof(blockData, uint64(segmentIndex))
	proof.SegmentIndex = uint64(segmentIndex)
	proof.HashSet = cachedHashSet
	copy(proof.Segment[:], base)
	proofBytes := encoding.Marshal(proof)
	// Create a signature for the announcement
	proofHash := crypto.HashBytes(proofBytes)
	var sk crypto.SecretKey
	copy(sk[:], h.account.PrivateKey)
	sig := crypto.SignHash(proofHash, sk)
	return append(proofBytes, sig[:]...), nil
}

func (h *storageHost) decodeProof(proof []byte) (StorageProof, error) {
	var sp StorageProof
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

func (h *storageHost)storeReposize(proof StorageProof) error {
	pk := string(proof.PublicKey)
	dbKey := fmt.Sprintf("storage_used_%s", pk)
	value := int64ToBytes(int64(proof.RepoSize))
	return h.db.Put([]byte(dbKey), value)
}

func (h *storageHost)VerifyStorageProof(data []byte) (bool, error) {
	proof, err := h.decodeProof(data)
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
	return ret, nil
}

func (h *storageHost) ProvideStorageProof() error {
	syncState := h.getBlockSyncState(h.chainId)
	if !syncState {
		return errGetBlockSyncState
	}
	proof, err := h.createStorageProof()
	if err != nil {
		return errCreateStorageProof
	}
	timeNow := time.Now().Unix()
	transaction, err := types.NewInvokeContract(innerCommon.NameToIndex(h.accountName),
		innerCommon.NameToIndex("root"), h.chainId,
		"owner", "reg_proof", []string{string(proof)}, 0, timeNow)
	if err != nil {
		return err
	}
	err = event.Send(event.ActorNil, event.ActorTxPool, transaction)
	if err != nil {
		return err
	}
	return nil
}
