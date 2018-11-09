package walletclient

import (
	"context"
	"errors"
	"github.com/ecoball/go-ecoball/core/types"
	clientCommon "github.com/ecoball/go-ecoball/client/common"
	"github.com/ecoball/go-ecoball/client/rpc"
	"github.com/ecoball/go-ecoball/http/request"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	walletHttp "github.com/ecoball/go-ecoball/walletserver/http"
	"encoding/hex"
)

type WalletClientConf struct {
	AccountName string
	WalletName  string
	Collateral  int
}

type WalletClient struct {
	Conf WalletClientConf
	ctx context.Context
}

func NewWalletClient(acc, wallet string, col int) *WalletClient {
	conf := WalletClientConf{
		AccountName: acc,
		WalletName: wallet,
		Collateral: col,
	}
	return &WalletClient{
		Conf: conf,
		ctx: context.Background(),
	}
}

//other query method
func getMainChainHash() (innerCommon.Hash, error) {
	var result string
	err := rpc.NodeGet("/query/mainChainHash", &result)
	if nil != err {
		return innerCommon.Hash{}, err
	}

	hash := new(innerCommon.Hash)
	return hash.FormHexString(result), nil
}

//other wallet method
func getPublicKeys() (walletHttp.Keys, error) {
	var result walletHttp.Keys
	err := rpc.WalletGet("/wallet/getPublicKeys", &result)
	if nil == err {
		return result, nil
	}
	return walletHttp.Keys{}, err
}

func getRequiredKeys(chainHash innerCommon.Hash, permission string, account string) ([]innerCommon.Address, error) {
	//var result string
	pubAdd := request.PubKeyAddress{Addresses: []innerCommon.Address{}}
	requestData := request.PermissionPublicKeys{Name: account, Permission: permission, ChainHash: chainHash}
	err := rpc.NodePost("/query/getRequiredKeys", &requestData, &pubAdd)
	if nil == err {
		return pubAdd.Addresses, nil
	}

	return []innerCommon.Address{}, err
}

func getAccountInfo(acc string) (*state.Account, error) {
	chainHash, err := getMainChainHash()
	if err != nil {
		return nil, err
	}
	var result state.Account
	requestData := request.AccountName{Name: acc, ChainHash: chainHash}
	err = rpc.NodePost("/query/getAccountInfo", &requestData, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func signTransaction(chainHash innerCommon.Hash, publickeys walletHttp.Keys, rawData []byte) (walletHttp.SignTransaction, error) {
	var result walletHttp.SignTransaction
	requestData := walletHttp.RawTransactionData{PublicKeys: publickeys, RawData: rawData}
	err := rpc.WalletPost("/wallet/signTransaction", &requestData, &result)
	if nil == err {
		return result, nil
	}
	return walletHttp.SignTransaction{}, err
}

func (w *WalletClient) CheckCollateral() bool {
	state, err := getAccountInfo(w.Conf.AccountName)
	if err != nil {
		return false
	}
	if state.Votes.Staked < uint64(w.Conf.Collateral) {
		return false
	}
	return true
}

func (w *WalletClient) InvokeContract(transaction *types.Transaction) (string, error) {
	chainId, err := getMainChainHash()
	if err != nil {
		return "", err
	}
	pkKeys, err :=  getPublicKeys()
	if err != nil {
		return "", err
	}
	reqKeys, err := getRequiredKeys(chainId, "owner", w.Conf.AccountName)
	if err != nil {
		return "", err
	}
	publickeys := clientCommon.IntersectionKeys(pkKeys, reqKeys)
	if 0 == len(publickeys.KeyList) {
		return "", errors.New("no publickeys")
	}
	dataPays, err := signTransaction(chainId, publickeys, transaction.Hash[:])
	if err != nil {
		return "", err
	}
	for _, v := range dataPays.Signature {
		transaction.AddSignature(v.PublicKey.Key, v.SignData)
	}

	datas, err := transaction.Serialize()
	if err != nil {
		return "", err
	}

	trx_str := hex.EncodeToString(datas)
	var resultPays rpc.SimpleResult
	err = rpc.NodePost("/invokeContract", &trx_str, &resultPays)
	return transaction.Hash.HexString(), err
}

func (w *WalletClient) Transer(transaction *types.Transaction) (string, error) {
	chainId, err := getMainChainHash()
	if err != nil {
		return "", err
	}
	pkKeys, err :=  getPublicKeys()
	if err != nil {
		return "", err
	}
	reqKeys, err := getRequiredKeys(chainId, "owner", w.Conf.AccountName)
	if err != nil {
		return "", err
	}
	publickeys := clientCommon.IntersectionKeys(pkKeys, reqKeys)
	if 0 == len(publickeys.KeyList) {
		return "", errors.New("no publickeys")
	}
	dataPays, err := signTransaction(chainId, publickeys, transaction.Hash[:])
	if err != nil {
		return "", err
	}
	for _, v := range dataPays.Signature {
		transaction.AddSignature(v.PublicKey.Key, v.SignData)
	}

	datas, err := transaction.Serialize()
	if err != nil {
		return "", err
	}

	trx_str := hex.EncodeToString(datas)
	var resultPays rpc.SimpleResult
	err = rpc.NodePost("/invokeContract", &trx_str, &resultPays)
	return transaction.Hash.HexString(), err
}

