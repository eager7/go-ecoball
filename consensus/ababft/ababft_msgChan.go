package ababft

import (
	"github.com/ecoball/go-ecoball/common/message"
	"reflect"
	"github.com/ecoball/go-ecoball/net/message/pb"
)

func ConsensusABABFTThread(actorC *ActorABABFT) {
	for {
		select {
		case <-actorC.msgStop:
			{
				log.Info("Stop ABABFT Thread")
				return
			}
		case msgIn :=<- actorC.msgChan:
			switch msg := msgIn.(type) {
			case message.ABABFTStart:
				{
					ProcessSTART(actorC)
					continue
				}
			case pb.SignaturePreBlockA:
				{
					ProcessSignPreBlk(actorC,msg)
					continue
				}
			case PreBlockTimeout:
				{
					ProcessPreBlkTimeout(actorC)
					continue
				}
			case pb.BlockFirstRound:
				{
					ProcessBlkF(actorC,msg)
					continue
				}
			case TxTimeout:
				{
					ProcessTxTimeout(actorC)
					continue
				}
			case pb.SignatureBlkFA:
				{
					ProcessSignBlkF(actorC,msg)
					continue
				}
			case SignTxTimeout:
				{
					ProcessSignTxTimeout(actorC)
					continue
				}
			case pb.BlockSecondRound:
				{
					ProcessBlkS(actorC,msg)
					continue
				}
			case BlockSTimeout:
				{
					ProcessBlkSTimeout(actorC)
					continue
				}
			case pb.REQSynA:
				{
					ProcessREQSyn(actorC,msg)
					continue
				}
			case pb.REQSynSolo:
				{
					ProcessREQSynSolo(actorC,msg)
					continue
				}
			case pb.BlockSynA:
				{
					ProcessBlkSyn(actorC,msg)
					continue
				}
			case pb.TimeoutMsg:
				{
					ProcessTimeoutMsg(actorC,msg)
					continue
				}
			case *message.RegChain:
				{
					go actorC.serviceABABFT.GenNewChain(msg.ChainID,msg.Address)
					continue
				}
			default :
				log.Debug(msg)
				log.Warn("unknown message", reflect.TypeOf(msgIn))
				continue
			}

		}
	}

}
