package example

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/core/state"
	"time"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"github.com/ecoball/go-ecoball/http/common/abi"
)

func Main(ledger ledger.Ledger) {
	root := common.NameToIndex("root")
	tokenContract, err := types.NewDeployContract(root, root, config.ChainHash, state.Active, types.VmNative, "system control", nil, nil, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(tokenContract.SetSignature(&config.Root))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, tokenContract))
	time.Sleep(time.Second * 2)

	invoke, err := types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "new_account", []string{"worker", common.AddressFromPubKey(config.Worker.PublicKey).HexString()}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)
	time.Sleep(time.Second * 2)

	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "new_account", []string{"worker1", common.AddressFromPubKey(config.Worker1.PublicKey).HexString()}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)
	time.Sleep(time.Second * 2)

	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "new_account", []string{"worker2", common.AddressFromPubKey(config.Worker2.PublicKey).HexString()}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)
	time.Sleep(time.Second * 2)


	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "pledge", []string{"root", "worker", "100", "100"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)
	time.Sleep(time.Second * 2)

	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "pledge", []string{"root", "worker1", "100", "100"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)
	time.Sleep(time.Second * 2)

	invoke, err = types.NewInvokeContract(root, root, config.ChainHash, state.Owner, "pledge", []string{"root", "worker2", "100", "100"}, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Root)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(interval)
	time.Sleep(time.Second * 2)

	path := os.Getenv("GOPATH")

	// abi file, common for contract file1 and file2
	abifile, err := os.OpenFile(path + "/src/github.com/ecoball/go-ecoball/test/contract/testToken/simple_token.abi", os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("open file simple_token.abi failed")
		return
	}

	defer abifile.Close()
	abidata, err := ioutil.ReadAll(abifile)
	if err != nil {
		fmt.Println("read contract filr err: ", err.Error())
		return
	}

	//contract file2 data
	file2, err := os.OpenFile(path + "/src/github.com/ecoball/go-ecoball/test/contract/testToken/inline_action2.wasm", os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("open file inline_action2.wasm failed")
		return
	}

	defer file2.Close()
	data2, err := ioutil.ReadAll(file2)
	if err != nil {
		fmt.Println("read contract filr err: ", err.Error())
		return
	}


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

	// deploy second contract
	contract, err := types.NewDeployContract(common.NameToIndex("worker"), common.NameToIndex("worker"), config.ChainHash, state.Owner, types.VmWasm, "test", data2, abibyte, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(contract.SetSignature(&config.Worker))
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, contract))
	time.Sleep(time.Millisecond * 1500)


	contractGet, err := ledger.GetContract(config.ChainHash, common.NameToIndex("worker"))
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

	// second contract create
	create := []byte(`["worker1", "800", "xxx"]`)

	argbyte, err := abi.CheckParam(abiDef, "create", create)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	var parameters2 []string

	parameters2 = append(parameters2, string(argbyte[:]))

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("worker"), config.ChainHash, state.Owner, "create", parameters2, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 2500)

	// second contract issue
	issue := []byte(`{"to": "worker1", "amount": "100", "token_id": "xxx"}`)

	argbyte, err = abi.CheckParam(abiDef, "issue", issue)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	var issueParameters2 []string

	issueParameters2 = append(issueParameters2, string(argbyte[:]))

	invoke, err = types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("worker"), config.ChainHash, state.Owner, "issue", issueParameters2, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 2500)
}

func ParallelInvokeContract1(ledger ledger.Ledger) {
	time.Sleep(time.Second * 40)

	contractGet, err := ledger.GetContract(config.ChainHash, common.NameToIndex("worker"))
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

	trans := []byte(`{"from": "worker1", "to": "worker2", "amount": "20", "token_id": "xxx"}`)

	argbyte, err := abi.CheckParam(abiDef, "transfer", trans)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	var transferParameters []string

	transferParameters = append(transferParameters, string(argbyte[:]))

	invoke, err := types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("worker"), config.ChainHash, state.Owner, "transfer", transferParameters, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 2500)
}

func ParallelInvokeContract2(ledger ledger.Ledger) {
	time.Sleep(time.Second * 40)

	contractGet, err := ledger.GetContract(config.ChainHash, common.NameToIndex("worker"))
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

	trans := []byte(`{"from": "worker1", "to": "worker2", "amount": "20", "token_id": "xxx"}`)

	argbyte, err := abi.CheckParam(abiDef, "transfer", trans)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	var transferParameters []string

	transferParameters = append(transferParameters, string(argbyte[:]))

	invoke, err := types.NewInvokeContract(common.NameToIndex("worker1"), common.NameToIndex("worker"), config.ChainHash, state.Owner, "transfer", transferParameters, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker1)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 2500)
}

func ParallelInvokeContract3(ledger ledger.Ledger) {
	time.Sleep(time.Second * 40)

	contractGet, err := ledger.GetContract(config.ChainHash, common.NameToIndex("worker"))
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

	trans := []byte(`{"from": "worker2", "to": "worker1", "amount": "20", "token_id": "xxx"}`)

	argbyte, err := abi.CheckParam(abiDef, "transfer", trans)
	if err != nil {
		fmt.Errorf("can not find UnmarshalBinary abi file")
		return
	}

	var transferParameters []string

	transferParameters = append(transferParameters, string(argbyte[:]))

	invoke, err := types.NewInvokeContract(common.NameToIndex("worker2"), common.NameToIndex("worker"), config.ChainHash, state.Owner, "transfer", transferParameters, 0, time.Now().UnixNano())
	errors.CheckErrorPanic(err)
	invoke.SetSignature(&config.Worker2)
	errors.CheckErrorPanic(event.Send(event.ActorNil, event.ActorTxPool, invoke))
	time.Sleep(time.Millisecond * 2500)
}
