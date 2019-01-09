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
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/geneses"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/smartcontract"
	"github.com/ecoball/go-ecoball/smartcontract/context"
	"github.com/ecoball/go-ecoball/spectator/connect"
	"github.com/ecoball/go-ecoball/spectator/info"
	"github.com/gin-gonic/gin/json"
	"math/big"
	"time"
)

var log = elog.NewLogger("Chain Tx", elog.NoticeLog)

type ChainTx struct {
	ledger          ledger.Ledger
	CurrentHeader   LastHeader
	BlockStoreCache store.Storage
	BlockStore      store.Storage
	HeaderStore     store.Storage
	MapStore        store.Storage
	StateDB         *state.State
}

func NewTransactionChain(path string, ledger ledger.Ledger, option ...bool) (c *ChainTx, err error) {
	c = &ChainTx{ledger: ledger}

	c.BlockStore, err = store.NewLevelDBStore(path+config.StringBlock, 0, 0)
	c.BlockStoreCache, err = store.NewLevelDBStore(path+config.StringBlockCache, 0, 0)
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

	existed, err := c.RestoreCurrentHeader()
	if err != nil {
		return nil, err
	}
	if existed {
		if c.StateDB, err = state.NewState(path+config.StringState, c.CurrentHeader.Get().StateHash); err != nil {
			return nil, err
		}
	} else {
		if c.StateDB, err = state.NewState(path+config.StringState, common.Hash{}); err != nil {
			return nil, err
		}
	}

	c.StateDB.Type = state.FinalType

	//event.InitMsgDispatcher()

	return c, nil
}

/**
*  @brief  create a new block, this function will execute the transaction to rebuild mpt trie
*  @param  consensusData - the data of consensus module set
 */
func (c *ChainTx) NewBlock(ledger ledger.Ledger, txs []*types.Transaction, consensusData types.ConsData, timeStamp int64) (*types.Block, []*types.Transaction, error) {
	s, err := c.StateDB.StateCopy()
	if err != nil {
		return nil, nil, err
	}
	s.Type = state.CopyType

	var cpu, net float64
	for i := 0; i < len(txs); i++ {
		log.Notice("Handle Transaction:", txs[i].Type.String(), txs[i].Hash.HexString(), " in Copy DB")
		if _, cp, n, err := c.HandleTransaction(s, txs[i], timeStamp, c.CurrentHeader.Get().Receipt.BlockCpu, c.CurrentHeader.Get().Receipt.BlockNet); err != nil {
			log.Warn(txs[i].String())
			if err := event.Send(event.ActorLedger, event.ActorTxPool, message.DeleteTx{ChainID: txs[i].ChainID, Hash: txs[i].Hash}); err != nil {
				log.Warn("send transaction message failed:", err)
			}
			txs = append(txs[:i], txs[i+1:]...)
			return nil, txs, err
		} else {
			cpu += cp
			net += n
		}
	}
	block, err := types.NewBlock(c.CurrentHeader.Get().ChainID, c.CurrentHeader.Get(), s.GetHashRoot(), consensusData, txs, cpu, net, timeStamp)
	if err != nil {
		return nil, nil, err
	}
	return block, nil, nil
}

/**
*  @brief  if create a new block failed, then need to reset state DB
*  @param  hash - the root hash of mpt trie which need to reset
 */
func (c *ChainTx) ResetStateDB(header *types.Header) error {
	c.CurrentHeader.Set(header)
	return c.StateDB.Reset(header.StateHash)
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
		return errors.New("block verify signature failed")
	}
	for _, v := range block.Transactions {
		if err := c.CheckTransaction(v); err != nil {
			log.Warn(v.String())
			return err
		}
	}
	return nil
}

/**
*  @brief  save a block into levelDB, then push this block to p2p and tx pool module, and commit mpt trie into levelDB
*  @param  block - the block need to save
 */
func (c *ChainTx) SaveBlock(block *types.Block) (err error) {
	if block == nil {
		return errors.New("block is nil")
	}
	if c.blockExisted(block.Hash) {
		log.Warn("the block:", block.Height, "is existed")
		return nil
	}
	if c.CurrentHeader.Get() != nil && c.CurrentHeader.Get().Height+1 != block.Height {
		return errors.New(fmt.Sprintf("there maybe lost some blocks, the current block height is %d, the new block height is %d", c.CurrentHeader.Get().Height, block.Height))
	}

	stateHashRoot := c.StateDB.GetHashRoot()
	for i := 0; i < len(block.Transactions); i++ {
		log.Notice("Handle Transaction:", block.Transactions[i].Type.String(), block.Transactions[i].Hash.HexString(), " in final DB")
		if _, _, _, err = c.HandleTransaction(c.StateDB, block.Transactions[i], block.TimeStamp, c.CurrentHeader.Get().Receipt.BlockCpu, c.CurrentHeader.Get().Receipt.BlockNet); err != nil {
			log.Warn(block.Transactions[i].String())
			if err := c.StateDB.Reset(stateHashRoot); err != nil {
				log.Warn("reset state db failed:", err)
			}
			return err
		}
	}
	if c.StateDB.GetHashRoot().HexString() != block.StateHash.HexString() {
		log.Warn(block.String())
		if err := c.StateDB.Reset(stateHashRoot); err != nil {
			log.Warn("reset state db failed:", err)
		}
		return errors.New(fmt.Sprintf("hash mismatch:%s, %s", c.StateDB.GetHashRoot().HexString(), block.Hash.HexString()))
	}

	for _, t := range block.Transactions {
		c.MapStore.BatchPut(t.Hash.Bytes(), block.Hash.Bytes())
	}
	if err := c.MapStore.BatchCommit(); err != nil {
		if err := c.StateDB.Reset(stateHashRoot); err != nil {
			log.Warn("reset state db failed:", err)
		}
		return err
	}

	payload, err := block.Header.Serialize()
	if err != nil {
		if err := c.StateDB.Reset(stateHashRoot); err != nil {
			log.Warn("reset state db failed:", err)
		}
		return err
	}
	if err := c.HeaderStore.Put(block.Header.Hash.Bytes(), payload); err != nil {
		if err := c.StateDB.Reset(stateHashRoot); err != nil {
			log.Warn("reset state db failed:", err)
		}
		return err
	}
	payload, err = block.Serialize()
	if err != nil {
		if err := c.StateDB.Reset(stateHashRoot); err != nil {
			log.Warn("reset state db failed:", err)
		}
		return err
	}
	c.BlockStore.BatchPut(block.Hash.Bytes(), payload)
	if err := c.BlockStore.BatchCommit(); err != nil {
		if err := c.StateDB.Reset(stateHashRoot); err != nil {
			log.Warn("reset state db failed:", err)
		}
		return err
	}
	if err := c.StateDB.CommitToDB(); err != nil {
		if err := c.StateDB.Reset(stateHashRoot); err != nil {
			log.Warn("reset state db failed:", err)
		}
		return err
	}
	c.CurrentHeader.Set(block.Header)

	if err := c.MapStore.Put(common.Uint64ToBytes(block.Height), block.Hash.Bytes()); err != nil {
		if err := c.StateDB.Reset(stateHashRoot); err != nil {
			log.Warn("reset state db failed:", err)
		}
		return err
	}
	log.Notice("save block finished, height:", block.Height, "transaction number:", len(block.Transactions), block.Header.String())

	if block.Height != 1 {
		if err := connect.Notify(info.InfoBlock, block); err != nil {
			log.Warn("notify browser failed:", err)
		}
		if err := event.Send(event.ActorLedger, event.ActorTxPool, block); err != nil {
			log.Warn(err)
		}
		if err := event.Send(event.ActorLedger, event.ActorP2P, block); err != nil {
			log.Warn(err)
		}
	}
	return nil

}

/**
*  @brief  get a block by a hash value
*  @param  hash - the block's hash need to return
 */
func (c *ChainTx) GetBlock(hash common.Hash) (*types.Block, error) {
	dataBlock, err := c.BlockStore.Get(hash.Bytes())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("GetBlock error:%s", err.Error()))
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
		return nil, errors.New(fmt.Sprintf("can't find the block by height:%d", height))
	}
	return c.GetBlock(hash)
}

/**
*  @brief  create a genesis block with built-in account and contract, then save this block into block chain
 */
func (c *ChainTx) GenesesBlockInit(chainID common.Hash, addr common.Address) error {
	if c.CurrentHeader.Get() != nil {
		log.Debug("geneses block is existed:", c.CurrentHeader.Get().Height)
		return nil
	}

	tm, err := time.Parse("02/01/2006 15:04:05 PM", "21/02/1990 00:00:00 AM")
	if err != nil {
		return err
	}
	timeStamp := tm.UnixNano()
	hash := common.SingleHash([]byte("EcoBall Geneses Block"))
	conData, err := types.InitConsensusData(timeStamp)
	if err != nil {
		return err
	}
	if err := geneses.PresetContract(c.StateDB, timeStamp, addr); err != nil {
		return err
	}

	header := &types.Header{
		Version:    types.VersionHeader,
		ChainID:    chainID,
		TimeStamp:  timeStamp,
		Height:     1,
		ConsData:   *conData,
		PrevHash:   hash,
		MerkleHash: common.Hash{},
		StateHash:  c.StateDB.GetHashRoot(),
		Receipt:    types.BlockReceipt{BlockCpu: config.BlockCpuLimit, BlockNet: config.BlockNetLimit},
		Signatures: nil,
		Hash:       common.Hash{},
	}
	if err := header.ComputeHash(); err != nil {
		return err
	}
	if err := c.SaveBlock(&types.Block{Header: header}); err != nil {
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
			c.CurrentHeader.Set(header)
		}
	}
	log.Info("the block height is:", h, "hash:", c.CurrentHeader.Get().Hash.HexString())
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
	tx := block.GetTransaction(hash)
	if tx == nil {
		return nil, errors.New(fmt.Sprintf("can't find this tx:%s", hash.String()))
	}
	return tx, nil
}

/**
*  @brief  validity check of transaction, include signature verify, duplicate check and balance check
*  @param  tx - a transaction
 */
func (c *ChainTx) CheckTransaction(tx *types.Transaction) (err error) {
	if err := c.StateDB.CheckPermission(tx.From, tx.Permission, tx.Hash, tx.Signatures); err != nil {
		return err
	}
	if data, _ := c.MapStore.Get(tx.Hash.Bytes()); data != nil {
		return errors.New(errors.ErrDuplicatedTx.ErrorInfo())
	}
	switch tx.Type {
	case types.TxTransfer:
		if value, err := c.AccountGetBalance(tx.From, state.AbaToken); err != nil {
			return err
		} else if value.Sign() <= 0 {
			return errors.New(errors.ErrDoubleSpend.ErrorInfo())
		}
	case types.TxDeploy:
	case types.TxInvoke:
	default:
		return errors.New("check transaction unknown tx type")
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
			return errors.New(errors.ErrDuplicatedTx.ErrorInfo())
		}
		if value, err := s.AccountGetBalance(tx.From, state.AbaToken); err != nil {
			return err
		} else if value.Sign() <= 0 {
			log.Error(err)
			return errors.New(errors.ErrDoubleSpend.ErrorInfo())
		}
	case types.TxDeploy:
		if data, _ := c.MapStore.Get(tx.Addr.Bytes()); data != nil {
			return errors.New(errors.ErrDuplicatedTx.ErrorInfo())
		}
	case types.TxInvoke:
		if data, _ := c.MapStore.Get(tx.Hash.Bytes()); data != nil {
			return errors.New(errors.ErrDuplicatedTx.ErrorInfo())
		}
	default:
		return errors.New("check transaction unknown tx type")
	}

	return nil
}
func (c *ChainTx) CheckPermission(index common.AccountName, name string, hash common.Hash, sig []common.Signature) error {
	return c.StateDB.CheckPermission(index, name, hash, sig)
}

func (c *ChainTx) StoreSet(index common.AccountName, key, value []byte) (err error) {
	return c.StateDB.StoreSet(index, key, value)
}
func (c *ChainTx) StoreGet(index common.AccountName, key []byte) (value []byte, err error) {
	return c.StateDB.StoreGet(index, key)
}

//func (c *ChainTx) AddResourceLimits(from, to common.AccountName, cpu, net float32) error {
//	return c.StateDB.AddResourceLimits(from, to, cpu, net)
//}
func (c *ChainTx) SetContract(index common.AccountName, t types.VmType, des, code []byte, abi []byte) error {
	return c.StateDB.SetContract(index, t, des, code, abi)
}
func (c *ChainTx) GetContract(index common.AccountName) (*types.DeployInfo, error) {
	return c.StateDB.GetContract(index)
}
func (c *ChainTx) GetChainList() ([]state.Chain, error) {
	return c.StateDB.GetChainList()
}

/**
*  @brief  get the abi of contract
*  @param  indexAcc - the uuid of account
 */
func (c *ChainTx) GetContractAbi(index common.AccountName) ([]byte, error) {
	return c.StateDB.GetContractAbi(index)
}

func (c *ChainTx) AddPermission(index common.AccountName, perm state.Permission) error {
	return c.StateDB.AddPermission(index, perm)
}
func (c *ChainTx) FindPermission(index common.AccountName, name string) (string, error) {
	return c.StateDB.FindPermission(index, name)
}

/**
*  @brief  get a account's balance
*  @param  indexAcc - the uuid of account
*  @param  indexToken - the uuid of token
 */
func (c *ChainTx) AccountGetBalance(index common.AccountName, token string) (*big.Int, error) {
	return c.StateDB.AccountGetBalance(index, token)
}

/**
*  @brief  add a account's balance
*  @param  indexAcc - the uuid of account
*  @param  indexToken - the uuid of token
 */
func (c *ChainTx) AccountAddBalance(index common.AccountName, token string, value uint64) error {
	return c.StateDB.AccountAddBalance(index, token, new(big.Int).SetUint64(value))
}

/**
*  @brief  sub a account's balance
*  @param  indexAcc - the uuid of account
*  @param  indexToken - the uuid of token
 */
func (c *ChainTx) AccountSubBalance(index common.AccountName, token string, value uint64) error {
	return c.StateDB.AccountSubBalance(index, token, new(big.Int).SetUint64(value))
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
		payload, ok := tx.Payload.GetInstance().(*types.TransferInfo)
		if !ok {
			return nil, 0, 0, errors.New("transaction type error[transfer]")
		}
		if err := s.AccountSubBalance(tx.From, state.AbaToken, payload.Value); err != nil {
			return nil, 0, 0, err
		}
		if err := s.AccountAddBalance(tx.Addr, state.AbaToken, payload.Value); err != nil {
			return nil, 0, 0, err
		}
		if !tx.Receipt.IsBeSet() {
			tx.Receipt.Token = state.AbaToken
			tx.Receipt.Amount = payload.Value
		}
	case types.TxDeploy:
		if err := s.CheckPermission(tx.Addr, state.Active, tx.Hash, tx.Signatures); err != nil {
			return nil, 0, 0, err
		}
		payload, ok := tx.Payload.GetInstance().(*types.DeployInfo)
		if !ok {
			return nil, 0, 0, errors.New("transaction type error[deploy]")
		}
		if err := s.SetContract(tx.Addr, payload.TypeVm, payload.Describe, payload.Code, payload.Abi); err != nil {
			return nil, 0, 0, err
		}
	case types.TxInvoke:
		actionNew := types.NewAction(tx)
		trxContext, _ := context.NewTranscationContext(s, tx, cpuLimit, netLimit, timeStamp)
		_, err = smartcontract.DispatchAction(trxContext, actionNew, 0)
		if err != nil {
			return nil, 0, 0, errors.New(err.Error())
		}
		ret, err = json.Marshal(trxContext.Trace)
		if err != nil {
			return nil, 0, 0, errors.New(fmt.Sprintf("marshal trxContext.Trace failed:%d", err.Error()))
		}
	default:
		return nil, 0, 0, errors.New("the transaction's type error")
	}
	end := time.Now().UnixNano()
	if !tx.Receipt.IsBeSet() {
		cpu = float64(end-start) / 1000000.0
		tx.Receipt.Cpu = cpu
		data, err := tx.Serialize()
		if err != nil {
			return nil, 0, 0, err
		}
		net = float64(len(data))
		tx.Receipt.Net = net
		if err := tx.Receipt.ComputeHash(); err != nil {
			return nil, 0, 0, err
		}
	} else {
		cpu = tx.Receipt.Cpu
		net = tx.Receipt.Net
		tx.Receipt.Result = common.CopyBytes(ret)
	}

	//TODO:测试期间暂时关掉资源检测
	/*if err := s.RecoverResources(tx.From, timeStamp, cpuLimit, netLimit); err != nil {
		return nil, 0, 0, err
	}
	if err := s.SubResources(tx.From, cpu, net, cpuLimit, netLimit); err != nil {
		return nil, 0, 0, err
	}*/

	return ret, cpu, net, nil
}

func (c *ChainTx) blockExisted(hash common.Hash) bool {
	if _, err := c.BlockStore.Get(hash.Bytes()); err != nil {
		if err == store.ErrNotFound {
			if _, err := c.BlockStoreCache.Get(hash.Bytes()); err == nil { //在缓存中,还未入链,用于分片模式
				return true
			}
		}
		return false
	}
	return true
}
