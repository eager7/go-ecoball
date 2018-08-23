package transaction

import (
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common"
	"time"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/smartcontract"
	"github.com/ecoball/go-ecoball/common/config"
)

func (c *ChainTx) NewBlockWithoutHandle(ledger ledger.Ledger, txs []*types.Transaction, consensusData types.ConsensusData, timeStamp int64) (*types.Block, error) {
	var cpu, net float64
	for _, tx := range txs{
		cpu += tx.Receipt.Cpu
		net += tx.Receipt.Net
	}
	return types.NewBlock(config.ChainHash, c.CurrentHeader, c.StateDB.TempDB.GetHashRoot(), consensusData, txs, cpu, net, timeStamp)
}

func (c *ChainTx) SaveBlockWithoutHandle(block *types.Block) error {
	if block == nil {
		return errors.New(log, "block is nil")
	}
	for i := 0; i < len(block.Transactions); i++ {
		if _, _, _, err := c.HandleTransaction(c.StateDB.FinalDB, block.Transactions[i], block.TimeStamp, c.CurrentHeader.Receipt.BlockCpu, c.CurrentHeader.Receipt.BlockNet); err != nil {
			log.Error("Handle Transaction Error:", err)
			return err
		}
	}
	block.Header.StateHash = c.StateDB.FinalDB.GetHashRoot()

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
	c.StateDB.FinalDB.CommitToDB()
	c.CurrentHeader = block.Header
	return nil
}

func (c *ChainTx) HandleTransactionTxPool(s *state.State, tx *types.Transaction, timeStamp int64, cpuLimit, netLimit float64) (ret []byte, cpu, net float64, err error) {
	start := time.Now().UnixNano()
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
		if err := s.CheckPermission(tx.From, state.Active, tx.Hash, tx.Signatures); err != nil {
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
		service, err := smartcontract.NewContractService(s, tx, cpuLimit, netLimit, timeStamp)
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
		cpu = float64(end-start) / 1000000.0
		tx.Receipt.Cpu = cpu
	} else {
		cpu = tx.Receipt.Cpu
	}
	data, err := tx.Serialize()
	if err != nil {
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
		return nil, 0, 0, err
	}
	if err := s.SubResources(tx.From, cpu, net, cpuLimit, netLimit); err != nil {
		return nil, 0, 0, err
	}
	return ret, cpu, net, nil
}
