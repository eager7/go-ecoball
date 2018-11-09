package main

import (
	"fmt"
	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

const nBitsForKeypairDefault  =  1024

func main() {
	var err error
	var privKey ic.PrivKey
	privKey, _, err = ic.GenerateKeyPair(ic.RSA, nBitsForKeypairDefault)
	if err != nil {
		fmt.Println("failed to generate rsa key pair")
		return
	}

	skey, _ := ic.MarshalPrivateKey(privKey)
	fmt.Println("Private Key:", ic.ConfigEncodeKey(skey))
	pubKey := privKey.GetPublic()
	key, _ := ic.MarshalPublicKey(pubKey)
	fmt.Println("Public  Key:", ic.ConfigEncodeKey(key))
}

