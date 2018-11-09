package TBLS

import (
	"math/big"
	"golang.org/x/crypto/bn256"
	"fmt"
	"math/rand"
	"time"
	"bytes"
)

// var s = string("65000549695646603732796438742359905742825358107623003571877145026864184071783")
var s = string("65000549695646603732796438742359905742570406053903786389881062969044166799969")
var p, _ = new(big.Int).SetString(s,10)

type PriPoly struct{
	index    int
	coeffs   []*big.Int
	epochNum int
}

type PubPoly struct{
	index    int
	epochNum int
	coEffs   []*bn256.G2
}

type SijShareDKG struct{
	index int
	epochNum int
	Sij big.Int
	pubKeyPoly PubPoly
}

func SetPriShare(epochNum int, index, threshold int)*PriPoly{
	var private PriPoly
	private.coeffs = make([]*big.Int, 0)
	private.index = index
	private.epochNum = epochNum
	for i := 0; i < threshold; i++{
		randSeed := rand.New(rand.NewSource(time.Now().Unix()+int64(2*i*i+8*index*index*index+1024)))
		bigNum := new(big.Int).Rand(randSeed, p)
		// bigNum := new(big.Int).SetInt64(int64(index))
		private.coeffs = append(private.coeffs, bigNum)
	}
	return &private
}

func SetPubPolyByPrivate(private *PriPoly)*PubPoly{
	var public PubPoly
	public.coEffs = make([]*bn256.G2, 0)

	length := len(private.coeffs)

	for i := 0; i < length; i++{
		g1 := new(bn256.G2).ScalarBaseMult(private.coeffs[i])
		public.coEffs = append(public.coEffs,g1)
	}
	public.index = private.index

	return &public
}

func computeSij(priPoly *PriPoly, pubKeyShare *PubPoly, indexJ int, epochNum int)*SijShareDKG{
	var Sij = new(big.Int)
	var bigNum1, bigNum2 *big.Int

	bigNum1 = new(big.Int)
	bigNum2 = new(big.Int)

	Sij.SetString(priPoly.coeffs[0].String(),10)
	// in calculation, should use index+1 instead of index
	bigNum1.SetInt64(int64(indexJ+1))
	for i := 1; i < len(priPoly.coeffs); i++ {
		bigNum2.SetInt64(int64(i))
		bigNum2 = bigNum2.Exp(bigNum1, bigNum2, p)
		bigNum2.Mul(bigNum2, priPoly.coeffs[i])
		Sij = Sij.Add(Sij, bigNum2)
	}
	fmt.Printf("ABATBLS:\n %s,\n %s,\n %s\n",priPoly.coeffs[0].String(),priPoly.coeffs[1].String(),priPoly.coeffs[2].String())
	fmt.Printf("%d,%d,sij = %s\n",priPoly.index,indexJ,Sij.String())

	return &SijShareDKG{priPoly.index, epochNum,*Sij, *pubKeyShare }
}

func SijVerify(sij *SijShareDKG,pubShare *PubPoly, indexJ int, epochNow int, index int)*Complain{
	// in calculation, should use index+1 instead of index
	bignum1 := new(big.Int).SetInt64(int64(index+1))
	bignum2 := new(big.Int)
	bignum3 := new(big.Int)

	// g1 := pubShare.coEffs[0]
	g1 := new(bn256.G2).ScalarMult(pubShare.coEffs[0], new(big.Int).SetInt64(1))
	g2 := new(bn256.G2).ScalarBaseMult(&sij.Sij)

	for i := 1; i < len(pubShare.coEffs); i++{
		bignum2.SetInt64(int64(i))
		bignum3.Exp(bignum1,bignum2,p)

		g := new(bn256.G2).ScalarMult(pubShare.coEffs[i], bignum3)
		g1 = g1.Add(g,g1)
	}

	byte1 := g1.Marshal()
	byte2 := g2.Marshal()
	fmt.Printf("compare:\n")
	fmt.Println(byte1)
	fmt.Println(byte2)
	result := bytes.Compare(byte1, byte2)

	if result == 0 && sij.epochNum == epochNow{
		fmt.Println("compare pass")
		return nil
	}

	return &Complain{indexJ, pubShare.index, sij.epochNum,false }
}