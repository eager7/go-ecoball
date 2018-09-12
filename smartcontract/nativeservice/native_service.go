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
)

var log = elog.NewLogger("native", elog.NoticeLog)

type NativeService struct {
	state     state.InterfaceState
	tx        *types.Transaction
	method    string
	params    []string
	cpuLimit  float64
	netLimit  float64
	timeStamp int64
}

func NewNativeService(s state.InterfaceState, tx *types.Transaction, method string, params []string, cpuLimit, netLimit float64, timeStamp int64) (*NativeService, error) {
	ns := &NativeService{
		state:     s,
		tx:        tx,
		method:    method,
		params:    params,
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

func (ns *NativeService) RootExecute() ([]byte, error) {
	switch ns.method {
	case "new_account":
		index := common.NameToIndex(ns.params[0])
		addr := common.AddressFormHexString(ns.params[1])
		if _, err := ns.state.AddAccount(index, addr, ns.timeStamp); err != nil {
			return nil, err
		}
	case "set_account":
		index := common.NameToIndex(ns.params[0])
		perm := state.Permission{Keys: make(map[string]state.KeyFactor, 1), Accounts: make(map[string]state.AccFactor, 1)}
		if err := json.Unmarshal([]byte(ns.params[1]), &perm); err != nil {
			fmt.Println(ns.params[1])
			return nil, err
		}
		if err := ns.state.AddPermission(index, perm); err != nil {
			return nil, err
		}
	case "reg_prod":
		index := common.NameToIndex(ns.params[0])
		ns.state.RegisterProducer(index)
	case "vote":
		from := common.NameToIndex(ns.params[0])
		to1 := common.NameToIndex(ns.params[1])
		to2 := common.NameToIndex(ns.params[2])
		accounts := []common.AccountName{to1, to2}
		ns.state.ElectionToVote(from, accounts)
	case "reg_chain":
		index := common.NameToIndex(ns.params[0])
		consensus := ns.params[1]
		addr := common.AddressFormHexString(ns.params[2])
		data := []byte(index.String() + consensus + addr.HexString())
		hash := common.SingleHash(data)
		if err := ns.state.RegisterChain(index, hash, ns.tx.Hash, addr); err != nil {
			return nil, err
		}
		if  ns.state.StateType()== state.FinalType {
			if consensus == "solo" {
				msg := &message.RegChain{ChainID: hash, TxHash: ns.tx.Hash, Address: addr}
				event.Send(event.ActorNil, event.ActorConsensusSolo, msg)
			} else {
				log.Warn("not support now")
			}
		}

	case "pledge":
		from := common.NameToIndex(ns.params[0])
		to := common.NameToIndex(ns.params[1])
		cpu, err := strconv.ParseUint(ns.params[2], 10, 64)
		if err != nil {
			return nil, err
		}
		net, err := strconv.ParseUint(ns.params[3], 10, 64)
		if err != nil {
			return nil, err
		}

		//log.Debug(from, to, cpu, net)
		if err := ns.state.SetResourceLimits(from, to, cpu, net, ns.cpuLimit, ns.netLimit); err != nil {
			return nil, err
		}
		return nil, nil
	case "cancel_pledge":
		from := common.NameToIndex(ns.params[0])
		to := common.NameToIndex(ns.params[1])
		cpu, err := strconv.ParseUint(ns.params[2], 10, 64)
		if err != nil {
			return nil, err
		}
		net, err := strconv.ParseUint(ns.params[3], 10, 64)
		if err != nil {
			return nil, err
		}
		log.Debug(from, to, cpu, net)
		if err := ns.state.CancelDelegate(from, to, cpu, net, ns.cpuLimit, ns.netLimit); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(log, fmt.Sprintf("unknown method:%s", ns.method))
	}
	return nil, nil
}
