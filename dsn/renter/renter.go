package renter

import (
	"math/big"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/core/store"
	"os"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/dsn/crypto"
	"time"
	"errors"
	"github.com/ecoball/go-ecoball/core/types"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
	"context"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ipfs/go-ipfs/core"
	"path/filepath"
	"gx/ipfs/QmdE4gMduCKCGAcczM2F5ioYDfdeKuPix138wrES1YSr7f/go-ipfs-cmdkit/files"
	"path"
	"github.com/ecoball/go-ecoball/dsn/renter/pb"
)

var (
	ipfsNode *core.IpfsNode
	errUnSyncedStat = errors.New("Block is unsynced!")
	errCreateContract = errors.New("failed to create file contract")
	errCheckColFailed = errors.New("Checking collateral failed")
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
	StartAt     uint64
	Expiration  uint64
}

type RenterConf struct {
	AccountName   common.AccountName
	Redundancy    uint8
	Allowance     big.Int
	Collateral    big.Int
	MaxCollateral big.Int
	ChainId       common.Hash
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

func NewRenter(ctx context.Context,l ledger.Ledger ,ac account.Account, conf RenterConf) *Renter {
	r := Renter{
		account: ac,
		ledger: l,
		conf: conf,
		files: make(map[string]fileInfo, 64),
		ctx: ctx,
	}
	//TODO init db

	return &r
}

func (r *Renter) Start()  {
	r.loadFileInfo()
	r.getBlockSyncState(r.conf.ChainId)
}
func (r *Renter) estimateFee(fname string, conf RenterConf) big.Int {
	//TODO
	var fee big.Int
	return fee
}

func (r *Renter) getBlockSyncState(chainId common.Hash) bool {
	timerChan := time.NewTicker(10 * time.Second).C
	var syncState bool
	for {
		select {
		case timerChan:
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
	fc.StartAt = r.ledger.GetCurrentHeight(r.conf.ChainId)
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

func (r *Renter) InvokeFileContract(fname, cid string) error {
	if !r.isSynced {
		return errUnSyncedStat
	}
	fc, err := r.createFileContract(fname, cid)
	if err != nil {
		return  errCreateContract
	}
	timeNow := time.Now().Unix()
	transaction, err := types.NewInvokeContract(r.conf.AccountName,
		innerCommon.NameToIndex("root"), r.conf.ChainId,
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

	r.persistFileInfo(f)

	return nil
}

func (r *Renter)checkCollateral() bool {
	sacc, err := r.ledger.AccountGet(r.conf.ChainId, r.conf.AccountName)
	if err != nil {
		return false
	}
	//TODO much more checking
	if sacc.Votes.Staked > 0 {
		return true
	}
	return false
}

func (r *Renter) AddFile(fpath string) (string, error) {
	if !r.isSynced {
		return "", errUnSyncedStat
	}
	colState := r.checkCollateral()
	if !colState {
		return "", errCheckColFailed
	}
	//TODO erasure coding and add file
	adder, err := NewEcoAdder(r.ctx, ipfsNode.Pinning,ipfsNode.Blockstore, ipfsNode.DAG)
	if err != nil {
		return "", err
	}
	adder.SetRedundancy(r.conf.Redundancy)
	fpath = filepath.ToSlash(filepath.Clean(fpath))
	stat, err := os.Lstat(fpath)
	if err != nil {
		return "", err
	}
	af, err := files.NewSerialFile(path.Base(fpath), fpath, false, stat)
	if err != nil {
		return "", err
	}
	adder.AddFile(af)
	dagnode, err := adder.Finalize()
	r.InvokeFileContract(fpath, dagnode.String())
	return dagnode.String(), nil
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


func SetIpfsNode(node *core.IpfsNode)  {
	ipfsNode = node
}

func (fc *FileContract) Serialize() ([]byte, error) {
	var pfc pb.FcMessage
	pfc.PublicKey = fc.PublicKey
	pfc.Cid = fc.Cid
	pfc.LocalPath = fc.LocalPath
	pfc.FileSize = fc.FileSize
	pfc.Redundancy = uint32(fc.Redundancy)
	pfc.Funds, _ = fc.Funds.GobEncode()
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
	err = fc.Funds.GobDecode(pfc.Funds)
	if err != nil {
		return err
	}
	fc.StartAt = pfc.StartAt
	fc.Expiration = pfc.Expiration
	return nil
}