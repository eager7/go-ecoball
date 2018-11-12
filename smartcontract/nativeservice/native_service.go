package nativeservice

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/message"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/types"
	"strconv"
	"github.com/ecoball/go-ecoball/dsn/audit"
	dsnComm "github.com/ecoball/go-ecoball/dsn/common"
	"math/big"
	"github.com/ecoball/go-ecoball/smartcontract/context"
)

var log = elog.NewLogger("native", elog.NoticeLog)

type NativeService struct {
	state     state.InterfaceState
	tx        *types.Transaction
	context   *context.ApplyContext
	//method    string
	//params    []string
	cpuLimit  float64
	netLimit  float64
	timeStamp int64
}

func NewNativeService(s state.InterfaceState, tx *types.Transaction, context *context.ApplyContext, cpuLimit, netLimit float64, timeStamp int64) (*NativeService, error) {
	ns := &NativeService{
		state:     s,
		tx:        tx,
		//method:    method,
		//params:    params,
		context:	context,
		cpuLimit:  cpuLimit,
		netLimit:  netLimit,
		timeStamp: timeStamp,
	}
	return ns, nil
}

func (ns *NativeService) Execute() ([]byte, error) {
	switch ns.tx.Addr {
	case common.NameToIndex("root"):
		return ns.RootExecute()
	default:
		return nil, errors.New(log, "unknown native contract's owner")
	}
	return nil, nil
}

func (ns *NativeService)Println(s string) {
	ns.context.Action.Console += s + "\n"
}

func (ns *NativeService) RootExecute() ([]byte, error) {
	method := string(ns.tx.Payload.GetObject().(types.InvokeInfo).Method)
	params := ns.tx.Payload.GetObject().(types.InvokeInfo).Param
	switch method {
	case "new_account":
		index := common.NameToIndex(params[0])
		addr := common.AddressFormHexString(params[1])
		acc, err := ns.state.AddAccount(index, addr, ns.timeStamp)
		if err != nil {
			ns.Println(err.Error())
			return nil, err
		}

		ns.Println(fmt.Sprint("create account success"))

		// generate trx receipt
		data, err := acc.Serialize()
		if err != nil {
			return nil, err
		}
		ns.tx.Receipt.Accounts[0] = data
	case "set_account":
		index := common.NameToIndex(params[0])
		perm := state.Permission{Keys: make(map[string]state.KeyFactor, 1), Accounts: make(map[string]state.AccFactor, 1)}
		if err := json.Unmarshal([]byte(params[1]), &perm); err != nil {
			fmt.Println(params[1])
			ns.Println(err.Error())
			return nil, err
		}
		if err := ns.state.AddPermission(index, perm); err != nil {
			ns.Println(err.Error())
			return nil, err
		}

		ns.Println(fmt.Sprint("set account success"))

		// generate trx receipt
		acc := state.Account{
			Index:			index,
			Permissions: make(map[string]state.Permission, 1),
		}
		acc.Permissions[perm.PermName] = perm

		var err error
		data, err := acc.Serialize()
		if err != nil {
			return nil, err
		}
		ns.tx.Receipt.Accounts[0] = data

	case "reg_prod":
		index := common.NameToIndex(params[0])
		if err := ns.state.RegisterProducer(index); err != nil {
			ns.Println(err.Error())
			return nil, err
		}
		// generate trx receipt
		ns.tx.Receipt.Producer = uint64(index)

		ns.Println(fmt.Sprint("register producer success"))

	case "vote":

		from := common.NameToIndex(params[0])
		//to1 := common.NameToIndex(params[1])
		//to2 := common.NameToIndex(params[2])
		//accounts := []common.AccountName{to1, to2}
		var accounts []common.AccountName
		for i := 1; i < len(params); i++ {
			accounts = append(accounts, common.NameToIndex(params[i]))
		}
		ns.state.ElectionToVote(from, accounts)

		ns.Println(fmt.Sprint("vote success"))
	case "reg_chain":
		index := common.NameToIndex(params[0])
		consensus := params[1]
		addr := common.AddressFormHexString(params[2])
		data := []byte(index.String() + consensus + addr.HexString())
		hash := common.SingleHash(data)
		if err := ns.state.RegisterChain(index, hash, ns.tx.Hash, addr); err != nil {
			ns.Println(err.Error())
			return nil, err
		}
		if  ns.state.StateType()== state.FinalType {
			if consensus == "solo" {
				msg := &message.RegChain{ChainID: hash, TxHash: ns.tx.Hash, Address: addr}
				event.Send(event.ActorNil, event.ActorConsensusSolo, msg)
			} else if consensus == "ababft" {
				msg := &message.RegChain{ChainID: hash, TxHash: ns.tx.Hash, Address: addr}
				event.Send(event.ActorNil, event.ActorConsensus, msg)
			} else {
				log.Warn("not support now")
			}
		}

	case "pledge":
		from := common.NameToIndex(params[0])
		to := common.NameToIndex(params[1])
		cpu, err := strconv.ParseUint(params[2], 10, 64)
		if err != nil {
			ns.Println(err.Error())
			return nil, err
		}
		net, err := strconv.ParseUint(params[3], 10, 64)
		if err != nil {
			ns.Println(err.Error())
			return nil, err
		}

		//log.Debug(from, to, cpu, net)
		if err := ns.state.SetResourceLimits(from, to, cpu, net, ns.cpuLimit, ns.netLimit); err != nil {
			ns.Println(err.Error())
			return nil, err
		}

		ns.Println("pledge success!")

		// generate trx receipt
		accFrom, err := ns.state.GetAccountByName(from)
		if err != nil {
			return nil, err
		}

		fromAccount := state.Account{
			Index:			from,
			Tokens:			make(map[string]state.Token),
		}

		toAccount := state.Account{
			Index:			to,
		}

		balance := state.Token{
			Name:		state.AbaToken,
			Balance:	big.NewInt(int64(0 - (cpu + net))),
		}
		fromAccount.Tokens[state.AbaToken] = balance

		if from == to {
			fromAccount.Cpu.Staked = cpu
			fromAccount.Net.Staked = net

			data, err := fromAccount.Serialize()
			if err != nil {
				return nil, err
			}
			ns.tx.Receipt.Accounts[0] = data

		} else {
			fromAccount.Delegates = accFrom.Delegates
			toAccount.Cpu.Delegated = cpu
			toAccount.Net.Delegated = net

			data, err := fromAccount.Serialize()
			if err != nil {
				return nil, err
			}
			ns.tx.Receipt.Accounts[0] = data

			data1, err := toAccount.Serialize()
			if err != nil {
				return nil, err
			}
			ns.tx.Receipt.Accounts[1] = data1
		}

	case "cancel_pledge":
		from := common.NameToIndex(params[0])
		to := common.NameToIndex(params[1])
		cpu, err := strconv.ParseUint(params[2], 10, 64)
		if err != nil {
			ns.Println(err.Error())
			return nil, err
		}
		net, err := strconv.ParseUint(params[3], 10, 64)
		if err != nil {
			ns.Println(err.Error())
			return nil, err
		}
		log.Debug(from, to, cpu, net)
		if err := ns.state.CancelDelegate(from, to, cpu, net, ns.cpuLimit, ns.netLimit); err != nil {
			ns.Println(err.Error())
			return nil, err
		}

		ns.Println("cancel pledge")

		// generate trx receipt
		accFrom, err := ns.state.GetAccountByName(from)
		if err != nil {
			return nil, err
		}

		fromAccount := state.Account{
			Tokens:			make(map[string]state.Token),
			Index:			from,
		}

		toAccount := state.Account{
			Index:			to,
		}

		balance := state.Token{
			Name:		state.AbaToken,
			Balance:	big.NewInt(int64(cpu + net)),
		}
		fromAccount.Tokens[state.AbaToken] = balance

		if from == to {
			fromAccount.Cpu.Staked = 0 - cpu
			fromAccount.Net.Staked = 0 - net

			data, err := fromAccount.Serialize()
			if err != nil {
				return nil, err
			}
			ns.tx.Receipt.Accounts[0] = data
		} else {
			fromAccount.Delegates = accFrom.Delegates
			toAccount.Cpu.Delegated = 0 - cpu
			toAccount.Net.Delegated = 0 - net

			data, err := fromAccount.Serialize()
			if err != nil {
				return nil, err
			}
			ns.tx.Receipt.Accounts[0] = data

			data1, err := toAccount.Serialize()
			if err != nil {
				return nil, err
			}
			ns.tx.Receipt.Accounts[1] = data1
		}

	case dsnComm.FcMethodProof:
		audit.HandleStorageProof(params[0], ns.state)
	case dsnComm.FcMethodAn:
		audit.HandleStoreAnn(params[0], ns.state)
	case dsnComm.FcMethodFile:
		audit.HandleFileContract(params[0], ns.state)
	default:
		ns.Println(fmt.Sprintf("unknown method:%s", method))
		return nil, errors.New(log, fmt.Sprintf("unknown method:%s", method))
	}
	return nil, nil
}

