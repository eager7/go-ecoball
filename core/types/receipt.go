package types

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/pb"
	"math/big"
)

type TrxReceipt struct {
	From   common.AccountName
	Addr   common.AccountName
	Token  string
	Amount *big.Int
	Cpu    float64
	Net    float64
	Result []byte
	Hash   common.Hash
}

type BlockReceipt struct {
	BlockCpu float64
	BlockNet float64
}

func (r *TrxReceipt) ComputeHash() error {
	p, err := r.proto()
	if err != nil {
		return err
	}
	p.Hash = nil
	data, err := p.Marshal()
	if err != nil {
		return err
	}
	r.Hash, err = common.DoubleHash(data)
	if err != nil {
		return err
	}
	return nil
}

func (r *TrxReceipt) IsBeSet() bool {
	return !r.Hash.Equals(&common.Hash{})
}

func (r *TrxReceipt) Identify() mpb.Identify {
	return mpb.Identify_APP_MSG_TRANSACTION_RECEIPT
}

func (r *TrxReceipt) GetInstance() interface{} {
	return r
}

func (r *TrxReceipt) String() string {
	return common.JsonString(r)
}

func (r *TrxReceipt) Serialize() ([]byte, error) {
	p, err := r.proto()
	if err != nil {
		return nil, err
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *TrxReceipt) proto() (*pb.TrxReceipt, error) {
	amount, err := r.Amount.GobEncode()
	if err != nil {
		return nil, err
	}
	return &pb.TrxReceipt{
		Cpu:    r.Cpu,
		Net:    r.Net,
		Result: common.CopyBytes(r.Result),
		From:   uint64(r.From),
		Addr:   uint64(r.Addr),
		Token:  r.Token,
		Amount: amount,
		Hash:   r.Hash.Bytes(),
	}, nil
}

func (r *TrxReceipt) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var receipt pb.TrxReceipt
	if err := receipt.Unmarshal(data); err != nil {
		return err
	}
	amount := new(big.Int)
	if err := amount.GobDecode(receipt.Amount); err != nil {
		return errors.New(fmt.Sprintf("GobDecode err:%s", err.Error()))
	}
	r.Amount = amount
	r.Token = receipt.Token
	r.From = common.AccountName(receipt.From)
	r.Addr = common.AccountName(receipt.Addr)
	r.Cpu = receipt.Cpu
	r.Net = receipt.Net
	r.Result = common.CopyBytes(receipt.Result)
	r.Hash = common.NewHash(receipt.Hash)
	return nil
}
