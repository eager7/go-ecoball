package ledger

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
)

type Ledger interface {
	GetTxBlock(hash common.Hash) (*types.Block, error)
	NewTxBlock(txs []*types.Transaction, consensusData types.ConsensusData, timeStamp int64) (*types.Block, error)
	VerifyTxBlock(block *types.Block) error
	SaveTxBlock(block *types.Block) error
	GetTxBlockByHeight(height uint64) (*types.Block, error)
	CheckTransaction(tx *types.Transaction) error
	PreHandleTransaction(tx *types.Transaction, timeStamp int64) (ret []byte, cpu, net float64, err error)
	GetCurrentHeader() *types.Header
	GetCurrentHeight() uint64
	StateDB() *state.State
	ResetStateDB(header *types.Header) error

	AccountAdd(index common.AccountName, addr common.Address, timeStamp int64) (*state.Account, error)
	SetContract(index common.AccountName, t types.VmType, des, code []byte) error
	GetContract(index common.AccountName) (*types.DeployInfo, error)
	AccountGet(index common.AccountName) (*state.Account, error)
	AddPermission(index common.AccountName, perm state.Permission) error
	FindPermission(index common.AccountName, name string) (string, error)
	CheckPermission(index common.AccountName, name string, hash common.Hash, sig []common.Signature) error
	RequireResources(index common.AccountName, timeStamp int64) (float64, float64, error)
	GetProducerList() ([]common.AccountName, error)
	AccountGetBalance(index common.AccountName, token string) (uint64, error)
	AccountAddBalance(index common.AccountName, token string, value uint64) error
	AccountSubBalance(index common.AccountName, token string, value uint64) error

	//AddResourceLimits(from, to common.AccountName, cpu, net float32) error
	StoreGet(index common.AccountName, key []byte) ([]byte, error)
	StoreSet(index common.AccountName, key, value []byte) error

	TokenCreate(index common.AccountName, token string, maximum uint64) error
	TokenIsExisted(token string) bool
	Start()

	GetGenesesTime() int64

	GetChainTx() ChainInterface
}
