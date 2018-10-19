package types

import (
	"github.com/ecoball/go-ecoball/common"
	"math/big"
	"github.com/ecoball/go-ecoball/core/pb"
	"fmt"
	"github.com/ecoball/go-ecoball/common/errors"
)

type TransactionReceipt struct {
	TokenName	string
	From   		*big.Int
	To     		*big.Int
	Hash   		common.Hash
	Cpu    		float64
	Net    		float64
	Account 	[][]byte
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

	var accounts [][]byte
	for _, v := range r.Account {
		accounts = append(accounts, v)
	}
	p := &pb.TransactionReceipt{
		TokenName:	r.TokenName,
		From:   	from,
		To:     	to,
		Hash: 		r.Hash.Bytes(),
		Cpu: 		r.Cpu,
		Net: 		r.Net,
		Account:	accounts,
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
	for _, v := range receipt.Account {
		r.Account = append(r.Account, v)
	}
	r.Result = common.CopyBytes(receipt.Result)


	return nil
}