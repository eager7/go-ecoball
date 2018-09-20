package common

import (
	"github.com/ecoball/go-ecoball/core/types"
)

type NodeInstance interface {
	Start()
	MsgDispatch(msg interface{})
}

const (
	SD_CM_BLOCK = iota + 1
	SD_FINAL_BLOCK
	SD_MINOR_BLOCK
	SD_VIEWCHANGE_BLOCK
	SD_END
)

type NetPacket struct {
	ChainId    uint32
	PacketType uint32
	BlockType  uint16
	Step       uint16
	Packet     []byte
}

type CsPacket struct {
	PacketType uint32
	BlockType  uint16
	Step       uint16
	Packet     interface{}
}

func (c *CsPacket) Copyhead(p *NetPacket) {
	c.PacketType = p.PacketType
	c.BlockType = p.BlockType
	c.Step = p.Step
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
	PrepareRsp() uint16
	PrecommitRsp() uint16
	GetCandidate() *types.NodeInfo
}
