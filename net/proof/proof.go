package proof

import (
	"github.com/ecoball/go-ecoball/net/crypto"
	"errors"
)

var (
	errAlteredRevisionPayouts     = errors.New("file contract revision has altered payout volume")
	errInvalidStorageProof        = errors.New("provided storage proof is invalid")
)
type FileContractID crypto.Hash
type BlockHeight uint64 

//type SectorSize uint64(1<<22)
// SectorSize defines how large a sector should be in bytes. The sector
// size needs to be a power of two to be compatible with package
// merkletree. 4MB has been chosen for the live network because large
// sectors significantly reduce the tracking overhead experienced by the
// renter and the host.
//4Mib
var SectorSize uint64 = 1 << 22

// Storage obligations are broken up into ordered atomic sectors that are
// exactly 4MiB each. By saving the roots of each sector, storage proofs
// and modifications to the data can be made inexpensively by making use of
// the merkletree.CachedTree. Sectors can be appended, modified, or deleted
// and the host can recompute the Merkle root of the whole file without
// much computational or I/O expense.
var SectorRoots []crypto.Hash

type StorageProof struct {
	ParentID FileContractID
	Segment  [crypto.SegmentSize]byte
	HashSet  []crypto.Hash
}

type FileContract struct {
	FileSize       uint64
	FileMerkleRoot crypto.Hash
	WindowStart    BlockHeight
	WindowEnd      BlockHeight
	RevisionNumber uint64
}

//StorageProofSegment returns the segment to be used in the storage proof for a given file contract
func StorageProofSegmemt(id FileContractID) (uint64, error) {
	//TODO
	return 0, nil
}
// ReadSector will read a sector from the storage manager, returning the
// bytes that match the input sector root.
func ReadSector(sectorRoot crypto.Hash) ([]byte, error)  {
	//TODO
	return nil, nil
}
func CreateStoragePoof(fid FileContractID) (*StorageProof, error) {
	segmentIndex, err := StorageProofSegmemt(fid)
	if err != nil {
		return nil, err
	}
	sectorIndex := segmentIndex / (SectorSize / crypto.SegmentSize)
	// Pull the corresponding sector into memory.
	sectorRoot := SectorRoots[sectorIndex]
	sectorBytes, err := ReadSector(sectorRoot)
	if err != nil {
		return nil, nil
	}
	//Build the storage proof for just the sector
	sectorSegment := segmentIndex % (SectorSize / crypto.SegmentSize)
	base, cachedHashSet := crypto.MerkleProof(sectorBytes, sectorSegment)
	//Using the sector, build a cached root.
	log2SectorSize := uint64(0)
	for 1<<log2SectorSize < (SectorSize / crypto.SegmentSize) {
		log2SectorSize++
	}
	ct := crypto.NewCachedTree(log2SectorSize)
	ct.SetIndex(segmentIndex)
	for _, root := range SectorRoots {
		ct.Push(root)
	}
	hashSet := ct.Prove(base, cachedHashSet)
	sp := StorageProof{
		ParentID: fid,
		HashSet: hashSet,
	}
	copy(sp.Segment[:], base)

	return &sp, nil
}

func GetFileContract(fid FileContractID) (FileContract, error) {
	//TODO
	return FileContract{}, nil
}

func VerifyStorageProof(fid FileContractID, sp StorageProof) error {
	fc, err := GetFileContract(fid)
	if err != nil {
		return err
	}
	segmentIndex, err := StorageProofSegmemt(fid)
	if err != nil {
		return err
	}
	leaves := crypto.CalculateLeaves(fc.FileSize)
	segmentLen := uint64(crypto.SegmentSize)
	//If this segment chosen is the final segment, it should only be as long
	//as neccessary to complete the filesize
	if segmentIndex == leaves - 1 {
		segmentLen = fc.FileSize % crypto.SegmentSize
	}
	if segmentLen == 0 {
		segmentLen = uint64(crypto.SegmentSize)
	}

	verified := crypto.VerifySegment(
		sp.Segment[:segmentLen],
		sp.HashSet,
		leaves,
		segmentIndex,
		fc.FileMerkleRoot,
	)
	if verified && fc.FileSize > 0 {
		return errInvalidStorageProof
	}

	return nil
}