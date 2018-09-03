package renter

import (
	"math/big"
	//"github.com/ecoball/go-ecoball/net/proof"
	//"github.com/ecoball/go-ecoball/net/settlement"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"os"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/dsn/crypto"
	"time"
	"errors"
	"github.com/ecoball/go-ecoball/core/types"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
)

var (
	errUnSyncedStat = errors.New("Block is unsynced!")
	errCreateContract = errors.New("failed to create file contract")
)
// An allowance dictates how much the Renter is allowed to spend in a given
// period. Note that funds are spent on both storage and bandwidth
type Allowance struct {
	Funds       big.Int
	Period      uint64
	RenewWindow uint64
}
type FileContract struct {
	PublicKey   []byte
	Cid         string
	LocalPath   string
	FileSize    uint64
	Redundancy  uint8
	Funds       big.Int
	//Expiration  proof.BlockHeight
	StartAt     uint64
	// 0 : nerver expired
	Expiration  uint64
}

type RenterConf struct {
	AccountName   string
	Redundancy    uint8
	Allowance     big.Int
}

type fileInfo struct {
	name            string
	size            uint64
	transactionId   common.Hash
	fileId          string
	redundancy      uint8
	fee             big.Int
}


type Renter interface {
	Files() []fileInfo
	TotalCost() big.Int
	GetFile(cid string) error
	AddFile(fpath string) error
}

type renter struct {
	isSynced bool
	account  account.Account
	files    map[string]fileInfo
	ledger   ledger.Ledger
	chainId  common.Hash
	db       store.Storage
	conf     RenterConf
}

func NewRenter(l ledger.Ledger, aac account.Account, conf RenterConf)  {
	
}

func (r *renter) estimateFee(fname string, conf RenterConf) big.Int {
	//TODO
	var fee big.Int
	return fee
}

func (r *renter) getCurBlockHeight() uint64 {
	return 0
}

func (r *renter) getBlockSyncState(chainId common.Hash) bool {
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

func (r *renter)createFileContract(fname string, cid string) ([]byte, error) {
	fi, err := os.Stat(fname)
	if err != nil {
		return nil, err
	}
	var fc FileContract
	fc.LocalPath = fname
	fc.FileSize = uint64(fi.Size())
	fc.PublicKey = r.account.PublicKey
	fc.Cid = cid
	fc.StartAt = r.getCurBlockHeight()
	fc.Expiration = 0
	fc.Funds = r.estimateFee(fname, r.conf)
	fc.Redundancy = r.conf.Redundancy
	fcBytes := encoding.Marshal(fc)
	fcHash := crypto.HashBytes(fcBytes)
	var sk crypto.SecretKey
	copy(sk[:], r.account.PrivateKey)
	sig := crypto.SignHash(fcHash, sk)
	return append(fcBytes, sig[:]...), nil
}

func (r *renter) InvokeFileContract(fname, cid string) error {
	if !r.isSynced {
		return errUnSyncedStat
	}
	fc, err := r.createFileContract(fname, cid)
	if err != nil {
		return  errCreateContract
	}
	timeNow := time.Now().Unix()
	transaction, err := types.NewInvokeContract(innerCommon.NameToIndex(r.conf.AccountName),
		innerCommon.NameToIndex("root"), r.chainId,
		"owner", "reg_file", []string{string(fc)}, 0, timeNow)
	if err != nil {
		return err
	}
	err = event.Send(event.ActorNil, event.ActorTxPool, transaction)
	if err != nil {
		return err
	}

	fi, err := os.Stat(fname)
	if err != nil {
		return err
	}
	var f fileInfo
	f.size = uint64(fi.Size())
	f.name = fname
	f.fileId = cid
	f.redundancy = r.conf.Redundancy
	f.transactionId = transaction.Hash
	f.fee = r.estimateFee(fname, r.conf)
	r.files[cid] = f

	return nil
}

func (r *renter) Start()  {
	
}

func (r *renter) AddFile(fpath string) error {
	return nil
}

func (r *renter) GetFile(cid string) error {
	return nil
}

func (r *renter) Files() []fileInfo {
	var files []fileInfo
	for _, v := range r.files {
		files = append(files, v)
	}
	return files
}

func (r *renter) TotalCost() big.Int {
	fee := new(big.Int)
	for _, v := range r.files {
		fee = fee.Add(fee, &v.fee)
	}
	return *fee
}


