package client

import (
	"os"
	"math/big"
	"time"
	"errors"
	"context"
	"io"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/core/types"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	dsnComm "github.com/ecoball/go-ecoball/dsn/common"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/ecoball/go-ecoball/dsn/renter"
	"strconv"
	"github.com/ecoball/go-ecoball/client/rpc"
	//clientcommon "github.com/ecoball/go-ecoball/client/common"
	//"net/url"
	"path/filepath"
	ipfsshell "github.com/ipfs/go-ipfs-api"
	"fmt"
	//ecoclient "github.com/ecoball/go-ecoball/client/commands"
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


type RenterConf struct {
	AccountName   string
	Redundancy    uint8
	Allowance     string
	Collateral    string
	MaxCollateral string
	ChainId       string
	//StorePath     string
	DsnApiUrl	  string
	IpfsApiUrl    string
}

type Renter struct {
	conf         RenterConf
	client       *http.Client
	ipfsClient   *ipfsshell.Shell
	ctx          context.Context
}

func InitDefaultConf() RenterConf {
	chainId := config.ChainHash
	return RenterConf{
		AccountName: "dsn",
		Redundancy: 2,
		Allowance: "10",
		Collateral: "10",
		MaxCollateral: "20",
		ChainId: common.ToHex(chainId[:]),
		//StorePath: "/tmp/storage/rent",
		DsnApiUrl: "http://localhost:9000",
		IpfsApiUrl: "127.0.0.1:5011",
	}
}

func NewRenter(ctx context.Context, conf RenterConf) *Renter {
	r := Renter{
		conf: conf,
		client: &http.Client{},
		ipfsClient: ipfsshell.NewShell(conf.IpfsApiUrl),
		ctx: ctx,
	}
	return &r
}

func NewRcWithDefaultConf(ctx context.Context) *Renter {
	conf := InitDefaultConf()
	return NewRenter(ctx, conf)
}

func (r *Renter) estimateFee(fname string, conf RenterConf) *big.Int {
	//TODO
	var fee big.Int
	return &fee
}

func (r *Renter)createFileContract(fname string, cid string) ([]byte, error) {
	fi, err := os.Stat(fname)
	if err != nil {
		return nil, err
	}
	var fc renter.FileContract
	fc.LocalPath = fname
	fc.FileSize = uint64(fi.Size())
	//fc.PublicKey = r.account.PublicKey
	fc.Cid = cid
	//fc.StartAt = r.ledger.GetCurrentHeight(common.HexToHash(r.conf.ChainId))
	fc.StartAt = uint64(time.Now().Unix())
	fc.Expiration = 0
	fee := r.estimateFee(fname, r.conf)
	fc.Funds, _ = fee.GobEncode()
	fc.Redundancy = r.conf.Redundancy
	fc.AccountName = r.conf.AccountName
	fcBytes := encoding.Marshal(fc)
	/*annHash := sha256.Sum256(fcBytes)
	sig, err := r.account.Sign(annHash[:])
	if err !=  nil {
		return nil, err
	}
	return append(fcBytes, sig[:]...), nil*/
	return fcBytes, nil
}

func (r *Renter) PayForFile(fname, cid string) (*types.Transaction, error) {
	fi, err := os.Stat(fname)
	if err != nil {
		return nil, err
	}

	fee := fi.Size() * int64(r.conf.Redundancy) / 1024 * 1024 + 1
	var bf big.Int
	fun := bf.SetInt64(fee)
	timeNow := time.Now().UnixNano()
	tran, err := types.NewTransfer(common.NameToIndex(r.conf.AccountName),
		innerCommon.NameToIndex(dsnComm.RootAccount), common.HexToHash(r.conf.ChainId), "owner", fun, 0, timeNow)
	if err != nil {
		return nil, err
	}
	log.Debug("pay for ", fname, " ", fee)
	/*trn, err := tran.Serialize()
	if err != nil {
		return err
	}

	var resultKeys clientcommon.SimpleResult
	err = rpc.WalletGet("/wallet/getPublicKeys", &resultKeys)
	if err != nil {
		return err
	}

	pks := resultKeys.Result
	var retReqKeys clientcommon.SimpleResult
	values := url.Values{}
	values.Set("permission", "owner")
	values.Set("chainId", r.conf.ChainId)
	values.Set("keys", pks)
	values.Set("transaction", innerCommon.ToHex(trn))
	err = rpc.NodePost("/get_required_keys", values.Encode(), &retReqKeys)
	if err != nil {
		return err
	}

	var retTrn clientcommon.SimpleResult
	sigTrnReq := url.Values{}
	sigTrnReq.Set("keys", retReqKeys.Result)
	sigTrnReq.Set("data", innerCommon.ToHex(trn))
	err = rpc.WalletPost("/wallet/signTransaction", values.Encode(), &retTrn)
	if err != nil {
		return err
	}
	err = tran.Deserialize(innerCommon.FromHex(retTrn.Result))
	if err != nil {
		return err
	}

	data, err := tran.Serialize()
	if err != nil {
		return err
	}

	var retTfer clientcommon.SimpleResult
	ctcv := url.Values{}
	ctcv.Set("transaction", common.ToHex(data))
	err = rpc.NodePost("/transfer", values.Encode(), &retTfer)*/
	//fmt.Println(result.Result)
	return tran, nil
}
func (r *Renter) InvokeFileContract(fname, cid string) (*types.Transaction, error) {
	/*if !r.isSynced {
		return errUnSyncedStat
	}*/
	fc, err := r.createFileContract(fname, cid)
	if err != nil {
		return  nil, errCreateContract
	}

	timeNow := time.Now().UnixNano()
	transaction, err := types.NewInvokeContract(common.NameToIndex(r.conf.AccountName),
		innerCommon.NameToIndex(dsnComm.RootAccount), common.HexToHash(r.conf.ChainId),
		"owner", dsnComm.FcMethodFile, []string{string(fc)}, 0, timeNow)
	if err != nil {
		return nil, err
	}

	/*transaction.SetSignature(&config.Root)
	err = event.Send(event.ActorNil, event.ActorTxPool, transaction)
	if err != nil {
		return err
	}*/

	/*trn, err := transaction.Serialize()
	if err != nil {
		return err
	}*/

	/*var resultKeys clientcommon.SimpleResult
	err = rpc.WalletGet("/wallet/getPublicKeys", &resultKeys)
	if err != nil {
		return err
	}

	pks := resultKeys.Result
	var retReqKeys clientcommon.SimpleResult
	values := url.Values{}
	values.Set("permission", "owner")
	values.Set("chainId", r.conf.ChainId)
	values.Set("keys", pks)
	values.Set("transaction", innerCommon.ToHex(trn))
	err = rpc.NodePost("/get_required_keys", values.Encode(), &retReqKeys)
	if err != nil {
		return err
	}

	var retTrn clientcommon.SimpleResult
	sigTrnReq := url.Values{}
	sigTrnReq.Set("keys", retReqKeys.Result)
	sigTrnReq.Set("data", innerCommon.ToHex(trn))
	err = rpc.WalletPost("/wallet/signTransaction", values.Encode(), &retTrn)
	if err != nil {
		return err
	}*/

	/*pkKeys, err := ecoclient.GetPublicKeys()
	if err != nil {
		return err
	}

	reqKeys, err := ecoclient.GetRequiredKeys(innerCommon.HexToHash(r.conf.ChainId), pkKeys, "owner", transaction)
	if err != nil {
		return err
	}

	err = ecoclient.SignTransaction(innerCommon.HexToHash(r.conf.ChainId), reqKeys, transaction)
	if err != nil {
		return err
	}

	data, err := transaction.Serialize()
	if err != nil {
		return err
	}

	var retContract clientcommon.SimpleResult
	ctcv := url.Values{}
	ctcv.Set("transaction", common.ToHex(data))
	err = rpc.NodePost("/invokeContract", ctcv.Encode(), &retContract)
	fmt.Println("fileContract: ", retContract.Result)
	//var f fileInfo
	*//*f.size = uint64(fi.Size())
	f.name = fname
	f.fileId = cid
	f.redundancy = r.conf.Redundancy
	f.transactionId = transaction.Hash
	f.fee = *r.estimateFee(fname, r.conf)
	r.files[cid] = f*//*

	//r.payForFile(f)
	//r.persistFileInfo(f)

	return err*/

	return transaction, nil
}

func (r *Renter)CheckCollateral() bool {
	//sacc, err := r.ledger.AccountGet(common.HexToHash(r.conf.ChainId), common.NameToIndex(r.conf.AccountName))
	//if err != nil {
	//	return false
	//}
	//TODO much more checking
	//if sacc.Votes.Staked > 0 {
	//	return true
	//}
	url := r.conf.DsnApiUrl + "/dsn/accountstake?" + "name=" + r.conf.AccountName + "&chainid=" + r.conf.ChainId
	//url :="/dsn/accountstake/" + r.conf.AccountName + "/" + r.conf.ChainId
	rsp, err := r.client.Get(url)
	if err != nil {
		return false
	}
	defer rsp.Body.Close()

	out, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return false
	}
	var result renter.AccountStakeRsp
	if err := json.Unmarshal(out, &result); err != nil {
		return false
	}
	fmt.Println("col:", result.Stake)
	col, err := strconv.Atoi(r.conf.Collateral)
	if err != nil {
		return false
	}
	if result.Stake < uint64(col) {
		return false
	}
	return true
}

func (r *Renter) RscCodingReq(fpath, cid string) (string, error) {
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
	req := renter.RscReq{
		Cid: cid,
		Redundency: int(r.conf.Redundancy),
		IsDir: false,
		Chunk: PieceSize,
		FileSize: uint64(stat.Size()),
	}

	jreq, _ := json.Marshal(req)
	resp, err := rpc.NodeCall("DsnAddFile", []interface{}{string(jreq)})
	if err != nil {
		return "", err
	}
	result := resp["desc"].(string)
	if result != "success" {
		return "", errors.New(result)
	}
	newCid := resp["result"].(string)
	return newCid, nil
}

func (r *Renter) RscDecodingReq(cid string) error{
	resp, err := rpc.NodeCall("DsnCatFile", []interface{}{cid})
	if err != nil {
		return err
	}
	return rpc.EchoResult(resp)
}

func (r *Renter) AddFile(fpath string) (string, error) {
	file, err := os.Open(fpath)
	if err!= nil{
		return "", err
	}
	defer file.Close()

	return r.ipfsClient.Add(file)
}

func (r *Renter) CatFile(path string) (io.ReadCloser, error) {
	newPath := path + "/file"
	return r.ipfsClient.Cat(newPath)
}

func (r *Renter) GetFile(path, out string) error {
	newPath := path + "/file"
	return r.ipfsClient.Get(newPath, out)
}
