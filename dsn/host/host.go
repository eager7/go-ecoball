package host

import (
	"context"
	"errors"
	"io/ioutil"
	"math/big"
	"math/rand"
	"time"
	"github.com/ecoball/go-ecoball/common"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	dsnComm "github.com/ecoball/go-ecoball/dsn/common"
	"github.com/ecoball/go-ecoball/dsn/host/ipfs"
	"github.com/ecoball/go-ecoball/dsn/host/ipfs/api"
	dproof "github.com/ecoball/go-ecoball/dsn/host/proof"
	wc "github.com/ecoball/go-ecoball/client/walletclient"
)

var (
	dbPath                string = "/tmp/store/leveldb"
	contractDesc          string = "storage host"
	errGetBlockSyncState         = errors.New("failed to get block sync state")
	errCreateAnnouncement        = errors.New("failed to create announcement")
	errCreateStorageProof        = errors.New("failed to create proof")
	errCheckCol                  = errors.New("Checking collateral failed")
	log                          = elog.NewLogger("dsn-h", elog.DebugLog)
)

type StorageHostConf struct {
	TotalStorage  uint64
	Collateral    string
	MaxCollateral string
	AccountName   string
	ChainId       string
	WalletName    string
}

type StorageHost struct {
	isBlockSynced     bool
	announced         bool
	announceConfirmed bool
	wc                *wc.WalletClient
	conf              StorageHostConf
	ctx               context.Context
}

func InitDefaultConf() StorageHostConf {
	chainId := config.ChainHash
	return StorageHostConf{
		TotalStorage:  10 * 1024 * 1024,
		Collateral:    "10",
		MaxCollateral: "20",
		AccountName:   "dsn",
		WalletName:    "ecoball",
		ChainId:       common.ToHex(chainId[:]),
	}
}

func NewStorageHost(ctx context.Context, conf StorageHostConf) *StorageHost {
	return &StorageHost{
		conf:    conf,
		wc:      wc.NewWalletClient(conf.AccountName, conf.WalletName, 10),
		ctx:     ctx,
	}
}

func NewHostWithDefaultConf() *StorageHost {
	conf := InitDefaultConf()
	ctx := context.Background()
	return NewStorageHost(ctx, conf)
}
func (h *StorageHost) Start() error {
	if err := ipfs.Initialize(); err != nil {
		return err
	}
	go ipfs.DaemonRun()
	/*if err := h.proofLoop(); err != nil {
		return err
	}*/

	go h.proofLoop()
	return nil
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
	ok := h.wc.CheckCollateral()
	if !ok {
		return errors.New("Checking collateral failed")
	}
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
	id, err := h.wc.InvokeContract(transaction)
	if err != nil {
		return err
	}

	h.announced = true
	log.Debug("Invoke host announcement,id: ", id)
	return nil
}

func (h *StorageHost) TotalStorage() uint64 {
	return h.conf.TotalStorage
}

// createAnnouncement will take a storage host announcement and encode it, returning the
// exact []byte that should be added to the arbitrary data of a transaction
func (h *StorageHost) createAnnouncement() ([]byte, error) {
	//curBlockHeight := h.ledger.GetCurrentHeight(common.HexToHash(h.conf.ChainId))
	var afc dsnComm.HostAncContract
	//afc.PublicKey = h.account.PublicKey
	afc.TotalStorage = h.conf.TotalStorage
	//afc.StartAt = curBlockHeight
	afc.StartAt = uint64(time.Now().Unix())
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
	/*annHash := sha256.Sum256(annBytes)
	sig, err := h.account.Sign(annHash[:])
	if err !=  nil {
		return nil, err
	}
	return append(annBytes, sig[:]...), nil*/
	return annBytes, nil
}

func (h *StorageHost) createStorageProof() ([]byte, error) {
	repoStat, err := api.IpfsRepoStat(h.ctx)
	if err != nil {
		return nil, err
	}
	if repoStat.NumObjects == 0 {
		return nil, errors.New("have no object")
	}
	var proof dsnComm.StorageProof
	//proof.PublicKey = h.account.PublicKey
	proof.RepoSize = repoStat.RepoSize
	//proof.AtHeight = h.ledger.GetCurrentHeight(common.HexToHash(h.conf.ChainId))
	proof.AtHeight = uint64(time.Now().Unix())
	allCids, err := api.IpfsBlockAllKey(h.ctx)
	if err != nil {
		return nil, err
	}

	for {
		var proofCid string
		j := rand.Intn(int(repoStat.NumObjects))
		for k, cid := range allCids {
			if k == j {
				proofCid = cid
			}
		}
		block, err := api.IpfsBlockGet(h.ctx, proofCid)
		if err != nil {
			continue
		}
		proof.Cid = proofCid
		blockData, err := ioutil.ReadAll(block)
		if err != nil {
			continue
		}
		dataSize := len(blockData)
		if dataSize < dproof.SegmentSize {
			continue
		}
		var numberSegment int
		if dataSize % dproof.SegmentSize == 0 {
			numberSegment = dataSize / dproof.SegmentSize
		} else {
			numberSegment = dataSize / dproof.SegmentSize + 1
		}
		segmentIndex := rand.Intn(int(numberSegment))
		base, cachedHashSet := dproof.MerkleProof(blockData, uint64(segmentIndex))
		proof.SegmentIndex = uint64(segmentIndex)
		proof.HashSet = cachedHashSet
		copy(proof.Segment[:], base)
		proof.AccountName = h.conf.AccountName
		break
	}

	proofBytes := encoding.Marshal(proof)
	/*annHash := sha256.Sum256(proofBytes)
	sig, err := h.account.Sign(annHash[:])
	if err !=  nil {
		return nil, err
	}*/
	//return append(proofBytes, sig[:]...), nil
	return proofBytes, nil
}

func (h *StorageHost) ProvideStorageProof() error {
	//TODO
	//syncState := h.getBlockSyncState(h.chainId)
	//if !syncState {
	//	return errGetBlockSyncState
	//}

	proof, err := h.createStorageProof()
	if err != nil {
		log.Error(err.Error())
		return errCreateStorageProof
	}
	timeNow := time.Now().UnixNano()
	transaction, err := types.NewInvokeContract(common.NameToIndex(h.conf.AccountName),
		innerCommon.NameToIndex(dsnComm.RootAccount), common.HexToHash(h.conf.ChainId),
		"owner", dsnComm.FcMethodProof, []string{string(proof)}, 0, timeNow)
	if err != nil {
		return err
	}
	id, err := h.wc.InvokeContract(transaction)
	if err != nil {
		return err
	}
	log.Info("storage proof, id: ", id)
	return nil
}

func (h *StorageHost) proofLoop() error {
	//TODO period: move to config
	//timerChan := time.NewTicker(24 * time.Hour).C
	timerChan := time.NewTicker(2 * 60 * time.Second).C
	for {
		select {
		case <-timerChan:
			log.Debug("new storage proof...")
			if !h.announced {
				h.Announce()
				continue
			}
			if h.announced {
				h.ProvideStorageProof()
			}
		case <-h.ctx.Done():
			return h.ctx.Err()
		}
	}
}
