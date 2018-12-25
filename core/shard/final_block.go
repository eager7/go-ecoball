package shard

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/core/types"
)

type FinalBlockHeader struct {
	ChainID            common.Hash
	Version            uint32
	Height             uint64
	Timestamp          int64
	TrxCount           uint32
	PrevHash           common.Hash
	ProposalPubKey     []byte
	EpochNo            uint64
	CMBlockHash        common.Hash
	TrxRootHash        common.Hash
	StateDeltaRootHash common.Hash
	MinorBlocksHash    common.Hash
	StateHashRoot      common.Hash

	Hashes common.Hash
	*types.COSign
}

func (h *FinalBlockHeader) ComputeHash() error {
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

func (h *FinalBlockHeader) VerifySignature() (bool, error) {
	/*for _, v := range h.Signatures {
		b, err := secp256k1.Verify(h.Hash.Bytes(), v.SigData, v.PubKey)
		if err != nil || b != true {
			return false, err
		}
	}*/
	return true, nil
}

func (h *FinalBlockHeader) proto() (*pb.FinalBlockHeader, error) {
	pbHeader := &pb.FinalBlockHeader{
		ChainID:            h.ChainID.Bytes(),
		Version:            h.Version,
		PrevHash:           h.PrevHash.Bytes(),
		Height:             h.Height,
		Timestamp:          h.Timestamp,
		TrxCount:           h.TrxCount,
		ProposalPubKey:     common.CopyBytes(h.ProposalPubKey),
		EpochNo:            h.EpochNo,
		CMBlockHash:        h.CMBlockHash.Bytes(),
		TrxRootHash:        h.TrxRootHash.Bytes(),
		StateDeltaRootHash: h.StateDeltaRootHash.Bytes(),
		MinorBlocksHash:    h.MinorBlocksHash.Bytes(),
		StateHashRoot:      h.StateHashRoot.Bytes(),
		Hash:               h.Hashes.Bytes(),
		COSign:             h.COSign.Proto(),
	}
	return pbHeader, nil
}

func (h *FinalBlockHeader) unSignatureData() ([]byte, error) {
	pbHeader, err := h.proto()
	if err != nil {
		return nil, err
	}
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

func (h *FinalBlockHeader) Serialize() ([]byte, error) {
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

func (h *FinalBlockHeader) Deserialize(data []byte) error {
	var pbHeader pb.FinalBlockHeader
	if err := pbHeader.Unmarshal(data); err != nil {
		return err
	}

	h.ChainID = common.NewHash(pbHeader.ChainID)
	h.Version = pbHeader.Version
	h.Height = pbHeader.Height
	h.Timestamp = pbHeader.Timestamp
	h.TrxCount = pbHeader.TrxCount
	h.PrevHash = common.NewHash(pbHeader.PrevHash)
	h.ProposalPubKey = common.CopyBytes(pbHeader.ProposalPubKey)
	h.EpochNo = pbHeader.EpochNo
	h.CMBlockHash = common.NewHash(pbHeader.CMBlockHash)
	h.TrxRootHash = common.NewHash(pbHeader.TrxRootHash)
	h.StateDeltaRootHash = common.NewHash(pbHeader.StateDeltaRootHash)
	h.MinorBlocksHash = common.NewHash(pbHeader.MinorBlocksHash)
	h.StateHashRoot = common.NewHash(pbHeader.StateHashRoot)
	h.Hashes = common.NewHash(pbHeader.Hash)
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

func (h *FinalBlockHeader) JsonString() string {
	data, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return "hash:" + h.Hashes.HexString() + string(data)
}

func (h *FinalBlockHeader) Type() uint32 {
	return uint32(HeFinalBlock)
}
func (h FinalBlockHeader) GetObject() interface{} {
	return h
}

func (h *FinalBlockHeader) Hash() common.Hash {
	return h.Hashes
}
func (h *FinalBlockHeader) GetHeight() uint64 {
	return h.Height
}
func (h *FinalBlockHeader) GetChainID() common.Hash {
	return h.ChainID
}

type FinalBlock struct {
	FinalBlockHeader
	MinorBlocks []*MinorBlockHeader
}

func NewFinalBlock(header FinalBlockHeader, minorBlocks []*MinorBlockHeader) (*FinalBlock, error) {
	if err := header.ComputeHash(); err != nil {
		return nil, err
	}
	return &FinalBlock{
		FinalBlockHeader: header,
		MinorBlocks:      minorBlocks,
	}, nil
}

func (b *FinalBlock) SetSignature(account *account.Account) error {
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

func (b *FinalBlock) proto() (block *pb.FinalBlock, err error) {
	var pbBlock pb.FinalBlock
	pbBlock.Header, err = b.FinalBlockHeader.proto()
	if err != nil {
		return nil, err
	}

	for _, h := range b.MinorBlocks {
		pbHeader, err := h.proto()
		if err != nil {
			return nil, err
		}
		pbBlock.MinorBlocks = append(pbBlock.MinorBlocks, pbHeader)
	}

	return &pbBlock, nil
}

func (b *FinalBlock) Serialize() ([]byte, error) {
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

func (b *FinalBlock) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var pbBlock pb.FinalBlock
	if err := pbBlock.Unmarshal(data); err != nil {
		return errors.New(err.Error())
	}
	dataHeader, err := pbBlock.Header.Marshal()
	if err != nil {
		return err
	}
	err = b.FinalBlockHeader.Deserialize(dataHeader)
	if err != nil {
		return err
	}

	for _, h := range pbBlock.MinorBlocks {
		if bytes, err := h.Marshal(); err != nil {
			return err
		} else {
			header := new(MinorBlockHeader)
			if err := header.Deserialize(bytes); err != nil {
				return err
			}
			b.MinorBlocks = append(b.MinorBlocks, header)
		}
	}

	return nil
}

func (b FinalBlock) GetObject() interface{} {
	return b
}

func (b *FinalBlock) JsonString() string {
	data, err := json.Marshal(b)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return "hash:" + b.Hashes.HexString() + string(data)
}
