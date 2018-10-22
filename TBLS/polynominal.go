package TBLS

import "math/big"

type PriPoly struct{
	index int
	coeffs []*big.Int
}



func SetPriShare(index, threshold int)*PriPoly{
	var private PriPoly
	private.coeffs = make([]*big.Int, 0)
	private.index = index
	for i := 0; i < threshold; i++{
		//rand := rand.New(rand.NewSource(time.Now().Unix()))
		//bignum := new(big.Int).Rand(rand, p)
		bignum := new(big.Int).SetInt64(int64(index))
		private.coeffs = append(private.coeffs, bignum)
	}
	return &private
}