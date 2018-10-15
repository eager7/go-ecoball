package settlement

import "math/big"

const (
	BlkHeightPerDay = 172800
	BlkHeightPerHour = 7200
	KeyStorageAn = "store_an"
	KeyStorageTotal = "store_total"
	KeyStorageFile = "store_file"
	KeyStorageProof = "store_proof"
	RatioTotal = 10.0000
	RatioUsed = 10.0000
	RationOntime = 10.0000
)

type HostAnceSource struct {
	TotalStorage uint64
	StartAt      uint64
}

type Currency struct {
	i big.Int
}

type RenterFee struct {
	DownloadSpending Currency
	StorageSpending  Currency
	TotalCost        Currency
}

type DiskResource struct{
	TotalCapacity  uint64
	UsedCapacity   uint64
	TotalFileSize  uint64
	TotalFileCount uint64
	Hosts          []string
}

type ProofInfo struct {
	RepoSize uint64
	Snapshot []uint64
}

type fileInfo struct {
	FileSize    uint64
	Redundancy  uint8
}

type Files struct {
	AllFiles []fileInfo
}



