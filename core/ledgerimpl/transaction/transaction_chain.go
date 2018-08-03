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
	"github.com/ecoball/go-ecoball/core/bloom"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/geneses"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/store"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/smartcontract"
	"math/big"
	"time"
)

var log = elog.NewLogger("Chain Tx", elog.DebugLog)

type ChainTx struct {
	BlockStore     store.Storage
	HeaderStore    store.Storage
	TxsStore       store.Storage
	ConsensusStore store.Storage

	CurrentHeader *types.Header
	Geneses       *types.Header
	StateDB       *state.State
	ledger        ledger.Ledger
}

func NewTransactionChain(path string, ledger ledger.Ledger) (c *ChainTx, err error) {
	c = &ChainTx{ledger: ledger}
	c.BlockStore, err = store.NewLevelDBStore(path+config.StringBlock, 0, 0)
	if err != nil {
		return nil, err
	}
	c.HeaderStore, err = store.NewLevelDBStore(path+config.StringHeader, 0, 0)
	if err != nil {
		return nil, err
	}
	c.TxsStore, err = store.NewLevelDBStore(path+config.StringTxs, 0, 0)
	if err != nil {
		return nil, err
	}

	existed, err := c.RestoreCurrentHeader()
	if err != nil {
		return nil, err
	}
	if existed {
		if c.StateDB, err = state.NewState(path+config.StringState, c.CurrentHeader.StateHash); err != nil {
			return nil, err
		}
	} else {
		if c.StateDB, err = state.NewState(path+config.StringState, common.Hash{}); err != nil {
			return nil, err
		}
	}

	return c, nil
}

/**
*  @brief  create a new block, this function will execute the transaction to rebuild mpt trie
*  @param  consensusData - the data of consensus module set
 */
func (c *ChainTx) NewBlock(ledger ledger.Ledger, txs []*types.Transaction, consensusData types.ConsensusData, timeStamp int64) (*types.Block, error) {
	s, err := c.StateDB.CopyState()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(txs); i++ {
		if ret, _, _, err := c.HandleTransaction(s, txs[i], timeStamp); err != nil {
			log.Error("Handle Transaction Error:", err)
			txs[i].Show()
			return nil, err
		} else {
			log.Notice("Handle Transaction Result:", ret)
		}
	}
	log.Warn("NewBlock State", s.GetHashRoot().HexString())
	return types.NewBlock(c.CurrentHeader, s.GetHashRoot(), consensusData, txs, timeStamp)
}

/**
*  @brief  if create a new block failed, then need to reset state DB
*  @param  hash - the root hash of mpt trie which need to reset
 */
func (c *ChainTx) ResetStateDB(hash common.Hash) error {
	return c.StateDB.Reset(hash)
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
	var cpu float32
	cpuFlag := true
	var net float32
	netFlag := true
	for i := 0; i < len(block.Transactions); i++ {
		if _, c, n, err := c.HandleTransaction(c.StateDB, block.Transactions[i], block.TimeStamp); err != nil {
			log.Error("Handle Transaction Error:", err)
			return err
		} else {
			cpu += c
			net += n
		}
	}
	if cpu < (state.BlockCpuLimit / 10) {
		cpuFlag = true
	} else {
		cpuFlag = false
	}
	if net < (state.BlockNetLimit / 10) {
		netFlag = true
	} else {
		netFlag = false
	}
	c.StateDB.SetBlockLimits(cpuFlag, netFlag)
	if err := event.Publish(event.ActorLedger, block, event.ActorTxPool, event.ActorP2P); err != nil {
		log.Warn(err)
	}
	for _, t := range block.Transactions {
		payload, _ := t.Serialize()
		if t.Type == types.TxDeploy {
			c.TxsStore.BatchPut(common.IndexToBytes(t.Addr), payload)
		} else {
			c.TxsStore.BatchPut(t.Hash.Bytes(), payload)
		}
	}
	if err := c.TxsStore.BatchCommit(); err != nil {
		return err
	}
	if c.StateDB.GetHashRoot().HexString() != block.StateHash.HexString() {
		return errors.New(log, fmt.Sprintf("hash mismatch:%s, %s", c.StateDB.GetHashRoot().HexString(), block.Hash.HexString()))
	}

	payload, err := block.Header.Serialize()
	if err != nil {
		return err
	}
	if err := c.HeaderStore.Put(block.Header.Hash.Bytes(), payload); err != nil {
		return err
	}
	payload, _ = block.Serialize()
	c.BlockStore.BatchPut(block.Hash.Bytes(), payload)
	if err := c.BlockStore.BatchCommit(); err != nil {
		return err
	}
	c.StateDB.CommitToDB()
	log.Debug("block state:", block.Height, block.StateHash.HexString())
	log.Debug("state hash:", c.StateDB.GetHashRoot().HexString())

	c.CurrentHeader = block.Header
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
		log.Error(err)
		return nil, err
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
	log.Info("The geneses block is existed:", len(headers))
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
func (c *ChainTx) GenesesBlockInit() error {
	if c.CurrentHeader != nil {
		log.Debug("geneses block is existed")
		c.CurrentHeader.Show()
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

	if err := geneses.PresetContract(c.StateDB, timeStamp); err != nil {
		return err
	}

	hashState := c.StateDB.GetHashRoot()
	header, err := types.NewHeader(types.VersionHeader, 1, hash, hash, hashState, *conData, bloom.Bloom{}, timeStamp)
	if err != nil {
		return err
	}
	block := &types.Block{Header: header, CountTxs: 0, Transactions: nil}

	if err := block.SetSignature(&config.Root); err != nil {
		return err
	}

	if err := c.VerifyTxBlock(block); err != nil {
		return err
	}
	c.CurrentHeader = block.Header
	c.Geneses = block.Header //Store Geneses for timeStamp
	if err := c.SaveBlock(block); err != nil {
		log.Error("Save geneses block error:", err)
		return err
	}
	c.CurrentHeader = block.Header
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
		if header.Height == 1 {
			c.Geneses = header //Store Geneses for timeStamp
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
func (c *ChainTx) GetTransaction(key []byte) (*types.Transaction, error) {
	data, err := c.TxsStore.Get(key)
	if err != nil {
		return nil, err
	}
	tx := new(types.Transaction)
	if err := tx.Deserialize(data); err != nil {
		return nil, err
	}
	return tx, nil
}

/**
*  @brief  validity check of transaction, include signature verify, duplicate check and balance check
*  @param  tx - a transaction
 */
func (c *ChainTx) CheckTransaction(tx *types.Transaction) (err error) {
	result, err := tx.VerifySignature()
	if err != nil {
		return err
	} else if result == false {
		return errors.New(log, "tx verify signature failed")
	}
	if err := c.StateDB.CheckPermission(tx.From, tx.Permission, tx.Signatures); err != nil {
		return err
	}

	switch tx.Type {
	case types.TxTransfer:
		if data, _ := c.TxsStore.Get(tx.Hash.Bytes()); data != nil {
			return errors.New(log, errors.ErrDuplicatedTx.ErrorInfo())
		}
		if value, err := c.AccountGetBalance(tx.From, state.AbaToken); err != nil {
			return err
		} else if value.Sign() <= 0 {
			log.Error(err)
			return errors.New(log, errors.ErrDoubleSpend.ErrorInfo())
		}
	case types.TxDeploy:
		if data, _ := c.TxsStore.Get(common.IndexToBytes(tx.Addr)); data != nil {
			return errors.New(log, errors.ErrDuplicatedTx.ErrorInfo())
		}
		//hash := c.StateDB.GetHashRoot()
		//c.HandleTransaction(c, tx)
	case types.TxInvoke:
		if data, _ := c.TxsStore.Get(tx.Hash.Bytes()); data != nil {
			return errors.New(log, errors.ErrDuplicatedTx.ErrorInfo())
		}
	default:
		return errors.New(log, "check transaction unknown tx type")
	}

	return nil
}
func (c *ChainTx) CheckPermission(index common.AccountName, name string, sig []common.Signature) error {
	return c.StateDB.CheckPermission(index, name, sig)
}

/**
*  @brief  create a new account in mpt tree
*  @param  index - the uuid of account
*  @param  addr - the public key of account
 */
func (c *ChainTx) AccountAdd(index common.AccountName, addr common.Address, timeStamp int64) (*state.Account, error) {
	return c.StateDB.AddAccount(index, addr, timeStamp)
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
func (c *ChainTx) SetContract(index common.AccountName, t types.VmType, des, code []byte) error {
	return c.StateDB.SetContract(index, t, des, code)
}
func (c *ChainTx) GetContract(index common.AccountName) (*types.DeployInfo, error) {
	return c.StateDB.GetContract(index)
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
func (c *ChainTx) HandleTransaction(s *state.State, tx *types.Transaction, timeStamp int64) (ret []byte, cpu, net float32, err error) {
	start := time.Now().UnixNano()
	//log.Debug(start, c.Geneses.TimeStamp)
	//n := (start - c.Geneses.TimeStamp) / 1000000 / int64(config.TimeSlot)
	//m := (c.CurrentHeader.TimeStamp - c.Geneses.TimeStamp) / 1000000  / int64(config.TimeSlot)
	//log.Debug(n, m, n - m)
	//timeRecover := (timeStamp - c.CurrentHeader.TimeStamp - 2 * c.Geneses.TimeStamp) / 1000000 / int64(config.TimeSlot)
	switch tx.Type {
	case types.TxTransfer:
		payload, ok := tx.Payload.GetObject().(types.TransferInfo)
		if !ok {
			return nil, 0, 0, errors.New(log, "transaction type error[transfer]")
		}
		if err := s.AccountSubBalance(tx.From, state.AbaToken, payload.Value); err != nil {
			return nil, 0, 0, err
		}
		if err := s.AccountAddBalance(tx.Addr, state.AbaToken, payload.Value); err != nil {
			return nil, 0, 0, err
		}
	case types.TxDeploy:
		if err := s.CheckPermission(tx.From, state.Active, tx.Signatures); err != nil {
			return nil, 0, 0, err
		}
		payload, ok := tx.Payload.GetObject().(types.DeployInfo)
		if !ok {
			return nil, 0, 0, errors.New(log, "transaction type error[deploy]")
		}
		if err := s.SetContract(tx.Addr, payload.TypeVm, payload.Describe, payload.Code); err != nil {
			return nil, 0, 0, err
		}
	case types.TxInvoke:
		service, err := smartcontract.NewContractService(s, tx, timeStamp)
		if err != nil {
			return nil, 0, 0, err
		}
		ret, err = service.Execute()
		if err != nil {
			return nil, 0, 0, err
		}
	default:
		return nil, 0, 0, errors.New(log, "the transaction's type error")
	}
	end := time.Now().UnixNano()
	if tx.Receipt.Cpu == 0 {
		cpu = float32(end-start) / 1000000.0
		tx.Receipt.Cpu = cpu
	} else {
		cpu = tx.Receipt.Cpu
	}
	data, err := tx.Serialize()
	if err != nil {
		return nil, 0, 0, err
	}
	if tx.Receipt.Net == 0 {
		net = float32(len(data))
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
	//log.Warn("tx, Time", tx.From, timeStamp)
	if err := s.RecoverResources(tx.From, timeStamp); err != nil {
		return nil, 0, 0, err
	}
	//log.Warn("CPU, NET", cpu, net)
	if err := s.SubResources(tx.From, cpu, net); err != nil {
		return nil, 0, 0, err
	}
	//log.Warn("Handle Type", tx.Type.String(), s.GetHashRoot().HexString())
	return ret, cpu, net, nil
}

func (c *ChainTx) TokenExisted(token string) bool {
	return c.StateDB.TokenExisted(token)
}

func (c *ChainTx) TokenAllocation() error {
	if err := c.StateDB.AccountAddBalance(state.IndexAbaRoot, state.AbaToken, new(big.Int).SetUint64(2100000)); err != nil {
		return err
	}
	return nil
}
