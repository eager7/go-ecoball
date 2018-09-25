package renter

import (
	"os"
	"math/big"
	"time"
	"errors"
	"context"
	"crypto/sha256"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/core/types"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/dsn/renter/pb"
	"github.com/ecoball/go-ecoball/dsn/ipfs/api"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	dsnComm "github.com/ecoball/go-ecoball/dsn/common"
)

var (
	errUnSyncedStat = errors.New("Block is unsynced!")
	errCreateContract = errors.New("failed to create file contract")
	errCheckColFailed = errors.New("Checking collateral failed")
	log = elog.NewLogger("dsn-r", elog.DebugLog)
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
	Funds       []byte
	StartAt     uint64
	Expiration  uint64
	AccountName   string
}

type RenterConf struct {
	AccountName   string
	Redundancy    uint8
	Allowance     string
	Collateral    string
	MaxCollateral string
	ChainId       string
	StorePath     string
}

type fileInfo struct {
	name            string
	size            uint64
	transactionId   common.Hash
	fileId          string
	redundancy      uint8
	fee             big.Int
}

type Renter struct {
	isSynced bool
	account  account.Account
	files    map[string]fileInfo
	conf     RenterConf
	ledger   ledger.Ledger
	db       store.Storage
	ctx      context.Context
}

func InitDefaultConf() RenterConf {
	chainId := config.ChainHash
	return RenterConf{
		AccountName: "root",
		Redundancy: 1,
		Allowance: "100",
		Collateral: "10000",
		MaxCollateral: "20000",
		ChainId: common.ToHex(chainId[:]),
		StorePath: "/tmp/storage/rent",
	}
}

func NewRenter(ctx context.Context,l ledger.Ledger ,ac account.Account, conf RenterConf) *Renter {
	r := Renter{
		account: ac,
		ledger: l,
		conf: conf,
		files: make(map[string]fileInfo, 64),
		ctx: ctx,
	}
	r.db, _ = store.NewBlockStore(conf.StorePath)
	return &r
}

func (r *Renter) Start()  {
	r.loadFileInfo()
	//r.getBlockSyncState(common.HexToHash(r.conf.ChainId))
}
func (r *Renter) estimateFee(fname string, conf RenterConf) *big.Int {
	//TODO
	var fee big.Int
	return &fee
}

func (r *Renter) getBlockSyncState(chainId common.Hash) bool {
	timerChan := time.NewTicker(10 * time.Second).C
	var syncState bool
	for {
		select {
		case <-timerChan:
			//TODO get current block synced state
			if syncState {
				r.isSynced = true
				return true
			}
		case <-r.ctx.Done():
			return false
		}

	}
	return false
}

func (r *Renter)createFileContract(fname string, cid string) ([]byte, error) {
	fi, err := os.Stat(fname)
	if err != nil {
		return nil, err
	}
	var fc FileContract
	fc.LocalPath = fname
	fc.FileSize = uint64(fi.Size())
	fc.PublicKey = r.account.PublicKey
	fc.Cid = cid
	fc.StartAt = r.ledger.GetCurrentHeight(common.HexToHash(r.conf.ChainId))
	fc.Expiration = 0
	fee := r.estimateFee(fname, r.conf)
	fc.Funds, _ = fee.GobEncode()
	fc.Redundancy = r.conf.Redundancy
	fc.AccountName = r.conf.AccountName
	fcBytes := encoding.Marshal(fc)
	annHash := sha256.Sum256(fcBytes)
	sig, err := r.account.Sign(annHash[:])
	if err !=  nil {
		return nil, err
	}
	return append(fcBytes, sig[:]...), nil
}

func (r *Renter) payForFile(fc fileInfo) error {
	timeNow := time.Now().UnixNano()
	tran, err := types.NewTransfer(common.NameToIndex(r.conf.AccountName),
		innerCommon.NameToIndex(dsnComm.RootAccount), common.HexToHash(r.conf.ChainId), "owner", &fc.fee, 0, timeNow)
	if err != nil {
		return err
	}
	tran.SetSignature(&r.account)
	err = event.Send(event.ActorNil, event.ActorTxPool, tran)
	if err != nil {
		return err
	}
	return nil
}
func (r *Renter) InvokeFileContract(fname, cid string) error {
	if !r.isSynced {
		return errUnSyncedStat
	}
	fc, err := r.createFileContract(fname, cid)
	if err != nil {
		return  errCreateContract
	}
	timeNow := time.Now().UnixNano()
	transaction, err := types.NewInvokeContract(common.NameToIndex(r.conf.AccountName),
		innerCommon.NameToIndex(dsnComm.RootAccount), common.HexToHash(r.conf.ChainId),
		"owner", dsnComm.FcMethodFile, []string{string(fc)}, 0, timeNow)
	if err != nil {
		return err
	}
	transaction.SetSignature(&config.Root)
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
	f.fee = *r.estimateFee(fname, r.conf)
	r.files[cid] = f

	r.payForFile(f)
	r.persistFileInfo(f)

	return nil
}

func (r *Renter)checkCollateral() bool {
	sacc, err := r.ledger.AccountGet(common.HexToHash(r.conf.ChainId), common.NameToIndex(r.conf.AccountName))
	if err != nil {
		return false
	}
	//TODO much more checking
	if sacc.Votes.Staked > 0 {
		return true
	}
	return false
}

func (r *Renter) AddFile(fpath string, era int8) (string, error) {
	//TODO
	//if !r.isSynced {
	//	return "", errUnSyncedStat
	//}
	//colState := r.checkCollateral()
	//if !colState {
	//	return "", errCheckColFailed
	//}
	var redundancy uint8
	if era == -1 {
		redundancy = r.conf.Redundancy
	} else {
		redundancy = uint8(era)
	}
	cid, err := api.IpfsAddEraFile(r.ctx, fpath, redundancy)
	if err != nil {
		return "", err
	}
	r.InvokeFileContract(fpath, cid)
	return cid, nil
}

func (r *Renter) GetFile(cid string) error {
	//TODO
	return nil
}

func (r *Renter) Files() []fileInfo {
	var files []fileInfo
	for _, v := range r.files {
		files = append(files, v)
	}
	return files
}

func (r *Renter) TotalCost() big.Int {
	fee := new(big.Int)
	for _, v := range r.files {
		fee = fee.Add(fee, &v.fee)
	}
	return *fee
}

func (r *Renter) persistFileInfo(fi fileInfo) error {
	//TODO
	//r.db.Put()
	return nil
}

func (r *Renter) loadFileInfo() error {
	//TODO
	//r.db.Get()
	return nil
}


func (fc *FileContract) Serialize() ([]byte, error) {
	var pfc pb.FcMessage
	pfc.PublicKey = fc.PublicKey
	pfc.Cid = fc.Cid
	pfc.LocalPath = fc.LocalPath
	pfc.FileSize = fc.FileSize
	pfc.Redundancy = uint32(fc.Redundancy)
	//pfc.Funds, _ = fc.Funds.GobEncode()
	pfc.Funds = fc.Funds
	pfc.StartAt = fc.StartAt
	pfc.Expiration = fc.Expiration
	return pfc.Marshal()
}

func (fc *FileContract) Deserialize(data []byte) error {
	var pfc pb.FcMessage
	err := pfc.Unmarshal(data)
	if err != nil {
		return err
	}
	fc.PublicKey = pfc.PublicKey
	fc.Cid = pfc.Cid
	fc.LocalPath = pfc.LocalPath
	fc.FileSize = pfc.FileSize
	fc.Redundancy = uint8(pfc.Redundancy)
	//err = fc.Funds.GobDecode(pfc.Funds)
	//if err != nil {
	//	return err
	//}
	fc.Funds = pfc.Funds
	fc.StartAt = pfc.StartAt
	fc.Expiration = pfc.Expiration
	return nil
}