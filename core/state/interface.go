package state

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"math/big"
)

type InterfaceState interface {
	StateType() TypeState
	RegisterChain(index common.AccountName, hash, txHash common.Hash, addr common.Address) error
	GetChainList() ([]Chain, error)
	AddPermission(index common.AccountName, perm Permission) error
	CheckPermission(index common.AccountName, name string, hash common.Hash, signatures []common.Signature) error
	CheckAccountPermission(host common.AccountName, guest common.AccountName, permission string) error
	FindPermission(index common.AccountName, name string) (string, error)
	SetResourceLimits(from, to common.AccountName, cpuStaked, netStaked uint64, cpuLimit, netLimit float64) error
	SubResources(index common.AccountName, cpu, net float64, cpuLimit, netLimit float64) error
	CancelDelegate(from, to common.AccountName, cpuStaked, netStaked uint64, cpuLimit, netLimit float64) error
	RecoverResources(index common.AccountName, timeStamp int64, cpuLimit, netLimit float64) error
	RequireResources(index common.AccountName, cpuLimit, netLimit float64, timeStamp int64) (float64, float64, error)
	RegisterProducer(index common.AccountName, addr string, port uint32, payee common.AccountName) error
	UnRegisterProducer(index common.AccountName) error
	ElectionToVote(index common.AccountName, accounts []common.AccountName) error
	RequireVotingInfo() bool
	GetProducerList() ([]Elector, error)
	AccountGetBalance(index common.AccountName, token string) (*big.Int, error)
	AccountSubBalance(index common.AccountName, token string, value *big.Int) error
	AccountAddBalance(index common.AccountName, token string, value *big.Int) error
	CreateToken(symbol string, maxSupply *big.Int, creator, issuer common.AccountName) (*TokenInfo, error)
	IssueToken(to common.AccountName, amount *big.Int, symbol string) error
	GetTokenInfo(symbol string) (*TokenInfo, error)
	TokenExisted(name string) bool
	SetTokenInfo(symbol string, maxSupply, supply *big.Int, creator, issuer common.AccountName) (*TokenInfo, error)
	CopyState() (*State, error)
	AddAccount(index common.AccountName, addr common.Address, timeStamp int64) (*Account, error)
	SetContract(index common.AccountName, t types.VmType, des, code, abi[]byte) error
	GetContract(index common.AccountName) (*types.DeployInfo, error)
	StoreSet(index common.AccountName, key, value []byte) (err error)
	StoreGet(index common.AccountName, key []byte) (value []byte, err error)
	GetAccountByName(index common.AccountName) (*Account, error)
	GetAccountByAddr(addr common.Address) (*Account, error)
	GetHashRoot() common.Hash
	CommitToDB() error
	Reset(hash common.Hash) error
}
