package types

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/pb"
	"math/big"
)

type MinorBlockHeader struct {
	ChainID           common.Hash
	Version           uint32
	Height            uint64
	Timestamp         uint64
	PrevHash          common.Hash
	TrxHashRoot       common.Hash
	StateDeltaHash    common.Hash
	CMBlockHash       common.Hash
	ProposalPublicKey []byte
	ConsData          ConsensusData
	ShardId           uint32
	CMEpochNo         uint64

	Receipt BlockReceipt
	Hash    common.Hash
}

func (h *MinorBlockHeader) ComputeHash() error {
	data, err := h.unSignatureData()
	if err != nil {
		return err
	}
	h.Hash, err = common.DoubleHash(data)
	if err != nil {
		return err
	}
	return nil
}

func (h *MinorBlockHeader) ProtoBuf() (*pb.MinorBlockHeader, error) {
	pbCon, err := h.ConsData.ProtoBuf()
	if err != nil {
		return nil, err
	}
	pbHeader := &pb.MinorBlockHeader{
		ChainID:           h.Hash.Bytes(),
		Version:           h.Version,
		Height:            h.Height,
		Timestamp:         h.Timestamp,
		PrevHash:          h.PrevHash.Bytes(),
		TrxHashRoot:       h.TrxHashRoot.Bytes(),
		StateDeltaHash:    h.StateDeltaHash.Bytes(),
		CMBlockHash:       h.CMBlockHash.Bytes(),
		ProposalPublicKey: h.ProposalPublicKey,
		ConsData:          pbCon,
		ShardId:           h.ShardId,
		CMEpochNo:         h.CMEpochNo,
		Receipt:           nil,
		Hash:              nil,
	}
	return pbHeader, nil
}

func (h *MinorBlockHeader) unSignatureData() ([]byte, error) {
	pbHeader, err := h.ProtoBuf()
	if err != nil {
		return nil, err
	}
	data, err := pbHeader.Marshal()
	if err != nil {
		return nil, errors.New(log, fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
	}
	return data, nil
}

func (h *MinorBlockHeader) Serialize() ([]byte, error) {
	pbCon, err := h.ConsData.ProtoBuf()
	if err != nil {
		return nil, err
	}
	protoHeader := pb.MinorBlockHeader{
		ChainID:           h.Hash.Bytes(),
		Version:           h.Version,
		Height:            h.Height,
		Timestamp:         h.Timestamp,
		PrevHash:          h.PrevHash.Bytes(),
		TrxHashRoot:       h.TrxHashRoot.Bytes(),
		StateDeltaHash:    h.StateDeltaHash.Bytes(),
		CMBlockHash:       h.CMBlockHash.Bytes(),
		ProposalPublicKey: h.ProposalPublicKey,
		ConsData:          pbCon,
		ShardId:           h.ShardId,
		CMEpochNo:         h.CMEpochNo,
		Receipt: &pb.BlockReceipt{
			BlockCpu: h.Receipt.BlockCpu,
			BlockNet: h.Receipt.BlockNet,
		},
		Hash: h.Hash.Bytes(),
	}
	data, err := protoHeader.Marshal()
	if err != nil {
		return nil, errors.New(log, fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
	}
	return data, nil
}

func (h *MinorBlockHeader) Deserialize(data []byte) error {
	var pbHeader pb.MinorBlockHeader
	if err := pbHeader.Unmarshal(data); err != nil {
		return err
	}

	h.ChainID = common.NewHash(pbHeader.ChainID)
	h.Version = pbHeader.Version
	h.Height = pbHeader.Height
	h.Timestamp = pbHeader.Timestamp
	h.PrevHash = common.NewHash(pbHeader.PrevHash)
	h.TrxHashRoot = common.NewHash(pbHeader.TrxHashRoot)
	h.StateDeltaHash = common.NewHash(pbHeader.StateDeltaHash)
	h.CMBlockHash = common.NewHash(pbHeader.CMBlockHash)
	h.ProposalPublicKey = common.CopyBytes(pbHeader.ProposalPublicKey)
	h.ConsData = ConsensusData{}
	h.ShardId = pbHeader.ShardId
	h.CMEpochNo = pbHeader.CMEpochNo
	h.Hash = common.NewHash(pbHeader.Hash)
	h.Receipt = BlockReceipt{BlockNet: pbHeader.Receipt.BlockNet, BlockCpu: pbHeader.Receipt.BlockCpu}

	dataCon, err := pbHeader.ConsData.Marshal()
	if err != nil {
		return err
	}
	if err := h.ConsData.Deserialize(dataCon); err != nil {
		return err
	}

	return nil
}

func (h MinorBlockHeader) GetObject() interface{} {
	return h
}

func (h *MinorBlockHeader) JsonString() string {
	data, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(data)
}

func (h *MinorBlockHeader) Show() {
	log.Debug(h.JsonString())
}

func (h *MinorBlockHeader) Type() uint32 {
	return uint32(HeMinorBlock)
}

type AccountMinor struct {
	Balance big.Int
	Nonce   big.Int
}

func (a *AccountMinor) ProtoBuf() (*pb.AccountMinor, error) {
	balance, err := a.Balance.GobEncode()
	if err != nil {
		return nil, err
	}
	nonce, err := a.Nonce.GobEncode()
	if err != nil {
		return nil, err
	}
	return &pb.AccountMinor{
		Balance: balance,
		Nonce:   nonce,
	}, nil
}

type MinorBlock struct {
	Header       *MinorBlockHeader
	Transactions []*Transaction
	StateDelta   []AccountMinor
}

func NewMinorBlock() {
	
}

func (b *MinorBlock) ProtoBuf() (block *pb.MinorBlock, err error) {
	var pbBlock pb.MinorBlock
	pbBlock.Header, err = b.Header.ProtoBuf()
	if err != nil {
		return nil, err
	}
	var pbTxs []*pb.Transaction
	for _, tx := range b.Transactions {
		pbTx, err := tx.protoBuf()
		if err != nil {
			return nil, err
		}
		pbTxs = append(pbTxs, pbTx)
	}
	var pbStates []*pb.AccountMinor
	for _, acc := range b.StateDelta {
		pbState, err := acc.ProtoBuf()
		if err != nil {
			return nil, err
		}
		pbStates = append(pbStates, pbState)
	}
	pbBlock.Transactions = append(pbBlock.Transactions, pbTxs...)
	pbBlock.StateDelta = append(pbBlock.StateDelta, pbStates...)

	return &pbBlock, nil
}

func (b *MinorBlock) Serialize() ([]byte, error) {
	p, err := b.ProtoBuf()
	if err != nil {
		return nil, err
	}
	data, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *MinorBlock) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New(log, "input data's length is zero")
	}
	var pbBlock pb.MinorBlock
	if err := pbBlock.Unmarshal(data); err != nil {
		return err
	}
	dataHeader, err := pbBlock.Header.Marshal()
	if err != nil {
		return err
	}
	err = b.Header.Deserialize(dataHeader)
	if err != nil {
		return err
	}

	for _, tx := range pbBlock.Transactions {
		if bytes, err := tx.Marshal(); err != nil {
			return err
		} else {
			t := new(Transaction)
			if err := t.Deserialize(bytes); err != nil {
				return err
			}
			b.Transactions = append(b.Transactions, t)
		}
	}

	for _, acc := range pbBlock.StateDelta {
		balance := new(big.Int)
		if err := balance.GobDecode(acc.Balance); err != nil {
			return err
		}
		nonce := new(big.Int)
		if err := balance.GobDecode(acc.Nonce); err != nil {
			return err
		}
		state := AccountMinor{
			Balance: *balance,
			Nonce:   *nonce,
		}
		b.StateDelta = append(b.StateDelta, state)
	}
	return nil
}
