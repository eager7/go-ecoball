package ledger

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
)

var L Ledger

type Ledger interface {
	GetTxBlock(chainID common.Hash, hash common.Hash) (*types.Block, error)
	NewTxBlock(chainID common.Hash, txs []*types.Transaction, consData types.ConsData, timeStamp int64) (*types.Block, []*types.Transaction, error)
	VerifyTxBlock(chainID common.Hash, block *types.Block) error
	GetTxBlockByHeight(chainID common.Hash, height uint64) (*types.Block, error)
	CheckTransaction(chainID common.Hash, tx *types.Transaction) error
	PreHandleTransaction(chainID common.Hash, s *state.State, tx *types.Transaction, timeStamp int64) (ret []byte, cpu, net float64, err error)
	GetCurrentHeader(chainID common.Hash) *types.Header
	GetCurrentHeight(chainID common.Hash) uint64
	StateDB(chainID common.Hash) *state.State
	ResetStateDB(chainID common.Hash, header *types.Header) error
	SetContract(chainID common.Hash, index common.AccountName, t types.VmType, des, code []byte, abi []byte) error
	GetContract(chainID common.Hash, index common.AccountName) (*types.DeployInfo, error)
	AccountGet(chainID common.Hash, index common.AccountName) (*state.Account, error)
	FindPermission(chainID common.Hash, index common.AccountName, name string) (string, error)
	CheckPermission(chainID common.Hash, index common.AccountName, name string, hash common.Hash, sig []common.Signature) error
	GetChainList(chainID common.Hash) ([]state.Chain, error)
	RequireResources(chainID common.Hash, index common.AccountName, timeStamp int64) (float64, float64, error)
	GetProducerList(chainID common.Hash) ([]state.Elector, error)
	AccountAddBalance(chainID common.Hash, index common.AccountName, token string, value uint64) error
	StoreGet(chainID common.Hash, index common.AccountName, key []byte) (value []byte, err error)
	//StoreSet(index common.AccountName, key, value []byte) error
	GetTokenInfo(chainID common.Hash, token string) (*state.TokenInfo, error)
	//TokenCreate(index common.AccountName, token string, maximum uint64) error
	//TokenIsExisted(token string) bool

	//GetGenesesTime() int64

	GetTransaction(chainID, transactionId common.Hash) (*types.Transaction, error)

}
