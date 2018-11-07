package TBLS

import (
	"testing"
	"fmt"
	"github.com/ecoball/go-ecoball/net/network"
	"github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/net"
	"context"
	"github.com/ecoball/go-ecoball/core/pb"
	pb2 "github.com/ecoball/go-ecoball/net/message/pb"
	message2 "github.com/ecoball/go-ecoball/net/message"
	"time"
	"math/big"
	"golang.org/x/crypto/bn256"
	"bytes"
)

func TestTBLS(t *testing.T) {
	net.InitNetWork(context.Background())
	net.StartNetWork(nil)

	// build 3 worker
	workers := make([]common.Worker,7)
	workers[0].Pubkey = "A"
	workers[0].Port   = "1000"
	workers[0].Address = "aa"
	workers[1].Pubkey = "B"
	workers[1].Port   = "1001"
	workers[1].Address = "bb"
	workers[2].Pubkey = "C"
	workers[2].Port   = "1002"
	workers[2].Address = "cc"
	workers[3].Pubkey = "D"
	workers[3].Port   = "1003"
	workers[3].Address = "dd"
	workers[4].Pubkey = "E"
	workers[4].Port   = "1004"
	workers[4].Address = "ee"
	workers[5].Pubkey = "F"
	workers[5].Port   = "1005"
	workers[5].Address = "ff"
	workers[6].Pubkey = "G"
	workers[6].Port   = "1006"
	workers[6].Address = "gg"

	re := abaTBLS.StartTBLS(1,0,workers)
	print("return:", re)
	fmt.Print("ABATBLS:",abaTBLS.epochNum,abaTBLS.index,abaTBLS.threshold)

	StartDKG(abaTBLS.epochNum,0,&abaTBLS)
	fmt.Printf("ABATBLS:%s,%s,%s \n",abaTBLS.PrivatePoly.coeffs[0].String(),abaTBLS.PrivatePoly.coeffs[1].String(),abaTBLS.PrivatePoly.coeffs[2].String())

	abaTBLS1 := ABATBLS{}
	abaTBLS1.actorC = make(chan interface{})
	abaTBLS1.epochNum = 1
	abaTBLS1.threshold = 5
	abaTBLS1.index = 0

	abaTBLS2 := ABATBLS{}
	abaTBLS2.actorC = make(chan interface{})
	abaTBLS2.epochNum = 1
	abaTBLS2.threshold = 5
	abaTBLS2.index = 1

	abaTBLS3 := ABATBLS{}
	abaTBLS3.actorC = make(chan interface{})
	abaTBLS3.epochNum = 1
	abaTBLS3.threshold = 5
	abaTBLS3.index = 2

	abaTBLS4 := ABATBLS{}
	abaTBLS4.actorC = make(chan interface{})
	abaTBLS4.epochNum = 1
	abaTBLS4.threshold = 5
	abaTBLS4.index = 3

	abaTBLS7 := ABATBLS{}
	abaTBLS7.actorC = make(chan interface{})
	abaTBLS7.epochNum = 1
	abaTBLS7.threshold = 5
	abaTBLS7.index = 6

	abaTBLS1.workers = abaTBLS.workers
	abaTBLS2.workers = abaTBLS.workers
	abaTBLS3.workers = abaTBLS.workers
	abaTBLS4.workers = abaTBLS.workers
	abaTBLS7.workers = abaTBLS.workers

	abaTBLS1.netObject,_ = network.GetNetInstance()
	abaTBLS2.netObject,_ = network.GetNetInstance()
	abaTBLS3.netObject,_ = network.GetNetInstance()
	abaTBLS4.netObject,_ = network.GetNetInstance()
	abaTBLS7.netObject,_ = network.GetNetInstance()

	abaTBLS1.mapSignDKG = make(map[int][]byte)
	abaTBLS2.mapSignDKG = make(map[int][]byte)
	abaTBLS3.mapSignDKG = make(map[int][]byte)
	abaTBLS4.mapSignDKG = make(map[int][]byte)
	abaTBLS7.mapSignDKG = make(map[int][]byte)


	abaTBLS1.PrivatePoly = *SetPriShare(abaTBLS1.epochNum, 0, abaTBLS1.threshold)
	abaTBLS2.PrivatePoly = *SetPriShare(abaTBLS2.epochNum, 1, abaTBLS2.threshold)
	abaTBLS3.PrivatePoly = *SetPriShare(abaTBLS3.epochNum, 2, abaTBLS3.threshold)
	abaTBLS4.PrivatePoly = *SetPriShare(abaTBLS4.epochNum, 3, abaTBLS2.threshold)
	abaTBLS7.PrivatePoly = *SetPriShare(abaTBLS7.epochNum, 6, abaTBLS3.threshold)

	fmt.Printf("ABATBLS1:\n %s,\n %s,\n %s\n",abaTBLS1.PrivatePoly.coeffs[0].String(),abaTBLS1.PrivatePoly.coeffs[1].String(),abaTBLS1.PrivatePoly.coeffs[2].String(),abaTBLS1.PrivatePoly.coeffs[3].String(),abaTBLS1.PrivatePoly.coeffs[4].String())
	fmt.Printf("ABATBLS2:\n %s,\n %s,\n %s\n",abaTBLS2.PrivatePoly.coeffs[0].String(),abaTBLS2.PrivatePoly.coeffs[1].String(),abaTBLS2.PrivatePoly.coeffs[2].String(),abaTBLS2.PrivatePoly.coeffs[3].String(),abaTBLS2.PrivatePoly.coeffs[4].String())
	fmt.Printf("ABATBLS3:\n %s,\n %s,\n %s\n",abaTBLS3.PrivatePoly.coeffs[0].String(),abaTBLS4.PrivatePoly.coeffs[1].String(),abaTBLS3.PrivatePoly.coeffs[2].String(),abaTBLS3.PrivatePoly.coeffs[3].String(),abaTBLS3.PrivatePoly.coeffs[4].String())
	fmt.Printf("ABATBLS4:\n %s,\n %s,\n %s\n",abaTBLS4.PrivatePoly.coeffs[0].String(),abaTBLS4.PrivatePoly.coeffs[1].String(),abaTBLS4.PrivatePoly.coeffs[2].String(),abaTBLS4.PrivatePoly.coeffs[3].String(),abaTBLS4.PrivatePoly.coeffs[4].String())
	fmt.Printf("ABATBLS7:\n %s,\n %s,\n %s\n",abaTBLS7.PrivatePoly.coeffs[0].String(),abaTBLS7.PrivatePoly.coeffs[1].String(),abaTBLS7.PrivatePoly.coeffs[2].String(),abaTBLS7.PrivatePoly.coeffs[3].String(),abaTBLS7.PrivatePoly.coeffs[4].String())

	abaTBLS1.PubKeyShare = SetPubPolyByPrivate(&abaTBLS1.PrivatePoly)
	abaTBLS2.PubKeyShare = SetPubPolyByPrivate(&abaTBLS2.PrivatePoly)
	abaTBLS3.PubKeyShare = SetPubPolyByPrivate(&abaTBLS3.PrivatePoly)
	abaTBLS4.PubKeyShare = SetPubPolyByPrivate(&abaTBLS4.PrivatePoly)
	abaTBLS7.PubKeyShare = SetPubPolyByPrivate(&abaTBLS7.PrivatePoly)

	abaTBLS1.dealer = Dealer{}
	abaTBLS1.dealer.threshold = abaTBLS1.threshold
	abaTBLS1.dealer.index = abaTBLS1.index
	abaTBLS1.dealer.private = &abaTBLS1.PrivatePoly
	abaTBLS1.dealer.complain = 0
	abaTBLS1.dealer.QUAL = make([]int,len(workers))
	abaTBLS1.dealer.deal = make(map[int]*Deal)
	abaTBLS1.dealer.tagQUAL = make([]bool,len(workers))
	abaTBLS1.dealer.counterQUAL = make([]int,len(workers))
	abaTBLS1.Qual = make([]int,len(workers))
	for i := 0; i < len(abaTBLS.workers); i++ {
		abaTBLS1.dealer.deal[i] = new(Deal)
		abaTBLS1.dealer.deal[i].keyShare = new(SijNumDKG)
		abaTBLS1.dealer.QUAL[i] = 0
		abaTBLS1.dealer.tagQUAL[i] = false
		abaTBLS1.dealer.counterQUAL[i] = 0
		abaTBLS1.Qual[i] = 0
	}
	abaTBLS1.dealer.QUAL[abaTBLS1.index] = 1


	abaTBLS1.dealer.deal[abaTBLS1.index].pubKeyShare = abaTBLS1.PubKeyShare
	abaTBLS1.dealer.deal[1].pubKeyShare = abaTBLS2.PubKeyShare
	abaTBLS1.dealer.deal[2].pubKeyShare = abaTBLS3.PubKeyShare
	abaTBLS1.dealer.deal[3].pubKeyShare = abaTBLS4.PubKeyShare
	abaTBLS1.dealer.deal[6].pubKeyShare = abaTBLS7.PubKeyShare


	abaTBLS2.dealer = Dealer{}
	abaTBLS2.dealer.threshold = abaTBLS2.threshold
	abaTBLS2.dealer.index = abaTBLS2.index
	abaTBLS2.dealer.private = &abaTBLS2.PrivatePoly
	abaTBLS2.dealer.complain = 0
	abaTBLS2.dealer.QUAL = make([]int,len(workers))
	abaTBLS2.dealer.deal = make(map[int]*Deal)
	abaTBLS2.dealer.tagQUAL = make([]bool,len(workers))
	abaTBLS2.dealer.counterQUAL = make([]int,len(workers))
	abaTBLS2.Qual = make([]int,len(workers))
	for i := 0; i < len(abaTBLS.workers); i++ {
		abaTBLS2.dealer.deal[i] = new(Deal)
		abaTBLS2.dealer.deal[i].keyShare = new(SijNumDKG)
		abaTBLS2.dealer.QUAL[i] = 0
		abaTBLS2.dealer.tagQUAL[i] = false
		abaTBLS2.dealer.counterQUAL[i] = 0
		abaTBLS2.Qual[i] = 0
	}
	abaTBLS2.dealer.QUAL[abaTBLS2.index] = 1
	abaTBLS2.dealer.deal[abaTBLS2.index].pubKeyShare = abaTBLS2.PubKeyShare
	abaTBLS2.dealer.deal[0].pubKeyShare = abaTBLS1.PubKeyShare
	abaTBLS2.dealer.deal[2].pubKeyShare = abaTBLS3.PubKeyShare
	abaTBLS2.dealer.deal[3].pubKeyShare = abaTBLS4.PubKeyShare
	abaTBLS2.dealer.deal[6].pubKeyShare = abaTBLS7.PubKeyShare


	abaTBLS3.dealer = Dealer{}
	abaTBLS3.dealer.threshold = abaTBLS3.threshold
	abaTBLS3.dealer.index = abaTBLS3.index
	abaTBLS3.dealer.private = &abaTBLS3.PrivatePoly
	abaTBLS3.dealer.complain = 0
	abaTBLS3.dealer.QUAL = make([]int,len(workers))
	abaTBLS3.dealer.deal = make(map[int]*Deal)
	abaTBLS3.dealer.tagQUAL = make([]bool,len(workers))
	abaTBLS3.dealer.counterQUAL = make([]int,len(workers))
	abaTBLS3.Qual = make([]int,len(workers))
	for i := 0; i < len(abaTBLS.workers); i++ {
		abaTBLS3.dealer.deal[i] = new(Deal)
		abaTBLS3.dealer.deal[i].keyShare = new(SijNumDKG)
		abaTBLS3.dealer.QUAL[i] = 0
		abaTBLS3.dealer.tagQUAL[i] = false
		abaTBLS3.dealer.counterQUAL[i] = 0
		abaTBLS3.Qual[i] = 0
	}
	abaTBLS3.dealer.QUAL[abaTBLS3.index] = 1
	abaTBLS3.dealer.deal[abaTBLS3.index].pubKeyShare = abaTBLS3.PubKeyShare
	abaTBLS3.dealer.deal[0].pubKeyShare = abaTBLS1.PubKeyShare
	abaTBLS3.dealer.deal[1].pubKeyShare = abaTBLS2.PubKeyShare
	abaTBLS3.dealer.deal[3].pubKeyShare = abaTBLS4.PubKeyShare
	abaTBLS3.dealer.deal[6].pubKeyShare = abaTBLS7.PubKeyShare

	abaTBLS4.dealer = Dealer{}
	abaTBLS4.dealer.threshold = abaTBLS4.threshold
	abaTBLS4.dealer.index = abaTBLS4.index
	abaTBLS4.dealer.private = &abaTBLS4.PrivatePoly
	abaTBLS4.dealer.complain = 0
	abaTBLS4.dealer.QUAL = make([]int,len(workers))
	abaTBLS4.dealer.deal = make(map[int]*Deal)
	abaTBLS4.dealer.tagQUAL = make([]bool,len(workers))
	abaTBLS4.dealer.counterQUAL = make([]int,len(workers))
	abaTBLS4.Qual = make([]int,len(workers))
	for i := 0; i < len(abaTBLS.workers); i++ {
		abaTBLS4.dealer.deal[i] = new(Deal)
		abaTBLS4.dealer.deal[i].keyShare = new(SijNumDKG)
		abaTBLS4.dealer.QUAL[i] = 0
		abaTBLS4.dealer.tagQUAL[i] = false
		abaTBLS4.dealer.counterQUAL[i] = 0
		abaTBLS4.Qual[i] = 0
	}
	abaTBLS4.dealer.QUAL[abaTBLS4.index] = 1
	abaTBLS4.dealer.deal[abaTBLS4.index].pubKeyShare = abaTBLS4.PubKeyShare
	abaTBLS4.dealer.deal[0].pubKeyShare = abaTBLS1.PubKeyShare
	abaTBLS4.dealer.deal[1].pubKeyShare = abaTBLS2.PubKeyShare
	abaTBLS4.dealer.deal[2].pubKeyShare = abaTBLS3.PubKeyShare
	abaTBLS4.dealer.deal[6].pubKeyShare = abaTBLS7.PubKeyShare


	abaTBLS7.dealer = Dealer{}
	abaTBLS7.dealer.threshold = abaTBLS7.threshold
	abaTBLS7.dealer.index = abaTBLS7.index
	abaTBLS7.dealer.private = &abaTBLS7.PrivatePoly
	abaTBLS7.dealer.complain = 0
	abaTBLS7.dealer.QUAL = make([]int,len(workers))
	abaTBLS7.dealer.deal = make(map[int]*Deal)
	abaTBLS7.dealer.tagQUAL = make([]bool,len(workers))
	abaTBLS7.dealer.counterQUAL = make([]int,len(workers))
	abaTBLS7.Qual = make([]int,len(workers))
	for i := 0; i < len(abaTBLS.workers); i++ {
		abaTBLS7.dealer.deal[i] = new(Deal)
		abaTBLS7.dealer.deal[i].keyShare = new(SijNumDKG)
		abaTBLS7.dealer.QUAL[i] = 0
		abaTBLS7.dealer.tagQUAL[i] = false
		abaTBLS7.dealer.counterQUAL[i] = 0
		abaTBLS7.Qual[i] = 0
	}
	abaTBLS7.dealer.QUAL[abaTBLS7.index] = 1
	abaTBLS7.dealer.deal[abaTBLS7.index].pubKeyShare = abaTBLS7.PubKeyShare
	abaTBLS7.dealer.deal[0].pubKeyShare = abaTBLS1.PubKeyShare
	abaTBLS7.dealer.deal[1].pubKeyShare = abaTBLS2.PubKeyShare
	abaTBLS7.dealer.deal[2].pubKeyShare = abaTBLS3.PubKeyShare
	abaTBLS7.dealer.deal[3].pubKeyShare = abaTBLS4.PubKeyShare


	//fmt.Printf("abaTBLS.dealer.deal[0].keyShare.Sij:\n %s\n",abaTBLS.dealer.deal[0].keyShare.Sij.String())
	//fmt.Printf("abaTBLS.dealer.deal[1].keyShare.Sij:\n %s\n",abaTBLS.dealer.deal[1].keyShare.Sij.String())
	//fmt.Printf("abaTBLS.dealer.deal[2].keyShare.Sij:\n %s\n",abaTBLS.dealer.deal[2].keyShare.Sij.String())


	genSijMsg(1, 0, &abaTBLS1)
	genSijMsg(1, 1, &abaTBLS2)
	genSijMsg(1, 2, &abaTBLS3)
	genSijMsg(1, 3, &abaTBLS4)
	genSijMsg(1, 6, &abaTBLS7)

	/*
	sijTmp := computeSij( &abaTBLS2.PrivatePoly, abaTBLS2.PubKeyShare, 0, 1)
	abaTBLS1.dealer.deal[1].keyShare.Sij = &sijTmp.Sij
	abaTBLS1.dealer.deal[1].keyShare.index = 1
	abaTBLS1.dealer.QUAL[1] = 1
	sijTmp = computeSij( &abaTBLS3.PrivatePoly, abaTBLS3.PubKeyShare, 0, 1)
	abaTBLS1.dealer.deal[2].keyShare.Sij = &sijTmp.Sij
	abaTBLS1.dealer.deal[2].keyShare.index = 2
	abaTBLS1.dealer.QUAL[2] = 1
	sijTmp = computeSij( &abaTBLS1.PrivatePoly, abaTBLS1.PubKeyShare, 1, 1)
	abaTBLS2.dealer.deal[0].keyShare.Sij = &sijTmp.Sij
	abaTBLS2.dealer.deal[0].keyShare.index = 0
	abaTBLS2.dealer.QUAL[0] = 1
	sijTmp = computeSij( &abaTBLS3.PrivatePoly, abaTBLS3.PubKeyShare, 1, 1)
	abaTBLS2.dealer.deal[2].keyShare.Sij = &sijTmp.Sij
	abaTBLS2.dealer.deal[2].keyShare.index = 2
	abaTBLS2.dealer.QUAL[2] = 1
	sijTmp = computeSij( &abaTBLS1.PrivatePoly, abaTBLS1.PubKeyShare, 2, 1)
	abaTBLS3.dealer.deal[0].keyShare.Sij = &sijTmp.Sij
	abaTBLS3.dealer.deal[0].keyShare.index = 0
	abaTBLS3.dealer.QUAL[0] = 1
	sijTmp = computeSij( &abaTBLS2.PrivatePoly, abaTBLS2.PubKeyShare, 2, 1)
	abaTBLS3.dealer.deal[1].keyShare.Sij = &sijTmp.Sij
	abaTBLS3.dealer.deal[1].keyShare.index = 1
	abaTBLS3.dealer.QUAL[1] = 1
	*/
	fmt.Printf("abaTBLS1.dealer.deal[0].keyShare.Sij:\n %s\n",abaTBLS1.dealer.deal[0].keyShare.Sij.String())
	fmt.Printf("abaTBLS1.dealer.deal[1].keyShare.Sij:\n %s\n",abaTBLS1.dealer.deal[1].keyShare.Sij.String())
	fmt.Printf("abaTBLS1.dealer.deal[2].keyShare.Sij:\n %s\n",abaTBLS1.dealer.deal[2].keyShare.Sij.String())
	fmt.Printf("abaTBLS1.dealer.deal[3].keyShare.Sij:\n %s\n",abaTBLS1.dealer.deal[3].keyShare.Sij.String())
	fmt.Printf("abaTBLS1.dealer.deal[6].keyShare.Sij:\n %s\n",abaTBLS1.dealer.deal[6].keyShare.Sij.String())


	fmt.Print("abaTBLS1 QUAL:",abaTBLS1.dealer.QUAL,"\n")
	// node 2 send Sij to node 1
	var sijMsgPB = pb.SijPBShareDKG{}
	sijMsgPB.EpochNum = int64(abaTBLS2.epochNum)
	sijMsgPB.Index = int64(abaTBLS2.index)
	var err error
	sijTmp := computeSij( &abaTBLS2.PrivatePoly, abaTBLS2.PubKeyShare, 0, 1)
	fmt.Println("abaTBLS2.PubKeyShare:",abaTBLS2.PubKeyShare.coEffs)

	sijMsgPB.Sij, err = sijTmp.Sij.MarshalJSON()
	if err != nil {
		log.Debug("test:sijMsgPB sij serialization error:", err)
	}
	for j := 0; j < len(sijTmp.pubKeyPoly.coEffs); j++ {
		buf := sijTmp.pubKeyPoly.coEffs[j].Marshal()
		TmpPoint := pb.CurvePointPB{buf}
		sijMsgPB.PubPolyPB = append(sijMsgPB.PubPolyPB, &TmpPoint)
	}
	msgType := pb2.MsgType_APP_MSG_DKGSIJ
	msgData,err := sijMsgPB.Marshal()
	if err != nil {
		log.Debug("sijMsgPB serialization error:", err)
	}
	msgSend := message2.New(msgType, msgData) // EcoBallNetMsg
	ProcessSijMSGDKG(msgSend,&abaTBLS1)
	fmt.Print("abaTBLS1 QUAL:",abaTBLS1.dealer.QUAL,"\n")

	// node 3 send Sij to node 1
	var sijMsgPB1 = pb.SijPBShareDKG{}
	sijMsgPB1.EpochNum = int64(abaTBLS3.epochNum)
	sijMsgPB1.Index = int64(abaTBLS3.index)
	sijTmp1 := computeSij( &abaTBLS3.PrivatePoly, abaTBLS3.PubKeyShare, 0, 1)
	sijMsgPB1.Sij, err = sijTmp1.Sij.MarshalJSON()
	if err != nil {
		log.Debug("test:sijMsgPB sij serialization error:", err)
	}
	for j := 0; j < len(sijTmp1.pubKeyPoly.coEffs); j++ {
		buf := sijTmp1.pubKeyPoly.coEffs[j].Marshal()
		TmpPoint := pb.CurvePointPB{buf}
		sijMsgPB1.PubPolyPB = append(sijMsgPB1.PubPolyPB, &TmpPoint)
	}
	msgType1 := pb2.MsgType_APP_MSG_DKGSIJ
	msgData1,err := sijMsgPB1.Marshal()
	if err != nil {
		log.Debug("sijMsgPB serialization error:", err)
	}
	msgSend1 := message2.New(msgType1, msgData1) // EcoBallNetMsg
	ProcessSijMSGDKG(msgSend1,&abaTBLS1)
	fmt.Print("abaTBLS1 QUAL:",abaTBLS1.dealer.QUAL,"\n")

	// node 4 send Sij to node 1
	var sijMsgPB14 = pb.SijPBShareDKG{}
	sijMsgPB14.EpochNum = int64(abaTBLS4.epochNum)
	sijMsgPB14.Index = int64(abaTBLS4.index)
	sijTmp14 := computeSij( &abaTBLS4.PrivatePoly, abaTBLS4.PubKeyShare, 0, 1)
	sijMsgPB14.Sij, err = sijTmp14.Sij.MarshalJSON()
	if err != nil {
		log.Debug("test:sijMsgPB sij serialization error:", err)
	}
	for j := 0; j < len(sijTmp14.pubKeyPoly.coEffs); j++ {
		buf := sijTmp14.pubKeyPoly.coEffs[j].Marshal()
		TmpPoint := pb.CurvePointPB{buf}
		sijMsgPB14.PubPolyPB = append(sijMsgPB14.PubPolyPB, &TmpPoint)
	}
	msgType14 := pb2.MsgType_APP_MSG_DKGSIJ
	msgData14,err := sijMsgPB14.Marshal()
	if err != nil {
		log.Debug("sijMsgPB serialization error:", err)
	}
	msgSend14 := message2.New(msgType14, msgData14) // EcoBallNetMsg
	ProcessSijMSGDKG(msgSend14,&abaTBLS1)
	fmt.Print("abaTBLS1 QUAL:",abaTBLS1.dealer.QUAL,"\n")


	// node 7 send Sij to node 1
	var sijMsgPB17 = pb.SijPBShareDKG{}
	sijMsgPB17.EpochNum = int64(abaTBLS7.epochNum)
	sijMsgPB17.Index = int64(abaTBLS7.index)
	sijTmp17 := computeSij( &abaTBLS7.PrivatePoly, abaTBLS7.PubKeyShare, 0, 1)
	sijMsgPB17.Sij, err = sijTmp17.Sij.MarshalJSON()
	if err != nil {
		log.Debug("test:sijMsgPB sij serialization error:", err)
	}
	for j := 0; j < len(sijTmp17.pubKeyPoly.coEffs); j++ {
		buf := sijTmp17.pubKeyPoly.coEffs[j].Marshal()
		TmpPoint := pb.CurvePointPB{buf}
		sijMsgPB17.PubPolyPB = append(sijMsgPB17.PubPolyPB, &TmpPoint)
	}
	msgType17 := pb2.MsgType_APP_MSG_DKGSIJ
	msgData17,err := sijMsgPB17.Marshal()
	if err != nil {
		log.Debug("sijMsgPB serialization error:", err)
	}
	msgSend17 := message2.New(msgType17, msgData17) // EcoBallNetMsg
	ProcessSijMSGDKG(msgSend17,&abaTBLS1)
	fmt.Print("abaTBLS1 QUAL:",abaTBLS1.dealer.QUAL,"\n")



	abaTBLS2.dealer.QUAL = abaTBLS1.dealer.QUAL
	abaTBLS3.dealer.QUAL = abaTBLS1.dealer.QUAL
	abaTBLS4.dealer.QUAL = abaTBLS1.dealer.QUAL
	abaTBLS7.dealer.QUAL = abaTBLS1.dealer.QUAL
	abaTBLS1.dealer.tagQUAL[0] = true
	for i:=0;i<len(abaTBLS1.workers);i++ {
		abaTBLS1.dealer.counterQUAL[i] += abaTBLS1.dealer.QUAL[i]
	}
	fmt.Print("abaTBLS1 counterQUAL:",abaTBLS1.dealer.counterQUAL,"\n")



	// node 2 send msgQual to node 1
	var msgQual = pb.QualPBNLDKG{}
	msgQual.Index = int64(abaTBLS2.index)
	msgQual.EpochNum = int64(abaTBLS2.epochNum)
	msgQual.QUAL = make([]int32,len(abaTBLS2.workers))
	for i:=0;i<len(abaTBLS2.workers);i++ {
		msgQual.QUAL[i] = int32(abaTBLS2.dealer.QUAL[i])
	}
	// send the msg to peer
	msgType2 := pb2.MsgType_APP_MSG_DKGNLQUAL
	msgData2,err := msgQual.Marshal()
	if err != nil {
		log.Debug("MSG_DKGNLQUAL serialization error:", err)
	}
	msgSend2 := message2.New(msgType2, msgData2) // EcoBallNetMsg
	ProcessNLQUALDKG(msgSend2,&abaTBLS1)
	fmt.Print("abaTBLS1 counterQUAL:",abaTBLS1.dealer.counterQUAL,"\n")


	// node 3 send msgQual to node 1
	var msgQual1 = pb.QualPBNLDKG{}
	msgQual1.Index = int64(abaTBLS3.index)
	msgQual1.EpochNum = int64(abaTBLS3.epochNum)
	msgQual1.QUAL = make([]int32,len(abaTBLS3.workers))
	for i:=0;i<len(abaTBLS3.workers);i++ {
		msgQual1.QUAL[i] = int32(abaTBLS3.dealer.QUAL[i])
	}
	// send the msg to peer
	msgType3 := pb2.MsgType_APP_MSG_DKGNLQUAL
	msgData3,err := msgQual1.Marshal()
	if err != nil {
		log.Debug("MSG_DKGNLQUAL serialization error:", err)
	}
	msgSend3 := message2.New(msgType3, msgData3) // EcoBallNetMsg
	ProcessNLQUALDKG(msgSend3,&abaTBLS1)
	fmt.Print("abaTBLS1 counterQUAL:",abaTBLS1.dealer.counterQUAL,"\n")

	// node 4 send msgQual to node 1
	var msgQual14 = pb.QualPBNLDKG{}
	msgQual14.Index = int64(abaTBLS4.index)
	msgQual14.EpochNum = int64(abaTBLS4.epochNum)
	msgQual14.QUAL = make([]int32,len(abaTBLS4.workers))
	for i:=0;i<len(abaTBLS4.workers);i++ {
		msgQual14.QUAL[i] = int32(abaTBLS4.dealer.QUAL[i])
	}
	// send the msg to peer
	msgType34 := pb2.MsgType_APP_MSG_DKGNLQUAL
	msgData34,err := msgQual14.Marshal()
	if err != nil {
		log.Debug("MSG_DKGNLQUAL serialization error:", err)
	}
	msgSend34 := message2.New(msgType34, msgData34) // EcoBallNetMsg
	ProcessNLQUALDKG(msgSend34,&abaTBLS1)
	fmt.Print("abaTBLS1 counterQUAL:",abaTBLS1.dealer.counterQUAL,"\n")


	// node 7 send msgQual to node 1
	var msgQual17 = pb.QualPBNLDKG{}
	msgQual17.Index = int64(abaTBLS7.index)
	msgQual17.EpochNum = int64(abaTBLS7.epochNum)
	msgQual17.QUAL = make([]int32,len(abaTBLS7.workers))
	for i:=0;i<len(abaTBLS7.workers);i++ {
		msgQual17.QUAL[i] = int32(abaTBLS7.dealer.QUAL[i])
	}
	// send the msg to peer
	msgType37 := pb2.MsgType_APP_MSG_DKGNLQUAL
	msgData37,err := msgQual17.Marshal()
	if err != nil {
		log.Debug("MSG_DKGNLQUAL serialization error:", err)
	}
	msgSend37 := message2.New(msgType37, msgData37) // EcoBallNetMsg
	ProcessNLQUALDKG(msgSend37,&abaTBLS1)
	fmt.Print("abaTBLS1 counterQUAL:",abaTBLS1.dealer.counterQUAL,"\n")

	fmt.Print("abaTBLS1.Qual:",abaTBLS1.Qual,"\n")


	// node 1 generate the qual
	ProcessNLQualTimeout(&abaTBLS1)
	fmt.Print("abaTBLS1.Qual:",abaTBLS1.Qual,"\n")


	// node 1 send Qual to node 2 and 3
	var msgQual2 = pb.QualPBLDKG{}
	msgQual2.Index = int64(abaTBLS1.index)
	msgQual2.EpochNum = int64(abaTBLS1.epochNum)
	msgQual2.QUAL = make([]int32,len(abaTBLS1.workers))
	for i:=0;i<len(abaTBLS1.workers);i++ {
		msgQual2.QUAL[i] = int32(abaTBLS1.Qual[i])
	}
	// send the msg to peer
	msgType4 := pb2.MsgType_APP_MSG_DKGLQUAL
	msgData4,err := msgQual2.Marshal()
	msgSend4 := message2.New(msgType4, msgData4)


	sijTmp = computeSij( &abaTBLS1.PrivatePoly, abaTBLS1.PubKeyShare, 1, 1)
	abaTBLS2.dealer.deal[0].keyShare.Sij = &sijTmp.Sij
	abaTBLS2.dealer.deal[0].keyShare.index = 0
	abaTBLS2.dealer.QUAL[0] = 1
	sijTmp = computeSij( &abaTBLS3.PrivatePoly, abaTBLS3.PubKeyShare, 1, 1)
	abaTBLS2.dealer.deal[2].keyShare.Sij = &sijTmp.Sij
	abaTBLS2.dealer.deal[2].keyShare.index = 2
	abaTBLS2.dealer.QUAL[2] = 1
	sijTmp = computeSij( &abaTBLS4.PrivatePoly, abaTBLS4.PubKeyShare, 1, 1)
	abaTBLS2.dealer.deal[3].keyShare.Sij = &sijTmp.Sij
	abaTBLS2.dealer.deal[3].keyShare.index = 3
	abaTBLS2.dealer.QUAL[3] = 1
	sijTmp = computeSij( &abaTBLS7.PrivatePoly, abaTBLS7.PubKeyShare, 1, 1)
	abaTBLS2.dealer.deal[6].keyShare.Sij = &sijTmp.Sij
	abaTBLS2.dealer.deal[6].keyShare.index = 6
	abaTBLS2.dealer.QUAL[6] = 1



	sijTmp = computeSij( &abaTBLS1.PrivatePoly, abaTBLS1.PubKeyShare, 2, 1)
	abaTBLS3.dealer.deal[0].keyShare.Sij = &sijTmp.Sij
	abaTBLS3.dealer.deal[0].keyShare.index = 0
	abaTBLS3.dealer.QUAL[0] = 1
	sijTmp = computeSij( &abaTBLS2.PrivatePoly, abaTBLS2.PubKeyShare, 2, 1)
	abaTBLS3.dealer.deal[1].keyShare.Sij = &sijTmp.Sij
	abaTBLS3.dealer.deal[1].keyShare.index = 1
	abaTBLS3.dealer.QUAL[1] = 1
	sijTmp = computeSij( &abaTBLS4.PrivatePoly, abaTBLS4.PubKeyShare, 2, 1)
	abaTBLS3.dealer.deal[3].keyShare.Sij = &sijTmp.Sij
	abaTBLS3.dealer.deal[3].keyShare.index = 3
	abaTBLS3.dealer.QUAL[3] = 1
	sijTmp = computeSij( &abaTBLS7.PrivatePoly, abaTBLS7.PubKeyShare, 2, 1)
	abaTBLS3.dealer.deal[6].keyShare.Sij = &sijTmp.Sij
	abaTBLS3.dealer.deal[6].keyShare.index = 6
	abaTBLS3.dealer.QUAL[6] = 1


	sijTmp = computeSij( &abaTBLS1.PrivatePoly, abaTBLS1.PubKeyShare, 3, 1)
	abaTBLS4.dealer.deal[0].keyShare.Sij = &sijTmp.Sij
	abaTBLS4.dealer.deal[0].keyShare.index = 0
	abaTBLS4.dealer.QUAL[0] = 1
	sijTmp = computeSij( &abaTBLS2.PrivatePoly, abaTBLS2.PubKeyShare, 3, 1)
	abaTBLS4.dealer.deal[1].keyShare.Sij = &sijTmp.Sij
	abaTBLS4.dealer.deal[1].keyShare.index = 1
	abaTBLS4.dealer.QUAL[1] = 1
	sijTmp = computeSij( &abaTBLS3.PrivatePoly, abaTBLS3.PubKeyShare, 3, 1)
	abaTBLS4.dealer.deal[2].keyShare.Sij = &sijTmp.Sij
	abaTBLS4.dealer.deal[2].keyShare.index = 2
	abaTBLS4.dealer.QUAL[2] = 1
	sijTmp = computeSij( &abaTBLS7.PrivatePoly, abaTBLS7.PubKeyShare, 3, 1)
	abaTBLS4.dealer.deal[6].keyShare.Sij = &sijTmp.Sij
	abaTBLS4.dealer.deal[6].keyShare.index = 6
	abaTBLS4.dealer.QUAL[6] = 1



	sijTmp = computeSij( &abaTBLS1.PrivatePoly, abaTBLS1.PubKeyShare, 6, 1)
	abaTBLS7.dealer.deal[0].keyShare.Sij = &sijTmp.Sij
	abaTBLS7.dealer.deal[0].keyShare.index = 0
	abaTBLS7.dealer.QUAL[0] = 1
	sijTmp = computeSij( &abaTBLS2.PrivatePoly, abaTBLS2.PubKeyShare, 6, 1)
	abaTBLS7.dealer.deal[1].keyShare.Sij = &sijTmp.Sij
	abaTBLS7.dealer.deal[1].keyShare.index = 1
	abaTBLS7.dealer.QUAL[1] = 1
	sijTmp = computeSij( &abaTBLS3.PrivatePoly, abaTBLS3.PubKeyShare, 6, 1)
	abaTBLS7.dealer.deal[2].keyShare.Sij = &sijTmp.Sij
	abaTBLS7.dealer.deal[2].keyShare.index = 2
	abaTBLS7.dealer.QUAL[2] = 1
	sijTmp = computeSij( &abaTBLS4.PrivatePoly, abaTBLS4.PubKeyShare, 6, 1)
	abaTBLS7.dealer.deal[3].keyShare.Sij = &sijTmp.Sij
	abaTBLS7.dealer.deal[3].keyShare.index = 3
	abaTBLS7.dealer.QUAL[3] = 1


	fmt.Println("abaTBLS1.PubKeyDKG.coEffs:",abaTBLS1.PubKeyDKG.coEffs)
	fmt.Println("abaTBLS2.PubKeyDKG.coEffs:",abaTBLS2.PubKeyDKG.coEffs)
	fmt.Println("abaTBLS3.PubKeyDKG.coEffs:",abaTBLS3.PubKeyDKG.coEffs)
	fmt.Println("abaTBLS4.PubKeyDKG.coEffs:",abaTBLS4.PubKeyDKG.coEffs)
	fmt.Println("abaTBLS7.PubKeyDKG.coEffs:",abaTBLS7.PubKeyDKG.coEffs)




	fmt.Print("abaTBLS2.Qual:",abaTBLS2.Qual,"\n")
	abaTBLS2.timer2 = time.NewTimer((DefaultSijTime*4) * time.Second)
	ProcessLQUALDKG(msgSend4,&abaTBLS2)
	fmt.Print("abaTBLS2.Qual:",abaTBLS2.Qual,"\n")

	fmt.Print("abaTBLS3.Qual:",abaTBLS3.Qual,"\n")
	abaTBLS3.timer2 = time.NewTimer((DefaultSijTime*4) * time.Second)
	ProcessLQUALDKG(msgSend4,&abaTBLS3)
	fmt.Print("abaTBLS3.Qual:",abaTBLS3.Qual,"\n")

	fmt.Print("abaTBLS4.Qual:",abaTBLS4.Qual,"\n")
	abaTBLS4.timer2 = time.NewTimer((DefaultSijTime*4) * time.Second)
	ProcessLQUALDKG(msgSend4,&abaTBLS4)
	fmt.Print("abaTBLS4.Qual:",abaTBLS4.Qual,"\n")

	fmt.Print("abaTBLS7.Qual:",abaTBLS7.Qual,"\n")
	abaTBLS7.timer2 = time.NewTimer((DefaultSijTime*4) * time.Second)
	ProcessLQUALDKG(msgSend4,&abaTBLS7)
	fmt.Print("abaTBLS7.Qual:",abaTBLS7.Qual,"\n")


	fmt.Println("abaTBLS1.PubKeyDKG.coEffs:",abaTBLS1.PubKeyDKG.coEffs)
	fmt.Println("abaTBLS2.PubKeyDKG.coEffs:",abaTBLS2.PubKeyDKG.coEffs)
	fmt.Println("abaTBLS3.PubKeyDKG.coEffs:",abaTBLS3.PubKeyDKG.coEffs)
	fmt.Println("abaTBLS2.PubKeyDKG.coEffs:",abaTBLS4.PubKeyDKG.coEffs)
	fmt.Println("abaTBLS3.PubKeyDKG.coEffs:",abaTBLS7.PubKeyDKG.coEffs)


	var msg []byte
	msg = []byte(string("hello, world"))
	fmt.Printf("message:%s\n",msg)

	// sign the message
	abaTBLS = abaTBLS1
	signature1 := abaTBLS.SignPreTBLS(msg)
	abaTBLS = abaTBLS2
	signature2 := abaTBLS.SignPreTBLS(msg)
	abaTBLS = abaTBLS3
	signature3 := abaTBLS.SignPreTBLS(msg)
	abaTBLS = abaTBLS4
	signature4 := abaTBLS.SignPreTBLS(msg)
	abaTBLS = abaTBLS7
	//abaTBLS7.PrivateDKG.Add(abaTBLS7.PrivateDKG,new(big.Int).SetInt64(1))
	signature7 := abaTBLS.SignPreTBLS(msg)
	fmt.Println("signature1:",signature1)
	fmt.Println("signature2:",signature2)
	fmt.Println("signature3:",signature3)
	fmt.Println("signature4:",signature4)
	fmt.Println("signature7:",signature7)


	// verify the signature
	fmt.Println("abaTBLS1.PubKeyDKG.coEffs:",abaTBLS1.PubKeyDKG.coEffs)
	abaTBLS.VerifyPreTBLS(0,1,msg,signature1)
	fmt.Println("0,abaTBLS.mapSignDKG:",abaTBLS.mapSignDKG)
	fmt.Println("abaTBLS.validSignNum:",abaTBLS.validSignNum)
	abaTBLS.VerifyPreTBLS(1,1,msg,signature2)
	fmt.Println("1,abaTBLS.mapSignDKG:",abaTBLS.mapSignDKG)
	fmt.Println("abaTBLS.validSignNum:",abaTBLS.validSignNum)
	abaTBLS.VerifyPreTBLS(2,1,msg,signature3)
	fmt.Println("2,abaTBLS.mapSignDKG:",abaTBLS.mapSignDKG)
	fmt.Println("abaTBLS.validSignNum:",abaTBLS.validSignNum)
	abaTBLS.VerifyPreTBLS(3,1,msg,signature4)
	fmt.Println("3,abaTBLS.mapSignDKG:",abaTBLS.mapSignDKG)
	fmt.Println("abaTBLS.validSignNum:",abaTBLS.validSignNum)
	abaTBLS.VerifyPreTBLS(6,1,msg,signature7)
	fmt.Println("6,abaTBLS.mapSignDKG:",abaTBLS.mapSignDKG)
	fmt.Println("abaTBLS.validSignNum:",abaTBLS.validSignNum)
	fmt.Println("abaTBLS1.PubKeyDKG.coEffs:",abaTBLS.PubKeyDKG.coEffs)


	// the leader generate the TBLS signature
	abaTBLS.index = 0
	pubKeyTBLS, signTBLS := abaTBLS.GenerateTBLS()
	fmt.Println("pubKeyTBLS:",pubKeyTBLS)
	fmt.Println("signTBLS:",signTBLS)

	fmt.Println("abaTBLS.PubKeyQual:",abaTBLS.PubKeyQual)
	fmt.Println("abaTBLS1.PubKeyQual:",abaTBLS1.PubKeyQual)
	fmt.Println("abaTBLS2.PubKeyQual:",abaTBLS2.PubKeyQual)
	fmt.Println("abaTBLS3.PubKeyQual:",abaTBLS3.PubKeyQual)
	fmt.Println("abaTBLS4.PubKeyQual:",abaTBLS4.PubKeyQual)
	fmt.Println("abaTBLS7.PubKeyQual:",abaTBLS7.PubKeyQual)

	resultaa := abaTBLS.VerifyTBLS(1,msg,signTBLS)
	fmt.Println("resultaa:",resultaa)

	//	verify the TBLS signature
	abaTBLS = abaTBLS1
	result := abaTBLS.VerifyTBLS(1,msg,signTBLS)
	fmt.Println("verify TBLS result:",result)
	abaTBLS = abaTBLS2
	result1 := abaTBLS.VerifyTBLS(1,msg,signTBLS)
	fmt.Println("verify TBLS result:",result1)
	abaTBLS = abaTBLS3
	result2 := abaTBLS.VerifyTBLS(1,msg,signTBLS)
	fmt.Println("verify TBLS result:",result2)
	abaTBLS = abaTBLS4
	result3 := abaTBLS.VerifyTBLS(1,msg,signTBLS)
	fmt.Println("verify TBLS result:",result3)
	abaTBLS = abaTBLS7
	result7 := abaTBLS.VerifyTBLS(1,msg,signTBLS)
	fmt.Println("verify TBLS result:",result7)

	bigTmp := new(big.Int).SetInt64(0)
	bigTmp = bigTmp.Add(bigTmp,abaTBLS1.PrivatePoly.coeffs[0])
	bigTmp = bigTmp.Add(bigTmp,abaTBLS2.PrivatePoly.coeffs[0])
	bigTmp = bigTmp.Add(bigTmp,abaTBLS3.PrivatePoly.coeffs[0])
	bigTmp = bigTmp.Add(bigTmp,abaTBLS4.PrivatePoly.coeffs[0])
	bigTmp = bigTmp.Add(bigTmp,abaTBLS7.PrivatePoly.coeffs[0])
	// bigTmp.Mod(bigTmp,p)
	// check the private key
	var qual []int
	for indexJ := range abaTBLS.mapSignDKG {
		// in calculation, should use index+1 instead of index
		qual = append(qual, indexJ+1)
	}

	priRecal := new(big.Int).SetInt64(0)
	num := LagrangeBase(1, qual)
	num.Mul(num,abaTBLS1.PrivateDKG)
	priRecal.Add(priRecal,num)
	num = LagrangeBase(2, qual)
	num.Mul(num,abaTBLS2.PrivateDKG)
	priRecal.Add(priRecal,num)
	num = LagrangeBase(3, qual)
	num.Mul(num,abaTBLS3.PrivateDKG)
	priRecal.Add(priRecal,num)
	num = LagrangeBase(4, qual)
	num.Mul(num,abaTBLS4.PrivateDKG)
	priRecal.Add(priRecal,num)
	num = LagrangeBase(7, qual)
	num.Mul(num,abaTBLS7.PrivateDKG)
	priRecal.Add(priRecal,num)
	priRecal.Mod(priRecal,p)
	fmt.Printf("bigTmp:%s\n",bigTmp.String())
	fmt.Printf("priRecal:%s\n",priRecal.String())


	signPreTBLS := BLSSign(bigTmp, msg)
	fmt.Println("signPreTBLS:\n",signPreTBLS)

	bigTmp.Add(bigTmp,new(big.Int).SetInt64(100))
	bigTmp.Add(bigTmp,new(big.Int).SetInt64(-100))
	priTmp := new(bn256.G1).ScalarBaseMult(bigTmp)
	pointG1 := new(bn256.G1).ScalarBaseMult(new(big.Int).SetInt64(1))
	pointG2 := new(bn256.G2).ScalarBaseMult(new(big.Int).SetInt64(1))
	left := bn256.Pair(priTmp, pointG2)
	right := bn256.Pair(pointG1,abaTBLS.PubKeyQual)
	leftBytes := left.Marshal()
	fmt.Printf("left  : %x\n",leftBytes)
	rightBytes := right.Marshal()
	fmt.Printf("right : %x\n",rightBytes)
	fmt.Println("priTmp:",priTmp)
	fmt.Println("bytes.Compare(leftBytes, rightBytes):",bytes.Compare(leftBytes, rightBytes))


	/*
	sijShareDKG := new(SijShareDKG)
	sijShareDKG.index = 0
	sijShareDKG.Sij = *abaTBLS3.dealer.deal[0].keyShare.Sij
	sijShareDKG.pubKeyPoly = *abaTBLS3.dealer.deal[0].pubKeyShare
	result := SijVerify(sijShareDKG, &sijShareDKG.pubKeyPoly, sijShareDKG.index, abaTBLS1.epochNum,abaTBLS3.index )
	println("result:",result)
	*/


}
