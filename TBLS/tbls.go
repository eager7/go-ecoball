package TBLS

import (
	"math/big"
	"github.com/ecoball/go-ecoball/net/dispatcher"
	"github.com/ecoball/go-ecoball/net/network"
	"time"
	"golang.org/x/crypto/bn256"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/sharding/common"
)

type ABATBLS struct{
	actorC      chan interface{}
	index       int
	epochNum    int
	coEffs      []*big.Int
	PrivatePoly PriPoly
	PubKeyShare *PubPoly
	dealer      Dealer
	workers     []common.Worker
	threshold   int
	msgChan     <-chan interface{}
	netObject   network.EcoballNetwork
	timer       *time.Timer
	timer1      *time.Timer
	timer2      *time.Timer
	PrivateDKG  *big.Int
	PubKeyDKG   PubPoly
	PubKeyQual  *bn256.G2
	Qual        []int
	mapSignDKG map[int][]byte
	validSignNum int
}

var abaTBLS ABATBLS

// index start from 0
func (abaT *ABATBLS)StartTBLS(epochNum int, index int, workers []common.Worker) error {
	var err error
	// initialize/reset the abaTBLS
	abaTBLS = ABATBLS{}
	abaTBLS.actorC = make(chan interface{})
	abaTBLS.epochNum = epochNum
	abaTBLS.threshold = int(2*len(workers)/3)+1
	abaTBLS.index = index

	abaTBLS.PrivatePoly = *SetPriShare(abaTBLS.epochNum, index, abaTBLS.threshold)
	abaTBLS.workers = workers
	abaTBLS.PubKeyShare = SetPubPolyByPrivate(&abaTBLS.PrivatePoly)
	abaTBLS.dealer = Dealer{}
	abaTBLS.dealer.threshold = abaTBLS.threshold
	abaTBLS.dealer.index = abaTBLS.index
	abaTBLS.dealer.private = &abaTBLS.PrivatePoly
	abaTBLS.dealer.complain = 0
	abaTBLS.dealer.QUAL = make([]int,len(workers))
	abaTBLS.dealer.deal = make(map[int]*Deal)
	abaTBLS.dealer.tagQUAL = make([]bool,len(workers))
	abaTBLS.Qual = make([]int,len(workers))
	abaTBLS.dealer.counterQUAL = make([]int,len(workers))
	for i := 0; i < len(abaTBLS.workers); i++ {
		abaTBLS.dealer.deal[i] = new(Deal)
		abaTBLS.dealer.deal[i].keyShare = new(SijNumDKG)
		abaTBLS.dealer.QUAL[i] = 0
		abaTBLS.dealer.tagQUAL[i] = false
		abaTBLS.dealer.counterQUAL[i] = 0
		abaTBLS.Qual[i] = 0
		abaTBLS.dealer.deal[i].complain = make(map[int]*Complain)
	}
	abaTBLS.dealer.QUAL[abaTBLS.index] = 1
	abaTBLS.dealer.deal[abaTBLS.index].pubKeyShare = abaTBLS.PubKeyShare
	abaTBLS.PrivateDKG = &(big.Int{})
	abaTBLS.PubKeyDKG = PubPoly{}
	abaTBLS.mapSignDKG = make(map[int][]byte)
	abaTBLS.validSignNum = 0

	// get the network instance
	abaTBLS.netObject = network.GetNetInstance()
	msg := []pb.MsgType{
		pb.MsgType_APP_MSG_DKGSIJ,
		pb.MsgType_APP_MSG_DKGNLQUAL,
		pb.MsgType_APP_MSG_DKGLQUAL,
	}

	abaTBLS.msgChan, err = dispatcher.Subscribe(msg...)
	if err != nil {
		log.Error("fail to subcribe msg",err)
		return err
	}

	// start the DKG process
	StartDKG(abaTBLS.epochNum,index,&abaTBLS)
	// the DKG is successfully generated

	return nil

}

func (abaT *ABATBLS)SignPreTBLS(msg []byte)[]byte {
	if abaTBLS.PrivateDKG == nil {
		return nil
	}
	signPreTBLS := BLSSign(abaTBLS.PrivateDKG, msg)
	return signPreTBLS
}

func (abaT *ABATBLS)VerifyPreTBLS(indexJ int, epochNum int, msg []byte, sign []byte) bool {
	// step 1, verify the input signTBLS and msg (by the public key corresponding to the index )
	if epochNum != abaTBLS.epochNum {
		log.Info("wrong epoch number!")
		return false
	}
	if abaTBLS.mapSignDKG[indexJ] != nil {
		log.Info("already receive signTBLS from node ", indexJ)
		return false
	}
	result := VerifySignDKG(indexJ, msg, sign, &abaTBLS)
	if result == false {
		return false
	}
	// step 2, save the valid signTBLS
	abaTBLS.mapSignDKG[indexJ] = sign
	abaTBLS.validSignNum++
	return true
}

func (abaT *ABATBLS)GenerateTBLS() ( *bn256.G2, []byte) {
	if abaTBLS.index != 0 {
		// non-leader do not generate the TBLS signTBLS and block
		return nil,nil
	}
	// If enough, then generate the TBLS signTBLS
	if abaTBLS.validSignNum < abaTBLS.threshold {
		return nil,nil
	}
	// enough valid signTBLS, then generate the TBLS signTBLS
	//recover signTBLS
	signTBLS := RecoverSignature(&abaTBLS)
	pubKeyTBLS := abaTBLS.PubKeyQual

	return pubKeyTBLS, signTBLS
}

func (abaT *ABATBLS)VerifyTBLS(epochNum int, msg []byte, sign []byte) bool {
	//verify TBLS signature
	if abaTBLS.PubKeyQual == nil {
		return false
	}
	if epochNum != abaTBLS.epochNum {
		log.Info("wrong epoch number :", epochNum)
		return false
	}
	// verify the msg and signature by using public key
	check, err := VerifySignTBLS(abaTBLS.PubKeyQual, msg, sign)
	if check == false || err != nil {
		return false
	}
	return true
}

