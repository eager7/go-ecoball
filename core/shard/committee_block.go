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

type NodeInfo struct {
	PublicKey []byte
	Address   string
	Port      string
}

type CMBlockHeader struct {
	ChainID   common.Hash
	Version   uint32
	Height    uint64
	Timestamp int64
	PrevHash  common.Hash
	//ConsData  ConsensusData

	LeaderPubKey []byte
	Nonce        uint32
	Candidate    NodeInfo
	ShardsHash   common.Hash

	Hashes common.Hash
	*types.COSign
}

func (h *CMBlockHeader) ComputeHash() error {
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

func (h *CMBlockHeader) VerifySignature() (bool, error) {
	/*for _, v := range h.Signatures {
		b, err := secp256k1.Verify(h.Hash.Bytes(), v.SigData, v.PubKey)
		if err != nil || b != true {
			return false, err
		}
	}*/
	return true, nil
}

func (h *CMBlockHeader) proto() (*pb.CMBlockHeader, error) {
	return &pb.CMBlockHeader{
		ChainID:      h.ChainID.Bytes(),
		Version:      h.Version,
		Height:       h.Height,
		Timestamp:    h.Timestamp,
		PrevHash:     h.PrevHash.Bytes(),
		LeaderPubKey: common.CopyBytes(h.LeaderPubKey),
		Nonce:        h.Nonce,
		Candidate: &pb.NodeInfo{
			PublicKey: h.Candidate.PublicKey,
			Address:   h.Candidate.Address,
			Port:      h.Candidate.Port,
		},
		ShardsHash: h.ShardsHash.Bytes(),
		Hash:       h.Hashes.Bytes(),
		COSign: &pb.COSign{
			Step1: h.COSign.Step1,
			Step2: h.COSign.Step2,
		},
	}, nil
}

func (h *CMBlockHeader) unSignatureData() ([]byte, error) {
	pbHeader, err := h.proto()
	if err != nil {
		return nil, err
	}
	pbHeader.Hash = nil
	pbHeader.COSign = nil
	data, err := pbHeader.Marshal()
	if err != nil {
		return nil, errors.New(log, fmt.Sprintf("ProtoBuf Marshal error:%s", err.Error()))
	}
	return data, nil
}

func (h *CMBlockHeader) Serialize() ([]byte, error) {
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

func (h *CMBlockHeader) Deserialize(data []byte) error {
	var pbHeader pb.CMBlockHeader
	if err := pbHeader.Unmarshal(data); err != nil {
		return err
	}
	h.ChainID = common.NewHash(pbHeader.ChainID)
	h.Version = pbHeader.Version
	h.Height = pbHeader.Height
	h.Timestamp = pbHeader.Timestamp
	h.PrevHash = common.NewHash(pbHeader.PrevHash)
	h.LeaderPubKey = common.CopyBytes(pbHeader.LeaderPubKey)
	h.Nonce = pbHeader.Nonce
	h.Candidate = NodeInfo{
		PublicKey: common.CopyBytes(pbHeader.Candidate.PublicKey),
		Address:   pbHeader.Candidate.Address,
		Port:      pbHeader.Candidate.Port,
	}
	h.ShardsHash = common.NewHash(pbHeader.ShardsHash)
	h.Hashes = common.NewHash(pbHeader.Hash)
	h.COSign = &types.COSign{
		Step1: pbHeader.COSign.Step1,
		Step2: pbHeader.COSign.Step2,
	}
	return nil
}

func (h *CMBlockHeader) JsonString() string {
	data, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(data)
}

func (h *CMBlockHeader) Type() uint32 {
	return uint32(HeCmBlock)
}

func (h *CMBlockHeader) Hash() common.Hash {
	return h.Hashes
}

func (h *CMBlockHeader) GetHeight() uint64 {
	return h.Height
}

func (h *CMBlockHeader) GetChainID() common.Hash {
	return h.ChainID
}

func (h CMBlockHeader) GetObject() interface{} {
	return h
}

//Block Interface
type NodeAddr struct {
	Address string
	Port    string
}

type Shard struct {
	//Id         uint32
	Member     []NodeInfo
	MemberAddr []NodeAddr
}

func (s *Shard) proto() *pb.Shard {
	pbShard := pb.Shard{
		//Id:         s.Id,
		Member:     nil,
		MemberAddr: nil,
	}
	for _, n := range s.Member {
		pbNodeInfo := pb.NodeInfo{
			PublicKey: n.PublicKey,
			Address:   n.Address,
			Port:      n.Port,
		}
		pbShard.Member = append(pbShard.Member, &pbNodeInfo)
	}
	for _, n := range s.MemberAddr {
		pbNodeAddr := pb.NodeAddr{
			Address: n.Address,
			Port:    n.Port,
		}
		pbShard.MemberAddr = append(pbShard.MemberAddr, &pbNodeAddr)
	}
	return &pbShard
}

func (s *Shard) Deserialize(data []byte) error {
	var pbShard pb.Shard
	if err := pbShard.Unmarshal(data); err != nil {
		return err
	}
	//s.Id = pbShard.Id
	for _, v := range pbShard.Member {
		nodeInfo := NodeInfo{
			PublicKey: common.CopyBytes(v.PublicKey),
			Address:   v.Address,
			Port:      v.Port,
		}
		s.Member = append(s.Member, nodeInfo)
	}
	for _, v := range pbShard.MemberAddr {
		nodeAddr := NodeAddr{
			Address: v.Address,
			Port:    v.Port,
		}
		s.MemberAddr = append(s.MemberAddr, nodeAddr)
	}
	return nil
}

type CMBlock struct {
	CMBlockHeader
	Shards []Shard
}

func NewCmBlock(header CMBlockHeader, shards []Shard) (*CMBlock, error) {
	if err := header.ComputeHash(); err != nil {
		return nil, err
	}
	return &CMBlock{
		CMBlockHeader: header,
		Shards:        shards,
	}, nil
}

func (b *CMBlock) SetSignature(account *account.Account) error {
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

func (b *CMBlock) proto() (block *pb.CMBlock, err error) {
	var pbBlock pb.CMBlock
	pbBlock.Header, err = b.CMBlockHeader.proto()
	if err != nil {
		return nil, err
	}

	for _, shard := range b.Shards {
		pbShard := shard.proto()
		pbBlock.Shards = append(pbBlock.Shards, pbShard)
	}

	return &pbBlock, nil
}

func (b *CMBlock) Serialize() ([]byte, error) {
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

func (b *CMBlock) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New(log, "input data's length is zero")
	}
	var pbBlock pb.CMBlock
	if err := pbBlock.Unmarshal(data); err != nil {
		return errors.New(log, err.Error())
	}
	dataHeader, err := pbBlock.Header.Marshal()
	if err != nil {
		return err
	}

	err = b.CMBlockHeader.Deserialize(dataHeader)
	if err != nil {
		return err
	}

	for _, pbShard := range pbBlock.Shards {
		if bytes, err := pbShard.Marshal(); err != nil {
			return err
		} else {
			var s Shard
			if err := s.Deserialize(bytes); err != nil {
				return err
			}
			b.Shards = append(b.Shards, s)
		}
	}

	return nil
}

func (b *CMBlock) JsonString() string {
	data, err := json.Marshal(b)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return "hash:" + b.Hashes.HexString() + string(data)
}

func (b CMBlock) GetObject() interface{} {
	return b
}
