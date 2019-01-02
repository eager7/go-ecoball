package shard

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/ecoball/go-ecoball/core/trie"
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
		Candidate:    &pb.NodeInfo{PublicKey: h.Candidate.PublicKey, Address: h.Candidate.Address, Port: h.Candidate.Port},
		ShardsHash:   h.ShardsHash.Bytes(),
		Hash:         h.Hashes.Bytes(),
		COSign:       h.COSign.Proto(),
	}, nil
}

func (h *CMBlockHeader) unSignatureData() ([]byte, error) {
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

func (h *CMBlockHeader) Serialize() ([]byte, error) {
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

func (h *CMBlockHeader) Type() uint32 {
	return uint32(HeCmBlock)
}

func (h *CMBlockHeader) Identify() mpb.Identify {
	return mpb.Identify_APP_MSG_CM_BLOCK
}

func (h *CMBlockHeader) String() string {
	data, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return "hash:" + h.Hashes.HexString() + string(data)
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

func (h *CMBlockHeader) GetInstance() interface{} {
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

func (s *Shard) Hash() (common.Hash, error) {
	data, err := s.proto().Marshal()
	if err != nil {
		return common.Hash{}, errors.New(err.Error())
	}
	hash, err := common.DoubleHash(data)
	if err != nil {
		return common.Hash{}, err
	}
	return hash, nil
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
	var hashes []common.Hash
	for _, s := range shards {
		if hash, err := s.Hash(); err != nil {
			return nil, err
		} else {
			hashes = append(hashes, hash)
		}
	}
	merkleHash, _ := trie.GetMerkleRoot(hashes)
	header.ShardsHash = merkleHash
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
		return errors.New("input data's length is zero")
	}
	var pbBlock pb.CMBlock
	if err := pbBlock.Unmarshal(data); err != nil {
		return errors.New(err.Error())
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

func (b *CMBlock) String() string {
	data, err := json.Marshal(b)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return "hash:" + b.Hashes.HexString() + string(data)
}

func (b *CMBlock) GetInstance() interface{} {
	return b
}
