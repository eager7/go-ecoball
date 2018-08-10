// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball.
//
// The go-ecoball is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.
package wallet

import (
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
)

func createKey() (privateKey []byte, publicKey []byte, err error) {
	pri, errCreate := secp256k1.NewECDSAPrivateKey()
	if nil != errCreate {
		return nil, nil, errCreate
	}

	privateKey, err = secp256k1.FromECDSA(pri)
	if nil != err {
		return nil, nil, err
	}

	publicKey, err = secp256k1.FromECDSAPub(&pri.PublicKey)
	return
}

func GetPublicFromPrivate(privateKey []byte) (publicKey []byte, err error) {
	pri, errCreate := secp256k1.ToECDSA(privateKey)
	if nil != errCreate {
		return nil, errCreate
	}

	publicKey, err = secp256k1.FromECDSAPub(&pri.PublicKey)
	return
}

func signDigest(digest []byte, privateKey []byte) (signData []byte, err error) {
	return secp256k1.Sign(digest, privateKey)
}

func Verify(data []byte, publicKey []byte, signature []byte) (bool, error) {
	return secp256k1.Verify(data, signature, publicKey)
}
