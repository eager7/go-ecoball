package example

import (
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"math/big"
	"time"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"os"
)

var log = elog.NewLogger("example", elog.InfoLog)

func AddAccount(state *state.State) error {
	from := common.NewAddress(common.FromHex("01b1a6569a557eafcccc71e0d02461fd4b601aea"))
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	indexFrom := common.NameToIndex("from")
	indexAddr := common.NameToIndex("addr")
	if _, err := state.AddAccount(indexFrom, from, time.Now().UnixNano()); err != nil {
		return nil
	}
	if _, err := state.AddAccount(indexAddr, addr, time.Now().UnixNano()); err != nil {
		return nil
	}
	return nil
}

func TestInvoke(method string) *types.Transaction {
	indexFrom := common.NameToIndex("from")
	indexAddr := common.NameToIndex("addr")
	invoke, err := types.NewInvokeContract(indexFrom, indexAddr, "", method, []string{"01b1a6569a557eafcccc71e0d02461fd4b601aea", "Token.Test", "20000"}, 0, time.Now().Unix())
	if err != nil {
		panic(err)
		return nil
	}
	acc := account.Account{PrivateKey: config.Root.PrivateKey, PublicKey: config.Root.PublicKey, Alg: 0}
	if err := invoke.SetSignature(&acc); err != nil {
		panic(err)
	}
	return invoke
}

func TestDeploy(code []byte) *types.Transaction {
	indexFrom := common.NameToIndex("from")
	indexAddr := common.NameToIndex("addr")
	deploy, err := types.NewDeployContract(indexFrom, indexAddr, "", types.VmWasm, "test deploy", code, 0, time.Now().Unix())
	if err != nil {
		panic(err)
		return nil
	}
	acc := account.Account{PrivateKey: config.Root.PrivateKey, PublicKey: config.Root.PublicKey, Alg: 0}
	if err := deploy.SetSignature(&acc); err != nil {
		panic(err)
	}
	return deploy
}

func TestTransfer() *types.Transaction {
	indexFrom := common.NameToIndex("from")
	indexAddr := common.NameToIndex("addr")
	value := big.NewInt(100)
	tx, err := types.NewTransfer(indexFrom, indexAddr, "", value, 0, time.Now().Unix())
	if err != nil {
		fmt.Println(err)
		return nil
	}
	acc := account.Account{PrivateKey: config.Root.PrivateKey, PublicKey: config.Root.PublicKey, Alg: 0}
	if err := tx.SetSignature(&acc); err != nil {
		fmt.Println(err)
		return nil
	}
	return tx
}

func Ledger(path string) ledger.Ledger {
	os.RemoveAll(path)
	l, err := ledgerimpl.NewLedger(path)
	errors.CheckErrorPanic(err)
	return l
}

func SaveBlock(ledger ledger.Ledger, txs []*types.Transaction) *types.Block {
	con, err := types.InitConsensusData(TimeStamp())
	errors.CheckErrorPanic(err)
	block, _, err := ledger.NewTxBlock(txs, *con, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	block.SetSignature(&config.Root)
	errors.CheckErrorPanic(ledger.VerifyTxBlock(block))
	errors.CheckErrorPanic(ledger.SaveTxBlock(block))
	return block
}

func TimeStamp() int64 {
	tm, err := time.Parse("02/01/2006 15:04:05 PM", "21/02/1990 00:00:00 AM")
	errors.CheckErrorPanic(err)
	return tm.UnixNano()
}

func ConsensusData() types.ConsensusData {
	con, _ := types.InitConsensusData(TimeStamp())
	return *con
}

func ShowAccountInfo(s *state.State, index common.AccountName) {
	acc, err := s.GetAccountByName(index)
	errors.CheckErrorPanic(err)
	acc.Show(false)
}