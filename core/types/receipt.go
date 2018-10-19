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
	TokenName	string
	From   		*big.Int
	To     		*big.Int
	Hash   		common.Hash
	Cpu    		float64
	Net    		float64
	Accounts 	map[int][]byte
	Result 		[]byte
}

type BlockReceipt struct {
	BlockCpu float64
	BlockNet float64
}

func (r *TransactionReceipt) Serialize() ([]byte, error) {
	from, err := r.From.GobEncode()
	if err != nil {
		return nil, err
	}
	to, err := r.To.GobEncode()
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
		TokenName:	r.TokenName,
		From:   	from,
		To:     	to,
		Hash: 		r.Hash.Bytes(),
		Cpu: 		r.Cpu,
		Net: 		r.Net,
		Accounts:	accounts,
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
		return errors.New(log, "input data's length is zero")
	}
	var receipt pb.TransactionReceipt
	if err := receipt.Unmarshal(data); err != nil {
		return err
	}

	from := new(big.Int)
	if err := from.GobDecode(receipt.From); err != nil {
		return errors.New(log, fmt.Sprintf("GobDecode err:%s", err.Error()))
	}
	to := new(big.Int)
	if err := to.GobDecode(receipt.To); err != nil {
		return errors.New(log, fmt.Sprintf("GobDecode err:%s", err.Error()))
	}

	r.TokenName = receipt.TokenName
	r.From = from
	r.To = to
	r.Cpu = receipt.Cpu
	r.Net = receipt.Net
	r.Hash = common.NewHash(receipt.Hash)
	r.Accounts = make(map[int][]byte)
	for k, v := range receipt.Accounts {
		r.Accounts[k] = v
	}
	r.Result = common.CopyBytes(receipt.Result)


	return nil
}