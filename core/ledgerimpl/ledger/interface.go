package ledger

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
)

var L Ledger

type Ledger interface {
	GetTxBlock(chainID common.Hash, hash common.Hash) (*types.Block, error)
	//区块的生产和存储具有时间线性关系,不可并发,如果调用接口需要维护两者在单线程中运行,否则请使用actor生成和存储
	NewTxBlock(chainID common.Hash, txs []*types.Transaction, consData types.ConsData, timeStamp int64) (*types.Block, []*types.Transaction, error)
	SaveTxBlock(chainID common.Hash, block *types.Block) error
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
	FindPermission(chainID common.Hash, index common.AccountName, name string) (string, error)
	CheckPermission(chainID common.Hash, index common.AccountName, name string, hash common.Hash, sig []common.Signature) error
	GetChainList(chainID common.Hash) ([]state.Chain, error)
	QueryResources(chainID common.Hash, index common.AccountName, timeStamp int64) (float64, float64, error)
	QueryAccountInfo(chainID common.Hash, index common.AccountName, timeStamp int64) (string, error)
	GetProducerList(chainID common.Hash) ([]state.Elector, error)
	AccountAddBalance(chainID common.Hash, index common.AccountName, token string, value uint64) error
	StoreGet(chainID common.Hash, index common.AccountName, key []byte) (value []byte, err error)
	//StoreSet(index common.AccountName, key, value []byte) error
	GetTokenInfo(chainID common.Hash, token string) (*state.TokenInfo, error)
	GetTransaction(chainID, transactionId common.Hash) (*types.Transaction, error)
}
