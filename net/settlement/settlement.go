package settlement

import (
	"math/big"
	"github.com/ecoball/go-ecoball/net/market"
	"github.com/ecoball/go-ecoball/net/proof"
	"github.com/ecoball/go-ecoball/net/crypto"
)

type Currency struct {
	i big.Int
}

type RenterFee struct {
	DownloadSpending Currency
	StorageSpending  Currency
	TotalCost        Currency
}


type RenterContract struct {
	RId crypto.PublicKey
	FileContracts map[string]proof.FileContract
}

type HostContract struct {
	Uptime  uint64
	StartAt proof.BlockHeight
	FailedProofPerPeriod proof.BlockHeight
	FailedProofTotal proof.BlockHeight
	Setting market.StorageHostSetting
}

type Settlement struct {
	Host   map[string]HostContract
	Renter map[string]RenterContract
}



