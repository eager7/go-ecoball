package ababft

import (
	"github.com/ecoball/go-ecoball/common/message"
	"reflect"
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
			log.Info("receive the actorC.msgChan:", msgIn)
			switch msg := msgIn.(type) {
			case message.ABABFTStart:
				{
					log.Info("thread receive the ABABFTStart msg;")
					ProcessSTART(actorC)
					continue
				}
			case SignaturePreBlock:
				{
					ProcessSignPreBlk(actorC,msg)
					continue
				}
			case PreBlockTimeout:
				{
					ProcessPreBlkTimeout(actorC)
					continue
				}
			case BlockFirstRound:
				{
					ProcessBlkF(actorC,msg)
					continue
				}
			case TxTimeout:
				{
					ProcessTxTimeout(actorC)
					continue
				}
			case SignatureBlkF:
				{
					ProcessSignBlkF(actorC,msg)
					continue
				}
			case SignTxTimeout:
				{
					ProcessSignTxTimeout(actorC)
					continue
				}
			case BlockSecondRound:
				{
					ProcessBlkS(actorC,msg)
					continue
				}
			case BlockSTimeout:
				{
					ProcessBlkSTimeout(actorC)
					continue
				}
			case REQSyn:
				{
					ProcessREQSyn(actorC,msg)
					continue
				}
			case REQSynSolo:
				{
					ProcessREQSynSolo(actorC,msg)
					continue
				}
			case BlockSyn:
				{
					ProcessBlkSyn(actorC,msg)
					continue
				}
			case TimeoutMsg:
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
