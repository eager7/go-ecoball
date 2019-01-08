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

package state

import (
	"encoding/json"
	"fmt"
	. "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/core/trie"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/gogo/protobuf/proto"
	"math/big"
	"sort"
	"sync"
)

type Account struct {
	Index       AccountName           `json:"index"`
	Tokens      map[string]Token      `json:"token"` //map[token name]Token
	Elector     Elector               `json:"elector"`
	Resource    Resource              `json:"resource"`
	Contract    types.DeployInfo      `json:"contract"`
	TimeStamp   int64                 `json:"timestamp"`
	Delegates   []Delegate            `json:"delegate"`
	Permissions map[string]Permission `json:"permissions"` //map[perm name]Permission
	Hash        Hash                  `json:"hash"`        //此哈希是mpt树的根哈希
	mpt         *trie.Mpt             //用于存储合约数据,通过store set和get接口操作
	lock        sync.RWMutex          //用于整个结构体的锁,在调用getAccountByName后上锁,不对单独成员变量上锁
}

/**
 *  @brief create a new account, binding a char name with a address
 *  @param index - the unique id of account name created by common.NameToIndex()
 *  @param address - the account's public key
 */
func NewAccount(path string, index AccountName, addr Address, timeStamp int64) (acc *Account, err error) {
	res := Resource{Votes: struct {
		Staked    uint64                 `json:"staked_aba, omitempty"`
		Producers map[AccountName]uint64 `json:"producers, omitempty"`
	}{Staked: 0, Producers: make(map[AccountName]uint64, 1)}}
	acc = &Account{
		Index:       index,
		TimeStamp:   timeStamp,
		Tokens:      make(map[string]Token, 1),
		Permissions: make(map[string]Permission, 1),
		Resource:    res,
	}
	perm := NewPermission(Owner, "", 1, []KeyFactor{{Actor: addr, Weight: 1}}, []AccFactor{})
	acc.AddPermission(perm)
	perm = NewPermission(Active, Owner, 1, []KeyFactor{{Actor: addr, Weight: 1}}, []AccFactor{})
	acc.AddPermission(perm)

	return acc, nil
}

func (a *Account) proto() (*pb.Account, error) {
	var tokens []*pb.Token
	var keysToken []string
	for k := range a.Tokens {
		keysToken = append(keysToken, k)
	}
	sort.Strings(keysToken)
	for _, k := range keysToken {
		v := a.Tokens[k]
		balance, err := v.Balance.GobEncode()
		if err != nil {
			return nil, err
		}
		t := pb.Token{
			Name:    v.Name,
			Balance: balance,
		}
		tokens = append(tokens, &t)
	}

	var perms []*pb.Permission
	var keysPerm []string
	for _, perm := range a.Permissions {
		keysPerm = append(keysPerm, perm.PermName)
	}
	sort.Strings(keysPerm)
	for _, k := range keysPerm {
		perm := a.Permissions[k]
		var pbKeys []*pb.KeyWeight
		var pbAccounts []*pb.AccountWeight
		var keysKeys []string
		var keysAccount []string
		for k := range perm.Keys {
			keysKeys = append(keysKeys, k)
		}
		sort.Strings(keysKeys)
		for _, k := range keysKeys {
			key := perm.Keys[k]
			pbKey := &pb.KeyWeight{Actor: key.Actor.Bytes(), Weight: key.Weight}
			pbKeys = append(pbKeys, pbKey)
		}

		for k := range perm.Accounts {
			keysAccount = append(keysAccount, k)
		}
		sort.Strings(keysAccount)
		for _, k := range keysAccount {
			acc := perm.Accounts[k]
			pbAccount := &pb.AccountWeight{Actor: uint64(acc.Actor), Weight: acc.Weight, Permission: []byte(acc.Permission)}
			pbAccounts = append(pbAccounts, pbAccount)
		}

		pbPerm := &pb.Permission{
			PermName:  []byte(perm.PermName),
			Parent:    []byte(perm.Parent),
			Threshold: perm.Threshold,
			Keys:      pbKeys,
			Accounts:  pbAccounts,
		}
		perms = append(perms, pbPerm)
	}
	var delegates []*pb.Delegate
	for _, v := range a.Delegates {
		d := pb.Delegate{Index: uint64(v.Index), Cpu: v.CpuStaked, Net: v.NetStaked}
		delegates = append(delegates, &d)
	}
	var producers []*pb.Producer
	var keysVotes []float64
	for name := range a.Resource.Votes.Producers {
		keysVotes = append(keysVotes, float64(name))
	}
	sort.Float64s(keysVotes)
	for _, v := range keysVotes {
		producer := pb.Producer{AccountName: uint64(v), Amount: a.Resource.Votes.Producers[AccountName(v)]}
		producers = append(producers, &producer)
	}
	pbAcc := pb.Account{
		Index:       uint64(a.Index),
		TimeStamp:   a.TimeStamp,
		Tokens:      tokens,
		Permissions: perms,
		Contract: &pb.DeployInfo{
			TypeVm:   uint32(a.Contract.TypeVm),
			Describe: CopyBytes(a.Contract.Describe),
			Code:     CopyBytes(a.Contract.Code),
			Abi:      CopyBytes(a.Contract.Abi),
		},
		Delegates: delegates,
		Cpu: &pb.Res{
			Staked:    a.Resource.Cpu.Staked,
			Delegated: a.Resource.Cpu.Delegated,
			Used:      a.Resource.Cpu.Used,
			Available: a.Resource.Cpu.Available,
			Limit:     a.Resource.Cpu.Limit,
		},
		Net: &pb.Res{
			Staked:    a.Resource.Net.Staked,
			Delegated: a.Resource.Net.Delegated,
			Used:      a.Resource.Net.Used,
			Available: a.Resource.Net.Available,
			Limit:     a.Resource.Net.Limit,
		},
		Votes: &pb.Votes{
			Staked:    a.Resource.Votes.Staked,
			Producers: producers,
		},
		Elector: &pb.Elector{
			Index:   a.Elector.Index.Number(),
			Amount:  a.Elector.Amount,
			B64Pub:  a.Elector.B64Pub,
			Address: a.Elector.Address,
			Port:    a.Elector.Port,
			Payee:   a.Elector.Payee.Number(),
		},
		Hash: a.Hash.Bytes(),
	}
	return &pbAcc, nil
}
func (a *Account) Serialize() ([]byte, error) {
	p, err := a.proto()
	if err != nil {
		return nil, err
	}
	data, err := proto.Marshal(p)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (a *Account) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input Token's length is zero")
	}
	var pbAcc pb.Account
	if err := proto.Unmarshal(data, &pbAcc); err != nil {
		return err
	}
	a.Index = AccountName(pbAcc.Index)
	a.TimeStamp = pbAcc.TimeStamp

	a.Resource.Cpu.Staked = pbAcc.Cpu.Staked
	a.Resource.Cpu.Delegated = pbAcc.Cpu.Delegated
	a.Resource.Cpu.Used = pbAcc.Cpu.Used
	a.Resource.Cpu.Available = pbAcc.Cpu.Available
	a.Resource.Cpu.Limit = pbAcc.Cpu.Limit
	a.Resource.Net.Staked = pbAcc.Net.Staked
	a.Resource.Net.Delegated = pbAcc.Net.Delegated
	a.Resource.Net.Used = pbAcc.Net.Used
	a.Resource.Net.Available = pbAcc.Net.Available
	a.Resource.Net.Limit = pbAcc.Net.Limit
	a.Resource.Votes.Staked = pbAcc.Votes.Staked
	a.Elector.Index = AccountName(pbAcc.Elector.Index)
	a.Elector.Payee = AccountName(pbAcc.Elector.Payee)
	a.Elector.B64Pub = pbAcc.Elector.B64Pub
	a.Elector.Address = pbAcc.Elector.Address
	a.Elector.Port = pbAcc.Elector.Port
	a.Elector.Amount = pbAcc.Elector.Amount
	a.Resource.Votes.Producers = make(map[AccountName]uint64, 1)
	for _, v := range pbAcc.Votes.Producers {
		a.Resource.Votes.Producers[AccountName(v.AccountName)] = v.Amount
	}

	a.Hash = NewHash(pbAcc.Hash)
	a.Tokens = make(map[string]Token)
	a.Contract = types.DeployInfo{
		TypeVm:   types.VmType(pbAcc.Contract.TypeVm),
		Describe: CopyBytes(pbAcc.Contract.Describe),
		Code:     CopyBytes(pbAcc.Contract.Code),
		Abi:      CopyBytes(pbAcc.Contract.Abi),
	}
	a.Permissions = make(map[string]Permission, 1)
	for _, v := range pbAcc.Tokens {
		ac := Token{
			Name:    string(v.Name),
			Balance: new(big.Int),
		}
		if err := ac.Balance.GobDecode(v.Balance); err != nil {
			return err
		}
		a.Tokens[ac.Name] = ac
	}
	for _, v := range pbAcc.Delegates {
		a.Delegates = append(a.Delegates, Delegate{Index: AccountName(v.Index), CpuStaked: v.Cpu, NetStaked: v.Net})
	}
	for _, pbPerm := range pbAcc.Permissions {
		keys := make(map[string]KeyFactor, 1)
		for _, pbKey := range pbPerm.Keys {
			key := KeyFactor{Actor: NewAddress(pbKey.Actor), Weight: pbKey.Weight}
			keys[NewAddress(pbKey.Actor).HexString()] = key
		}
		accounts := make(map[string]AccFactor, 1)
		for _, pbAcc := range pbPerm.Accounts {
			acc := AccFactor{Actor: AccountName(pbAcc.Actor), Weight: pbAcc.Weight, Permission: string(pbAcc.Permission)}
			accounts[AccountName(pbAcc.Actor).String()] = acc
		}
		a.Permissions[string(pbPerm.PermName)] = Permission{
			PermName:  string(pbPerm.PermName),
			Parent:    string(pbPerm.Parent),
			Threshold: pbPerm.Threshold,
			Keys:      keys,
			Accounts:  accounts,
		}
	}

	return nil
}
func (a *Account) GetInstance() interface{} {
	return a
}
func (a *Account) Identify() mpb.Identify {
	return mpb.Identify_APP_MSG_STATE_OBJECT
}
func (a *Account) String() string {
	data, err := json.Marshal(a)
	if err != nil {
		fmt.Println(err)
	}
	return string(data)
}
func (a *Account) Clone() (*Account, error) {
	n := new(Account)
	data, err := a.Serialize()
	if err != nil {
		log.Warn(err)
		return nil, err
	}
	if err := n.Deserialize(data); err != nil {
		log.Warn(err)
		return nil, err
	}
	return n, nil
}
