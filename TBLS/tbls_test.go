package TBLS

import (
	"testing"
	"github.com/ecoball/go-ecoball/sharding/cell"
	"fmt"
	"github.com/ecoball/go-ecoball/net/network"
)

func TestTBLS(t *testing.T) {
	// build 3 worker
	workers := make([]cell.Worker,3)
	workers[0].Pubkey = "A"
	workers[0].Port   = "1000"
	workers[0].Address = "aa"
	workers[1].Pubkey = "B"
	workers[1].Port   = "1001"
	workers[1].Address = "bb"
	workers[2].Pubkey = "C"
	workers[2].Port   = "1002"
	workers[2].Address = "cc"

	re := StartTBLS(1, 0, workers )
	print("return:", re)
	fmt.Print("ABATBLS:",abaTBLS.epochNum,abaTBLS.index,abaTBLS.threshold)

	StartDKG(abaTBLS.epochNum,0,&abaTBLS)
	fmt.Printf("ABATBLS:%s,%s,%s \n",abaTBLS.PrivatePoly.coeffs[0].String(),abaTBLS.PrivatePoly.coeffs[1].String(),abaTBLS.PrivatePoly.coeffs[2].String())

	abaTBLS1 := ABATBLS{}
	abaTBLS1.actorC = make(chan interface{})
	abaTBLS1.epochNum = 1
	abaTBLS1.threshold = 3
	abaTBLS1.index = 0
	abaTBLS2 := ABATBLS{}
	abaTBLS2.actorC = make(chan interface{})
	abaTBLS2.epochNum = 1
	abaTBLS2.threshold = 3
	abaTBLS2.index = 1
	abaTBLS3 := ABATBLS{}
	abaTBLS3.actorC = make(chan interface{})
	abaTBLS3.epochNum = 1
	abaTBLS3.threshold = 3
	abaTBLS3.index = 2
	abaTBLS1.workers = abaTBLS.workers
	abaTBLS2.workers = abaTBLS.workers
	abaTBLS3.workers = abaTBLS.workers
	abaTBLS1.netObject = network.GetNetInstance()
	abaTBLS2.netObject = network.GetNetInstance()
	abaTBLS3.netObject = network.GetNetInstance()

	abaTBLS1.PrivatePoly = *SetPriShare(abaTBLS1.epochNum, 0, abaTBLS1.threshold)
	abaTBLS2.PrivatePoly = *SetPriShare(abaTBLS2.epochNum, 1, abaTBLS2.threshold)
	abaTBLS3.PrivatePoly = *SetPriShare(abaTBLS3.epochNum, 2, abaTBLS3.threshold)
	fmt.Printf("ABATBLS1:\n %s,\n %s,\n %s\n",abaTBLS1.PrivatePoly.coeffs[0].String(),abaTBLS1.PrivatePoly.coeffs[1].String(),abaTBLS1.PrivatePoly.coeffs[2].String())
	fmt.Printf("ABATBLS2:\n %s,\n %s,\n %s\n",abaTBLS2.PrivatePoly.coeffs[0].String(),abaTBLS2.PrivatePoly.coeffs[1].String(),abaTBLS2.PrivatePoly.coeffs[2].String())
	fmt.Printf("ABATBLS3:\n %s,\n %s,\n %s\n",abaTBLS3.PrivatePoly.coeffs[0].String(),abaTBLS3.PrivatePoly.coeffs[1].String(),abaTBLS3.PrivatePoly.coeffs[2].String())

	abaTBLS1.PubKeyShare = SetPubPolyByPrivate(&abaTBLS1.PrivatePoly)
	abaTBLS2.PubKeyShare = SetPubPolyByPrivate(&abaTBLS2.PrivatePoly)
	abaTBLS3.PubKeyShare = SetPubPolyByPrivate(&abaTBLS3.PrivatePoly)

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
	fmt.Printf("abaTBLS1.PubKeyShare:\n %s,\n %s,\n %s\n",abaTBLS1.PubKeyShare.coEffs[0].String(),abaTBLS1.PubKeyShare.coEffs[1].String(),abaTBLS1.PubKeyShare.coEffs[2].String())
	abaTBLS1.dealer.deal[abaTBLS1.index].pubKeyShare = abaTBLS1.PubKeyShare
	abaTBLS1.dealer.deal[1].pubKeyShare = abaTBLS2.PubKeyShare
	abaTBLS1.dealer.deal[2].pubKeyShare = abaTBLS3.PubKeyShare

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

	fmt.Printf("abaTBLS.dealer.deal[0].keyShare.Sij:\n %s\n",abaTBLS.dealer.deal[0].keyShare.Sij.String())
	fmt.Printf("abaTBLS.dealer.deal[1].keyShare.Sij:\n %s\n",abaTBLS.dealer.deal[1].keyShare.Sij.String())
	fmt.Printf("abaTBLS.dealer.deal[2].keyShare.Sij:\n %s\n",abaTBLS.dealer.deal[2].keyShare.Sij.String())


	genSijMsg(1, 0, &abaTBLS1)
	genSijMsg(1, 1, &abaTBLS2)
	genSijMsg(1, 2, &abaTBLS3)

	fmt.Printf("abaTBLS1.dealer.deal[0].keyShare.Sij:\n %s\n",abaTBLS1.dealer.deal[0].keyShare.Sij.String())
	fmt.Printf("abaTBLS1.dealer.deal[1].keyShare.Sij:\n %s\n",abaTBLS1.dealer.deal[1].keyShare.Sij.String())
	fmt.Printf("abaTBLS1.dealer.deal[2].keyShare.Sij:\n %s\n",abaTBLS1.dealer.deal[2].keyShare.Sij.String())
	fmt.Printf("abaTBLS2.dealer.deal[0].keyShare.Sij:\n %s\n",abaTBLS2.dealer.deal[0].keyShare.Sij.String())
	fmt.Printf("abaTBLS2.dealer.deal[1].keyShare.Sij:\n %s\n",abaTBLS2.dealer.deal[1].keyShare.Sij.String())
	fmt.Printf("abaTBLS2.dealer.deal[2].keyShare.Sij:\n %s\n",abaTBLS2.dealer.deal[2].keyShare.Sij.String())
	fmt.Printf("abaTBLS3.dealer.deal[0].keyShare.Sij:\n %s\n",abaTBLS3.dealer.deal[0].keyShare.Sij.String())
	fmt.Printf("abaTBLS3.dealer.deal[1].keyShare.Sij:\n %s\n",abaTBLS3.dealer.deal[1].keyShare.Sij.String())
	fmt.Printf("abaTBLS3.dealer.deal[2].keyShare.Sij:\n %s\n",abaTBLS3.dealer.deal[2].keyShare.Sij.String())

}
