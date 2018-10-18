package host

import (
	"time"
	"errors"
	"math/big"
	"context"
	"math/rand"
	"io/ioutil"
	"crypto/sha256"
	"github.com/ecoball/go-ecoball/dsn/crypto"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
	dproof "github.com/ecoball/go-ecoball/dsn/proof"
	"github.com/ecoball/go-ecoball/dsn/host/pb"
	"github.com/ecoball/go-ecoball/dsn/ipfs/api"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	dsnComm "github.com/ecoball/go-ecoball/dsn/common"
	"github.com/ecoball/go-ecoball/dsn/ipfs"
)

var (
	dbPath string = "/tmp/store/leveldb"
	contractDesc string = "storage host"
	errGetBlockSyncState = errors.New("failed to get block sync state")
	errCreateAnnouncement = errors.New("failed to create announcement")
	errCreateStorageProof = errors.New("failed to create proof")
	errCheckCol = errors.New("Checking collateral failed")
	log = elog.NewLogger("dsn-h", elog.DebugLog)
)

type StorageHostConf struct {
	TotalStorage  uint64
	Collateral    string
	MaxCollateral string
	AccountName   string
	ChainId       string
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
	PublicKey     []byte
	RepoSize      uint64
	Cid           string
	SegmentIndex  uint64
	Segment       [dproof.SegmentSize]byte
	HashSet       []crypto.Hash
	AtHeight      uint64
	AccountName   string
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

func InitDefaultConf() StorageHostConf {
	chainId := config.ChainHash
	return StorageHostConf{
		TotalStorage: 10*1024*1024,
		Collateral: "10000",
		MaxCollateral: "20000",
		AccountName: "root",
		ChainId: common.ToHex(chainId[:]),
	}
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
	if err := ipfs.Initialize(); err != nil {
		return err
	}
	if err := ipfs.DaemonRun(); err != nil {
		return err
	}
	if err := h.Announce(); err != nil {
		return err
	}

	if err := h.proofLoop(); err != nil {
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
		case <-timerChan:
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

	//TODO do block syncing and coll checking
	//syncState := h.getBlockSyncState(chainId)
	//if !syncState {
	//	return errGetBlockSyncState
	//}
	//colState := h.checkCollateral()
	//if !colState {
	//	return errCheckCol
	//}

	announcement, err := h.createAnnouncement()
	if err != nil {
		return errCreateAnnouncement
	}
	timeNow := time.Now().UnixNano()
	transaction, err := types.NewInvokeContract(common.NameToIndex(h.conf.AccountName),
		innerCommon.NameToIndex(dsnComm.RootAccount), chainId,
		"owner", dsnComm.FcMethodAn, []string{string(announcement)}, 0, timeNow)
	if err != nil {
		return err
	}
	transaction.SetSignature(&config.Root)
	err = event.Send(event.ActorNil, event.ActorTxPool, transaction)
	if err != nil {
		return err
	}
	log.Debug("Invoke host announcement")
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
	cv := new(big.Int)
	cv, ok := cv.SetString(h.conf.Collateral, 10)
	if !ok {
		return nil, errors.New("conf err")
	}
	afc.Collateral, _ = cv.GobEncode()
	bcv := new(big.Int)
	bcv, ok = cv.SetString(h.conf.MaxCollateral, 10)
	if !ok {
		return nil, errors.New("conf err")
	}
	afc.MaxCollateral, _ = bcv.GobEncode()
	afc.AccountName = h.conf.AccountName
	annBytes := encoding.Marshal(afc)
	annHash := sha256.Sum256(annBytes)
	sig, err := h.account.Sign(annHash[:])
	if err !=  nil {
		return nil, err
	}
	return append(annBytes, sig[:]...), nil
}

func (h *StorageHost)createStorageProof() ([]byte, error) {
	repoStat, err := api.IpfsRepoStat(h.ctx)
	if err != nil {
		return nil, err
	}
	var proof StorageProof
	proof.PublicKey = h.account.PublicKey
	proof.RepoSize = repoStat.RepoSize
	proof.AtHeight = h.ledger.GetCurrentHeight(common.HexToHash(h.conf.ChainId))
	allCids, err := api.IpfsBlockAllKey(h.ctx)
	if err != nil {
		return nil, err
	}
	j := rand.Intn(int(repoStat.NumObjects))
	var proofCid string
	for k, cid := range allCids {
		if k == j {
			proofCid = cid
		}
	}
	block, err := api.IpfsBlockGet(h.ctx, proofCid)
	if err != nil {
		return nil, err
	}
	proof.Cid = proofCid
	blockData, err := ioutil.ReadAll(block)
	if err != nil {
		return nil, err
	}
	dataSize := len(blockData)
	numberSegment := dataSize / dproof.SegmentSize
	segmentIndex := rand.Intn(int(numberSegment))
	base, cachedHashSet := dproof.MerkleProof(blockData, uint64(segmentIndex))
	proof.SegmentIndex = uint64(segmentIndex)
	proof.HashSet = cachedHashSet
	copy(proof.Segment[:], base)
	proof.AccountName = h.conf.AccountName
	proofBytes := encoding.Marshal(proof)
	annHash := sha256.Sum256(proofBytes)
	sig, err := h.account.Sign(annHash[:])
	if err !=  nil {
		return nil, err
	}
	return append(proofBytes, sig[:]...), nil
}

func (h *StorageHost) ProvideStorageProof() error {
	//TODO
	//syncState := h.getBlockSyncState(h.chainId)
	//if !syncState {
	//	return errGetBlockSyncState
	//}

	proof, err := h.createStorageProof()
	if err != nil {
		return errCreateStorageProof
	}
	timeNow := time.Now().UnixNano()
	transaction, err := types.NewInvokeContract(common.NameToIndex(h.conf.AccountName),
		innerCommon.NameToIndex(dsnComm.RootAccount), common.HexToHash(h.conf.ChainId),
		"owner", dsnComm.FcMethodProof, []string{string(proof)}, 0, timeNow)
	if err != nil {
		return err
	}
	transaction.SetSignature(&config.Root)
	err = event.Send(event.ActorNil, event.ActorTxPool, transaction)
	if err != nil {
		return err
	}
	return nil
}

func (h *StorageHost) proofLoop() error {
	//TODO period: move to config
	timerChan := time.NewTicker(24 * time.Hour).C
	for {
		select {
		case <-timerChan:
			err := h.ProvideStorageProof()
			if err != nil {
				return err
			}
		case <-h.ctx.Done():
			return h.ctx.Err()
		}
	}
}

func (an *HostAncContract) marshal() *pb.Announcement {
	return &pb.Announcement{
		PublicKey:an.PublicKey,
		StartAt:an.StartAt,
		Collateral:an.Collateral,
		MaxCollateral:an.MaxCollateral,
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
	an.Collateral = pm.Collateral
	an.MaxCollateral = pm.MaxCollateral
	return nil
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

