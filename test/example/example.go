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
	"github.com/ecoball/go-ecoball/http/commands"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
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
	deploy, err := types.NewDeployContract(indexFrom, indexAddr, config.ChainHash, "", types.VmWasm, "test deploy", code, 0, time.Now().UnixNano())
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
	l, err := ledgerimpl.NewLedger(path)
	errors.CheckErrorPanic(err)
	return l
}

func SaveBlock(ledger ledger.Ledger, txs []*types.Transaction, chainID common.Hash) *types.Block {
	con, err := types.InitConsensusData(TimeStamp())
	errors.CheckErrorPanic(err)
	block, err := ledger.NewTxBlock(chainID, txs, *con, time.Now().UnixNano())
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
	//_, err := ledger.AccountGet(common.NameToIndex("worker1"))
	//if err == nil {
	//	log.Panic("Please Delete Block Database First, Then Restart Program")
	//}

	//set smart contract for root delegate
	time.Sleep(time.Second * 15)
	log.Warn("Start Voting Producer")
	contract, err := types.NewDeployContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, types.VmNative, "system control", nil, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Root))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))
	time.Sleep(time.Millisecond * 500)

	contract, err = types.NewDeployContract(common.NameToIndex("delegate"), common.NameToIndex("delegate"), config.ChainHash, state.Owner, types.VmNative, "system control", nil, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Delegate))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))
	time.Sleep(time.Millisecond * 500)

	//create account worker1, worker2
	time.Sleep(time.Second * 5)
	invoke, err := types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, "new_account", []string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 1, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//transfer worker1, worker2 aba token
	time.Sleep(time.Second * 5)
	transfer, err := types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("worker1"), config.ChainHash, state.Owner, new(big.Int).SetUint64(10000), 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(time.Millisecond * 500)

	transfer, err = types.NewTransfer(common.NameToIndex("root"), common.NameToIndex("worker2"), config.ChainHash, state.Owner, new(big.Int).SetUint64(10000), 1, time.Now().Unix())
	errors.CheckErrorPanic(err)
	transfer.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, transfer))
	time.Sleep(time.Millisecond * 500)

	//delegate for worker1 and worker2 cpu,net
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("delegate"), config.ChainHash, state.Active, "pledge", []string{"root", "worker1", "500", "500"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	invoke, err = types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("delegate"), config.ChainHash, state.Active, "pledge", []string{"root", "worker2", "500", "500"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	//worker1 and worker2 delegate aba to get votes
	time.Sleep(time.Second * 5)
	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("delegate"), config.ChainHash, state.Active, "pledge", []string{"worker1", "worker1", "4000", "4000"}, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), common.NameToIndex("delegate"), config.ChainHash, state.Active, "pledge", []string{"worker2", "worker2", "4000", "4000"}, 0, time.Now().Unix())
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
}

func CreateAccountBlock(chainID common.Hash) {
	log.Info("-----------------------------CreateAccountBlock")
	root := common.NameToIndex("root")
	tokenContract, err := types.NewDeployContract(root, root, chainID, state.Active, types.VmNative, "system control", nil, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(tokenContract.SetSignature(&config.Root))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, tokenContract))
	time.Sleep(time.Second * 2)

	invoke, err := types.NewInvokeContract(root, root, chainID, state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().UnixNano())
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
	delegate := common.NameToIndex("delegate")

	tokenContract, err := types.NewDeployContract(delegate, delegate, chainID, "active", types.VmNative, "system control", nil, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	tokenContract.SetSignature(&config.Delegate)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, tokenContract))
	time.Sleep(time.Second * 2)

	invoke, err := types.NewInvokeContract(root, delegate, chainID, "owner", "pledge", []string{"root", "worker1", "100", "100"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), delegate, chainID, "owner", "pledge", []string{"worker1", "worker1", "200", "200"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker2"), delegate, chainID, "owner", "pledge", []string{"worker2", "worker2", "100", "100"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)
	time.Sleep(time.Second * 2)
}
func CreateNewChain(chainID common.Hash) {
	log.Info("-----------------------------CreateNewChain")
	invoke, err := types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("root"), chainID, "active", "reg_chain", []string{"worker1", "solo"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)
	time.Sleep(time.Second * 5)
}

func checkParam(abi commands.ABI, method string, arg []byte) (error){
	var f interface{}

	if err := json.Unmarshal(arg, &f); err != nil {
		return err
	}

	m := f.(map[string]interface{})

	var fields []commands.FieldDef
	for _, action := range abi.Actions {
		// first: find method
		if action.Name == "transfer" {
			fmt.Println("find transfer")
			for _, struction := range abi.Structs {
				// second: find struct
				if struction.Name == action.Type {
					fields = struction.Fields
				}
			}
			break
		}
	}

	args := make([]wasmservice.ParamTV, len(fields))
	for i, field := range fields {
		v := m[field.Name]
		if v != nil {
			args[i].Ptype = field.Type

			switch vv := v.(type) {
			case string:
				if field.Type == "string" || field.Type == "account_name" || field.Type == "asset" {
					args[i].Pval = vv
				} else {
					return fmt.Errorf("error, can't match abi struct field ty")
				}
				fmt.Println(field.Name, "is ", field.Type, "", vv)

			//case int:
			//	fmt.Println(field.Name, "is int", vv)

			//case []interface{}:
			//	fmt.Println(field.Name, "is an array:")
			//	for i, u := range vv {
			//		fmt.Println(i, u)
			//	}
			default:
				return fmt.Errorf("error, ", field.Name, "is of a type I don’t know how to handle")
			}
		} else {
			return fmt.Errorf("error, can't match abi struct field name")
		}

	}

	return nil
}

func InvokeContract() {
	time.Sleep(time.Second * 15)
	log.Warn("Start Invoke contract")

	//file data
	file, err := os.OpenFile("/home/ubuntu/go/src/github.com/ecoball/go-ecoball/test/token/token.wasm", os.O_RDONLY, 0666)
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

	contract, err := types.NewDeployContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, types.VmWasm, "test", data, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Root))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))
	time.Sleep(time.Millisecond * 500)


	systemABI := []byte(`
{
  "types": [],
  "structs": [{
      "name": "nonce",
       "base": "",
       "fields": [
          {"name":"value", "type":"string"}
      ]
    },{
      "name": "transfer",
      "base": "",
      "fields": [
         {"name":"from", "type":"account_name"},
         {"name":"to", "type":"account_name"},
         {"name":"quantity", "type":"asset"},
         {"name":"memo", "type":"string"}
      ]
    },{
     "name": "issue",
     "base": "",
     "fields": [
        {"name":"to", "type":"account_name"},
        {"name":"quantity", "type":"asset"}
     ]
    },{
      "name": "account",
      "base": "",
      "fields": [
        {"name":"currency", "type":"uint64"},
        {"name":"balance", "type":"uint64"}
      ]
    },{
      "name": "currency_stats",
      "base": "",
      "fields": [
        {"name":"currency", "type":"uint64"},
        {"name":"supply", "type":"uint64"}
      ]
    },{
      "name": "delegatebw",
      "base": "",
      "fields": [
         {"name":"from", "type":"account_name"},
         {"name":"receiver", "type":"account_name"},
         {"name":"stake_net", "type":"asset"},
         {"name":"stake_cpu", "type":"asset"},
         {"name":"stake_storage", "type":"asset"}
      ]
    },{
      "name": "undelegatebw",
      "base": "",
      "fields": [
         {"name":"from", "type":"account_name"},
         {"name":"receiver", "type":"account_name"},
         {"name":"unstake_net", "type":"asset"},
         {"name":"unstake_cpu", "type":"asset"},
         {"name":"unstake_bytes", "type":"uint64"}
      ]
    },{
      "name": "refund",
      "base": "",
      "fields": [
         {"name":"owner", "type":"account_name"}
      ]
    },{
      "name": "delegated_bandwidth",
      "base": "",
      "fields": [
         {"name":"from", "type":"account_name"},
         {"name":"to", "type":"account_name"},
         {"name":"net_weight", "type":"asset"},
         {"name":"cpu_weight", "type":"asset"},
         {"name":"storage_stake", "type":"asset"},
         {"name":"storage_bytes", "type":"uint64"}
      ]
    },{
      "name": "total_resources",
      "base": "",
      "fields": [
         {"name":"owner", "type":"account_name"},
         {"name":"net_weight", "type":"uint64"},
         {"name":"cpu_weight", "type":"uint64"},
         {"name":"storage_stake", "type":"uint64"},
         {"name":"storage_bytes", "type":"uint64"}
      ]
    },{
      "name": "eosio_parameters",
      "base": "",
      "fields": [
         {"name":"target_block_size", "type":"uint32"},
         {"name":"max_block_size", "type":"uint32"},
         {"name":"target_block_acts_per_scope", "type":"uint32"},
         {"name":"max_block_acts_per_scope", "type":"uint32"},
         {"name":"target_block_acts", "type":"uint32"},
         {"name":"max_block_acts", "type":"uint32"},
         {"name":"max_storage_size", "type":"uint64"},
         {"name":"max_transaction_lifetime", "type":"uint32"},
         {"name":"max_transaction_exec_time", "type":"uint32"},
         {"name":"max_authority_depth", "type":"uint16"},
         {"name":"max_inline_depth", "type":"uint16"},
         {"name":"max_inline_action_size", "type":"uint32"},
         {"name":"max_generated_transaction_size", "type":"uint32"},
         {"name":"percent_of_max_inflation_rate", "type":"uint32"},
         {"name":"storage_reserve_ratio", "type":"uint32"}
      ]
    },{
      "name": "eosio_global_state",
      "base": "eosio_parameters",
      "fields": [
         {"name":"total_storage_bytes_reserved", "type":"uint64"},
         {"name":"total_storage_stake", "type":"uint64"},
         {"name":"payment_per_block", "type":"uint64"}
      ]
    },{
      "name": "producer_info",
      "base": "",
      "fields": [
         {"name":"owner",              "type":"account_name"},
         {"name":"total_votes",        "type":"uint128"},
         {"name":"prefs",              "type":"eosio_parameters"},
         {"name":"packed_key",         "type":"uint8[]"},
         {"name":"per_block_payments", "type":"uint64"},
         {"name":"last_claim_time",    "type":"uint32"}
      ]
    },{
      "name": "regproducer",
      "base": "",
      "fields": [
        {"name":"producer",     "type":"account_name"},
        {"name":"producer_key", "type":"bytes"},
        {"name":"prefs",        "type":"eosio_parameters"}
      ]
    },{
      "name": "unregprod",
      "base": "",
      "fields": [
        {"name":"producer",     "type":"account_name"}
      ]
    },{
      "name": "regproxy",
      "base": "",
      "fields": [
        {"name":"proxy",     "type":"account_name"}
      ]
    },{
      "name": "unregproxy",
      "base": "",
      "fields": [
        {"name":"proxy",     "type":"account_name"}
      ]
    },{
      "name": "voteproducer",
      "base": "",
      "fields": [
        {"name":"voter",     "type":"account_name"},
        {"name":"proxy",     "type":"account_name"},
        {"name":"producers", "type":"account_name[]"}
      ]
    },{
      "name": "voter_info",
      "base": "",
      "fields": [
        {"name":"owner",             "type":"account_name"},
        {"name":"proxy",             "type":"account_name"},
        {"name":"last_update",       "type":"uint32"},
        {"name":"is_proxy",          "type":"uint32"},
        {"name":"staked",            "type":"uint64"},
        {"name":"unstaking",         "type":"uint64"},
        {"name":"unstake_per_week",  "type":"uint64"},
        {"name":"proxied_votes",     "type":"uint128"},
        {"name":"producers",         "type":"account_name[]"},
        {"name":"deferred_trx_id",   "type":"uint32"},
        {"name":"last_unstake",      "type":"uint32"}
      ]
    },{
      "name": "claimrewards",
      "base": "",
      "fields": [
        {"name":"owner",   "type":"account_name"}
      ]
    }
  ],
  "actions": [{
      "name": "transfer",
      "type": "transfer"
    },{
      "name": "issue",
      "type": "issue"
    },{
      "name": "delegatebw",
      "type": "delegatebw"
    },{
      "name": "undelegatebw",
      "type": "undelegatebw"
    },{
      "name": "refund",
      "type": "refund"
    },{
      "name": "regproducer",
      "type": "regproducer"
    },{
      "name": "unregprod",
      "type": "unregprod"
    },{
      "name": "regproxy",
      "type": "regproxy"
    },{
      "name": "unregproxy",
      "type": "unregproxy"
    },{
      "name": "voteproducer",
      "type": "voteproducer"
    },{
      "name": "claimrewards",
      "type": "claimrewards"
    },{
      "name": "nonce",
      "type": "nonce"
    }
  ],
  "tables": [
  ]
}
`)

	var abiDef commands.ABI
	json.Unmarshal(systemABI, &abiDef)

	transfer := []byte(`{"from": "gm2tsojvgene", "to": "hellozhongxh", "quantity": "100.0000 EOS", "memo": "nothing"}`)

	checkParam(abiDef, "transfer", transfer)

	//test param
	time.Sleep(time.Second * 5)
	params, err := commands.ParseParams("string:foo,int32:2147483647")
	if err != nil {
		return
	}

	data, err = json.Marshal(params)
	if err != nil {
		return
	}
	log.Debug("ParseParams: ", string(data))

	argbyte, err := commands.BuildWasmContractParam(params)
	if err != nil {
		//t.Errorf("build wasm contract param failed:%s", err)
		//return
		return
	}
	log.Debug("BuildWasmContractParam: ", string(argbyte))

	var parameters []string

	parameters = append(parameters, string(argbyte[:]))

	invoke, err := types.NewInvokeContract(common.NameToIndex("root"), common.NameToIndex("root"), config.ChainHash, state.Owner, "test", parameters, 0, time.Now().Unix())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 500)
}

func Wait() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(interrupt)
	sig := <-interrupt
	log.Info("ecoball received signal:", sig)
}
