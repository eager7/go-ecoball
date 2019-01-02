package shard

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/core/trie"
	"github.com/ecoball/go-ecoball/core/types"
)

var log = elog.NewLogger("core-shard", elog.NoticeLog)

type MinorBlockHeader struct {
	ChainID           common.Hash
	Version           uint32
	Height            uint64
	Timestamp         int64
	PrevHash          common.Hash
	TrxHashRoot       common.Hash
	StateRootHash     common.Hash
	StateDeltaHash    common.Hash
	CMBlockHash       common.Hash
	ProposalPublicKey []byte
	ShardId           uint32
	CMEpochNo         uint64

	Receipt types.BlockReceipt
	Hashes  common.Hash
	*types.COSign
}

func (h *MinorBlockHeader) ComputeHash() error {
	data, err := h.unSignatureData()
	if err != nil {
		return err
	}
	h.Hashes, err = common.DoubleHash(data)
	if err != nil {
		return err
	}
	return nil
}

func (h *MinorBlockHeader) VerifySignature() (bool, error) {
	/*for _, v := range h.Signatures {
		b, err := secp256k1.Verify(h.Hash.Bytes(), v.SigData, v.PubKey)
		if err != nil || b != true {
			return false, err
		}
	}*/
	return true, nil
}

func (h *MinorBlockHeader) proto() (*pb.MinorBlockHeader, error) {
	pbHeader := &pb.MinorBlockHeader{
		ChainID:           h.ChainID.Bytes(),
		Version:           h.Version,
		Height:            h.Height,
		Timestamp:         h.Timestamp,
		PrevHash:          h.PrevHash.Bytes(),
		TrxHashRoot:       h.TrxHashRoot.Bytes(),
		StateDeltaHash:    h.StateDeltaHash.Bytes(),
		StateRootHash:     h.StateRootHash.Bytes(),
		CMBlockHash:       h.CMBlockHash.Bytes(),
		ProposalPublicKey: h.ProposalPublicKey,
		ShardId:           h.ShardId,
		CMEpochNo:         h.CMEpochNo,
		Receipt: &pb.BlockReceipt{
			BlockCpu: h.Receipt.BlockCpu,
			BlockNet: h.Receipt.BlockNet,
		},
		Hash:   h.Hashes.Bytes(),
		COSign: h.COSign.Proto(),
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
	pbHeader.COSign.Sign1 = nil
	pbHeader.COSign.Sign2 = nil
	pbHeader.COSign.Step1 = 0
	pbHeader.COSign.Step2 = 0
	data, err := pbHeader.Marshal()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
	}
	return data, nil
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
	return "hash:" + h.Hashes.HexString() + string(data)
}

func (h *MinorBlockHeader) Type() uint32 {
	return uint32(HeMinorBlock)
}

func (h *MinorBlockHeader) Hash() common.Hash {
	return h.Hashes
}
func (h *MinorBlockHeader) GetHeight() uint64 {
	return h.Height
}
func (h *MinorBlockHeader) GetChainID() common.Hash {
	return h.ChainID
}

func (h *MinorBlockHeader) Identify() mpb.Identify {
	return mpb.Identify_APP_MSG_MINOR_BLOCK
}
func (h *MinorBlockHeader) String() string {
	data, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return "hash:" + h.Hashes.HexString() + string(data)
}
func (h MinorBlockHeader) GetInstance() interface{} {
	return h
}
func (h *MinorBlockHeader) Serialize() ([]byte, error) {
	pbHeader, err := h.proto()
	if err != nil {
		return nil, err
	}
	data, err := pbHeader.Marshal()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
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
	h.StateRootHash = common.NewHash(pbHeader.StateRootHash)
	h.CMBlockHash = common.NewHash(pbHeader.CMBlockHash)
	h.ProposalPublicKey = common.CopyBytes(pbHeader.ProposalPublicKey)
	h.ShardId = pbHeader.ShardId
	h.CMEpochNo = pbHeader.CMEpochNo
	h.Hashes = common.NewHash(pbHeader.Hash)
	h.Receipt = types.BlockReceipt{BlockCpu: pbHeader.Receipt.BlockCpu, BlockNet: pbHeader.Receipt.BlockNet}
	h.COSign = &types.COSign{
		TPubKey: pbHeader.COSign.TPubKey,
		Step1:   pbHeader.COSign.Step1,
		Sign1:   nil,
		Step2:   pbHeader.COSign.Step2,
		Sign2:   nil,
	}
	h.COSign.Sign1 = append(h.COSign.Sign1, pbHeader.COSign.Sign1...)
	h.COSign.Sign2 = append(h.COSign.Sign2, pbHeader.COSign.Sign2...)

	return nil
}

type AccountMinor struct {
	Type    types.TxType
	Receipt types.TransactionReceipt
}

func (a *AccountMinor) proto() (*pb.AccountMinor, error) {
	data, err := a.Receipt.Serialize()
	if err != nil {
		return nil, err
	}
	return &pb.AccountMinor{
		AccountData: data,
		Type:        uint64(a.Type),
	}, nil
}

func (a *AccountMinor) Hash() (common.Hash, error) {
	if p, err := a.proto(); err != nil {
		return common.Hash{}, err
	} else {
		if data, err := p.Marshal(); err != nil {
			return common.Hash{}, errors.New(err.Error())
		} else {
			return common.DoubleHash(data)
		}
	}
}

type MinorBlock struct {
	MinorBlockHeader
	Transactions []*types.Transaction
	StateDelta   []*AccountMinor
}

func NewMinorBlock(header MinorBlockHeader, prevHeader *MinorBlockHeader, txs []*types.Transaction, cpu, net float64) (*MinorBlock, error) {
	var hashes []common.Hash
	var sDelta []*AccountMinor
	for _, tx := range txs {
		delta := AccountMinor{Type: tx.Type, Receipt: tx.Receipt}
		if h, err := delta.Hash(); err != nil {
			return nil, err
		} else {
			hashes = append(hashes, h)
		}
		sDelta = append(sDelta, &delta)
	}
	merkleHash, err := trie.GetMerkleRoot(hashes)
	if err != nil {
		return nil, err
	}
	header.StateDeltaHash = merkleHash
	if err := header.ComputeHash(); err != nil {
		return nil, err
	}
	block := &MinorBlock{
		MinorBlockHeader: header,
		Transactions:     txs,
		StateDelta:       sDelta,
	}
	if err := block.SetReceipt(prevHeader, cpu, net); err != nil {
		return nil, err
	}
	return block, nil
}

func (b *MinorBlock) SetSignature(account *account.Account) error {
	sigData, err := account.Sign(b.Hashes.Bytes())
	if err != nil {
		return err
	}
	sig := common.Signature{}
	sig.SigData = common.CopyBytes(sigData)
	sig.PubKey = common.CopyBytes(account.PublicKey)
	//t.Signatures = append(t.Signatures, sig)
	return nil
}

func (b *MinorBlock) SetReceipt(prevHeader *MinorBlockHeader, cpu, net float64) error {
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
	log.Info("the new block limit is :", cpuLimit, netLimit)
	b.MinorBlockHeader.Receipt.BlockCpu = config.BlockCpuLimit
	b.MinorBlockHeader.Receipt.BlockNet = config.BlockNetLimit
	return nil
}

func (b *MinorBlock) GetTransaction(hash common.Hash) (*types.Transaction, error) {
	for _, tx := range b.Transactions {
		if tx.Hash.Equals(&hash) {
			return tx, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("can't find the tx:%s", hash.HexString()))
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

func (b MinorBlock) GetObject() interface{} {
	return b
}

func (b *MinorBlock) JsonString() string {
	data, err := json.Marshal(b)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return "hash:" + b.Hashes.HexString() + string(data)
}

func (b *MinorBlock) String() string {
	data, err := json.Marshal(b)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return "hash:" + b.Hashes.HexString() + string(data)
}
func (b MinorBlock) GetInstance() interface{} {
	return b
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
		return errors.New("input data's length is zero")
	}
	var pbBlock pb.MinorBlock
	if err := pbBlock.Unmarshal(data); err != nil {
		return errors.New(err.Error())
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
		receipt := types.TransactionReceipt{}
		if err := receipt.Deserialize(acc.AccountData); err != nil {
			return err
		}
		stateDelta := AccountMinor{
			Type:    types.TxType(acc.Type),
			Receipt: receipt,
		}
		b.StateDelta = append(b.StateDelta, &stateDelta)
	}
	return nil
}
