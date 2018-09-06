package host

import (
	"time"
	"errors"
	"math/big"
	"github.com/ecoball/go-ecoball/dsn/crypto"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	"github.com/ipfs/go-ipfs/core/corerepo"
	"context"
	"math/rand"
	dproof "github.com/ecoball/go-ecoball/dsn/proof"
	"github.com/ipfs/go-ipfs/core"
)

var (
	dbPath string = "/tmp/store/leveldb"
	contractDesc string = "storage host"
	errGetBlockSyncState = errors.New("failed to get block sync state")
	errCreateAnnouncement = errors.New("failed to create announcement")
	errCreateStorageProof = errors.New("failed to create proof")
	errCheckCol = errors.New("Checking collateral failed")
	ipfsNode *core.IpfsNode
)

type HostAncContract struct {
	PublicKey     []byte
	TotalStorage  uint64
	StartAt       uint64
	Collateral    big.Int
	MaxCollateral big.Int
}

type StorageHostConf struct {
	TotalStorage  uint64
	Collateral    big.Int
	MaxCollateral big.Int
	AccountName   common.AccountName
	ChainId       common.Hash
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
	account           account.Account
	ledger            ledger.Ledger
	conf              StorageHostConf
	ctx               context.Context
}

func NewStorageHost(ctx context.Context, l ledger.Ledger,acc account.Account ,conf StorageHostConf) StorageHost {
	return &storageHost{
		account:    acc,
		ledger:     l,
		conf:       conf,
		ctx:        ctx,
	}
}

func (h *storageHost) Start() error {
	//db, err := store.NewBlockStore(dbPath)
	//if err != nil {
	//	return err
	//}
	//h.db = db
	err := h.Announce()
	if err != nil {
		return err
	}
	err = h.proofLoop()
	if err != nil {
		return err
	}
	return nil
}

func (h *storageHost)checkCollateral() bool {
	sacc, err :=h.ledger.AccountGet(h.conf.ChainId, h.conf.AccountName)
	if err != nil {
		return false
	}
	//TODO much more checking
	if sacc.Votes.Staked > 0 {
		return true
	}
	return false
}

func (h *storageHost) getBlockSyncState(chainId common.Hash) bool {
	timerChan := time.NewTicker(10 * time.Second).C
	var isSynced bool
	for {
		select {
		case timerChan:
			//TODO get current block synced state
			if isSynced {
				return true
			}
		case <-h.ctx.Done():
			return false
		}
	}
	return false
}

// Announce creates a storage host announcement transaction
func (h *storageHost) Announce() error {
	syncState := h.getBlockSyncState(h.conf.ChainId)
	if !syncState {
		return errGetBlockSyncState
	}
	colState := h.checkCollateral()
	if !colState {
		return errCheckCol
	}
	announcement, err := h.createAnnouncement()
	if err != nil {
		return errCreateAnnouncement
	}
	timeNow := time.Now().Unix()
	transaction, err := types.NewInvokeContract(h.conf.AccountName,
		innerCommon.NameToIndex("root"), h.conf.ChainId,
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
	return h.conf.TotalStorage
}

// createAnnouncement will take a storage host announcement and encode it, returning the
// exact []byte that should be added to the arbitrary data of a transaction
func (h *storageHost) createAnnouncement() (signedAnnounce []byte, err error) {
	curBlockHeight := h.ledger.GetCurrentHeight(h.conf.ChainId)
	annBytes := encoding.Marshal(HostAncContract{
		PublicKey:    h.account.PublicKey,
		TotalStorage: h.conf.TotalStorage,
		StartAt:      curBlockHeight,
		Collateral:   h.conf.Collateral,
		MaxCollateral:h.conf.MaxCollateral,
	})

	// Create a signature for the announcement
	annHash := crypto.HashBytes(annBytes)
	var sk crypto.SecretKey
	copy(sk[:], h.account.PrivateKey)
	sig := crypto.SignHash(annHash, sk)
	return append(annBytes, sig[:]...), nil
}
func (h *storageHost)createStorageProof() ([]byte, error) {
	repoStat, err := corerepo.RepoStat(h.ctx, ipfsNode)
	if err != nil {
		return nil, err
	}
	var proof StorageProof
	proof.PublicKey = h.account.PublicKey
	proof.RepoSize = repoStat.RepoSize
	proof.AtHeight = h.ledger.GetCurrentHeight(h.conf.ChainId)
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

func (h *storageHost) ProvideStorageProof() error {
	//syncState := h.getBlockSyncState(h.chainId)
	//if !syncState {
	//	return errGetBlockSyncState
	//}
	proof, err := h.createStorageProof()
	if err != nil {
		return errCreateStorageProof
	}
	timeNow := time.Now().Unix()
	transaction, err := types.NewInvokeContract(h.conf.AccountName,
		innerCommon.NameToIndex("root"), h.conf.ChainId,
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

func (h *storageHost) proofLoop() error {
	timerChan := time.NewTicker(24 * time.Hour).C
	for {
		select {
		case timerChan:
			err := h.ProvideStorageProof()
			if err != nil {
				return err
			}
		case <-h.ctx.Done():
			return h.ctx.Err()
		}
	}
}

func SetIpfsNode(node *core.IpfsNode)  {
	ipfsNode = node
}