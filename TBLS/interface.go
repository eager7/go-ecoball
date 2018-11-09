package TBLS

import (
	"github.com/ecoball/go-ecoball/sharding/common"
	"golang.org/x/crypto/bn256"
)

type TBLSAPI interface {
	// StartTBLS is used to create/reset the DKG
	StartTBLS(epochNum int, index int, workers []common.Worker) error
	// SignPreTBLS is used to sign the message, which will be sent to the leader to generate the TBLS signature
	SignPreTBLS(msg []byte)[]byte
	//  VerifyPreTBLS is used by the leader to verify the preTBLS signature
	VerifyPreTBLS(indexJ int, epochNum int, msg []byte, sign []byte) bool
	// Generate the TBLS signature by the leader
	GenerateTBLS() ( *bn256.G2, []byte)
	// VerifyTBLS is used by node (non-leader) to verify the TBLS signature
	VerifyTBLS(epochNum int, msg []byte, sign []byte) bool
}
