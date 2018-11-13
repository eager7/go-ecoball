package client

import (
	"os"
	"math/big"
	"time"
	"errors"
	"context"
	"io"
	"net/http"
	"mime/multipart"
	"path/filepath"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/core/types"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	dsnComm "github.com/ecoball/go-ecoball/dsn/common"
	ipfsshell "github.com/ipfs/go-ipfs-api"
	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/ecoball/go-ecoball/http/response"
	 serial "gx/ipfs/QmYyFh6g1C9uieTpH8CR8PpWBUQjvMDJTsRhJWx5qkXy39/go-ipfs-config/serialize"
	 "path"
	 ecoballConfig "github.com/ecoball/go-ecoball/common/config"
	 ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	 manet "gx/ipfs/QmV6FjemM1K8oXjrvuq3wuVWWoU2TLDPmNnKrxHzY3v6Ai/go-multiaddr-net"
)

var (
	errUnSyncedStat = errors.New("Block is unsynced!")
	errCreateContract = errors.New("failed to create file contract")
	errCheckColFailed = errors.New("Checking collateral failed")
	log = elog.NewLogger("dsn-r", elog.DebugLog)
)

// An allowance dictates how much the DsnClient is allowed to spend in a given
// period. Note that funds are spent on both storage and bandwidth
type Allowance struct {
	Funds       big.Int
	Period      uint64
	RenewWindow uint64
}

type RenterConf struct {
	AccountName   string
	WalletName    string
	Redundancy    uint8
	Allowance     string
	Collateral    string
	MaxCollateral string
	ChainId       string
	DsnApiUrl	  string
	IpfsApiUrl    string
}

type DsnClient struct {
	Conf         RenterConf
	client       *http.Client
	ipfsClient   *ipfsshell.Shell
	ctx          context.Context
}

func InitDefaultConf() RenterConf {
	chainId := config.ChainHash
	return RenterConf{
		AccountName: "dsn",
		WalletName: "ecoball",
		Redundancy: 2,
		Allowance: "10",
		Collateral: "10",
		MaxCollateral: "20",
		ChainId: common.ToHex(chainId[:]),
		DsnApiUrl: "http://localhost:20678",
		IpfsApiUrl: "127.0.0.1:5011",
	}
}

func NewRenter(ctx context.Context, conf RenterConf) *DsnClient {


	cfg, err := serial.Load(path.Join(ecoballConfig.IpfsDir, "config"))
	if err != nil {
		return nil
	}

	apiMaddr, err := ma.NewMultiaddr(cfg.Addresses.API)
	if err != nil {
		return nil
	}
	_, ip_port, err := manet.DialArgs(apiMaddr)
	if err != nil {
		return nil
	}
	
	r := DsnClient{
		Conf: conf,
		client: &http.Client{},
		ipfsClient: ipfsshell.NewShell(ip_port),
		ctx: ctx,
	}
	return &r
}

func NewRcWithDefaultConf(ctx context.Context) *DsnClient {
	conf := InitDefaultConf()
	return NewRenter(ctx, conf)
}

func (r *DsnClient) estimateFee(fname string, conf RenterConf) *big.Int {
	//TODO
	var fee big.Int
	return &fee
}

func (r *DsnClient)createFileContract(fname, cid, payid string) ([]byte, error) {
	fi, err := os.Stat(fname)
	if err != nil {
		return nil, err
	}
	var fc dsnComm.FileContract
	fc.LocalPath = fname
	fc.FileSize = uint64(fi.Size())
	//fc.PublicKey = r.account.PublicKey
	fc.Cid = cid
	//fc.StartAt = r.ledger.GetCurrentHeight(common.HexToHash(r.conf.ChainId))
	fc.StartAt = uint64(time.Now().Unix())
	fc.Expiration = 0
	fee := r.estimateFee(fname, r.Conf)
	fc.Funds, _ = fee.GobEncode()
	fc.Redundancy = r.Conf.Redundancy
	fc.AccountName = r.Conf.AccountName
	fc.PayId = payid
	fcBytes := encoding.Marshal(fc)
	/*annHash := sha256.Sum256(fcBytes)
	sig, err := r.account.Sign(annHash[:])
	if err !=  nil {
		return nil, err
	}
	return append(fcBytes, sig[:]...), nil*/
	return fcBytes, nil
}

func (r *DsnClient)createFileContractWeb(fname string, size uint64, cid string) ([]byte, error) {

	var fc dsnComm.FileContract
	fc.LocalPath = fname
	
	fc.FileSize = size
	//fc.PublicKey = r.account.PublicKey
	fc.Cid = cid
	//fc.StartAt = r.ledger.GetCurrentHeight(common.HexToHash(r.Conf.ChainId))
	fc.StartAt = uint64(time.Now().Unix())
	fc.Expiration = 0
	fee := r.estimateFee(fname, r.Conf)
	fc.Funds, _ = fee.GobEncode()
	fc.Redundancy = r.Conf.Redundancy
	fc.AccountName = r.Conf.AccountName
	fcBytes := encoding.Marshal(fc)
	/*annHash := sha256.Sum256(fcBytes)
	sig, err := r.account.Sign(annHash[:])
	if err !=  nil {
		return nil, err
	}
	return append(fcBytes, sig[:]...), nil*/
	return fcBytes, nil
}

func (r *DsnClient) PayForFile(fname string) (*types.Transaction, error) {
	fi, err := os.Stat(fname)
	if err != nil {
		return nil, err
	}

	fee := fi.Size() * int64(r.Conf.Redundancy) / 1024 * 1024 + 1
	//var bf big.Int
	//fun := bf.SetInt64(fee)
	bigValue := big.NewInt(fee)
	timeNow := time.Now().UnixNano()
	tran, err := types.NewTransfer(common.NameToIndex(r.Conf.AccountName),
		innerCommon.NameToIndex(dsnComm.RootAccount), common.HexToHash(r.Conf.ChainId), "owner", bigValue, 0, timeNow)
	if err != nil {
		return nil, err
	}
	return tran, nil
}
func (r *DsnClient) InvokeFileContract(fname, cid, payId string) (*types.Transaction, error) {
	/*if !r.isSynced {
		return errUnSyncedStat
	}*/
	fc, err := r.createFileContract(fname, cid, payId)
	if err != nil {
		return  nil, errCreateContract
	}

	timeNow := time.Now().UnixNano()
	transaction, err := types.NewInvokeContract(common.NameToIndex(r.Conf.AccountName),
		innerCommon.NameToIndex(dsnComm.RootAccount), common.HexToHash(r.Conf.ChainId),
		"owner", dsnComm.FcMethodFile, []string{string(fc)}, 0, timeNow)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (r *DsnClient) InvokeFileContractWeb(fname string , size uint64, cid string) (*types.Transaction, error) {
	/*if !r.isSynced {
		return errUnSyncedStat
	}*/
	fc, err := r.createFileContractWeb(fname,size, cid)
	if err != nil {
		return  nil, errCreateContract
	}

	timeNow := time.Now().UnixNano()
	transaction, err := types.NewInvokeContract(common.NameToIndex(r.Conf.AccountName),
		innerCommon.NameToIndex(dsnComm.RootAccount), common.HexToHash(r.Conf.ChainId),
		"owner", dsnComm.FcMethodFile, []string{string(fc)}, 0, timeNow)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *DsnClient) RscCodingReq(fpath, cid string) (string, error) {
	fp := filepath.ToSlash(filepath.Clean(fpath))
	stat, err := os.Lstat(fp)
	if err != nil {
		panic(err)
	}

	var PieceSize uint64
	if stat.Size() < dsnComm.EraDataPiece * (256 * 1024) {
		PieceSize = uint64(stat.Size() / dsnComm.EraDataPiece)
	} else {
		PieceSize = uint64(256 * 1024)
	}
	req := dsnComm.RscReq{
		Cid: cid,
		Redundency: int(r.Conf.Redundancy),
		IsDir: false,
		Chunk: PieceSize,
		FileSize: uint64(stat.Size()),
	}
	var result response.DsnEraCoding
	err = rpc.NodePost("/dsn/eracode", &req, &result)
	return result.Cid, nil
}

func (r *DsnClient) RscCodingReqWeb(size int64, cid string) (string, error) {

	var PieceSize uint64
	if size < dsnComm.EraDataPiece * (256 * 1024) {
		PieceSize = uint64(size / dsnComm.EraDataPiece)
	} else {
		PieceSize = uint64(256 * 1024)
	}

	req := dsnComm.RscReq{
		Cid: cid,
		Redundency: int(r.Conf.Redundancy),
		IsDir: false,
		Chunk: PieceSize,
		FileSize: uint64(size),
	}
	var result response.DsnEraCoding
	err := rpc.NodePost("/dsn/eracode", &req, &result)
	if err != nil{
		return "", err
	}
	return result.Cid, nil
	
}

func (r *DsnClient) RscDecodingReq(cid string) error{
	var result response.DsnEraDecoding
	err := rpc.NodeGet("/dsn/eradecode", &result)
	 if err != nil {
	 	return err
	 }
	 //TODO
	 return nil
}

func (r *DsnClient) AddFile(fpath string) (string, string, error) {
	file, err := os.Open(fpath)
	if err!= nil{
		return "","", err
	}
	defer file.Close()

	return r.ipfsClient.Add(file)
}

func (r *DsnClient) HttpAddFile( mulFile *multipart.FileHeader) (string, string, error) {
	// name := mulFile.Header.Get("file")
	// fmt.Println(name)
	src, err := mulFile.Open()
	if err != nil {
		return "","", err 
	}
	defer src.Close()
	return r.ipfsClient.Add(src)

}

func (r *DsnClient) CatFile(path string) (io.ReadCloser, error) {
	newPath := path + "/file"
	return r.ipfsClient.Cat(newPath)
}

func (r *DsnClient) GetFile(path, out string) error {
	newPath := path + "/file"
	return r.ipfsClient.Get(newPath, out)
}

