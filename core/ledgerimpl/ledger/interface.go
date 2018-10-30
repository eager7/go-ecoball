package ledger

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/core/shard"
)

var L Ledger

type Ledger interface {
	//GetShardBlock() (types.Payload, error)

	GetTxBlock(chainID common.Hash, hash common.Hash) (*types.Block, error)
	NewTxBlock(chainID common.Hash, txs []*types.Transaction, consensusData types.ConsensusData, timeStamp int64) (*types.Block, error)
	VerifyTxBlock(chainID common.Hash, block *types.Block) error
	//SaveTxBlock(block *types.Block) error
	GetTxBlockByHeight(chainID common.Hash, height uint64) (*types.Block, error)
	CheckTransaction(chainID common.Hash, tx *types.Transaction) error
	PreHandleTransaction(chainID common.Hash, tx *types.Transaction, timeStamp int64) (ret []byte, cpu, net float64, err error)
	GetCurrentHeader(chainID common.Hash) *types.Header
	GetCurrentHeight(chainID common.Hash) uint64
	StateDB(chainID common.Hash) *state.State
	ResetStateDB(chainID common.Hash, header *types.Header) error

	SetContract(chainID common.Hash, index common.AccountName, t types.VmType, des, code []byte, abi []byte) error
	GetContract(chainID common.Hash, index common.AccountName) (*types.DeployInfo, error)
	AccountGet(chainID common.Hash, index common.AccountName) (*state.Account, error)
	//AddPermission(index common.AccountName, perm state.Permission) error
	FindPermission(chainID common.Hash, index common.AccountName, name string) (string, error)
	CheckPermission(chainID common.Hash, index common.AccountName, name string, hash common.Hash, sig []common.Signature) error
	GetChainList(chainID common.Hash) ([]state.Chain, error)
	RequireResources(chainID common.Hash, index common.AccountName, timeStamp int64) (float64, float64, error)
	GetProducerList(chainID common.Hash) ([]common.AccountName, error)
	//AccountGetBalance(index common.AccountName, token string) (uint64, error)
	AccountAddBalance(chainID common.Hash, index common.AccountName, token string, value uint64) error
	//AccountSubBalance(index common.AccountName, token string, value uint64) error

	//AddResourceLimits(from, to common.AccountName, cpu, net float32) error
	StoreGet(chainID common.Hash, index common.AccountName, key []byte) (value []byte, err error)
	//StoreSet(index common.AccountName, key, value []byte) error

	GetTokenInfo(chainID common.Hash, token string) (*state.TokenInfo, error)
	//TokenCreate(index common.AccountName, token string, maximum uint64) error
	//TokenIsExisted(token string) bool

	//GetGenesesTime() int64
	GetChainTx(chainID common.Hash) ChainInterface

	GetTransaction(chainID, transactionId common.Hash) (*types.Transaction, error)

	ShardPreHandleTransaction(chainID common.Hash, tx *types.Transaction, timeStamp int64) (ret []byte, cpu, net float64, err error)
	SaveShardBlock(chainID common.Hash, block shard.BlockInterface) (err error)
	GetShardBlockByHash(chainID common.Hash, typ shard.HeaderType, hash common.Hash) (shard.BlockInterface, error)
	GetShardBlockByHeight(chainID common.Hash, typ shard.HeaderType, height uint64) (shard.BlockInterface, error)
	GetLastShardBlock(chainID common.Hash, typ shard.HeaderType) (shard.BlockInterface, error)
	GetLastShardBlockById(chainID common.Hash, shardId uint32) (shard.BlockInterface, error)
	NewCmBlock(chainID common.Hash, timeStamp int64, shards []shard.Shard) (*shard.CMBlock, error)
	NewMinorBlock(chainID common.Hash, txs []*types.Transaction, timeStamp int64) (*shard.MinorBlock, error)
	NewFinalBlock(chainID common.Hash, timeStamp int64, hashes []common.Hash) (*shard.FinalBlock, error)
	NewViewChangeBlock(chainID common.Hash, timeStamp int64, round uint16) (*shard.ViewChangeBlock, error)
	GetShardId(chainID common.Hash) (uint32, error)
}
