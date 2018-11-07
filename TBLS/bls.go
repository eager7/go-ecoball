package TBLS

import (
	"math/big"
	"golang.org/x/crypto/bn256"
	"crypto/sha256"
	"bytes"
	"errors"
	"fmt"
)

type BLSKey struct{
	index int
	private *big.Int
	public *bn256.G2
}

func BLSSign(private *big.Int, msg []byte)[]byte{
	hashPoint := HashToPoint(msg)
	sig := new(bn256.G1).ScalarMult(hashPoint, private)
	point := new(bn256.G1)
	point.Unmarshal(sig.Marshal())
	return point.Marshal()
}

func HashToPoint(msg []byte)*bn256.G1{
	hash := sha256.Sum256(msg)
	hashNum := new(big.Int).SetBytes(hash[:])
	return new(bn256.G1).ScalarBaseMult(hashNum)
}

func VerifySignTBLS(public *bn256.G2, msg, sig []byte)(bool, error){
	sigPoint, result := new(bn256.G1).Unmarshal(sig)
	if result == false {
		err := errors.New("signature unmarshal err")
		return false, err
	}

	pointG2 := new(bn256.G2).ScalarBaseMult(new(big.Int).SetInt64(1))
	left := bn256.Pair(sigPoint, pointG2)
	right := bn256.Pair(HashToPoint(msg), public)

	leftBytes := left.Marshal()
	//fmt.Printf("left  : %x\n",leftBytes)
	rightBytes := right.Marshal()
	//fmt.Printf("right : %x\n",rightBytes)
	if bytes.Compare(leftBytes, rightBytes) != 0{
		err := errors.New("signature verify failed")
		return false, err
	}
	return true, nil
}

//compute group signature of TBLS
func RecoverSignature(abaTBLS *ABATBLS)[]byte{
	var signPoint = new(bn256.G1)
	var qual []int
	for indexJ := range abaTBLS.mapSignDKG {
		// in calculation, should use index+1 instead of index
		qual = append(qual, indexJ+1)
	}
	var tag int
	tag = 0
	for indexJ,value := range abaTBLS.mapSignDKG {
		if tag == 0 {
			tag = 1
			signPoint.Unmarshal(value)
			num := LagrangeBase(indexJ+1, qual)
			signPoint = new(bn256.G1).ScalarMult(signPoint, num)
			//fmt.Println("0,indexJ:",indexJ,value)
			if signPoint == nil {
				fmt.Println("error")
				return nil
			}
		} else {
			sig, _ := new(bn256.G1).Unmarshal(value)
			if sig == nil {
				fmt.Println("error")
			}
			num := LagrangeBase(indexJ+1, qual)
			// fmt.Println("indexJ:",indexJ,qual,value)
			// fmt.Printf("num %s:",num.String())

			sig = new(bn256.G1).ScalarMult(sig, num)
			signPoint = new(bn256.G1).Add(signPoint,sig)
		}
	}
	return signPoint.Marshal()
}

//compute lj(x)
func LagrangeBase(index int,QUAL []int)(*big.Int){
	bigNum2 := new(big.Int).SetInt64(int64(index))
	bigNum3 := new(big.Int).SetInt64(1)
	bigNum4 := new(big.Int).SetInt64(1)
	bigNum5 := new(big.Int).SetInt64(1)
	bigNum6 := new(big.Int).SetInt64(1)
	bigNum7 := new(big.Int).SetInt64(1)
	for i := 0; i < len(QUAL); i++{
		if index == QUAL[i] {
			continue
		}
		bigNum3.SetInt64(int64(QUAL[i])) //xj
		bigNum3.Neg(bigNum3)             //-xm
		bigNum4.Add(bigNum2, bigNum3)    //xj-xm

		bigNum5.Mul(bigNum5, bigNum3)
		bigNum6.Mul(bigNum6, bigNum4)
	}
	// bigNum7 is the inverse of bigNum6 under mod p
	bigNum7.ModInverse(bigNum6,p)

	// bigNum8 := new(big.Int).SetInt64(1)
	// bigNum8.Mul(bigNum7,bigNum6)
	// fmt.Printf("bigNum7,bigNum8:%s,%s\n",bigNum7.String(),bigNum8.String())
	// bigNum8.Mod(bigNum8,p)
	// fmt.Printf("bigNum7,bigNum8:%s,%s\n",bigNum7.String(),bigNum8.String())
	bigNum5.Mul(bigNum5, bigNum7)
	// fmt.Printf("bigNum5:%s\n",bigNum5.String())
	bigNum5.Mod(bigNum5,p)
	// fmt.Printf("bigNum5:%s\n",bigNum5.String())

	return bigNum5
}

/*
func LagrangeBase(index int,QUAL []int)(*big.Int){
	bigNum2 := new(big.Int).SetInt64(int64(index))
	bigNum3 := new(big.Int).SetInt64(1)
	bigNum4 := new(big.Int).SetInt64(1)
	bigNum5 := new(big.Int).SetInt64(1)
	bigNum6 := new(big.Int).SetInt64(1)

	for i := 0; i < len(QUAL); i++{
		if index == QUAL[i] {
			continue
		}
		bigNum3.SetInt64(int64(QUAL[i])) //xj
		bigNum3.Neg(bigNum3)             //-xm
		bigNum4.Add(bigNum2, bigNum3)    //xj-xm

		bigNum5.Mul(bigNum5, bigNum3)
		bigNum6.Mul(bigNum6, bigNum4)
	}
	bigNum5.Div(bigNum5, bigNum6)

	return bigNum5
}
 */