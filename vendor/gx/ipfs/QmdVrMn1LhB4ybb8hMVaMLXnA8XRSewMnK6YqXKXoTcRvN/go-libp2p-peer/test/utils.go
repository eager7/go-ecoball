package testutil

import (
	"io"
	"math/rand"
	"sync/atomic"
	"time"

	mh "gx/ipfs/QmPnFwZ2JXKnXgMw8CdBPxn7FWh6LLdjUjxV1fKHuJnkr8/go-multihash"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	ci "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

var generatedPairs int64 = 0

func RandPeerID() (peer.ID, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, 16)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	h, _ := mh.Sum(buf, mh.SHA2_256, -1)
	return peer.ID(h), nil
}

func RandTestKeyPair(bits int) (ci.PrivKey, ci.PubKey, error) {
	seed := time.Now().UnixNano()

	// workaround for low time resolution
	seed += atomic.AddInt64(&generatedPairs, 1) << 32

	r := rand.New(rand.NewSource(seed))
	return ci.GenerateKeyPairWithReader(ci.RSA, bits, r)
}

func SeededTestKeyPair(seed int64) (ci.PrivKey, ci.PubKey, error) {
	r := rand.New(rand.NewSource(seed))
	return ci.GenerateKeyPairWithReader(ci.RSA, 512, r)
}
