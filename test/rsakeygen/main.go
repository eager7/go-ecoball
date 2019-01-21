package main

import (
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
)

const nBitsForKeyPairDefault = 1024

func main() {
	var err error
	var privKey crypto.PrivKey
	privKey, _, err = crypto.GenerateKeyPair(crypto.RSA, nBitsForKeyPairDefault)
	if err != nil {
		fmt.Println("failed to generate rsa key pair")
		return
	}

	skey, _ := crypto.MarshalPrivateKey(privKey)
	fmt.Println("Private Key:", crypto.ConfigEncodeKey(skey))
	pubKey := privKey.GetPublic()
	key, _ := crypto.MarshalPublicKey(pubKey)
	fmt.Println("Public  Key:", crypto.ConfigEncodeKey(key))

	b, _ := pubKey.Bytes()
	pub := crypto.ConfigEncodeKey(b)
	id, err := IdFromPublicKey(pub)
	if err != nil {
		return
	}
	fmt.Println("Id Key:", id.Pretty())
}

func IdFromPublicKey(pubKey string) (peer.ID, error) {
	key, err := crypto.ConfigDecodeKey(pubKey)
	if err != nil {
		return "", errors.New(err.Error())
	}
	pk, err := crypto.UnmarshalPublicKey(key)
	if err != nil {
		return "", errors.New(err.Error())
	}
	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return "", errors.New(err.Error())
	}
	return id, nil
}
