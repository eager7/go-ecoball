package TBLS

import "math/big"

type ABATBLS struct{
	actorc       chan interface{}
	index int
	coeffs []*big.Int
}

var abaTBLS ABATBLS
func StartTBLS(ViewNum int,index int, threshold int, ) {
	abaTBLS = 	ABATBLS{}
	abaTBLS.actorc = make(chan interface{})
	
}
