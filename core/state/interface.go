package state

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"math/big"
)

type InterfaceState interface {
	StateType() TypeState
	RegisterChain(index common.AccountName, hash common.Hash) error
	GetChainList() ([]Chain, error)
	AddPermission(index common.AccountName, perm Permission) error
	CheckPermission(index common.AccountName, name string, hash common.Hash, signatures []common.Signature) error
	FindPermission(index common.AccountName, name string) (string, error)
	SetResourceLimits(from, to common.AccountName, cpuStaked, netStaked uint64, cpuLimit, netLimit float64) error
	SubResources(index common.AccountName, cpu, net float64, cpuLimit, netLimit float64) error
	CancelDelegate(from, to common.AccountName, cpuStaked, netStaked uint64, cpuLimit, netLimit float64) error
	RecoverResources(index common.AccountName, timeStamp int64, cpuLimit, netLimit float64) error
	RequireResources(index common.AccountName, cpuLimit, netLimit float64, timeStamp int64) (float64, float64, error)
	RegisterProducer(index common.AccountName) error
	UnRegisterProducer(index common.AccountName) error
	ElectionToVote(index common.AccountName, accounts []common.AccountName) error
	RequireVotingInfo() bool
	GetProducerList() ([]common.AccountName, error)
	AccountGetBalance(index common.AccountName, token string) (*big.Int, error)
	AccountSubBalance(index common.AccountName, token string, value *big.Int) error
	AccountAddBalance(index common.AccountName, token string, value *big.Int) error
	TokenExisted(name string) bool
	CopyState() (*State, error)
	AddAccount(index common.AccountName, addr common.Address, timeStamp int64) (*Account, error)
	SetContract(index common.AccountName, t types.VmType, des, code []byte) error
	GetContract(index common.AccountName) (*types.DeployInfo, error)
	StoreSet(index common.AccountName, key, value []byte) (err error)
	StoreGet(index common.AccountName, key []byte) (value []byte, err error)
	GetAccountByName(index common.AccountName) (*Account, error)
	GetAccountByAddr(addr common.Address) (*Account, error)
	GetHashRoot() common.Hash
	CommitToDB() error
	Reset(hash common.Hash) error
}
