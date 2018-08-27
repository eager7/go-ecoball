package market

import (
	"github.com/ecoball/go-ecoball/net/crypto"
	"github.com/ecoball/go-ecoball/net/proof"
)

type StorageHostSetting struct {
	PaymentAddr  crypto.Hash
	TotalStorage uint64
	WindowSize   proof.BlockHeight

	RevisionNumber uint64
	Version        string
}

func CreateAnnouncement(pk crypto.PublicKey, sk crypto.SecretKey) (signedAnnounce []byte, err error) {
	//TODO
	return nil, nil
}

func DecodeAnnouncement(fullAnnouncement []byte) (spk crypto.PublicKey, err error) {
	//TODO
	var sk [crypto.PublicKeySize]byte
	return sk, nil
}