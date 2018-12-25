package types

import (
	"github.com/ecoball/go-ecoball/common"
	"math/big"
	"github.com/ecoball/go-ecoball/core/pb"
	"fmt"
	"github.com/ecoball/go-ecoball/common/errors"
	"sort"
)

type TransactionReceipt struct {
	From   		common.AccountName
	To     		common.AccountName
	TokenName	string
	Amount 		*big.Int
	Hash   		common.Hash
	Cpu    		float64
	Net    		float64
	NewToken 	[]byte
	Accounts 	map[int][]byte
	Producer	uint64
	Result 		[]byte
}

type BlockReceipt struct {
	BlockCpu float64
	BlockNet float64
}

func (r *TransactionReceipt) Serialize() ([]byte, error) {
	amount, err := r.Amount.GobEncode()
	if err != nil {
		return nil, err
	}

	var keysAcc []int
	for k := range r.Accounts {
		keysAcc = append(keysAcc, k)
	}
	sort.Ints(keysAcc)

	var accounts [][]byte
	for _, k := range keysAcc {
		accounts = append(accounts, r.Accounts[k])
	}
	p := &pb.TransactionReceipt{
		From:   	uint64(r.From),
		To:     	uint64(r.To),
		TokenName:	r.TokenName,
		Amount:		amount,
		Hash: 		r.Hash.Bytes(),
		Cpu: 		r.Cpu,
		Net: 		r.Net,
		NewToken:	r.NewToken,
		Accounts:	accounts,
		Producer:	r.Producer,
		Result: 	common.CopyBytes(r.Result),
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *TransactionReceipt) Deserialize(data []byte) (error) {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var receipt pb.TransactionReceipt
	if err := receipt.Unmarshal(data); err != nil {
		return err
	}

	amount := new(big.Int)
	if err := amount.GobDecode(receipt.Amount); err != nil {
		return errors.New(fmt.Sprintf("GobDecode err:%s", err.Error()))
	}
	r.Amount = amount
	r.TokenName = receipt.TokenName
	r.From = common.AccountName(receipt.From)
	r.To = common.AccountName(receipt.To)
	r.Cpu = receipt.Cpu
	r.Net = receipt.Net
	r.Hash = common.NewHash(receipt.Hash)
	r.NewToken = receipt.NewToken
	r.Accounts = make(map[int][]byte)
	for k, v := range receipt.Accounts {
		r.Accounts[k] = v
	}
	r.Producer = receipt.Producer
	r.Result = common.CopyBytes(receipt.Result)


	return nil
}