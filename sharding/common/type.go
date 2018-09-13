package common

type NodeInstance interface {
	Start()
	MsgDispatch(msg interface{})
}

const (
	CS_PREPARE_BLOCK = iota
	CS_PRECOMMIT_BLOCK
	CS_COMMIT_BLOCK
	CS_END
)

const (
	SD_CM_BLOCK = iota
	SD_FINAL_BLOCK
	SD_MINOR_BLOCK
	SD_VIEWCHANGE_BLOCK
	SD_END
)

type SdPacket struct {
	BlockType uint16
	Packet    []byte
}

type CsPacket struct {
	Round     uint16
	BlockType uint16
	Packet    []byte
}

type CsView struct {
	EpochNo     uint64
	FinalHeight uint64
	MinorHeight uint64
}

func (v1 *CsView) Equal(v2 *CsView) bool {
	return v1.EpochNo == v2.EpochNo && v1.FinalHeight == v2.FinalHeight && v1.MinorHeight == v2.MinorHeight
}

type ConsensusInstance interface {
	GetCsView() *CsView
	MakeCsPacket(round uint16) *CsPacket
	GetCsBlock() interface{}
	CacheBlock(packet *CsPacket) *CsView
	PrepareRsp() uint16
	PrecommitRsp() uint16
	UpdateBlock(csp *CsPacket)
}
