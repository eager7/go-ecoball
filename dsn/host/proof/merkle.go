package proof

import (
	"bytes"

	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/dsn/host/proof/merkletree"
	"github.com/ecoball/go-ecoball/dsn/common/crypto"
)

const (
	// SegmentSize is the chunk size that is used when taking the Merkle root
	// of a file. 64 is chosen because bandwidth is scarce and it optimizes for
	// the smallest possible storage proofs. Using a larger base, even 256
	// bytes, would result in substantially faster hashing, but the bandwidth
	// tradeoff was deemed to be more important, as blockchain space is scarce.
	SegmentSize = 64
)

// MerkleTree wraps merkletree.Tree, changing some of the function definitions
// to assume sia-specific constants and return sia-specific types.
type MerkleTree struct {
	merkletree.Tree
}

// NewTree returns a MerkleTree, which can be used for getting Merkle roots and
// Merkle proofs on data. See merkletree.Tree for more details.
func NewTree() *MerkleTree {
	return &MerkleTree{*merkletree.New(crypto.NewHash())}
}

// PushObject encodes and adds the hash of the encoded object to the tree as a
// leaf.
func (t *MerkleTree) PushObject(obj interface{}) {
	t.Push(encoding.Marshal(obj))
}

// Root is a redefinition of merkletree.Tree.Root, returning a Hash instead of
// a []byte.
func (t *MerkleTree) Root() (h crypto.Hash) {
	copy(h[:], t.Tree.Root())
	return
}

// CachedMerkleTree wraps merkletree.CachedTree, changing some of the function
// definitions to assume sia-specific constants and return sia-specific types.
type CachedMerkleTree struct {
	merkletree.CachedTree
}

// NewCachedTree returns a CachedMerkleTree, which can be used for getting
// Merkle roots and proofs from data that has cached subroots. See
// merkletree.CachedTree for more details.
func NewCachedTree(height uint64) *CachedMerkleTree {
	return &CachedMerkleTree{*merkletree.NewCachedTree(crypto.NewHash(), height)}
}

// Prove is a redefinition of merkletree.CachedTree.Prove, so that Sia-specific
// types are used instead of the generic types used by the parent package. The
// base is not a return value because the base is used as input.
func (ct *CachedMerkleTree) Prove(base []byte, cachedHashSet []crypto.Hash) []crypto.Hash {
	// Turn the input in to a proof set that will be recognized by the high
	// level tree.
	cachedProofSet := make([][]byte, len(cachedHashSet)+1)
	cachedProofSet[0] = base
	for i := range cachedHashSet {
		cachedProofSet[i+1] = cachedHashSet[i][:]
	}
	_, proofSet, _, _ := ct.CachedTree.Prove(cachedProofSet)

	// convert proofSet to base and hashSet
	hashSet := make([]crypto.Hash, len(proofSet)-1)
	for i, proof := range proofSet[1:] {
		copy(hashSet[i][:], proof)
	}
	return hashSet
}

// Push is a redefinition of merkletree.CachedTree.Push, with the added type
// safety of only accepting a hash.
func (ct *CachedMerkleTree) Push(h crypto.Hash) {
	ct.CachedTree.Push(h[:])
}

// PushSubTree is a redefinition of merkletree.CachedTree.PushSubTree, with the
// added type safety of only accepting a hash.
func (ct *CachedMerkleTree) PushSubTree(height int, h crypto.Hash) error {
	return ct.CachedTree.PushSubTree(height, h[:])
}

// Root is a redefinition of merkletree.CachedTree.Root, returning a Hash
// instead of a []byte.
func (ct *CachedMerkleTree) Root() (h crypto.Hash) {
	copy(h[:], ct.CachedTree.Root())
	return
}

// CalculateLeaves calculates the number of leaves that would be pushed from
// data of size 'dataSize'.
func CalculateLeaves(dataSize uint64) uint64 {
	numSegments := dataSize / SegmentSize
	if dataSize == 0 || dataSize%SegmentSize != 0 {
		numSegments++
	}
	return numSegments
}

// MerkleRoot returns the Merkle root of the input data.
func MerkleRoot(b []byte) crypto.Hash {
	t := NewTree()
	buf := bytes.NewBuffer(b)
	for buf.Len() > 0 {
		t.Push(buf.Next(SegmentSize))
	}
	return t.Root()
}

// MerkleProof builds a Merkle proof that the data at segment 'proofIndex' is a
// part of the Merkle root formed by 'b'.
func MerkleProof(b []byte, proofIndex uint64) (base []byte, hashSet []crypto.Hash) {
	// Create the tree.
	t := NewTree()
	t.SetIndex(proofIndex)

	// Fill the tree.
	buf := bytes.NewBuffer(b)
	for buf.Len() > 0 {
		t.Push(buf.Next(SegmentSize))
	}

	// Get the proof and convert it to a base + hash set.
	_, proof, _, _ := t.Prove()
	if len(proof) == 0 {
		// There's no proof, because there's no data. Return blank values.
		return nil, nil
	}

	base = proof[0]
	hashSet = make([]crypto.Hash, len(proof)-1)
	for i, p := range proof[1:] {
		copy(hashSet[i][:], p)
	}
	return base, hashSet
}

// VerifySegment will verify that a segment, given the proof, is a part of a
// Merkle root.
func VerifySegment(base []byte, hashSet []crypto.Hash, numSegments, proofIndex uint64, root crypto.Hash) bool {
	// convert base and hashSet to proofSet
	proofSet := make([][]byte, len(hashSet)+1)
	proofSet[0] = base
	for i := range hashSet {
		proofSet[i+1] = hashSet[i][:]
	}
	return merkletree.VerifyProof(crypto.NewHash(), root[:], proofSet, proofIndex, numSegments)
}