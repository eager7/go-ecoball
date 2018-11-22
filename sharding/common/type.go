package common

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/net/network"
)

type NodeInstance interface {
	Start()
	MsgDispatch(msg interface{})
	SetNet(n network.EcoballNetwork)
}

//BlockType
const (
	SD_CM_BLOCK = iota + 1
	SD_FINAL_BLOCK
	SD_MINOR_BLOCK
	SD_VIEWCHANGE_BLOCK
	SD_END
	SD_SYNC
)

type NetPacket struct {
	ChainId    uint32
	PacketType pb.MsgType
	BlockType  uint16
	Step       uint16
	Packet     []byte
}

type CsPacket struct {
	PacketType pb.MsgType
	BlockType  uint16
	Step       uint16
	Packet     interface{}
}

func (c *CsPacket) CopyHeader(p *NetPacket) {
	c.PacketType = p.PacketType
	c.BlockType = p.BlockType
	c.Step = p.Step
}

func (p *NetPacket) CopyHeader(c *CsPacket) {
	p.PacketType = c.PacketType
	p.BlockType = c.BlockType
	p.Step = c.Step
}

func (p1 *NetPacket) DupHeader(p2 *NetPacket) {
	p1.ChainId = p2.ChainId
	p1.PacketType = p2.PacketType
	p1.BlockType = p2.BlockType
	p1.Step = p2.Step
}

type CsView struct {
	EpochNo     uint64
	FinalHeight uint64
	MinorHeight uint64
	Round       uint16
}

func (v1 *CsView) Equal(v2 *CsView) bool {
	return v1.EpochNo == v2.EpochNo && v1.FinalHeight == v2.FinalHeight && v1.MinorHeight == v2.MinorHeight && v1.Round == v2.Round
}

type ConsensusInstance interface {
	GetCsView() *CsView
	MakeNetPacket(round uint16) *NetPacket
	GetCsBlock() interface{}
	CheckBlock(bl interface{}, bLeader bool) bool
	PrepareRsp() uint32
	PrecommitRsp() uint32
	GetCandidate() *cs.NodeInfo
	GetCosign() *types.COSign
}

type WorkerId struct {
	Pubkey  string
	Address string
	Port    string
}
