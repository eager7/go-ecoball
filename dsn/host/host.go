package host

import (
	"time"
	"errors"
	"math/big"
	"github.com/ecoball/go-ecoball/dsn/crypto"
	"github.com/ecoball/go-ecoball/account"
	//"github.com/ecoball/go-ecoball/dsn/common/ecoding"
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
	"github.com/ecoball/go-ecoball/dsn/host/pb"
	"bytes"
	"encoding/binary"
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

type StorageHostConf struct {
	TotalStorage  uint64
	Collateral    string
	MaxCollateral string
	//AccountName   common.AccountName
	//ChainId       common.Hash
	AccountName   string
	ChainId       string
}

type HostAncContract struct {
	PublicKey     []byte
	TotalStorage  uint64
	StartAt       uint64
	Collateral    big.Int
	MaxCollateral big.Int
	AccountName   string
}

type StorageProof struct {
	PublicKey     []byte
	RepoSize      uint64
	Cid           string
	SegmentIndex  uint64
	Segment       [dproof.SegmentSize]byte
	HashSet       []crypto.Hash
	AtHeight      uint64
	AccountName   string
}


type StorageHoster interface {
	Announce() error
	TotalStorage() uint64
	ProvideStorageProof() error
}

type StorageHost struct {
	isBlockSynced     bool
	announced         bool
	announceConfirmed bool
	account           account.Account
	ledger            ledger.Ledger
	conf              StorageHostConf
	ctx               context.Context
}

func NewStorageHost(ctx context.Context, l ledger.Ledger,acc account.Account ,conf StorageHostConf) *StorageHost {
	return &StorageHost{
		account:    acc,
		ledger:     l,
		conf:       conf,
		ctx:        ctx,
	}
}

func (h *StorageHost) Start() error {
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

func (h *StorageHost)checkCollateral() bool {
	chainId := common.HexToHash(h.conf.ChainId)
	accName := common.NameToIndex(h.conf.AccountName)
	sacc, err :=h.ledger.AccountGet(chainId, accName)
	if err != nil {
		return false
	}
	//TODO much more checking
	if sacc.Votes.Staked > 0 {
		return true
	}
	return false
}

func (h *StorageHost) getBlockSyncState(chainId common.Hash) bool {
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
func (h *StorageHost) Announce() error {
	chainId := common.HexToHash(h.conf.ChainId)
	syncState := h.getBlockSyncState(chainId)
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
	transaction, err := types.NewInvokeContract(common.NameToIndex(h.conf.AccountName),
		innerCommon.NameToIndex("root"), chainId,
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

func (h *StorageHost) TotalStorage() uint64 {
	return h.conf.TotalStorage
}

// createAnnouncement will take a storage host announcement and encode it, returning the
// exact []byte that should be added to the arbitrary data of a transaction
func (h *StorageHost) createAnnouncement() ([]byte, error) {
	curBlockHeight := h.ledger.GetCurrentHeight(common.HexToHash(h.conf.ChainId))
	var afc HostAncContract
	afc.PublicKey = h.account.PublicKey
	afc.TotalStorage = h.conf.TotalStorage
	afc.StartAt = curBlockHeight
	var cv, bcv *big.Int
	cv, ok := cv.SetString(h.conf.Collateral, 10)
	if !ok {
		return nil, errors.New("conf err")
	}
	afc.Collateral = *cv
	bcv, ok = cv.SetString(h.conf.MaxCollateral, 10)
	if !ok {
		return nil, errors.New("conf err")
	}
	afc.MaxCollateral = *bcv
	afc.AccountName = h.conf.AccountName
	annBytes, err := afc.Serialize()
	if err != nil {
		return nil, err
	}

	annHash := crypto.HashBytes(annBytes)
	var sk crypto.SecretKey
	copy(sk[:], h.account.PrivateKey)
	sig := crypto.SignHash(annHash, sk)
	return append(annBytes, sig[:]...), nil
}
func (h *StorageHost)createStorageProof() ([]byte, error) {
	repoStat, err := corerepo.RepoStat(h.ctx, ipfsNode)
	if err != nil {
		return nil, err
	}
	var proof StorageProof
	proof.PublicKey = h.account.PublicKey
	proof.RepoSize = repoStat.RepoSize
	proof.AtHeight = h.ledger.GetCurrentHeight(common.HexToHash(h.conf.ChainId))
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
	proof.Cid = proofCid.String()
	blockData := block.RawData()
	dataSize := len(blockData)
	numberSegment := dataSize / dproof.SegmentSize
	segmentIndex := rand.Intn(int(numberSegment))
	base, cachedHashSet := dproof.MerkleProof(blockData, uint64(segmentIndex))
	proof.SegmentIndex = uint64(segmentIndex)
	proof.HashSet = cachedHashSet
	copy(proof.Segment[:], base)
	proof.AccountName = h.conf.AccountName
	proofBytes, err := proof.Serialize()
	if err != nil {
		return nil, err
	}
	proofHash := crypto.HashBytes(proofBytes)
	var sk crypto.SecretKey
	copy(sk[:], h.account.PrivateKey)
	sig := crypto.SignHash(proofHash, sk)
	return append(proofBytes, sig[:]...), nil
}

func (h *StorageHost) ProvideStorageProof() error {
	//syncState := h.getBlockSyncState(h.chainId)
	//if !syncState {
	//	return errGetBlockSyncState
	//}
	proof, err := h.createStorageProof()
	if err != nil {
		return errCreateStorageProof
	}
	timeNow := time.Now().Unix()
	transaction, err := types.NewInvokeContract(common.NameToIndex(h.conf.AccountName),
		innerCommon.NameToIndex("root"), common.HexToHash(h.conf.ChainId),
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

func (h *StorageHost) proofLoop() error {
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

func (an *HostAncContract) marshal() *pb.Announcement {
	bCol, _ := an.Collateral.GobEncode()
	bMcol, _ := an.MaxCollateral.GobEncode()
	return &pb.Announcement{
		PublicKey:an.PublicKey,
		StartAt:an.StartAt,
		Collateral:bCol,
		MaxCollateral:bMcol,
	}
}

func (an *HostAncContract) Serialize() ([]byte, error) {
	pm := an.marshal()
	return pm.Marshal()
}

func (an *HostAncContract) Deserialize(data []byte) error {
	//pm := new(pb.Announcement)
	var pm pb.Announcement
	err := pm.Unmarshal(data)
	if err != nil {
		return err
	}
	an.PublicKey = pm.PublicKey
	an.StartAt = pm.StartAt
	an.Collateral.GobDecode(pm.Collateral)
	an.MaxCollateral.GobDecode(pm.MaxCollateral)
	return nil
}

func (an *HostAncContract) SeriStateStore() []byte {
	b1 := make([]byte, 8)
	binary.BigEndian.PutUint64(b1, an.TotalStorage)
	b2 := make([]byte, 8)
	binary.BigEndian.PutUint64(b2, an.StartAt)
	b3 := an.Collateral.Bytes()
	b4 := an.MaxCollateral.Bytes()
	buff := make([]byte, 16 + len(b3) + len(b4))
	var offset int = 0
	copy(buff, b1)
	offset = offset + 8
	copy(buff[offset:], b2)
	offset = offset + 8
	copy(buff[offset:], b3)
	offset = offset + len(b3)
	copy(buff[offset:], b4)
	return buff
}
func (st *StorageProof) Serialize() ([]byte, error) {
	var sp pb.Proof
	sp.PublicKey = st.PublicKey
	sp.RepoSize = st.RepoSize
	sp.Cid = st.Cid
	sp.SegmentIndex = st.SegmentIndex
	copy(sp.Segment, st.Segment[:])
	for k, v:= range st.HashSet {
		copy(sp.HashSet[k * crypto.HashSize:], v[:])
	}
	sp.AtHeight = st.AtHeight
	return sp.Marshal()
}

func (st *StorageProof) Deserialize(data []byte) error {
	var sp pb.Proof
	err := sp.Unmarshal(data)
	if err != nil {
		return err
	}
	st.PublicKey = sp.PublicKey
	st.RepoSize = sp.RepoSize
	st.Cid = sp.Cid
	st.SegmentIndex = sp.SegmentIndex
	copy(st.Segment[:], sp.Segment)
	i := 0
	setCount := len(sp.HashSet) / crypto.HashSize
	for i < setCount {
		copy(st.HashSet[i][:], sp.HashSet[:(i + 1) * crypto.HashSize])
		i++
	}
	st.AtHeight = sp.AtHeight
	return nil
}

