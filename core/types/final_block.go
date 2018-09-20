package types

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/pb"
)

type FinalBlockHeader struct {
	ChainID            common.Hash
	Version            uint32
	Height             uint64
	Timestamp          int64
	TrxCount           uint32
	PrevHash           common.Hash
	//ConsData           ConsensusData
	ProposalPubKey     []byte
	EpochNo            uint64
	CMBlockHash        common.Hash
	TrxRootHash        common.Hash
	StateDeltaRootHash common.Hash
	MinorBlocksHash    common.Hash
	StateHashRoot      common.Hash

	hash common.Hash
	*COSign
}

func (h *FinalBlockHeader) ComputeHash() error {
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

func (h *FinalBlockHeader) proto() (*pb.FinalBlockHeader, error) {
	/*if h.ConsData.Payload == nil {
		return nil, errors.New(log, "the minor block header's consensus data is nil")
	}
	pbCon, err := h.ConsData.ProtoBuf()
	if err != nil {
		return nil, err
	}*/
	pbHeader := &pb.FinalBlockHeader{
		ChainID:            h.ChainID.Bytes(),
		Version:            h.Version,
		PrevHash:           h.PrevHash.Bytes(),
		Height:             h.Height,
		Timestamp:          h.Timestamp,
		TrxCount:           h.TrxCount,
		ProposalPubKey:     common.CopyBytes(h.ProposalPubKey),
		EpochNo:          h.EpochNo,
		CMBlockHash:        h.CMBlockHash.Bytes(),
		TrxRootHash:        h.TrxRootHash.Bytes(),
		StateDeltaRootHash: h.StateDeltaRootHash.Bytes(),
		MinorBlocksHash:    h.MinorBlocksHash.Bytes(),
		StateHashRoot:      h.StateHashRoot.Bytes(),
		//ConsData:           pbCon,
		Hash:               h.hash.Bytes(),
		COSign: &pb.COSign{
			Step1: h.COSign.Step1,
			Step2: h.COSign.Step2,
		},
	}
	return pbHeader, nil
}

func (h *FinalBlockHeader) unSignatureData() ([]byte, error) {
	pbHeader, err := h.proto()
	if err != nil {
		return nil, err
	}
	pbHeader.Hash = nil
	data, err := pbHeader.Marshal()
	if err != nil {
		return nil, errors.New(log, fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
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
		return nil, errors.New(log, fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
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
	//h.ConsData = ConsensusData{}
	h.StateDeltaRootHash = common.NewHash(pbHeader.StateDeltaRootHash)
	h.MinorBlocksHash = common.NewHash(pbHeader.MinorBlocksHash)
	h.StateHashRoot = common.NewHash(pbHeader.StateHashRoot)
	h.hash = common.NewHash(pbHeader.Hash)
	h.COSign = &COSign{
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

func (h *FinalBlockHeader) JsonString() string {
	data, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(data)
}

func (h *FinalBlockHeader) Type() uint32 {
	return uint32(HeFinalBlock)
}

func (h *FinalBlockHeader) Hash() common.Hash {
	return h.hash
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
		return errors.New(log, "input data's length is zero")
	}
	var pbBlock pb.FinalBlock
	if err := pbBlock.Unmarshal(data); err != nil {
		return err
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
	return string(data)
}
