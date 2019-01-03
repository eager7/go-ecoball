package ledger

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/common/message/mpb"
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
	GetChainTx(chainID common.Hash) ChainInterface

	GetTransaction(chainID, transactionId common.Hash) (*types.Transaction, error)

	ShardPreHandleTransaction(chainID common.Hash, s *state.State, tx *types.Transaction, timeStamp int64) (ret []byte, cpu, net float64, err error)
	/**
	 *  @brief save the block into levelDB, the minor block just store, but not handle
	 *  @param chainID - the chain id
	 *  @param block - the interface of block
	 */
	SaveShardBlock(chainID common.Hash, block shard.BlockInterface) (err error)
	/**
	 *  @brief get the shard block by hash
	 *  @param typ - the type of block
	 *  @param hash - the hash of block
	 *  @param expectFinalize - if the block has a high probability of finalizer, get it first from BlockStore, otherwise get it first from BlockStoreCache
	 */
	GetShardBlockByHash(chainID common.Hash, typ mpb.Identify, hash common.Hash, finalizer bool) (shard.BlockInterface, bool, error)
	/**
	 *  @brief get the shard block by height, if the block type is not minor block, the shardID is 0
	 *  @param typ - the type of block
	 *  @param height - the height of block
	 *  @param shardID - the shardId of minor block
	 */
	GetShardBlockByHeight(chainID common.Hash, typ mpb.Identify, height uint64, shardID uint32) (shard.BlockInterface, bool, error)
	/**
	 *  @brief get the last shard block by type, the minor block is local shard
	 *  @param typ - the type of block
	 */
	GetLastShardBlock(chainID common.Hash, typ mpb.Identify) (shard.BlockInterface, bool, error)
	/**
	 *  @brief create a new committee block
	 *  @param timeStamp - the time of block create
	 *  @param shards - the shard info
	 */
	NewCmBlock(chainID common.Hash, timeStamp int64, shards []shard.Shard) (*shard.CMBlock, error)
	/**
	 *  @brief create a new minor block
	 *  @param timeStamp - the time of block create
	 *  @param txs - the transactions of block contains
	 */
	NewMinorBlock(chainID common.Hash, txs []*types.Transaction, timeStamp int64) (*shard.MinorBlock, []*types.Transaction, error)
	/**
	 *  @brief create a new final block, this func will read levelDB to get minor block
	 *  @param timeStamp - the time of block create
	 *  @param hashes - the minor blocks' hash of contains in final block
	 */
	NewFinalBlock(chainID common.Hash, timeStamp int64, hashes []common.Hash) (*shard.FinalBlock, error)
	/**
	 *  @brief create a new view change block
	 *  @param timeStamp - the time of block create
	 *  @param round - the number of round
	 */
	NewViewChangeBlock(chainID common.Hash, timeStamp int64, round uint16) (*shard.ViewChangeBlock, error)
	/**
	 *  @brief get the local shard id, the id will update when the node receive cm block
	 *  @return the shard id
	 */
	GetShardId(chainID common.Hash) (uint32, error)
	/**
	 *  @brief check the block's signature, hash, state hash and transaction
	 *  @param block - the interface of block
	 */
	CheckShardBlock(chainID common.Hash, block shard.BlockInterface) error
}
