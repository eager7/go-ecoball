package TBLS

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/pb"
	message2 "github.com/ecoball/go-ecoball/net/message"
	npb "github.com/ecoball/go-ecoball/net/message/pb"
	"golang.org/x/crypto/bn256"
	"math/big"
	"time"
)
var (
	log = elog.NewLogger("tbls.dkg", elog.DebugLog)
)

const DefaultSijTime = 1

type Complain struct {
	index int
	from int
	epochNum int
	status bool
}

type SijNumDKG struct {
	index int
	Sij *big.Int
}

// Qual from unleader
type QualNLDKG struct {
	index int
	epochNum int
	QUAL []int
}

type QualLDKG struct {
	index int
	epochNum int
	QUAL []int
}

type Deal struct {
	keyShare *SijNumDKG
	complain map[int]*Complain
	pubKeyShare *PubPoly
	status bool
}

type Dealer struct{
	private *PriPoly
	deal map[int]*Deal
	index int
	threshold int
	complain int
	QUAL []int
	counterQUAL []int // for leader
	tagQUAL []bool // for leader
}

func StartDKG(epochNum int, index int, abaTBLS *ABATBLS) {
	// var err error
	// generate and send the corresponding messages Sij
	genSijMsg(epochNum, index, abaTBLS)

	go dkgRoutine(abaTBLS)


}

func genSijMsg(epochNum int, index int, abaTBLS *ABATBLS) {
	var err error
	workers := abaTBLS.workers
	workerNum := len(workers)
	for i:=0; i< workerNum; i++ {
		if i != index {
			// compute the Sij, SijShareDKG
			sijMsg := computeSij( &abaTBLS.PrivatePoly, abaTBLS.PubKeyShare, i, epochNum)
			// convert to SijPBShareDKG
			var sijMsgPB = pb.SijPBShareDKG{}
			sijMsgPB.EpochNum = int64(sijMsg.epochNum)
			sijMsgPB.Index = int64(sijMsg.index)
			sijMsgPB.Sij, err = sijMsg.Sij.MarshalJSON()
			if err != nil {
				log.Debug("sijMsgPB sij serialization error:", err)
			}
			for j := 0; j < len(sijMsg.pubKeyPoly.coEffs); j++ {
				buf := sijMsg.pubKeyPoly.coEffs[j].Marshal()
				TmpPoint := pb.CurvePointPB{buf}
				sijMsgPB.PubPolyPB = append(sijMsgPB.PubPolyPB, &TmpPoint)
			}
			// send the msg to peer
			msgType := npb.MsgType_APP_MSG_DKGSIJ
			msgData,err := sijMsgPB.Marshal()
			if err != nil {
				log.Debug("sijMsgPB serialization error:", err)
			}
			msgSend := message2.New(msgType, msgData) // EcoBallNetMsg
			abaTBLS.netObject.SendMsgToPeer(workers[i].Address, workers[i].Port, []byte(workers[i].Pubkey), msgSend)
		} else {
			sijMsg := computeSij( &abaTBLS.PrivatePoly, abaTBLS.PubKeyShare, i, epochNum)
			abaTBLS.dealer.deal[index].keyShare.index = index
			abaTBLS.dealer.deal[index].keyShare.Sij = &sijMsg.Sij
			abaTBLS.dealer.QUAL[index] = 1
			abaTBLS.dealer.deal[index].status = false
		}
	}
}

func dkgRoutine(abaTBLS *ABATBLS) {
	log.Debug("start committee routine")
	// timer
	abaTBLS.timer = time.NewTimer(DefaultSijTime * time.Second)
	abaTBLS.timer2 = time.NewTimer((DefaultSijTime*4) * time.Second)
	if abaTBLS.index == 0 {
		abaTBLS.timer1 = time.NewTimer((DefaultSijTime*2) * time.Second)
	}
	for {
		select {
		case msg := <- abaTBLS.msgChan:
			msgIn, ok := msg.(message2.EcoBallNetMsg)
			if !ok {
				log.Error("can't parse msg")
				continue
			}
			ProcessDKGMsg(msgIn, abaTBLS)
		case <-abaTBLS.timer.C:
			abaTBLS.timer.Stop()
			ProcessSijTimeout(abaTBLS)
		case <-abaTBLS.timer1.C:
			ProcessNLQualTimeout(abaTBLS)
			abaTBLS.timer1.Stop()
		case <-abaTBLS.timer2.C:
			ProcessQualTimeout()
			abaTBLS.timer2.Stop()
		}
	}
}

func ProcessDKGMsg(msg message2.EcoBallNetMsg,abaTBLS *ABATBLS) {
	switch msg.Type() {
	case npb.MsgType_APP_MSG_DKGSIJ:
		ProcessSijMSGDKG(msg,abaTBLS)
	case npb.MsgType_APP_MSG_DKGNLQUAL:
		ProcessNLQUALDKG(msg,abaTBLS)
	case npb.MsgType_APP_MSG_DKGLQUAL:
		ProcessLQUALDKG(msg,abaTBLS)
	default:
		log.Error("wrong actor message")
	}
}

func ProcessSijMSGDKG(msg message2.EcoBallNetMsg,abaTBLS *ABATBLS) {
	var err error
	// deserialize the message
	var sijMsgPB = pb.SijPBShareDKG{}
	sijMsgPB.Unmarshal(msg.Data())
	// convert to SijShareDKG
	var sijShareDKG SijShareDKG
	sijShareDKG.epochNum = int(sijMsgPB.EpochNum)
	sijShareDKG.index = int(sijMsgPB.Index)
	err = sijShareDKG.Sij.UnmarshalJSON(sijMsgPB.Sij)
	if err != nil {
		log.Debug("sijMsgPB sij deserialization error:", err)
		return
	}
	var pubKeyPoly = new(PubPoly)
	pubKeyPoly.coEffs = make([]*bn256.G2,abaTBLS.threshold)
	for j := 0; j < abaTBLS.threshold; j++ {
		// var TmpPoint *bn256.G2
		// TmpPoint.Unmarshal(sijMsgPB.PubPolyPB[j].CurvePoint)
		TmpPoint,_ := new(bn256.G2).Unmarshal(sijMsgPB.PubPolyPB[j].CurvePoint)
		sijShareDKG.pubKeyPoly.coEffs = append(sijShareDKG.pubKeyPoly.coEffs, TmpPoint)
		pubKeyPoly.coEffs[j] = TmpPoint
	}
	sijShareDKG.pubKeyPoly.index = sijShareDKG.index
	sijShareDKG.pubKeyPoly.epochNum = sijShareDKG.epochNum
	pubKeyPoly.index = sijShareDKG.index
	pubKeyPoly.epochNum = sijShareDKG.epochNum
	// to compute
	indexIn := sijShareDKG.index

	//abaTBLS.dealer.deal[indexIn].pubKeyShare = &sijShareDKG.pubKeyPoly
	abaTBLS.dealer.deal[indexIn].pubKeyShare = pubKeyPoly

	if abaTBLS.dealer.deal[indexIn].pubKeyShare != nil {
		complain := SijVerify(&sijShareDKG, &sijShareDKG.pubKeyPoly, indexIn, abaTBLS.epochNum, abaTBLS.index)
		if complain != nil {
			abaTBLS.dealer.deal[indexIn].complain[abaTBLS.dealer.index] = complain
			abaTBLS.dealer.complain++
			return
		}
		abaTBLS.dealer.deal[indexIn].keyShare.index = indexIn
		abaTBLS.dealer.deal[indexIn].keyShare.Sij = &sijShareDKG.Sij
		abaTBLS.dealer.QUAL[indexIn] = 1
		abaTBLS.dealer.deal[indexIn].status = false
	}

}

func ProcessSijTimeout(abaTBLS *ABATBLS) {
	// compute the Qual in ProcessSijMSGDKG
	// send the Qual message to the leader
	if abaTBLS.index != 0 {
		var msgQual = pb.QualPBNLDKG{}
		msgQual.Index = int64(abaTBLS.index)
		msgQual.EpochNum = int64(abaTBLS.epochNum)
		msgQual.QUAL = make([]int32,len(abaTBLS.workers))
		for i:=0;i<len(abaTBLS.workers);i++ {
			msgQual.QUAL[i] = int32(abaTBLS.dealer.QUAL[i])
		}
		// send the msg to peer
		msgType := npb.MsgType_APP_MSG_DKGNLQUAL
		msgData,err := msgQual.Marshal()
		if err != nil {
			log.Debug("MSG_DKGNLQUAL serialization error:", err)
		}
		msgSend := message2.New(msgType, msgData) // EcoBallNetMsg
		abaTBLS.netObject.SendMsgToPeer(abaTBLS.workers[0].Address, abaTBLS.workers[0].Port, []byte(abaTBLS.workers[0].Pubkey), msgSend)
	} else {
		abaTBLS.dealer.tagQUAL[0] = true
		for i:=0;i<len(abaTBLS.workers);i++ {
			abaTBLS.dealer.counterQUAL[i] += abaTBLS.dealer.QUAL[i]
		}
	}
}

func ProcessNLQUALDKG(msg message2.EcoBallNetMsg,abaTBLS *ABATBLS) {
	var err error
	// deserialize the message
	var msgQual = pb.QualPBNLDKG{}
	err = msgQual.Unmarshal(msg.Data())
	if err != nil {
		log.Debug("QualPBNLDKG deserialization error:", err)
		return
	}
	var indexIn = int(msgQual.Index)
	// fmt.Println("indexIn:",indexIn,abaTBLS.dealer.tagQUAL[indexIn],msgQual.EpochNum)
	if abaTBLS.dealer.tagQUAL[indexIn]==false && abaTBLS.epochNum==int(msgQual.EpochNum) {
		abaTBLS.dealer.tagQUAL[indexIn] = true
		for i:=0;i<len(abaTBLS.workers);i++ {
			abaTBLS.dealer.counterQUAL[i] += int(msgQual.QUAL[i])
		}
	}
}

func ProcessNLQualTimeout(abaTBLS *ABATBLS) {
	// var err error
	var counterValid int
	counterValid =0
	// leader to calculate the public key of DKG
	if abaTBLS.index == 0 {
		//calculate the qual
		for i:=0;i<len(abaTBLS.workers);i++ {
			if abaTBLS.dealer.counterQUAL[i] >= abaTBLS.threshold {
				abaTBLS.Qual[i] = 1
				counterValid++
			} else {
				abaTBLS.Qual[i] = 0
			}
		}
		if counterValid < abaTBLS.threshold {
			log.Debug("Qual fail: not enough nodes in the Qual set")
			// to add process
		} else {
			// broadcast the Qual message (by leader)
			// will add the DKG public key in message APP_MSG_DKGLQUAL
			var msgQual = pb.QualPBLDKG{}
			msgQual.Index = int64(abaTBLS.index)
			msgQual.EpochNum = int64(abaTBLS.epochNum)
			msgQual.QUAL = make([]int32,len(abaTBLS.workers))
			for i:=0;i<len(abaTBLS.workers);i++ {
				msgQual.QUAL[i] = int32(abaTBLS.Qual[i])
			}
			// send the msg to peer
			msgType := npb.MsgType_APP_MSG_DKGLQUAL
			msgData,err := msgQual.Marshal()
			if err != nil {
				log.Debug("MSG_DKGLQUAL serialization error:", err)
			}
			msgSend := message2.New(msgType, msgData) // EcoBallNetMsg
			err = abaTBLS.netObject.BroadcastMessage(msgSend)
			if err != nil {
				log.Debug("fail to send MSG_DKGLQUAL message:",err)
			}
			// leader to calculate the PrivateDKG and PubKeyDKG
			abaTBLS.PrivateDKG, abaTBLS.PubKeyDKG.coEffs = computePriKeyDKG(abaTBLS)
		}
	}

}

func ProcessQualTimeout() {
	log.Error("fail to receive the qual from the leader")
}

func computePriKeyDKG(abaTBLS *ABATBLS)(*big.Int, []*bn256.G2){
	coEffs := make([]*bn256.G2, abaTBLS.threshold)
	priKey := new(big.Int).SetInt64(0)

	for i := 0; i < len(abaTBLS.Qual); i++ {
		if abaTBLS.Qual[i]>0 {
			if abaTBLS.dealer.QUAL[i] == 0 {
				// means there is missing
				priKey = nil
				break
			}
			priKey.Add(priKey, abaTBLS.dealer.deal[i].keyShare.Sij)
		}
	}
	var tag int
	tag = 0
	for i := 0; i < len(abaTBLS.Qual); i++ {
		if abaTBLS.Qual[i] > 0 {
			if abaTBLS.dealer.QUAL[i] == 0 {
				// means there is missing
				for j:=0;j<abaTBLS.threshold;j++ {
					coEffs[j] = nil
				}
				break
			}
			if tag ==0 {
				for j:=0; j < abaTBLS.threshold; j++ {
					coEffs[j] = new(bn256.G2).ScalarMult(abaTBLS.dealer.deal[i].pubKeyShare.coEffs[j], new(big.Int).SetInt64(1))
				}
				tag = 1
			} else {
				for j:=0; j < abaTBLS.threshold; j++ {
					coEffs[j] = new(bn256.G2).Add(coEffs[j], abaTBLS.dealer.deal[i].pubKeyShare.coEffs[j])
				}
			}
		}
	}
	abaTBLS.PubKeyQual = coEffs[0]

	// maybe need to send the sij again

	return priKey, coEffs
}

func ComputePubKeyDKG(index int, pubPoly []*bn256.G2)*bn256.G2{
	var pubKeyData bn256.G2
	var pubKey *bn256.G2
	pubKey = &pubKeyData
	pubKeyData = *pubPoly[0]
	// in calculation, should use index+1 instead of index
	bigIndex := new(big.Int).SetInt64(int64(index+1))
	for i := 1; i < len(pubPoly); i++{
		bigNum := new(big.Int).SetInt64(int64(i))
		exp := new(big.Int).Exp(bigIndex, bigNum, p)
		point := new(bn256.G2).ScalarMult(pubPoly[i], exp)
		pubKey = new(bn256.G2).Add(pubKey, point)
	}
	return pubKey
}

func ProcessLQUALDKG(msg message2.EcoBallNetMsg,abaTBLS *ABATBLS) {
	var err error
	// deserialize the message
	var msgQual = pb.QualPBLDKG{}
	err = msgQual.Unmarshal(msg.Data())
	if err != nil {
		log.Debug("QualPBLDKG deserialization error:", err)
		return
	}

	if 0 == int(msgQual.Index) && abaTBLS.epochNum == int(msgQual.EpochNum) {
		for i:=0; i<len(msgQual.QUAL);i++ {
			abaTBLS.Qual[i] = int(msgQual.QUAL[i])
		}
	}

	// calculate the PrivateDKG and PubKeyDKG
	abaTBLS.PrivateDKG, abaTBLS.PubKeyDKG.coEffs = computePriKeyDKG(abaTBLS)

	// stop the timer2 for waiting the Qual from leader
	abaTBLS.timer2.Stop()
}

func VerifySignDKG (indexJ int, msg [] byte, sign []byte, abatbls *ABATBLS) bool {
	//verify single DKG signature
	// 1. generate the public
	if abatbls.PrivateDKG == nil {
		return false
	}
	var pubKeyDKG *bn256.G2
	pubKeyDKG = ComputePubKeyDKG(indexJ, abatbls.PubKeyDKG.coEffs)
	// 2. verify the msg and signature by using public key
	check, err := VerifySignTBLS(pubKeyDKG, msg, sign)
	if check == false || err != nil {
		return false
	}
	return true
}
