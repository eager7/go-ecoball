package shard

import (
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/net/message/pb"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/consensus"
	"github.com/ecoball/go-ecoball/sharding/net"
	"github.com/ecoball/go-ecoball/sharding/simulate"
)

func (s *shard) verifyPacket(p *sc.NetPacket) {
	log.Debug("verify packet ", p.BlockType)
	if p.PacketType == pb.MsgType_APP_MSG_CONSENSUS_PACKET {
		s.verifyConsensusPacket(p)
	} else if p.PacketType == pb.MsgType_APP_MSG_SHARDING_PACKET {
		s.verifyShardingPacket(p)
	} else if p.PacketType == pb.MsgType_APP_MSG_SYNC_REQUEST {
		s.verifySyncRequest(p)
	} else if p.PacketType == pb.MsgType_APP_MSG_SYNC_RESPONSE {
		s.verifySyncResponse(p)
	} else {
		log.Error("wrong packet type")
		return
	}
}

func (s *shard) verifyConsensusPacket(p *sc.NetPacket) {
	if p.Step >= consensus.StepNIL || p.Step < consensus.StepPrePare {
		log.Error("wrong step ", p.Step)
		return
	}

	var csp *sc.CsPacket

	if p.BlockType == sc.SD_MINOR_BLOCK {
		csp = s.ns.VerifyMinorPacket(p)
	} else {
		log.Error("wrong block type")
		return
	}

	if csp != nil {
		s.ppc <- csp
	}
}

func (s *shard) verifyShardingPacket(p *sc.NetPacket) {
	var csp *sc.CsPacket

	if p.BlockType == sc.SD_CM_BLOCK {
		csp = s.ns.VerifyCmPacket(p)
	} else if p.BlockType == sc.SD_FINAL_BLOCK {
		csp = s.ns.VerifyFinalPacket(p)
	} else if p.BlockType == sc.SD_VIEWCHANGE_BLOCK {
		csp = s.ns.VerifyViewChangePacket(p)
	} else if p.BlockType == sc.SD_MINOR_BLOCK {
		csp = s.ns.VerifyMinorPacket(p)
	} else {
		log.Error("wrong block type")
		return
	}

	if csp != nil {
		s.ppc <- csp
	}
}

func (c *shard) verifySyncResponse(p *sc.NetPacket) {
	var csp *sc.CsPacket

	if p.BlockType == sc.SD_SYNC {
		csp = c.ns.VerifySyncResponsePacket(p)
	} else {
		log.Error("wrong block type")
		return
	}

	if csp != nil {
		c.ppc <- csp
	}
}

//TODO
func (c *shard) verifySyncRequest(p *sc.NetPacket) {
	var csp *sc.CsPacket

	if p.BlockType == sc.SD_SYNC {
		csp = c.ns.VerifySyncRequestPacket(p)
	} else {
		log.Error("wrong block type")
		return
	}

	if csp != nil {
		c.ppc <- csp
	}
}

func (s *shard) consensusCb(bl interface{}) {
	switch blockType := bl.(type) {
	case *cs.MinorBlock:
		s.commitMinorBlock(bl.(*cs.MinorBlock))
	default:
		log.Error("consensus call back wrong packet type ", blockType)
	}
}

func (s *shard) processRetransTimeout() {
	s.cs.ProcessRetransPacket()
}

func (s *shard) recvConsensusPacket(packet *sc.CsPacket) {
	s.fsm.Execute(ActRecvConsensusPacket, packet)
}

func (s *shard) recvShardingPacket(packet *sc.CsPacket) {
	s.fsm.Execute(ActRecvShardingPacket, packet)
}

/*
func (s *shard) SyncResponseDecode(syncData *sc.SyncResponseData) (*sc.SyncResponsePacket)   {

	blockType := syncData.BlockType
	len := syncData.Len
	data := syncData.Data

	fmt.Println("len = ", len)
	fmt.Println("data = ", data)

	var list []cs.Payload
	for i := 0; i < int(len); i++ {
		blockInterface, err := cs.BlockDeserialize(data[i], cs.HeaderType(blockType))
		if err != nil {
			log.Error("minor block deserialize err")
			return nil
		}
		list = append(list,blockInterface)
	}
	csp := &sc.SyncResponsePacket{
		uint(len),
		blockType,
		list,
	}

	return csp
}

//TODO, make sure TellBlock will be all right
func (s *shard) dealSyncResponse(response *sc.SyncResponsePacket) {
	blocks := response.Blocks
	for _, block := range blocks {
		simulate.TellBlock(block.(cs.BlockInterface))
	}
}

func (s *shard) DealSyncRequestHelperTest(request *sc.SyncRequestPacket) (*sc.NetPacket)  {
	from := request.FromHeight
	to := request.ToHeight

	fmt.Println("from = ", from)
	if to < 0 {
		to = 20
	}
	fmt.Println("to = ", to)

	var response sc.SyncResponsePacket
	for i := from; i <= to; i++ {

		header := cs.MinorBlockHeader {
			Version: 213,
			Height: 21392,
			Timestamp:    time.Now().UnixNano(),

			COSign:       nil,



	}
		cosign := &ty.COSign{}
		cosign.Step1 = 1
		cosign.Step2 = 0

		header.COSign = cosign

		minorBlock := cs.MinorBlock {
			MinorBlockHeader: header,
			Transactions: nil  ,
			StateDelta: nil ,
		}
		response.Blocks = append(response.Blocks, &minorBlock)

	}

	data := response.Encode(uint8(cs.HeMinorBlock))

	csp := &sc.NetPacket{
		PacketType: netmsg.APP_MSG_SYNC_RESPONSE,
		BlockType: sc.SD_SYNC,
	}
	jsonData,err := json.Marshal(data)
	if err != nil {
		log.Error("GetLastShardBlock error", err)
		return nil
	}
	csp.Packet = jsonData

	return csp
}

func (s *shard) DealSyncRequestHelper(request *sc.SyncRequestPacket) (*sc.NetPacket)  {
	from := request.FromHeight
	to := request.ToHeight
	blockType := cs.HeaderType(request.BlockType)

	fmt.Println("from = ", from)
	if to < 0 {
		lastBlock, err := s.ns.Ledger.GetLastShardBlock(config.ChainHash, blockType)
		if err != nil {
			log.Error("GetLastShardBlock error", err)
			return nil
		}
		to = int64(lastBlock.GetHeight())
	}
	if to > from + 10 {
		to = from + 10
	}


	fmt.Println("to = ", to)

	var response sc.SyncResponsePacket
	for i := from; i <= to; i++ {
		blockInterface, err := s.ns.Ledger.GetShardBlockByHeight(config.ChainHash, blockType, uint64(i))
		if err == nil {
			minorBlock := blockInterface.GetObject().(cs.Payload)
			response.Blocks = append(response.Blocks, minorBlock)
		}
	}

	data := response.Encode(uint8(blockType))

	csp := &sc.NetPacket{
		PacketType: netmsg.APP_MSG_SYNC_RESPONSE,
		BlockType: sc.SD_SYNC,
	}
	jsonData,err := json.Marshal(data)
	if err != nil {
		log.Error("GetLastShardBlock error", err)
		return nil
	}
	csp.Packet = jsonData

	return csp
}

//TODO, Restrict max block counts
func (s *shard) dealSyncRequest(request *sc.SyncRequestPacket) {

	worker := request.Worker
	csp := s.DealSyncRequestHelper(request)

	net.Np.SendSyncResponse(csp, worker)

}

func (s *shard)  recvSyncRequestPacket(packet *sc.CsPacket){
	requestPacket := packet.Packet.(*sc.SyncRequestPacket)
	s.dealSyncRequest(requestPacket)
}

func (s *shard)  recvSyncResponsePacket(packet *sc.CsPacket){
	data := packet.Packet.(sc.SyncResponseData)

	p := s.SyncResponseDecode(&data)
	s.dealSyncResponse(p)
}
*/

func (s *shard) processShardingPacket(csp interface{}) {

	p := csp.(*sc.CsPacket)
	switch p.BlockType {
	case sc.SD_CM_BLOCK:
		cm := p.Packet.(*cs.CMBlock)
		last := s.ns.GetLastCMBlock()
		if last == nil {
			panic("last cm block not exist")
			return
		}

		if last.Height >= cm.Height {
			log.Debug("old cm packet ", cm.Height)
			return
		} else if cm.Height > last.Height+1 {
			log.Debug("cm block last ", last.Height, " recv ", cm.Height, " need sync")
			s.fsm.Execute(ActChainNotSync, nil)
			return
		}

		simulate.TellBlock(cm)
		s.ns.SaveLastCMBlock(cm)
		s.broadcastCommitteePacket(p)

		s.fsm.Execute(ActProductMinorBlock, nil)
	case sc.SD_FINAL_BLOCK:
		final := p.Packet.(*cs.FinalBlock)
		lastcm := s.ns.GetLastCMBlock()
		if lastcm == nil {
			panic("last cm block not exist")
			return
		}

		lastfinal := s.ns.GetLastFinalBlock()
		if lastfinal == nil {
			panic("last final block not exist")
			return
		}

		if lastcm.Height > final.EpochNo {
			log.Debug("old final packet epoch", final.EpochNo, " last epoch ", lastcm.Height)
			return
		} else if lastcm.Height < final.EpochNo {
			log.Debug("final block epoch ", final.EpochNo, " last epoch ", lastcm.Height, " need sync")
			s.fsm.Execute(ActChainNotSync, nil)
			return
		}

		if lastfinal.Height >= final.Height {
			log.Debug("old final packet ", final.Height)
			return
		} else if final.Height > lastfinal.Height+1 {
			log.Debug("final block last ", lastfinal.Height, " recv ", final.Height, " need sync")
			s.fsm.Execute(ActChainNotSync, nil)
			return
		}

		if !s.ns.CheckMinorBlockInPool(final) {
			log.Debug("miss minor block , need sync")
			s.fsm.Execute(ActChainNotSync, nil)
			return
		}

		simulate.TellBlock(final)
		s.ns.SaveLastFinalBlock(final)
		s.broadcastCommitteePacket(p)

		if final.Height%sc.DefaultEpochFinalBlockNumber != 0 {
			s.fsm.Execute(ActProductMinorBlock, nil)
		}
	case sc.SD_VIEWCHANGE_BLOCK:
		vc := p.Packet.(*cs.ViewChangeBlock)
		last := s.ns.GetLastViewchangeBlock()
		lastfinal := s.ns.GetLastFinalBlock()
		if lastfinal == nil || last == nil {
			panic("last block is nil")
			return
		}

		if last.Height >= vc.Height {
			log.Debug("old vc packet height ", vc.Height, " last height ", last.Height)
			return
		} else if vc.Height > last.Height+1 {
			log.Debug("vc packet height ", vc.Height, " last height ", last.Height, " need sync")
			s.fsm.Execute(ActChainNotSync, nil)
		}

		if vc.FinalBlockHeight < lastfinal.Height {
			log.Debug("wrong vc packet final height ", vc.FinalBlockHeight, " last final height", lastfinal.Height)
			return
		} else if vc.FinalBlockHeight == lastfinal.Height {
			if last.FinalBlockHeight == vc.FinalBlockHeight {
				if last.Round >= vc.Round {
					log.Debug("old vc packet vc round ", vc.Round, " last round ", last.Round)
					return
				} else if vc.Round > last.Round+1 {
					log.Debug("vc round ", vc.Round, " last round ", last.Round, " need sync")
					s.fsm.Execute(ActChainNotSync, nil)
					return
				}
			} else {
				if vc.Round > 1 {
					log.Debug("vc round ", vc.Round, " need sync")
					s.fsm.Execute(ActChainNotSync, nil)
					return
				} else if vc.Round < 1 {
					log.Debug("wrong round ", vc.Round)
					return
				}
			}
		} else {
			log.Debug("last final ", lastfinal.Height, " recv view change final height ", vc.FinalBlockHeight, " need sync")
			s.fsm.Execute(ActChainNotSync, nil)
			return
		}

		simulate.TellBlock(vc)
		s.ns.SaveLastViewchangeBlock(vc)
		s.broadcastCommitteePacket(p)
	case sc.SD_MINOR_BLOCK:
		minor := p.Packet.(*cs.MinorBlock)
		if !s.ns.SaveMinorBlockToPool(minor) {
			return
		}
		simulate.TellBlock(minor)
		net.Np.TransitBlock(p)

	default:
		log.Error("block type error ", p.BlockType)
		return
	}

}

func (s *shard) broadcastCommitteePacket(p *sc.CsPacket) {
	net.Np.TransitBlock(p)
}
