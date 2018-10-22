package shard

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/pb"
	"math/big"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/common/config"
)

var log = elog.NewLogger("sharding", elog.NoticeLog)

type MinorBlockHeader struct {
	ChainID           common.Hash
	Version           uint32
	Height            uint64
	Timestamp         int64
	PrevHash          common.Hash
	TrxHashRoot       common.Hash
	StateDeltaHash    common.Hash
	CMBlockHash       common.Hash
	ProposalPublicKey []byte
	//ConsData          ConsensusData
	ShardId           uint32
	CMEpochNo         uint64

	Receipt types.BlockReceipt
	hash    common.Hash
	*types.COSign
}

func (h *MinorBlockHeader) ComputeHash() error {
	data, err := h.unSignatureData()
	if err != nil {
		return err
	}
	h.hash, err = common.DoubleHash(data)
	if err != nil {
		return err
	}
	return nil
}

func (h *MinorBlockHeader) proto() (*pb.MinorBlockHeader, error) {
	/*if h.ConsData.Payload == nil {
		return nil, errors.New(log, "the minor block header's consensus data is nil")
	}
	pbCon, err := h.ConsData.ProtoBuf()
	if err != nil {
		return nil, err
	}*/
	pbHeader := &pb.MinorBlockHeader{
		ChainID:           h.ChainID.Bytes(),
		Version:           h.Version,
		Height:            h.Height,
		Timestamp:         h.Timestamp,
		PrevHash:          h.PrevHash.Bytes(),
		TrxHashRoot:       h.TrxHashRoot.Bytes(),
		StateDeltaHash:    h.StateDeltaHash.Bytes(),
		CMBlockHash:       h.CMBlockHash.Bytes(),
		ProposalPublicKey: h.ProposalPublicKey,
		//ConsData:          pbCon,
		ShardId:           h.ShardId,
		CMEpochNo:         h.CMEpochNo,
		Receipt: &pb.BlockReceipt{
			BlockCpu: h.Receipt.BlockCpu,
			BlockNet: h.Receipt.BlockNet,
		},
		Hash: h.hash.Bytes(),
		COSign: &pb.COSign{
			Step1: h.COSign.Step1,
			Step2: h.COSign.Step2,
		},
	}
	return pbHeader, nil
}

func (h *MinorBlockHeader) unSignatureData() ([]byte, error) {
	pbHeader, err := h.proto()
	if err != nil {
		return nil, err
	}
	pbHeader.Receipt = nil
	pbHeader.Hash = nil
	pbHeader.COSign = nil
	data, err := pbHeader.Marshal()
	if err != nil {
		return nil, errors.New(log, fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
	}
	return data, nil
}

func (h *MinorBlockHeader) Serialize() ([]byte, error) {
	pbHeader, err := h.proto()
	if err != nil {
		return nil, err
	}
	data, err := pbHeader.Marshal()
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
	//h.ConsData = ConsensusData{}
	h.ShardId = pbHeader.ShardId
	h.CMEpochNo = pbHeader.CMEpochNo
	h.hash = common.NewHash(pbHeader.Hash)
	h.Receipt = types.BlockReceipt{BlockCpu: pbHeader.Receipt.BlockCpu, BlockNet: pbHeader.Receipt.BlockNet}
	h.COSign = &types.COSign{
		Step1: pbHeader.COSign.Step1,
		Step2: pbHeader.COSign.Step2,
	}
	/*dataCon, err := pbHeader.ConsData.Marshal()
	if err != nil {
		return err
	}
	if err := h.ConsData.Deserialize(dataCon); err != nil {
		return err
	}*/

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

func (h *MinorBlockHeader) Type() uint32 {
	return uint32(HeMinorBlock)
}

func (h *MinorBlockHeader) Hash() common.Hash {
	return h.hash
}
func (h *MinorBlockHeader) GetHeight() uint64 {
	return h.Height
}
func (h *MinorBlockHeader) GetChainID() common.Hash {
	return h.ChainID
}

type AccountMinor struct {
	Accounts *state.Account
	//Account common.AccountName
	//Balance *big.Int
	//Nonce   uint64
}

func (a *AccountMinor) proto() (*pb.AccountMinor, error) {
	data, err := a.Accounts.Serialize()
	if err != nil {
		return nil, err
	}
	return &pb.AccountMinor{
		AccountData: data,
		Account:     0,
		Balance:     nil,
		Nonce:       0,
	}, nil
}

type MinorBlock struct {
	MinorBlockHeader
	Transactions []*types.Transaction
	StateDelta   []*AccountMinor
}

func NewMinorBlock(header MinorBlockHeader, prevHeader *types.Header, txs []*types.Transaction, cpu, net float64) (*MinorBlock, error) {
	if err := header.ComputeHash(); err != nil {
		return nil, err
	}
	var sDelta []*AccountMinor
	for _, tx := range txs {
		//receipt := tx.Receipt
		switch tx.Type {
		case types.TxDeploy:
			/*acc := state.Account{
				Index:       0,
				TimeStamp:   0,
				Tokens:      nil,
				Permissions: nil,
				Contract:    types.DeployInfo{
					TypeVm:   0,
					Describe: nil,
					Code:     nil,
					Abi:      nil,
				},
				Delegates:   nil,
				Resource:    state.Resource{
					Ram: struct {
						Quota float64 `json:"quota"`
						Used  float64 `json:"used"`
					}{},
					Net: struct {
						Staked    uint64  `json:"staked_aba, omitempty"`
						Delegated uint64  `json:"delegated_aba, omitempty"`
						Used      float64 `json:"used_byte, omitempty"`
						Available float64 `json:"available_byte, omitempty"`
						Limit     float64 `json:"limit_byte, omitempty"`
					}{
						Staked:    0,
						Delegated: 0,
						Used:      0,
						Available: 0,
						Limit:     0,
					},
					Cpu: struct {
						Staked    uint64  `json:"staked_aba, omitempty"`
						Delegated uint64  `json:"delegated_aba, omitempty"`
						Used      float64 `json:"used_ms, omitempty"`
						Available float64 `json:"available_ms, omitempty"`
						Limit     float64 `json:"limit_ms, omitempty"`
					}{
						Staked:    0,
						Delegated: 0,
						Used:      0,
						Available: 0,
						Limit:     0,
					},
					Votes: struct {
						Staked    uint64                        `json:"staked_aba, omitempty"`
						Producers map[common.AccountName]uint64 `json:"producers, omitempty"`
					}{
						Staked:    0,
						Producers: nil,
					},
				},
				Hash:        common.Hash{},
			}*/
		case types.TxInvoke:
		case types.TxTransfer:
		default:
			return nil, errors.New(log, "unknown transaction type")
		}
	}
	block := &MinorBlock{
		MinorBlockHeader: header,
		Transactions:     txs,
		StateDelta:       sDelta,
	}
	fmt.Println("block.StateDelta:",block.StateDelta)
	if err := block.SetReceipt(prevHeader, cpu, net); err != nil {
		return nil, err
	}
	return block, nil
}

func (b *MinorBlock) SetReceipt(prevHeader *types.Header, cpu, net float64) error {
	if prevHeader == nil {
		return nil
	}
	var cpuLimit, netLimit float64
	if cpu < (config.BlockCpuLimit / 10) {
		cpuLimit = prevHeader.Receipt.BlockCpu * 1.01
		if cpuLimit > config.VirtualBlockCpuLimit {
			cpuLimit = config.VirtualBlockCpuLimit
		}
	} else {
		cpuLimit = prevHeader.Receipt.BlockCpu * 0.99
		if cpuLimit < config.BlockCpuLimit {
			cpuLimit = config.BlockCpuLimit
		}
	}
	if net < (config.BlockNetLimit / 10) {
		netLimit = prevHeader.Receipt.BlockNet * 1.01
		if netLimit > config.VirtualBlockNetLimit {
			netLimit = config.VirtualBlockNetLimit
		}
	} else {
		netLimit = prevHeader.Receipt.BlockNet * 0.99
		if netLimit < config.BlockNetLimit {
			netLimit = config.BlockNetLimit
		}
	}
	b.MinorBlockHeader.Receipt.BlockCpu = cpuLimit
	b.MinorBlockHeader.Receipt.BlockNet = netLimit
	return nil
}

func (b *MinorBlock) proto() (block *pb.MinorBlock, err error) {
	var pbBlock pb.MinorBlock
	pbBlock.Header, err = b.MinorBlockHeader.proto()
	if err != nil {
		return nil, err
	}

	for _, tx := range b.Transactions {
		pbTx, err := tx.ProtoBuf()
		if err != nil {
			return nil, err
		}
		pbBlock.Transactions = append(pbBlock.Transactions, pbTx)
	}
	for _, acc := range b.StateDelta {
		pbState, err := acc.proto()
		if err != nil {
			return nil, err
		}
		pbBlock.StateDelta = append(pbBlock.StateDelta, pbState)
	}

	return &pbBlock, nil
}

func (b *MinorBlock) Serialize() ([]byte, error) {
	p, err := b.proto()
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
	err = b.MinorBlockHeader.Deserialize(dataHeader)
	if err != nil {
		return err
	}

	for _, tx := range pbBlock.Transactions {
		if bytes, err := tx.Marshal(); err != nil {
			return err
		} else {
			t := new(types.Transaction)
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
		account := new(state.Account)
		if err := account.Deserialize(acc.AccountData); err != nil {
			return err
		}
		state := AccountMinor{
			Accounts: account,
		}
		b.StateDelta = append(b.StateDelta, &state)
	}
	return nil
}

func (b MinorBlock) GetObject() interface{} {
	return b
}

func (b *MinorBlock) JsonString() string {
	data, err := json.Marshal(b)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(data)
}
