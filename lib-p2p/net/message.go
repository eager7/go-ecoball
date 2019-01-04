package net

import (
	cryptoRand "crypto/rand"
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/core/types"
	"io"
	"math/big"
	"math/rand"
)

func (i *Instance) NewMessage(msg types.EcoMessage) (*mpb.Message, error) {
	data, err := msg.Serialize()
	if err != nil {
		return nil, err
	}
	nonce := RandomUint64()
	i.msgFilter.Add(nonce, struct{}{})
	m := &mpb.Message{
		Nonce:    nonce,
		Identify: msg.Identify(),
		Payload:  data,
	}
	return m, nil
}

func RandomUint64() uint64 {
	b := make([]byte, 8)
	if _, err := io.ReadFull(cryptoRand.Reader, b); err == nil {
		return new(big.Int).SetBytes(b).Uint64()
	}
	rand.Seed(rand.Int63())
	return uint64(rand.Int63())
}

func (i *Instance) MessageFilter(msg *mpb.Message) bool {
	return i.msgFilter.Contains(msg.Nonce)
}
