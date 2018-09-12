package example

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/http/common/abi"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"
	"encoding/hex"
	"strconv"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
	"bytes"
)

var interval = time.Millisecond * 100
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
	invoke, err := types.NewInvokeContract(indexFrom, indexAddr, config.ChainHash, "", method, []string{"01b1a6569a557eafcccc71e0d02461fd4b601aea", "Token.Test", "20000"}, 0, time.Now().UnixNano())
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
	deploy, err := types.NewDeployContract(indexFrom, indexAddr, config.ChainHash, "", types.VmWasm, "test deploy", code, nil, 0, time.Now().UnixNano())
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
	indexFrom := common.NameToIndex("root")
	indexAddr := common.NameToIndex("delegate")
	value := big.NewInt(100)
	tx, err := types.NewTransfer(indexFrom, indexAddr, config.ChainHash, "", value, 0, time.Now().UnixNano())
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
	l, err := ledgerimpl.NewLedger(path, config.ChainHash, common.AddressFromPubKey(config.Root.PublicKey))
	errors.CheckErrorPanic(err)
	return l
}

func SaveBlock(ledger ledger.Ledger, txs []*types.Transaction, chainID common.Hash) *types.Block {
	con, err := types.InitConsensusData(TimeStamp())
	errors.CheckErrorPanic(err)
	headerPayload := &types.CMBlockHeader{LeaderPubKey:config.Root.PublicKey}
	block, err := ledger.NewTxBlock(chainID, txs, headerPayload, *con, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	block.SetSignature(&config.Root)
	errors.CheckErrorPanic(ledger.VerifyTxBlock(block.ChainID, block))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorLedger, block))
	time.Sleep(time.Millisecond * 500)
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

func AutoGenerateTransaction(ledger ledger.Ledger) {
	for {
		time.Sleep(time.Second * 2)
		if ledger.StateDB(config.ChainHash).RequireVotingInfo() {
			elog.Log.Info("Start Consensus Module")
			break
		}
	}
	for {
		time.Sleep(time.Second * 5)
		nonce := uint64(1)
		nonce++
		transfer, err := types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("delegate"), config.ChainHash, "active", new(big.Int).SetUint64(1), nonce, time.Now().UnixNano())
		errors.CheckErrorPanic(err)
		transfer.SetSignature(&config.Root)

		errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	}
}

func VotingProducer(ledger ledger.Ledger) {
	//set smart contract for root delegate
	time.Sleep(time.Second * 5)
	log.Warn("Start Voting Producer")
	contract, err := types.NewDeployContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, types.VmNative, "system control", nil, nil, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Root))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))
	time.Sleep(time.Millisecond * 500)

	/*contract, err = types.NewDeployContract(common.NameToIndex("delegate"), common.NameToIndex("delegate"), config.ChainHash, state.Owner, types.VmNative, "system control", nil, nil, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Delegate))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))
	time.Sleep(time.Millisecond * 500)*/

	invoke, err := types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, "new_account", []string{"delegate", common.AddressFromPubKey(config.Delegate.PublicKey).HexString()}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//create account worker1, worker2
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, "new_account", []string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 1, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//transfer worker1, worker2 aba token
	time.Sleep(time.Second * 5)
	transfer, err := types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("worker1"), config.ChainHash, state.Owner, new(big.Int).SetUint64(10000), 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(time.Millisecond * 500)

	transfer, err = types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("worker2"), config.ChainHash, state.Owner, new(big.Int).SetUint64(10000), 1, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(time.Millisecond * 500)

	//delegate for worker1 and worker2 cpu,net
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Active, "pledge", []string{"root", "worker1", "500", "500"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Active, "pledge", []string{"root", "worker2", "500", "500"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//worker1 and worker2 delegate aba to get votes
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("root"), config.ChainHash, state.Active, "pledge", []string{"worker1", "worker1", "4000", "4000"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), common.NameToIndex("root"), config.ChainHash, state.Active, "pledge", []string{"worker2", "worker2", "4000", "4000"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//worker1, worker2 register to producer
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("root"), config.ChainHash, state.Active, "reg_prod", []string{"worker1"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), common.NameToIndex("root"), config.ChainHash, state.Active, "reg_prod", []string{"worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//worker1, worker2 voting to be producer
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("root"), config.ChainHash, state.Active, "vote", []string{"worker1", "worker1", "worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), common.NameToIndex("root"), config.ChainHash, state.Active, "vote", []string{"worker1", "worker1", "worker2"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	go AutoGenerateTransaction(ledger)
}

func CreateAccountBlock(chainID common.Hash) {
	log.Info("-----------------------------CreateAccountBlock")
	root := common.NameToIndex("root")
	tokenContract, err := types.NewDeployContract(root, root, chainID, state.Active, types.VmNative, "system control", nil, nil, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(tokenContract.SetSignature(&config.Root))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, tokenContract))
	time.Sleep(time.Second * 2)

	invoke, err := types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"delegate", common.AddressFromPubKey(config.Delegate.PublicKey).HexString()}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	invoke, err = types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	invoke, err = types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 1, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	invoke, err = types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"worker3", common.AddressFromPubKey(config.Worker3.PublicKey).HexString()}, 2, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	perm := state.NewPermission(state.Active, state.Owner, 2, []state.KeyFactor{}, []state.AccFactor{{Actor: common.NameToIndex("worker1"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker2"), Weight: 1, Permission: "active"}, {Actor: common.NameToIndex("worker3"), Weight: 1, Permission: "active"}})
	param, err := json.Marshal(perm)
	errors.CheckErrorPanic(err)
	invoke, err = types.NewInvokeContract(root, root, chainID, state.Active, "set_account", []string{"root", string(param)}, 0, time.Now().UnixNano())
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	time.Sleep(time.Second * 2)
}

func TokenTransferBlock(chainID common.Hash) {
	log.Info("-----------------------------TokenTransferBlock")
	root := common.NameToIndex("root")
	transfer, err := types.NewTransfer(root, common.NameToIndex("worker1"), chainID, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(interval)

	transfer, err = types.NewTransfer(root, common.NameToIndex("delegate"), chainID, "active", new(big.Int).SetUint64(10000), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(interval)

	transfer, err = types.NewTransfer(root, common.NameToIndex("worker2"), chainID, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(interval)

	transfer, err = types.NewTransfer(root, common.NameToIndex("worker3"), chainID, "active", new(big.Int).SetUint64(500), 101, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(interval)

	time.Sleep(time.Second * 2)
}
func PledgeContract(chainID common.Hash) {
	log.Info("-----------------------------PledgeContract")
	root := common.NameToIndex("root")

	/*delegate := common.NameToIndex("delegate")

	tokenContract, err := types.NewDeployContract(delegate, delegate, chainID, "active", types.VmNative, "system control", nil, nil, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	tokenContract.SetSignature(&config.Delegate)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, tokenContract))
	time.Sleep(time.Second * 2)*/

	invoke, err := types.NewInvokeContract(root, root, chainID, "owner", "pledge", []string{"root", "worker1", "100", "100"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), root, chainID, "owner", "pledge", []string{"worker1", "worker1", "200", "200"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), root, chainID, "owner", "pledge", []string{"worker2", "worker2", "100", "100"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)
	time.Sleep(time.Second * 2)
}
func CreateNewChain(chainID common.Hash) {
	log.Info("-----------------------------CreateNewChain")
	invoke, err := types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("root"), chainID, "active", "reg_chain", []string{"worker1", "solo", common.AddressFromPubKey(config.Root.PublicKey).HexString()}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)
	time.Sleep(time.Second * 5)
}

func checkParam(abiDef abi.ABI, method string, arg []byte) ([]byte, error){
	var f interface{}

	if err := json.Unmarshal(arg, &f); err != nil {
		return nil, err
	}

	m := f.(map[string]interface{})

	var fields []abi.FieldDef
	for _, action := range abiDef.Actions {
		// first: find method
		if string(action.Name) == method {
			//fmt.Println("find ", method)
			for _, struction := range abiDef.Structs {
				// second: find struct
				if struction.Name == action.Type {
					fields = struction.Fields
				}
			}
			break
		}
	}

	if fields == nil {
		return nil, errors.New(log, "can not find method " + method)
	}

	args := make([]wasmservice.ParamTV, len(fields))
	for i, field := range fields {
		v := m[field.Name]
		if v != nil {
			args[i].Ptype = field.Type

			switch vv := v.(type) {
			case string:
			//	if field.Type == "string" || field.Type == "account_name" || field.Type == "asset" {
			//		args[i].Pval = vv
			//	} else {
			//		return nil, errors.New(log, fmt.Sprintln("can't match abi struct field type ", field.Type))
			//	}
			//	fmt.Println(field.Name, "is ", field.Type, "", vv)
			//case float64:
				switch field.Type {
				case "string","account_name","asset":
					args[i].Pval = vv
				case "int8":
					const INT8_MAX = int8(^uint8(0) >> 1)
					const INT8_MIN = ^INT8_MAX
					a, err := strconv.ParseInt(vv, 10, 8)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int8 range"))
					}
					if a >= int64(INT8_MIN) && a <= int64(INT8_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int8 range"))
					}
				case "int16":
					const INT16_MAX = int16(^uint16(0) >> 1)
					const INT16_MIN = ^INT16_MAX
					a, err := strconv.ParseInt(vv, 10, 16)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int16 range"))
					}
					if a >= int64(INT16_MIN) && a <= int64(INT16_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int16 range"))
					}
				case "int32":
					const INT32_MAX = int32(^uint32(0) >> 1)
					const INT32_MIN = ^INT32_MAX
					a, err := strconv.ParseInt(vv, 10, 32)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int32 range"))
					}
					if a >= int64(INT32_MIN) && a <= int64(INT32_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int32 range"))
					}
				case "int64":
					const INT64_MAX = int64(^uint64(0) >> 1)
					const INT64_MIN = ^INT64_MAX
					a, err := strconv.ParseInt(vv, 10, 64)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int64 range"))
					}
					if a >= INT64_MIN && a <= INT64_MAX {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of int64 range"))
					}

				case "uint8":
					const UINT8_MIN uint8 = 0
					const UINT8_MAX = ^uint8(0)
					a, err := strconv.ParseUint(vv, 10, 8)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint8 range"))
					}
					if a >= uint64(UINT8_MIN) && a <= uint64(UINT8_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint8 range"))
					}
				case "uint16":
					const UINT16_MIN uint16 = 0
					const UINT16_MAX = ^uint16(0)
					a, err := strconv.ParseUint(vv, 10, 16)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint16 range"))
					}
					if a >= uint64(UINT16_MIN) && a <= uint64(UINT16_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint16 range"))
					}
				case "uint32":
					const UINT32_MIN uint32 = 0
					const UINT32_MAX = ^uint32(0)
					a, err := strconv.ParseUint(vv, 10, 32)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint32 range"))
					}
					if a >= uint64(UINT32_MIN) && a <= uint64(UINT32_MAX) {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint32 range"))
					}
				case "uint64":
					const UINT64_MIN uint64 = 0
					const UINT64_MAX = ^uint64(0)
					a, err := strconv.ParseUint(vv, 10, 64)
					if err != nil {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint64 range"))
					}
					if a >= UINT64_MIN && a <= UINT64_MAX {
						args[i].Pval = vv
					} else {
						return nil, errors.New(log, fmt.Sprintln(vv, "is out of uint64 range"))
					}

				default:
					return nil, errors.New(log, fmt.Sprintln("can't match abi struct field type ", field.Type))
				}
				//
				//if field.Type == "int8" || field.Type == "int16" || field.Type == "int32" {
				//	args[i].Pval = strconv.FormatInt(int64(vv), 10)
				//} else if field.Type == "uint8" || field.Type == "uint16" || field.Type == "uint32" {
				//	args[i].Pval = strconv.FormatUint(uint64(vv), 10)
				//} else {
				//	return nil, errors.New(log, fmt.Sprintln("can't match abi struct field type ", field.Type))
				//}
				fmt.Println(field.Name, "is ", field.Type, "", vv)
				//case []interface{}:
				//	fmt.Println(field.Name, "is an array:")
				//	for i, u := range vv {
				//		fmt.Println(i, u)
				//	}
			default:
				return nil, errors.New(log, fmt.Sprintln("can't match abi struct field type: ", v))
			}
		} else {
			return nil, errors.New(log, "can't match abi struct field name:  " + field.Name)
		}

	}

	bs, err := json.Marshal(args)
	if err != nil {
		return nil, errors.New(log, "json.Marshal failed")
	}
	return bs, nil
}

func InvokeContract(ledger ledger.Ledger) {
	time.Sleep(time.Second * 2)
	log.Warn("Start Invoke contract")

	path := os.Getenv("GOPATH")

	//file data
	file, err := os.OpenFile(path + "/src/github.com/ecoball/go-ecoball/test/abaToken/program.wasm", os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("open file failed")
		return
	}

	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("read contract filr err: ", err.Error())
		return
	}

	abifile, err := os.OpenFile(path + "/src/github.com/ecoball/go-ecoball/test/abaToken/token.abi", os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("open file failed")
		return
	}

	defer abifile.Close()
	abidata, err := ioutil.ReadAll(abifile)
	if err != nil {
		fmt.Println("read contract filr err: ", err.Error())
		return
	}

//	simpleABI := []byte(`
//{
// "types": [],
// "structs": [{
//     "name": "transfer",
//     "base": "",
//     "fields": [
//        {"name":"from", "type":"string"},
//        {"name":"to", "type":"string"},
//        {"name":"quantity", "type":"string"},
//        {"name":"memo", "type":"int32"}
//     ]
//   },{
//      "name": "account",
//      "base": "",
//      "fields": [
//        {"name":"balance", "type":"asset"}
//      ]
//    },{
//      "name": "currency_stats",
//      "base": "",
//      "fields": [
//        {"name":"supply", "type":"asset"},
//        {"name":"max_supply", "type":"asset"},
//        {"name":"issuer", "type":"account_name"}
//      ]
//    }
// ],
// "actions": [{
//     "name": "transfer",
//     "type": "transfer"
//   }
// ],
// "tables": [{
//	"name": "accounts",
//	"type": "account",
//	"index_type": "i64",
//	"key_names" : ["currency"],
//	"key_types" : ["uint64"]
//	},{
//	"name": "stat",
//	"type": "currency_stats",
//	"index_type": "i64",
//	"key_names" : ["currency"],
//	"key_types" : ["uint64"]
//	}
// ]
//}
//`)
//
//	var simpleAbi abi.ABI
//	if err = json.Unmarshal(simpleABI, &simpleAbi); err != nil {
//		fmt.Errorf("ABI Unmarshal failed")
//		return
//	}

	var contractAbi abi.ABI
	if err = json.Unmarshal(abidata, &contractAbi); err != nil {
		fmt.Errorf("ABI Unmarshal failed")
		return
	}

	abibyte, err := abi.MarshalBinary(contractAbi)
	if err != nil {
		fmt.Errorf("ABI MarshalBinary failed")
		return
	}
	fmt.Println("abibyte: ", hex.EncodeToString(abibyte))

	contract, err := types.NewDeployContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, types.VmWasm, "test", data, abibyte, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Root))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))
	time.Sleep(time.Millisecond * 1500)


	contractGet, err := ledger.GetContract(config.ChainHash, common.NameToIndex("root"))
	if err != nil {
		fmt.Errorf("can not find contract abi file")
		return
	}

	var abiDef abi.ABI
	err = abi.UnmarshalBinary(contractGet.Abi, &abiDef)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	//var abiDef abi.ABI
	//json.Unmarshal(abiByte, &abiDef)

	//transfer := []byte(`{"from": "gm2tsojvgene", "to": "hellozhongxh", "quantity": "100.0000 EOS", "memo": "-100"}`)
	create := []byte(`{"creator": "user1", "max_supply": "800", "token_id": "xyx"}`)

	argbyte, err := checkParam(abiDef, "create", create)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	//test param
	//time.Sleep(time.Second * 5)
	//params, err := commands.ParseParams("string:foo,int32:2147483647")
	//if err != nil {
	//	return
	//}
	//
	//data, err = json.Marshal(params)
	//if err != nil {
	//	return
	//}
	//log.Debug("ParseParams: ", string(data))
	//
	//argbyte, err := commands.BuildWasmContractParam(params)
	//if err != nil {
	//	//t.Errorf("build wasm contract param failed:%s", err)
	//	//return
	//	return
	//}
	//log.Debug("BuildWasmContractParam: ", string(argbyte))

	var parameters []string

	parameters = append(parameters, string(argbyte[:]))

	invoke, err := types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, "test", parameters, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 2500)

}

func getContractTable(contractName string, accountName string, abiDef abi.ABI, tableName string) ([]byte, error){
	var fields []abi.FieldDef
	for _, table := range abiDef.Tables {
		if string(table.Name) == tableName {
			for _, struction := range abiDef.Structs {
				if struction.Name == table.Type {
					fields = struction.Fields
				}
			}
		}
	}

	if fields == nil {
		return nil, errors.New(log, "can not find struct of table  " + tableName)
	}

	key := []byte("xyx")
	storage, err := ledger.L.StoreGet(config.ChainHash, common.NameToIndex(contractName), key)
	fmt.Println("xyx: " + string(storage))

	key = []byte(accountName)
	storage, err = ledger.L.StoreGet(config.ChainHash, common.NameToIndex(contractName), key)
	fmt.Println(fields[0].Name + ": " + string(storage))

	type QueryRes struct {
		Balance 	string	`json:"balance"`
	}

	resp := QueryRes{string(storage)}

	js, _ := json.Marshal(resp)
	fmt.Println("json format: ", string(js))


	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)

	type json_object struct {
		Name 	string		`json:"name"`
		Value	string		`json:"value"`
	}

	//var v map[string]string
	table := make(map[string]string, len(fields))
	
	for i, _ := range fields {
		table[fields[i].Name] = string(storage)
		//table[i].Name = fields[i].Name
		//table[i].Value = string(storage)
	}

	if err := enc.Encode(&table); err != nil {
		fmt.Errorf("Encode failed")
	}
	fmt.Println("encode: " + string(buf.Bytes()))


	return nil, err
}

func QueryContractData(ledger ledger.Ledger) {
	time.Sleep(time.Second * 10)

	log.Warn("Query Contract Data")

	contractGet, err := ledger.GetContract(config.ChainHash, common.NameToIndex("root"))
	if err != nil {
		fmt.Errorf("can not find contract abi file")
		return
	}

	var abiDef abi.ABI
	err = abi.UnmarshalBinary(contractGet.Abi, &abiDef)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	getContractTable("root", "root", abiDef, "accounts")
}

func Wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	log.Info("ecoball received signal:", sig)
}
