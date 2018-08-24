package market

import (
	"github.com/ecoball/go-ecoball/net/proof"
	"github.com/ecoball/go-ecoball/net/settlement"
)

type FileInfo struct {
	LocalPath string
	FileSize uint64
	Redundancy float64
	Expiration proof.BlockHeight
}

type Renter interface {
	Contracts() []settlement.RenterContract
	CurrentPeriod() proof.BlockHeight
	PeriodSpending() settlement.Currency
	DelteFile(path string) error
	DownLoad() error
	File(path string)
}



