// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package transaction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/geneses"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/core/trie"
	"github.com/ecoball/go-ecoball/core/types"
	dsnstore "github.com/ecoball/go-ecoball/dsn/host/block"
	//shardCommon "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"github.com/ecoball/go-ecoball/smartcontract"
	"github.com/ecoball/go-ecoball/smartcontract/context"
	"github.com/ecoball/go-ecoball/spectator/connect"
	"github.com/ecoball/go-ecoball/spectator/info"
	"math/big"
	"time"
)

const keyLastCm = "lastCmHeader"
const keyLastMinor = "lastMinorHeader"
const keyLastFinal = "lastFinalHeader"
const keyLastVC = "lastVCHeader"

var log = elog.NewLogger("Chain Tx", elog.NoticeLog)

type StateDatabase struct {
	FinalDB *state.State //final database in levelDB
	TempDB  *state.State //temp database used for tx pool pre-handle transaction
}

type LastHeaders struct {
	Finalizer   bool
	CmHeader    *shard.CMBlockHeader
	MinorHeader *shard.MinorBlockHeader
	FinalHeader *shard.FinalBlockHeader
	VCHeader    *shard.ViewChangeBlockHeader
}

type ChainTx struct {
	ledger          ledger.Ledger
	BlockStoreCache store.Storage
	BlockStore      store.Storage
	HeaderStore     store.Storage
	MapStore        store.Storage
	CurrentHeader   *types.Header
	Geneses         *types.Header
	StateDB         StateDatabase

	LastHeader LastHeaders
	shardId    uint32
}

func NewTransactionChain(path string, ledger ledger.Ledger, shard bool) (c *ChainTx, err error) {
	c = &ChainTx{ledger: ledger}
	if config.DsnStorage {
		c.BlockStore, err = dsnstore.NewDsnStore(path + config.StringBlock)
	} else {
		c.BlockStore, err = store.NewLevelDBStore(path+config.StringBlock, 0, 0)
		c.BlockStoreCache, err = store.NewLevelDBStore(path+config.StringBlockCache, 0, 0)
	}
	if err != nil {
		return nil, err
	}
	c.HeaderStore, err = store.NewLevelDBStore(path+config.StringHeader, 0, 0)
	if err != nil {
		return nil, err
	}
	c.MapStore, err = store.NewLevelDBStore(path+config.StringTxs, 0, 0)
	if err != nil {
		return nil, err
	}

	if shard {
		existed, err := c.RestoreCurrentShardHeader()
		if err != nil {
			return nil, err
		}
		if existed {
			if c.StateDB.FinalDB, err = state.NewState(path+config.StringState, c.LastHeader.FinalHeader.StateHashRoot); err != nil {
				return nil, err
			}
		} else {
			if c.StateDB.FinalDB, err = state.NewState(path+config.StringState, common.Hash{}); err != nil {
				return nil, err
			}
		}
	} else {
		existed, err := c.RestoreCurrentHeader()
		if err != nil {
			return nil, err
		}
		if existed {
			if c.StateDB.FinalDB, err = state.NewState(path+config.StringState, c.CurrentHeader.StateHash); err != nil {
				return nil, err
			}
		} else {
			if c.StateDB.FinalDB, err = state.NewState(path+config.StringState, common.Hash{}); err != nil {
				return nil, err
			}
		}
	}

	c.StateDB.FinalDB.Type = state.FinalType

	event.InitMsgDispatcher()

	return c, nil
}

/**
*  @brief  create a new block, this function will execute the transaction to rebuild mpt trie
*  @param  consensusData - the data of consensus module set
 */
func (c *ChainTx) NewBlock(ledger ledger.Ledger, txs []*types.Transaction, consensusData types.ConsensusData, timeStamp int64) (*types.Block, []*types.Transaction, error) {
	//// every 30 blocks issue reward
	//if ledger.GetCurrentHeight(config.ChainHash) % 30 == 0 {
	//	c.StateDB.FinalDB.IssueToken(common.NameToIndex("saving"), big.NewInt(100), state.AbaToken)
	//
	//	produces, err := ledger.GetProducerList(config.ChainHash)
	//	if err != nil {
	//		fmt.Println(err)
	//		return nil, err
	//	}
	//
	//	value := 100 / len(produces)
	//	for _, producer := range produces {
	//		c.StateDB.FinalDB.IssueToken(producer, big.NewInt(int64(value)), state.AbaToken)
	//	}
	//}

	s, err := c.StateDB.FinalDB.CopyState()
	if err != nil {
		return nil, nil, err
	}
	s.Type = state.CopyType

	var cpu, net float64
	for i := 0; i < len(txs); i++ {
		log.Notice("Handle Transaction:", txs[i].Type.String(), txs[i].Hash.HexString(), " in Copy DB")
		if _, cp, n, err := c.HandleTransaction(s, txs[i], timeStamp, c.CurrentHeader.Receipt.BlockCpu, c.CurrentHeader.Receipt.BlockNet); err != nil {
			log.Warn(txs[i].JsonString())
			event.Send(event.ActorLedger, event.ActorTxPool, message.DeleteTx{ChainID: txs[i].ChainID, Hash: txs[i].Hash})
			txs = append(txs[:i], txs[i+1:]...)
			return nil, txs, err
		} else {
			cpu += cp
			net += n
		}
	}
	block, err := types.NewBlock(c.CurrentHeader.ChainID, c.CurrentHeader, s.GetHashRoot(), consensusData, txs, cpu, net, timeStamp)
	if err != nil {
		//c.ResetStateDB(c.CurrentHeader)
		return nil, nil, err
	}
	return block, nil, nil
}

/**
*  @brief  if create a new block failed, then need to reset state DB
*  @param  hash - the root hash of mpt trie which need to reset
 */
func (c *ChainTx) ResetStateDB(header *types.Header) error {
	c.CurrentHeader = header
	return c.StateDB.FinalDB.Reset(header.StateHash)
}

/**
*  @brief  check block's signature and all transactions
*  @param  block - the block need to verify
 */
func (c *ChainTx) VerifyTxBlock(block *types.Block) error {
	result, err := block.VerifySignature()
	if err != nil {
		log.Error("Block VerifySignature Failed")
		return err
	}
	if result == false {
		return errors.New(log, "block verify signature failed")
	}
	for _, v := range block.Transactions {
		if err := c.CheckTransaction(v); err != nil {
			log.Warn(v.JsonString())
			return err
		}
	}
	return nil
}

/**
*  @brief  save a block into levelDB, then push this block to p2p and tx pool module, and commit mpt trie into levelDB
*  @param  block - the block need to save
 */
func (c *ChainTx) SaveBlock(block *types.Block) error {
	if block == nil {
		return errors.New(log, "block is nil")
	}
	//check block is existed
	if c.blockExisted(block.Hash) {
		log.Warn("the block:", block.Height, "is existed")
		return nil
	}

	stateHashRoot := c.StateDB.FinalDB.GetHashRoot()
	for i := 0; i < len(block.Transactions); i++ {
		log.Notice("Handle Transaction:", block.Transactions[i].Type.String(), block.Transactions[i].Hash.HexString(), " in final DB")
		if _, _, _, err := c.HandleTransaction(c.StateDB.FinalDB, block.Transactions[i], block.TimeStamp, c.CurrentHeader.Receipt.BlockCpu, c.CurrentHeader.Receipt.BlockNet); err != nil {
			log.Warn(block.Transactions[i].JsonString())
			c.StateDB.FinalDB.Reset(stateHashRoot)
			return err
		}
	}
	if block.Height != 1 {
		connect.Notify(info.InfoBlock, block)
		if err := event.Publish(event.ActorLedger, block, event.ActorTxPool, event.ActorP2P); err != nil {
			log.Warn(err)
		}
	}

	for _, t := range block.Transactions {
		c.MapStore.BatchPut(t.Hash.Bytes(), block.Hash.Bytes())
	}
	if err := c.MapStore.BatchCommit(); err != nil {
		c.StateDB.FinalDB.Reset(stateHashRoot)
		return err
	}
	if c.StateDB.FinalDB.GetHashRoot().HexString() != block.StateHash.HexString() {
		log.Warn(block.JsonString(true))
		c.StateDB.FinalDB.Reset(stateHashRoot)
		return errors.New(log, fmt.Sprintf("hash mismatch:%s, %s", c.StateDB.FinalDB.GetHashRoot().HexString(), block.Hash.HexString()))
	}

	payload, err := block.Header.Serialize()
	if err != nil {
		c.StateDB.FinalDB.Reset(stateHashRoot)
		return err
	}
	if err := c.HeaderStore.Put(block.Header.Hash.Bytes(), payload); err != nil {
		c.StateDB.FinalDB.Reset(stateHashRoot)
		return err
	}
	payload, err = block.Serialize()
	if err != nil {
		c.StateDB.FinalDB.Reset(stateHashRoot)
		return err
	}
	c.BlockStore.BatchPut(block.Hash.Bytes(), payload)
	if err := c.BlockStore.BatchCommit(); err != nil {
		c.StateDB.FinalDB.Reset(stateHashRoot)
		return err
	}
	c.StateDB.FinalDB.CommitToDB()
	log.Debug("block state:", block.Height, block.StateHash.HexString())
	log.Notice(block.JsonString(false))
	c.CurrentHeader = block.Header

	//c.lockBlock.Lock()
	//c.BlockMap[block.Hash.HexString()] = BlockCache{Height: block.Height}
	//if data, err := json.Marshal(c.BlockMap); err != nil {
	//	return errors.New(log, err.Error())
	//} else {
	//	c.HeaderStore.Put([]byte(keyBlockMap), data)
	//}
	//c.lockBlock.Unlock()
	c.MapStore.Put(common.Uint64ToBytes(block.Height), block.Hash.Bytes())

	return nil
}

/**
*  @brief  return the highest block's hash
 */
func (c *ChainTx) GetTailBlockHash() common.Hash {
	return c.CurrentHeader.Hash
}

/**
*  @brief  get a block by a hash value
*  @param  hash - the block's hash need to return
 */
func (c *ChainTx) GetBlock(hash common.Hash) (*types.Block, error) {
	dataBlock, err := c.BlockStore.Get(hash.Bytes())
	if err != nil {
		return nil, errors.New(log, fmt.Sprintf("GetBlock error:%s", err.Error()))
	}
	block := new(types.Block)
	if err := block.Deserialize(dataBlock); err != nil {
		return nil, err
	}
	return block, nil
}

/**
*  @brief  get a block by a height value
*  @param  height - the block's height need to return
 */
func (c *ChainTx) GetBlockByHeight(height uint64) (*types.Block, error) {
	headers, err := c.HeaderStore.SearchAll()
	if err != nil {
		return nil, err
	}
	if len(headers) == 0 {
		return nil, nil
	}
	log.Debug("The geneses block is existed:", len(headers))
	var hash common.Hash
	for _, v := range headers {
		header := new(types.Header)
		if err := header.Deserialize([]byte(v)); err != nil {
			return nil, err
		}
		if header.Height == height {
			hash = header.Hash
			break
		}
	}
	if hash.Equals(&common.Hash{}) {
		return nil, errors.New(log, fmt.Sprintf("can't find the block by height:%d", height))
	}
	return c.GetBlock(hash)
}

/**
*  @brief  create a genesis block with built-in account and contract, then save this block into block chain
 */
func (c *ChainTx) GenesesBlockInit(chainID common.Hash, addr common.Address) error {
	if c.CurrentHeader != nil {
		log.Debug("geneses block is existed:", c.CurrentHeader.Height)
		return nil
	}

	tm, err := time.Parse("02/01/2006 15:04:05 PM", "21/02/1990 00:00:00 AM")
	if err != nil {
		return err
	}
	timeStamp := tm.UnixNano()

	//TODO start
	SecondInMs := int64(1000)
	BlockIntervalInMs := int64(15000)
	timeStamp = int64((timeStamp*SecondInMs-SecondInMs)/BlockIntervalInMs) * BlockIntervalInMs
	timeStamp = timeStamp / SecondInMs
	//TODO end

	hash := common.NewHash([]byte("EcoBall Geneses Block"))
	conData := types.GenesesBlockInitConsensusData(timeStamp)

	if err := geneses.PresetContract(c.StateDB.FinalDB, timeStamp, addr); err != nil {
		return err
	}

	header, err := types.NewHeader(types.VersionHeader, chainID, 1, chainID, hash,
		c.StateDB.FinalDB.GetHashRoot(), *conData, bloom.Bloom{}, config.BlockCpuLimit, config.BlockNetLimit, timeStamp)
	if err != nil {
		return err
	}
	block := &types.Block{Header: header, CountTxs: 0, Transactions: nil}
	if err := c.VerifyTxBlock(block); err != nil {
		return err
	}
	if err := c.SaveBlock(block); err != nil {
		log.Error("Save geneses block error:", err)
		return err
	}
	return nil
}

/**
*  @brief  restore the highest block's header from levelDB
*  @return bool - if can't find block in levelDB, return false, otherwise return true
 */
func (c *ChainTx) RestoreCurrentHeader() (bool, error) {
	headers, err := c.HeaderStore.SearchAll()
	if err != nil {
		return false, err
	}
	if len(headers) == 0 {
		return false, nil
	}
	log.Info("The geneses block is existed:", len(headers))
	var h uint64 = 0
	for _, v := range headers {
		header := new(types.Header)
		if err := header.Deserialize([]byte(v)); err != nil {
			return false, err
		}
		if header.Height > h {
			h = header.Height
			c.CurrentHeader = header
		}
	}
	log.Info("the block height is:", h, "hash:", c.CurrentHeader.Hash.HexString())
	return true, nil
}

/**
*  @brief  get a transaction from levelDB by a hash
*  @param  key - the hash of transaction
 */
func (c *ChainTx) GetTransaction(hash common.Hash) (*types.Transaction, error) {
	blockKey, err := c.MapStore.Get(hash.Bytes())
	if err != nil {
		return nil, err
	}
	blockData, err := c.BlockStore.Get(blockKey)
	block := types.Block{}
	if err := block.Deserialize(blockData); err != nil {
		return nil, err
	}
	return block.GetTransaction(hash)
}

/**
*  @brief  validity check of transaction, include signature verify, duplicate check and balance check
*  @param  tx - a transaction
 */
func (c *ChainTx) CheckTransaction(tx *types.Transaction) (err error) {
	if err := c.StateDB.FinalDB.CheckPermission(tx.From, tx.Permission, tx.Hash, tx.Signatures); err != nil {
		return err
	}
	if data, _ := c.MapStore.Get(tx.Hash.Bytes()); data != nil {
		return errors.New(log, errors.ErrDuplicatedTx.ErrorInfo())
	}
	switch tx.Type {
	case types.TxTransfer:
		if value, err := c.AccountGetBalance(tx.From, state.AbaToken); err != nil {
			return err
		} else if value.Sign() <= 0 {
			return errors.New(log, errors.ErrDoubleSpend.ErrorInfo())
		}
	case types.TxDeploy:
	case types.TxInvoke:
	default:
		return errors.New(log, "check transaction unknown tx type")
	}

	return nil
}
func (c *ChainTx) CheckTransactionWithDB(s *state.State, tx *types.Transaction) (err error) {
	if err := s.CheckPermission(tx.From, tx.Permission, tx.Hash, tx.Signatures); err != nil {
		return err
	}

	switch tx.Type {
	case types.TxTransfer:
		if data, _ := c.MapStore.Get(tx.Hash.Bytes()); data != nil {
			return errors.New(log, errors.ErrDuplicatedTx.ErrorInfo())
		}
		if value, err := s.AccountGetBalance(tx.From, state.AbaToken); err != nil {
			return err
		} else if value.Sign() <= 0 {
			log.Error(err)
			return errors.New(log, errors.ErrDoubleSpend.ErrorInfo())
		}
	case types.TxDeploy:
		if data, _ := c.MapStore.Get(tx.Addr.Bytes()); data != nil {
			return errors.New(log, errors.ErrDuplicatedTx.ErrorInfo())
		}
	case types.TxInvoke:
		if data, _ := c.MapStore.Get(tx.Hash.Bytes()); data != nil {
			return errors.New(log, errors.ErrDuplicatedTx.ErrorInfo())
		}
	default:
		return errors.New(log, "check transaction unknown tx type")
	}

	return nil
}
func (c *ChainTx) CheckPermission(index common.AccountName, name string, hash common.Hash, sig []common.Signature) error {
	return c.StateDB.FinalDB.CheckPermission(index, name, hash, sig)
}

/**
*  @brief  create a new account in mpt tree
*  @param  index - the uuid of account
*  @param  addr - the public key of account
 */
func (c *ChainTx) AccountAdd(index common.AccountName, addr common.Address, timeStamp int64) (*state.Account, error) {
	return c.StateDB.FinalDB.AddAccount(index, addr, timeStamp)
}
func (c *ChainTx) StoreSet(index common.AccountName, key, value []byte) (err error) {
	return c.StateDB.FinalDB.StoreSet(index, key, value)
}
func (c *ChainTx) StoreGet(index common.AccountName, key []byte) (value []byte, err error) {
	return c.StateDB.FinalDB.StoreGet(index, key)
}

//func (c *ChainTx) AddResourceLimits(from, to common.AccountName, cpu, net float32) error {
//	return c.StateDB.AddResourceLimits(from, to, cpu, net)
//}
func (c *ChainTx) SetContract(index common.AccountName, t types.VmType, des, code []byte, abi []byte) error {
	return c.StateDB.FinalDB.SetContract(index, t, des, code, abi)
}
func (c *ChainTx) GetContract(index common.AccountName) (*types.DeployInfo, error) {
	return c.StateDB.FinalDB.GetContract(index)
}
func (c *ChainTx) GetChainList() ([]state.Chain, error) {
	return c.StateDB.FinalDB.GetChainList()
}

/**
*  @brief  get the abi of contract
*  @param  indexAcc - the uuid of account
 */
func (c *ChainTx) GetContractAbi(index common.AccountName) ([]byte, error) {
	return c.StateDB.FinalDB.GetContractAbi(index)
}

func (c *ChainTx) AddPermission(index common.AccountName, perm state.Permission) error {
	return c.StateDB.FinalDB.AddPermission(index, perm)
}
func (c *ChainTx) FindPermission(index common.AccountName, name string) (string, error) {
	return c.StateDB.FinalDB.FindPermission(index, name)
}

/**
*  @brief  get a account's balance
*  @param  indexAcc - the uuid of account
*  @param  indexToken - the uuid of token
 */
func (c *ChainTx) AccountGetBalance(index common.AccountName, token string) (*big.Int, error) {
	return c.StateDB.FinalDB.AccountGetBalance(index, token)
}

/**
*  @brief  add a account's balance
*  @param  indexAcc - the uuid of account
*  @param  indexToken - the uuid of token
 */
func (c *ChainTx) AccountAddBalance(index common.AccountName, token string, value uint64) error {
	return c.StateDB.FinalDB.AccountAddBalance(index, token, new(big.Int).SetUint64(value))
}

/**
*  @brief  sub a account's balance
*  @param  indexAcc - the uuid of account
*  @param  indexToken - the uuid of token
 */
func (c *ChainTx) AccountSubBalance(index common.AccountName, token string, value uint64) error {
	return c.StateDB.FinalDB.AccountSubBalance(index, token, new(big.Int).SetUint64(value))
}

/**
*  @brief  handle transaction with transaction's type
*  @param  ledger - the interface of ledger impl
*  @param  tx - a transaction
 */
func (c *ChainTx) HandleTransaction(s *state.State, tx *types.Transaction, timeStamp int64, cpuLimit, netLimit float64) (ret []byte, cpu, net float64, err error) {
	start := time.Now().UnixNano()
	switch tx.Type {
	case types.TxTransfer:
		payload, ok := tx.Payload.GetObject().(types.TransferInfo)
		if !ok {
			event.PublishTrxRes(tx.Hash, "transaction type error[transfer]")
			return nil, 0, 0, errors.New(log, "transaction type error[transfer]")
		}
		if err := s.AccountSubBalance(tx.From, state.AbaToken, payload.Value); err != nil {
			event.PublishTrxRes(tx.Hash, err.Error())
			return nil, 0, 0, err
		}
		if err := s.AccountAddBalance(tx.Addr, state.AbaToken, payload.Value); err != nil {
			event.PublishTrxRes(tx.Hash, err.Error())
			return nil, 0, 0, err
		}

		//tx.Receipt.From.Balance, _ = s.AccountGetBalance(tx.From, state.AbaToken)
		//tx.Receipt.To.Balance, _ = s.AccountGetBalance(tx.Addr, state.AbaToken)
		tx.Receipt.TokenName = state.AbaToken
		tx.Receipt.Amount = payload.Value

	case types.TxDeploy:
		if err := s.CheckPermission(tx.Addr, state.Active, tx.Hash, tx.Signatures); err != nil {
			event.PublishTrxRes(tx.Hash, err.Error())
			return nil, 0, 0, err
		}
		payload, ok := tx.Payload.GetObject().(types.DeployInfo)
		if !ok {
			event.PublishTrxRes(tx.Hash, "transaction type error[deploy]")
			return nil, 0, 0, errors.New(log, "transaction type error[deploy]")
		}
		if err := s.SetContract(tx.Addr, payload.TypeVm, payload.Describe, payload.Code, payload.Abi); err != nil {
			event.PublishTrxRes(tx.Hash, err.Error())
			return nil, 0, 0, err
		}

		// generate trx receipt
		acc := state.Account{
			Index:    tx.Addr,
			Contract: payload,
		}
		if data, err := acc.Serialize(); err != nil {
			event.PublishTrxRes(tx.Hash, err.Error())
			return nil, 0, 0, err
		} else {
			tx.Receipt.Accounts[0] = data
		}

	case types.TxInvoke:
		actionNew, _ := types.NewAction(tx)
		trxContext, _ := context.NewTranscationContext(s, tx, cpuLimit, netLimit, timeStamp)
		ret, err = smartcontract.DispatchAction(trxContext, actionNew, 0)
		if err != nil {
			event.PublishTrxRes(tx.Hash, err.Error())
			return nil, 0, 0, errors.New(log, err.Error())
		}

		// update state change in trx receipt
		for i, acc := range trxContext.Accounts {
			tx.Receipt.Accounts[i] = trxContext.AccountDelta[acc]
		}

		js, err := json.Marshal(trxContext.Trace)
		if err != nil {
			event.PublishTrxRes(tx.Hash, err.Error())
			return nil, 0, 0, errors.New(log, err.Error())
		}
		//fmt.Println("json format: ", string(js))

		event.PublishTrxRes(tx.Hash, string(js))
	default:
		event.PublishTrxRes(tx.Hash, "the transaction's type error")
		return nil, 0, 0, errors.New(log, "the transaction's type error")
	}
	end := time.Now().UnixNano()
	if tx.Receipt.Cpu == 0 {
		cpu = float64(end-start) / 1000000.0
		tx.Receipt.Cpu = cpu
	} else {
		cpu = tx.Receipt.Cpu
	}
	data, err := tx.Serialize()
	if err != nil {
		event.PublishTrxRes(tx.Hash, err.Error())
		return nil, 0, 0, err
	}
	if tx.Receipt.Net == 0 {
		net = float64(len(data))
		tx.Receipt.Net = net
	} else {
		net = tx.Receipt.Net
	}
	if tx.Receipt.Hash.IsNil() {
		tx.Receipt.Hash = tx.Hash
	}
	if tx.Receipt.Result == nil {
		tx.Receipt.Result = common.CopyBytes(ret)
	}
	if err := s.RecoverResources(tx.From, timeStamp, cpuLimit, netLimit); err != nil {
		event.PublishTrxRes(tx.Hash, err.Error())
		return nil, 0, 0, err
	}
	if err := s.SubResources(tx.From, cpu, net, cpuLimit, netLimit); err != nil {
		event.PublishTrxRes(tx.Hash, err.Error())
		return nil, 0, 0, err
	}
	log.Debug("result:", ret, "cpu:", cpu, "us net:", net/1000, "byte")

	switch tx.Type {
	case types.TxTransfer:
		event.PublishTrxRes(tx.Hash, "transfer success!")
	case types.TxDeploy:
		event.PublishTrxRes(tx.Hash, "contract deploy success!")
	case types.TxInvoke:
		event.PublishTrxRes(tx.Hash, "contract invoke success!")
	default:
		event.PublishTrxRes(tx.Hash, "the transaction's type error")
	}

	return ret, cpu, net, nil
}

//ShardBlock
func (c *ChainTx) RestoreCurrentShardHeader() (result bool, err error) {
	data, err := c.HeaderStore.Get([]byte(keyLastCm))
	if err != nil {
		log.Warn("get last committee header error:", err)
	} else if data != nil {
		data, err = c.HeaderStore.Get(data)
		payload, err := shard.HeaderDeserialize(data)
		if err != nil {
			return false, err
		}
		header, ok := payload.GetObject().(shard.CMBlockHeader)
		if ok {
			result = true
			c.LastHeader.CmHeader = &header
		}
	}

	data, err = c.HeaderStore.Get([]byte(keyLastMinor))
	if err != nil {
		log.Warn("get last minor header error:", err)
	} else if data != nil {
		data, err = c.HeaderStore.Get(data)
		payload, err := shard.HeaderDeserialize(data)
		if err != nil {
			return false, err
		}
		header, ok := payload.GetObject().(shard.MinorBlockHeader)
		if ok {
			result = true
			c.LastHeader.MinorHeader = &header
			c.shardId = header.ShardId
		}
	}

	data, err = c.HeaderStore.Get([]byte(keyLastFinal))
	if err != nil {
		log.Warn("get last final header error:", err)
	} else if data != nil {
		data, err = c.HeaderStore.Get(data)
		payload, err := shard.HeaderDeserialize(data)
		if err != nil {
			return false, err
		}
		header, ok := payload.GetObject().(shard.FinalBlockHeader)
		if ok {
			result = true
			c.LastHeader.FinalHeader = &header
		}
	}

	data, err = c.HeaderStore.Get([]byte(keyLastVC))
	if err != nil {
		log.Warn("get last final header error:", err)
	} else if data != nil {
		data, err = c.HeaderStore.Get(data)
		payload, err := shard.HeaderDeserialize(data)
		if err != nil {
			return false, err
		}
		header, ok := payload.GetObject().(shard.ViewChangeBlockHeader)
		if ok {
			result = true
			c.LastHeader.VCHeader = &header
		}
	}

	blocks, err := c.BlockStoreCache.SearchAll()
	if err != nil {
		return result, err
	}
	log.Info("there are ", len(blocks), "block not finalizer")
	if len(blocks) == 0 {
		c.LastHeader.Finalizer = true
		return result, nil
	} else {
		done := true
		for hashes := range blocks {
			hash := common.BytesToHash([]byte(hashes))
			if c.LastHeader.MinorHeader.Hashes.Equals(&hash) {
				done = false
				break
			}
		}
		c.LastHeader.Finalizer = done
	}

	log.Debug("finalizer:", c.LastHeader.Finalizer)
	log.Debug("cm block:", c.LastHeader.CmHeader.JsonString())
	log.Debug("minor block:", c.LastHeader.MinorHeader.JsonString())
	log.Debug("final block:", c.LastHeader.FinalHeader.JsonString())
	return result, nil
}

/**
 *  @brief create the geneses block of shard, include cm block, minor block, final block and vc block
 *  @param chainID - the chain id
 *  @param addr - the root of this chain
 */
func (c *ChainTx) GenesesShardBlockInit(chainID common.Hash, addr common.Address) error {
	if c.LastHeader.CmHeader != nil || c.LastHeader.MinorHeader != nil || c.LastHeader.FinalHeader != nil || c.LastHeader.VCHeader != nil {
		log.Debug("geneses shard block is existed")
		return nil
	}

	tm, err := time.Parse("02/01/2006 15:04:05 PM", "21/02/1990 00:00:00 AM")
	if err != nil {
		return err
	}
	timeStamp := tm.UnixNano()

	prevHash := common.NewHash([]byte("EcoBall Geneses Block"))
	if err := geneses.PresetShardContract(c.StateDB.FinalDB, timeStamp, addr); err != nil {
		return err
	}

	headerCM := shard.CMBlockHeader{ //Init Committee Block
		ChainID:      chainID,
		Version:      types.VersionHeader,
		Height:       1,
		Timestamp:    timeStamp,
		PrevHash:     prevHash,
		LeaderPubKey: addr.Bytes(),
		Nonce:        0,
		Candidate:    shard.NodeInfo{},
		ShardsHash:   common.Hash{},
		COSign:       &types.COSign{Step1: 0, Step2: 0},
	}
	block, err := shard.NewCmBlock(headerCM, []shard.Shard{})
	if err := c.SaveShardBlock(block); err != nil {
		return err
	}
	c.LastHeader.CmHeader = &block.CMBlockHeader

	headerMinor := shard.MinorBlockHeader{ //Init Minor Block
		ChainID:           chainID,
		Version:           types.VersionHeader,
		Height:            1,
		Timestamp:         timeStamp,
		PrevHash:          prevHash,
		TrxHashRoot:       common.Hash{},
		StateDeltaHash:    c.StateDB.FinalDB.GetHashRoot(),
		CMBlockHash:       common.Hash{},
		ProposalPublicKey: nil,
		ShardId:           0,
		CMEpochNo:         headerCM.Height,
		Receipt:           types.BlockReceipt{BlockCpu: config.BlockCpuLimit, BlockNet: config.BlockNetLimit},
		COSign:            &types.COSign{},
	}
	if err := headerMinor.ComputeHash(); err != nil {
		return err
	}
	blockMinor := &shard.MinorBlock{
		MinorBlockHeader: headerMinor,
		Transactions:     nil,
		StateDelta:       nil,
	}
	if err := c.SaveShardBlock(blockMinor); err != nil {
		return err
	}
	c.LastHeader.MinorHeader = &blockMinor.MinorBlockHeader
	c.LastHeader.Finalizer = true

	headerFinal := shard.FinalBlockHeader{ //Init Final Block
		ChainID:            chainID,
		Version:            types.VersionHeader,
		Height:             1,
		Timestamp:          timeStamp,
		TrxCount:           0,
		PrevHash:           prevHash,
		ProposalPubKey:     nil,
		EpochNo:            headerCM.Height,
		CMBlockHash:        common.Hash{},
		TrxRootHash:        common.Hash{},
		StateDeltaRootHash: common.Hash{},
		MinorBlocksHash:    common.Hash{},
		StateHashRoot:      c.StateDB.FinalDB.GetHashRoot(),
		COSign:             &types.COSign{},
	}
	blockFinal, err := shard.NewFinalBlock(headerFinal, nil)
	if err != nil {
		return err
	}
	if err := c.SaveShardBlock(blockFinal); err != nil {
		return err
	}
	c.LastHeader.FinalHeader = &blockFinal.FinalBlockHeader

	headerVC := shard.ViewChangeBlockHeader{ //Init ViewChange Block
		ChainID:          chainID,
		Version:          types.VersionHeader,
		Height:           1,
		Timestamp:        timeStamp,
		PrevHash:         prevHash,
		CMEpochNo:        headerCM.Height,
		FinalBlockHeight: headerFinal.Height,
		Round:            0,
		Candidate:        shard.NodeInfo{},
		COSign:           &types.COSign{},
	}
	blockVC, err := shard.NewVCBlock(headerVC)
	if err != nil {
		return err
	}
	if err := c.SaveShardBlock(blockVC); err != nil {
		return err
	}
	c.LastHeader.VCHeader = &blockVC.ViewChangeBlockHeader
	return nil
}

/**
 *  @brief save the block into levelDB, the minor block just store, but not handle
 *  @param block - the interface of block
 */
func (c *ChainTx) SaveShardBlock(block shard.BlockInterface) (err error) {
	if block == nil {
		return errors.New(log, "the block is nil")
	}
	if c.blockExisted(block.Hash()) {
		log.Warn("the block:", shard.HeaderType(block.Type()).String(), "height:", block.GetHeight(), "hash:", block.Hash().HexString(), "is existed")
		return nil
	}

	stateHashRoot := c.StateDB.FinalDB.GetHashRoot() //used to reset mpt tire when the delta tx handle failed
	switch shard.HeaderType(block.Type()) {
	case shard.HeCmBlock:
		if cm, ok := block.GetObject().(shard.CMBlock); !ok {
			return errors.New(log, fmt.Sprintf("type asserts error:%s", shard.HeCmBlock.String()))
		} else {
			if err := c.HeaderStore.Put([]byte(keyLastCm), block.Hash().Bytes()); err != nil {
				return err
			}
			if err := c.storeBlock(0, &cm.CMBlockHeader, block, true); err != nil {
				return err
			}
			c.LastHeader.CmHeader = &cm.CMBlockHeader
			defer c.updateShardId()
		}

	case shard.HeMinorBlock:
		minor, ok := block.GetObject().(shard.MinorBlock)
		if !ok {
			return errors.New(log, fmt.Sprintf("type asserts error:%s", shard.HeMinorBlock.String()))
		}
		if c.shardId == minor.ShardId {
			if err := c.HeaderStore.Put([]byte(keyLastMinor), block.Hash().Bytes()); err != nil {
				return err
			}
			c.LastHeader.MinorHeader = &minor.MinorBlockHeader
		}
		for _, tx := range minor.Transactions {
			c.MapStore.BatchPut(tx.Hash.Bytes(), minor.Hashes.Bytes())
		}
		if err := c.MapStore.BatchCommit(); err != nil {
			return errors.New(log, err.Error())
		}
		if err := c.storeBlock(minor.ShardId, &minor.MinorBlockHeader, block, false); err != nil {
			return err
		}
	case shard.HeFinalBlock:
		final, ok := block.GetObject().(shard.FinalBlock)
		if !ok {
			return errors.New(log, fmt.Sprintf("type asserts error:%s", shard.HeFinalBlock.String()))
		}
		for _, minorHeader := range final.MinorBlocks { //Handle Minor Headers
			blockInterface, _, err := c.GetShardBlockByHash(shard.HeMinorBlock, minorHeader.Hashes, false)
			if err != nil {
				return err
			}
			minorBlock, ok := blockInterface.GetObject().(shard.MinorBlock)
			if !ok {
				return errors.New(log, "the type assertion failed")
			}
			if minorBlock.Hashes.Equals(&c.LastHeader.MinorHeader.Hashes) {
				c.LastHeader.Finalizer = true //set the minor block flag and persistence store for RestoreCurrentShardHeader
				//c.MapStore.Put([]byte(keyLastFinalizer), common.Uint64ToBytes(minorHeader.Height))
			}
			for _, delta := range minorBlock.StateDelta { //Handle tx or delta
				if tx, err := minorBlock.GetTransaction(delta.Receipt.Hash); err != nil {
					return err
				} else {
					if err := c.HandleDeltaState(c.StateDB.FinalDB, delta, tx, minorBlock.MinorBlockHeader.Timestamp, c.LastHeader.MinorHeader.Receipt.BlockCpu, c.LastHeader.MinorHeader.Receipt.BlockNet); err != nil {
						c.StateDB.FinalDB.Reset(stateHashRoot)
						return err
					}
				}
			}
			if err := c.recombinationBlockStore(&minorBlock); err != nil {
				return err
			}
		}

		if final.StateHashRoot != c.StateDB.FinalDB.GetHashRoot() { //check mpt tire's hash
			log.Warn(common.JsonString(c.StateDB.FinalDB.Params), common.JsonString(c.StateDB.FinalDB.Accounts))
			return errors.New(log, fmt.Sprintf("the final block state hash root is not eqaul, receive:%s, local:%s", final.StateHashRoot.HexString(), c.StateDB.FinalDB.GetHashRoot().HexString()))
		}

		c.LastHeader.FinalHeader = &final.FinalBlockHeader
		if err := c.HeaderStore.Put([]byte(keyLastFinal), block.Hash().Bytes()); err != nil {
			return err
		}
		if err := c.storeBlock(0, &final.FinalBlockHeader, block, true); err != nil {
			return err
		}
	case shard.HeViewChange:
		vc, ok := block.GetObject().(shard.ViewChangeBlock)
		if !ok {
			return errors.New(log, fmt.Sprintf("type asserts error:%s", shard.HeViewChange.String()))
		}

		c.LastHeader.VCHeader = &vc.ViewChangeBlockHeader
		if err := c.HeaderStore.Put([]byte(keyLastVC), block.Hash().Bytes()); err != nil {
			return err
		}
		if err := c.storeBlock(0, &vc.ViewChangeBlockHeader, block, true); err != nil {
			return err
		}
	default:
		return errors.New(log, fmt.Sprintf("unknown header type:%d", block.Type()))
	}

	/*log.Debug("store block into levelDB")
	if err := c.HeaderStore.Put(block.Hash().Bytes(), headerData); err != nil {
		return err
	}

	payload, err := shard.Serialize(block)
	if err != nil {
		return err
	}
	if err := c.BlockStore.Put(block.Hash().Bytes(), payload); err != nil {
		return err
	}

	log.Debug("commit to mpt")
	c.StateDB.FinalDB.CommitToDB()

	log.Debug("set block cache data")
	cacheKey := pb.BlockCacheKey{Type: block.Type(), ShardID: blockCache.ShardID, Height: blockCache.Height}
	key, err := cacheKey.Marshal()
	if err != nil {
		return errors.New(log, err.Error())
	}
	cacheData, err := blockCache.Marshal()
	if err != nil {
		return errors.New(log, err.Error())
	}
	log.Debug("commit to block map")
	c.MapStore.Put(key, cacheData)

	log.Notice("save "+shard.HeaderType(block.Type()).String()+" block", block.JsonString())
	if block.GetHeight() != 1 {
		go connect.Notify(info.ShardBlock, block)
		if err := event.Publish(event.ActorLedger, block, event.ActorTxPool); err != nil {
			log.Warn(err)
		}
	}
	log.Notice("Shard ", c.shardId, "Save", shard.HeaderType(block.Type()).String(), "Height", block.GetHeight(), "State Hash:", c.StateDB.FinalDB.GetHashRoot().HexString())
	*/
	return nil
}

func (c *ChainTx) storeBlock(shardID uint32, header shard.HeaderInterface, block shard.BlockInterface, finalizer bool) error {
	if headerData, err := shard.Serialize(header); err != nil {
		return err
	} else {
		if err := c.HeaderStore.Put(block.Hash().Bytes(), headerData); err != nil {
			return err
		}
	}

	if blockBody, err := shard.Serialize(block); err != nil {
		return err
	} else {
		if finalizer {
			if err := c.BlockStore.Put(block.Hash().Bytes(), blockBody); err != nil {
				return err
			}
			//c.BlockStoreCache.Delete(block.Hash().Bytes())
		} else {
			if err := c.BlockStoreCache.Put(block.Hash().Bytes(), blockBody); err != nil {
				return err
			}
		}
	}
	c.StateDB.FinalDB.CommitToDB()

	blockCache := pb.BlockCache{ShardID: shardID, Finalizer: false, Height: block.GetHeight(), HeaderType: block.Type(), Hash: block.Hash().Bytes()} //used to store the map of height and hash
	cacheKey := pb.BlockCacheKey{Type: block.Type(), ShardID: shardID, Height: block.GetHeight()}
	key, err := cacheKey.Marshal()
	if err != nil {
		return errors.New(log, err.Error())
	}
	cacheData, err := blockCache.Marshal()
	if err != nil {
		return errors.New(log, err.Error())
	}
	c.MapStore.Put(key, cacheData)

	go connect.Notify(info.ShardBlock, block)
	if err := event.Publish(event.ActorLedger, block, event.ActorTxPool); err != nil {
		log.Warn(err)
	}

	log.Notice("Shard ", c.shardId, "Save", shard.HeaderType(block.Type()).String(), "Height", block.GetHeight(), "State Hash:", c.StateDB.FinalDB.GetHashRoot().HexString(), "block:", block.JsonString())

	return nil
}

func (c *ChainTx) recombinationBlockStore(block shard.BlockInterface) error {
	if blockBody, err := shard.Serialize(block); err != nil {
		return err
	} else {
		if err := c.BlockStore.Put(block.Hash().Bytes(), blockBody); err != nil {
			return err
		}
		c.BlockStoreCache.Delete(block.Hash().Bytes())
	}
	return nil
}

/**
 *  @brief get the shard block by hash
 *  @param typ - the type of block
 *  @param hash - the hash of block
 *  @param expectFinalize - if the block has a high probability of finalizer, get it first from BlockStore, otherwise get it first from BlockStoreCache
 */
func (c *ChainTx) GetShardBlockByHash(typ shard.HeaderType, hash common.Hash, expectFinalize bool) (block shard.BlockInterface, finalizer bool, err error) {
	if expectFinalize {
		dataBlock, err := c.BlockStore.Get(hash.Bytes())
		if err != nil {
			if err == store.ErrNotFound {
				dataBlock, err = c.BlockStoreCache.Get(hash.Bytes())
				if err != nil {
					return nil, false, err
				}
				finalizer = false
			} else {
				log.Warn(hash.HexString(), typ.String())
				return nil, false, err
			}
		} else {
			finalizer = true
		}
		block, err = shard.BlockDeserialize(dataBlock)
		if err != nil {
			return nil, false, err
		}
	} else {
		dataBlock, err := c.BlockStoreCache.Get(hash.Bytes())
		if err != nil {
			if err == store.ErrNotFound {
				dataBlock, err = c.BlockStore.Get(hash.Bytes())
				if err != nil {
					return nil, false, err
				}
				finalizer = true
			} else {
				log.Warn(hash.HexString(), typ.String())
				return nil, false, err
			}
		} else {
			finalizer = false
		}
		block, err = shard.BlockDeserialize(dataBlock)
		if err != nil {
			return nil, false, err
		}
	}
	return block, finalizer, nil
}

/**
 *  @brief get the shard block by height, if the block type is not minor block, the shardID is 0
 *  @param typ - the type of block
 *  @param height - the height of block
 *  @param shardID - the shardId of minor block
 */
func (c *ChainTx) GetShardBlockByHeight(typ shard.HeaderType, height uint64, shardID uint32) (shard.BlockInterface, bool, error) {
	keyPb := pb.BlockCacheKey{Type: uint32(typ), ShardID: shardID, Height: height}
	key, err := keyPb.Marshal()
	if err != nil {
		return nil, false, errors.New(log, err.Error())
	}
	data, err := c.MapStore.Get(key)
	if err != nil {
		return nil, false, errors.New(log, err.Error())
	}
	var blockCachePb pb.BlockCache
	if err := blockCachePb.Unmarshal(data); err != nil {
		return nil, false, errors.New(log, err.Error())
	}
	return c.GetShardBlockByHash(typ, common.NewHash(blockCachePb.Hash), true)
}

/**
 *  @brief get the final block by epochNo
 *  @param epochNo - the epochNo of block
 */
/*
func (c *ChainTx) GetFinalBlocksByEpochNo(epochNo uint64) (finalBlocks []shard.BlockInterface, num int, err error) {
	if epochNo < 1 {
		return nil, 0, errors.New(log, fmt.Sprintf("epochNo is invalid, %d < 1", epochNo))
	}

	// calculate height range in epoch
	var heightMax, heightMin uint64
	if epochNo > 1 {
		heightMin = (epochNo-2)*shardCommon.DefaultEpochFinalBlockNumber + 1
		heightMax = (epochNo - 1) * shardCommon.DefaultEpochFinalBlockNumber
	} else {
		heightMin = 0
		heightMax = 0
	}

	c.lockBlock.RLock()
	defer c.lockBlock.RUnlock()
	num = 0

	c.lockBlock.RLock()
	for k, v := range c.BlockMap {
		if v.Type == shard.HeFinalBlock {
			if v.Height >= heightMin && v.Height <= heightMax {
				block, err := c.GetShardBlockByHash(shard.HeFinalBlock, common.HexToHash(k))
				if err != nil {
					c.lockBlock.RUnlock()
					return nil, 0, errors.New(log, fmt.Sprintf("can't find this block:[type]%s, [height]%d", shard.HeFinalBlock, v.Height))
				}

				finalBlocks = append(finalBlocks, block)
				num++
			}
		}
	}
	c.lockBlock.RUnlock()

	return finalBlocks, num, nil
}
*/
/**
 *  @brief get the last shard block by type, the minor block is local shard
 *  @param typ - the type of block
 */
func (c *ChainTx) GetLastShardBlock(typ shard.HeaderType) (shard.BlockInterface, bool, error) {
	switch typ {
	case shard.HeFinalBlock:
		if c.LastHeader.FinalHeader != nil {
			return c.GetShardBlockByHash(typ, c.LastHeader.FinalHeader.Hashes, true)
		}
	case shard.HeMinorBlock:
		if c.LastHeader.MinorHeader != nil {
			return c.GetShardBlockByHash(typ, c.LastHeader.MinorHeader.Hashes, true)
		}
	case shard.HeCmBlock:
		if c.LastHeader.CmHeader != nil {
			return c.GetShardBlockByHash(typ, c.LastHeader.CmHeader.Hashes, true)
		}
	case shard.HeViewChange:
		if c.LastHeader.VCHeader != nil {
			return c.GetShardBlockByHash(typ, c.LastHeader.VCHeader.Hashes, true)
		}
	default:
		return nil, false, errors.New(log, fmt.Sprintf("unknown block type:%d", typ))
	}
	return nil, false, errors.New(log, "can't find the last block")
}

/**
 *  @brief create a new minor block
 *  @param timeStamp - the time of block create
 *  @param txs - the transactions of block contains
 */
func (c *ChainTx) NewMinorBlock(txs []*types.Transaction, timeStamp int64) (*shard.MinorBlock, []*types.Transaction, error) {
	if c.LastHeader.Finalizer == false {
		lastMinor, finalizer, err := c.GetLastShardBlock(shard.HeMinorBlock)
		if err != nil {
			return nil, nil, err
		}
		if  finalizer != false {
			return nil, nil, errors.New(log, "the last minor block's finalizer state error")
		}
		block, _ := lastMinor.GetObject().(shard.MinorBlock)
		log.Info("new minor block:", block.GetHeight(), block.JsonString())
		return &block, nil, nil
	}

	s, err := c.StateDB.FinalDB.CopyState()
	if err != nil {
		return nil, nil, err
	}
	s.Type = state.CopyType
	var hashes []common.Hash
	var cpu, net float64
	for i := 0; i < len(txs); i++ {
		log.Notice("Handle Transaction:", txs[i].Type.String(), txs[i].Hash.HexString(), " in Copy DB")
		if _, cp, n, err := c.HandleTransaction(s, txs[i], timeStamp, c.LastHeader.MinorHeader.Receipt.BlockCpu, c.LastHeader.MinorHeader.Receipt.BlockNet); err != nil {
			log.Warn(txs[i].JsonString())
			event.Send(event.ActorLedger, event.ActorTxPool, message.DeleteTx{ChainID: txs[i].ChainID, Hash: txs[i].Hash})
			txs = append(txs[:i], txs[i+1:]...)
			return nil, txs, err
		} else {
			hashes = append(hashes, txs[i].Hash)
			cpu += cp
			net += n
		}
	}
	merkleHash, err := trie.GetMerkleRoot(hashes)
	if err != nil {
		return nil, nil, err
	}

	shardID, err := c.GetShardId()
	if err != nil {
		return nil, nil, err
	}
	header := shard.MinorBlockHeader{
		ChainID:           c.LastHeader.MinorHeader.ChainID,
		Version:           c.LastHeader.MinorHeader.Version,
		Height:            c.LastHeader.MinorHeader.Height + 1,
		Timestamp:         timeStamp,
		PrevHash:          c.LastHeader.MinorHeader.Hashes,
		TrxHashRoot:       merkleHash,
		StateDeltaHash:    s.GetHashRoot(),
		CMBlockHash:       c.LastHeader.CmHeader.Hashes,
		ProposalPublicKey: nil,
		ShardId:           shardID,
		CMEpochNo:         c.LastHeader.CmHeader.Height,
		Receipt:           types.BlockReceipt{},
		COSign:            &types.COSign{},
	}
	block, err := shard.NewMinorBlock(header, c.LastHeader.MinorHeader, txs, cpu, net)
	if err != nil {
		return nil, nil, err
	}
	c.LastHeader.Finalizer = false
	log.Info("new minor block:", block.GetHeight(), block.JsonString())
	return block, nil, nil
}

/**
 *  @brief create a new committee block
 *  @param timeStamp - the time of block create
 *  @param shards - the shard info
 */
func (c *ChainTx) NewCmBlock(timeStamp int64, shards []shard.Shard) (*shard.CMBlock, error) {
	/*
		if c.LastHeader.CmHeader.Height > 3 {
			// get cmBlock
			epochNo := c.LastHeader.CmHeader.Height - 3
			cmBlock, err := c.GetShardBlockByHeight(shard.HeCmBlock, epochNo, 0)
			if err != nil {
				return nil, err
			}

			finalBlocks, total, err := c.GetFinalBlocksByEpochNo(epochNo)
			if err != nil {
				return nil, err
			}

			shards := cmBlock.GetObject().(shard.CMBlock).Shards
			num := make([]int, len(shards) + 1)	// [0] is committee
			num[0] = total

			// calculate how many minorblock of shard had produced during this epoch
			for _, fBlock := range finalBlocks {
				for _, mBlock := range fBlock.GetObject().(shard.FinalBlock).MinorBlocks {
					num[mBlock.ShardId] += 1		// shard id start from 1
				}
			}

			// calculate how many minor block
			var totalNum int = 0
			for _, number := range num {
				totalNum += number
			}
			log.Debug("totalNum: ", totalNum)
			rewardEveryBlock := 10000 / totalNum		// reward(!0) every block

			// calculate reward every producer
			for i, s := range shards {
				reward := num[i + 1] * rewardEveryBlock / len(s.Member)		// reward(!0) every producer
				for _, member := range s.Member {
					log.Debug(string(member.PublicKey) + " get reward ", reward)
				}
			}
		}*/

	header := shard.CMBlockHeader{
		ChainID:      c.LastHeader.CmHeader.ChainID,
		Version:      c.LastHeader.CmHeader.Version,
		Height:       c.LastHeader.CmHeader.Height + 1,
		Timestamp:    timeStamp,
		PrevHash:     c.LastHeader.CmHeader.Hashes,
		LeaderPubKey: c.LastHeader.CmHeader.LeaderPubKey,
		Nonce:        c.LastHeader.CmHeader.Nonce + 1,
		Candidate:    shard.NodeInfo{PublicKey: nil, Address: "", Port: ""},
		ShardsHash:   common.Hash{},
		COSign:       &types.COSign{},
	}
	block, err := shard.NewCmBlock(header, shards)
	if err != nil {
		return nil, err
	}
	return block, nil
}

/**
 *  @brief create a final block, this func will exec the transactions in minor block to rebuild state hash
 *  @param timeStamp - the time of block create
 *  @param minorBlocks - the minor block
 */
func (c *ChainTx) newFinalBlock(timeStamp int64, minorBlocks []*shard.MinorBlock) (*shard.FinalBlock, error) {
	log.Debug("new final block")
	var TrxCount uint32
	var hashesTxs []common.Hash
	var hashesState []common.Hash
	var hashesMinor []common.Hash
	for _, m := range minorBlocks {
		hashesTxs = append(hashesTxs, m.TrxHashRoot)
		hashesState = append(hashesState, m.StateDeltaHash)
		hashesMinor = append(hashesMinor, m.Hashes)
	}
	TrxRootHash, err := trie.GetMerkleRoot(hashesTxs)
	if err != nil {
		return nil, err
	}
	StateDeltaRootHash, err := trie.GetMerkleRoot(hashesState)
	if err != nil {
		return nil, err
	}
	MinorBlocksHash, err := trie.GetMerkleRoot(hashesMinor)
	if err != nil {
		return nil, err
	}
	s, err := c.StateDB.FinalDB.CopyState()
	if err != nil {
		return nil, err
	}
	s.Type = state.CopyType
	var headers []*shard.MinorBlockHeader
	for _, block := range minorBlocks {
		TrxCount += uint32(len(block.Transactions))
		headers = append(headers, &block.MinorBlockHeader)
		for _, delta := range block.StateDelta {
			tx, err := block.GetTransaction(delta.Receipt.Hash)
			if err != nil {
				return nil, err
			}
			if err := c.HandleDeltaState(s, delta, tx, block.MinorBlockHeader.Timestamp,
				c.LastHeader.MinorHeader.Receipt.BlockCpu, c.LastHeader.MinorHeader.Receipt.BlockNet); err != nil {
				return nil, err
			}
		}
	}

	header := shard.FinalBlockHeader{
		ChainID:            c.LastHeader.FinalHeader.ChainID,
		Version:            c.LastHeader.FinalHeader.Version,
		Height:             c.LastHeader.FinalHeader.Height + 1,
		Timestamp:          timeStamp,
		TrxCount:           TrxCount,
		PrevHash:           c.LastHeader.FinalHeader.Hashes,
		ProposalPubKey:     nil,
		EpochNo:            c.LastHeader.CmHeader.Height,
		CMBlockHash:        c.LastHeader.CmHeader.Hashes,
		TrxRootHash:        TrxRootHash,
		StateDeltaRootHash: StateDeltaRootHash,
		MinorBlocksHash:    MinorBlocksHash,
		StateHashRoot:      s.GetHashRoot(),
		COSign:             &types.COSign{},
	}
	block, err := shard.NewFinalBlock(header, headers)
	if err != nil {
		return nil, err
	}
	log.Info("new final block:", block.Height, block.JsonString())
	return block, nil
}

/**
 *  @brief create a new final block, this func will read levelDB to get minor block
 *  @param timeStamp - the time of block create
 *  @param hashes - the minor blocks' hash of contains in final block
 */
func (c *ChainTx) NewFinalBlock(timeStamp int64, hashes []common.Hash) (*shard.FinalBlock, error) {
	var minorBlocks []*shard.MinorBlock
	for _, hash := range hashes {
		if b, _, err := c.GetShardBlockByHash(shard.HeMinorBlock, hash, false); err != nil {
			log.Warn(err)
		} else {
			if B, ok := b.GetObject().(shard.MinorBlock); ok {
				minorBlocks = append(minorBlocks, &B)
			} else {
				return nil, errors.New(log, "the type is error")
			}
		}
	}
	return c.newFinalBlock(timeStamp, minorBlocks)
}

/**
 *  @brief create a new view change block
 *  @param timeStamp - the time of block create
 *  @param round - the number of round
 */
func (c *ChainTx) NewViewChangeBlock(timeStamp int64, round uint16) (*shard.ViewChangeBlock, error) {
	header := shard.ViewChangeBlockHeader{
		ChainID:          c.LastHeader.VCHeader.ChainID,
		Version:          types.VersionHeader,
		Height:           c.LastHeader.VCHeader.Height + 1,
		Timestamp:        timeStamp,
		PrevHash:         c.LastHeader.VCHeader.Hashes,
		CMEpochNo:        c.LastHeader.CmHeader.Height,
		FinalBlockHeight: c.LastHeader.FinalHeader.Height,
		Round:            round,
		Candidate:        shard.NodeInfo{},
		COSign:           &types.COSign{},
	}
	block, err := shard.NewVCBlock(header)
	if err != nil {
		return nil, err
	}
	return block, nil
}

/**
 *  @brief update the local shard id
 *  @param block - the interface of block
 */
func (c *ChainTx) updateShardId() (uint32, error) {
	cm, _, err := c.GetLastShardBlock(shard.HeCmBlock)
	if err != nil {
		return 0, err
	}
	block, ok := cm.GetObject().(shard.CMBlock)
	if !ok {
		return 0, errors.New(log, "type error")
	}
	for index, s := range block.Shards {
		for _, node := range s.Member {
			if bytes.Equal(simulate.GetNodePubKey(), node.PublicKey) {
				c.shardId = uint32(index) + 1
				return uint32(index) + 1, nil
			}
		}
	}
	log.Warn(fmt.Sprintf("can't find the public key:%s", simulate.GetNodePubKey()))
	return 0, nil
}

/**
 *  @brief get the local shard id, the id will update when the node receive cm block
 *  @return the shard id
 */
func (c *ChainTx) GetShardId() (uint32, error) {
	if c.shardId == 0 {
		return c.updateShardId()
	} else {
		return c.shardId, nil
	}
}

/**
 *  @brief check the block's signature, hash, state hash and transaction
 *  @param block - the interface of block
 */
func (c *ChainTx) CheckShardBlock(block shard.BlockInterface) error {
	if c.blockExisted(block.Hash()) {
		return errors.New(log, fmt.Sprintf("the block %s is existed:%s-%d", shard.HeaderType(block.Type()).String(), block.Hash().HexString(), block.GetHeight()))
	}

	result, err := block.VerifySignature()
	if err != nil {
		log.Error("Block VerifySignature Failed")
		return err
	}
	if result == false {
		return errors.New(log, "block verify signature failed")
	}

	switch block.Type() {
	case uint32(shard.HeMinorBlock):
		//TODO:State Hash Check
		minorBlock, ok := block.GetObject().(shard.MinorBlock)
		if !ok {
			return errors.New(log, "the block type is not minor block")
		}

		for _, v := range minorBlock.Transactions { //Check Transaction
			if err := c.CheckTransaction(v); err != nil {
				return err
			}
		}
		newBlock, _, err := c.NewMinorBlock(minorBlock.Transactions, minorBlock.Timestamp) //check state hash
		if err != nil {
			return err
		}
		if !newBlock.StateDeltaHash.Equals(&minorBlock.StateDeltaHash) {
			return errors.New(log, fmt.Sprintf("the state hash is not equal:%s, %s", minorBlock.StateDeltaHash.HexString(), newBlock.StateDeltaHash.HexString()))
		}
	case uint32(shard.HeCmBlock):
	case uint32(shard.HeFinalBlock):
		//TODO:State Hash Check
		finalBlock, ok := block.GetObject().(shard.FinalBlock)
		if !ok {
			return errors.New(log, "block type error")
		}
		var hashes []common.Hash
		for _, v := range finalBlock.MinorBlocks {
			hashes = append(hashes, v.Hashes)
		}
		newBlock, err := c.NewFinalBlock(finalBlock.Timestamp, hashes)
		if err != nil {
			return err
		}
		if !newBlock.StateHashRoot.Equals(&finalBlock.StateHashRoot) {
			return errors.New(log, fmt.Sprintf("the state hash is not equal:%s, %s", finalBlock.StateHashRoot.HexString(), newBlock.StateHashRoot.HexString()))
		}
	case uint32(shard.HeViewChange):
	default:
		return errors.New(log, "unknown header type")
	}

	return nil
}

func (c *ChainTx) blockExisted(hash common.Hash) bool {
	if data, err := c.BlockStore.Get(hash.Bytes()); err != nil {
		if err == store.ErrNotFound {
			return false
		} else {
			panic(err)
		}
	} else if data == nil {
		return false
	} else {
		return true
	}
}

/**
 *  @brief handle tx's receipt to sync state trie, if the tx is contract invoke, then handle the tx
 *  @param s - the mpt trie object
 *  @param delta - the tx's receipt data
 *  @param tx - the transaction, used to contract invoke
 *  @param timeStamp - the timeStamp
 *  @param cpuLimit, netLimit - the limit of cpu and net
 */
func (c *ChainTx) HandleDeltaState(s *state.State, delta *shard.AccountMinor, tx *types.Transaction, timeStamp int64, cpuLimit, netLimit float64) (err error) {
	switch delta.Type {
	case types.TxTransfer:
		log.Info("handle delta in ", s.Type.String(), common.JsonString(delta))
		if err := s.AccountSubBalance(delta.Receipt.From, state.AbaToken, delta.Receipt.Amount); err != nil {
			return err
		}
		if err := s.AccountAddBalance(delta.Receipt.To, state.AbaToken, delta.Receipt.Amount); err != nil {
			return err
		}
		if err := s.RecoverResources(delta.Receipt.From, timeStamp, cpuLimit, netLimit); err != nil {
			return err
		}
		if err := s.SubResources(delta.Receipt.From, delta.Receipt.Cpu, delta.Receipt.Net, cpuLimit, netLimit); err != nil {
			return err
		}
	case types.TxDeploy:
		if len(delta.Receipt.Accounts) != 1 {
			return errors.New(log, "deploy delta's account len is not 1")
		}
		acc := new(state.Account)
		if err := acc.Deserialize(delta.Receipt.Accounts[0]); err != nil {
			return err
		}
		if err := s.SetContract(delta.Receipt.To, acc.Contract.TypeVm, acc.Contract.Describe, acc.Contract.Code, acc.Contract.Abi); err != nil {
			return err
		}
		if err := s.RecoverResources(delta.Receipt.From, timeStamp, cpuLimit, netLimit); err != nil {
			return err
		}
		if err := s.SubResources(delta.Receipt.From, delta.Receipt.Cpu, delta.Receipt.Net, cpuLimit, netLimit); err != nil {
			return err
		}
	case types.TxInvoke:
		/*if delta.Receipt.NewToken != nil {
			token := new(state.TokenInfo)
			if err := token.Deserialize(delta.Receipt.NewToken); err != nil {
				return err
			}
			if err := s.CommitToken(token); err != nil {
				return err
			}
		}
		for _, data := range delta.Receipt.Accounts {
			acc := new(state.Account)
			if err := acc.Deserialize(data); err != nil {
				return err
			}
			accState, err := s.GetAccountByName(acc.Index)
			if err != nil {
				return err
			}
			if accState == nil {
				accState = acc
			}
			if acc.Tokens != nil {
				for k, v := range acc.Tokens {
					accState.Tokens[k] = v
				}
			}
			if acc.Permissions != nil {
				for k, v := range acc.Permissions {
					accState.Permissions[k] = v
				}
			}
			if acc.Cpu.Limit != 0 {
				accState.Cpu.Limit = acc.Cpu.Limit
				//accState.Cpu.Available = acc.Cpu.Available
				accState.Cpu.Staked = acc.Cpu.Staked
				//accState.Cpu.Used = acc.Cpu.Used
				accState.Cpu.Delegated = acc.Cpu.Delegated

			}
			if acc.Net.Limit != 0 {
				accState.Net.Limit = acc.Net.Limit
				accState.Net.Delegated = acc.Net.Delegated
				accState.Net.Staked = acc.Net.Staked
				//accState.Net.Available = acc.Net.Available
			}
			if acc.TimeStamp != 0 {
				accState.TimeStamp = acc.TimeStamp
			}
			//if acc.Delegates
			s.CommitAccount(accState)
		}*/
		log.Info("handle tx in ", s.Type.String(), common.JsonString(delta))
		_, _, _, err := c.HandleTransaction(s, tx, timeStamp, cpuLimit, netLimit)
		if err != nil {
			return err
		}
	default:
		return errors.New(log, "unknown transaction type")
	}

	return nil
}
