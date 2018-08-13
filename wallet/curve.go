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
	"errors"

	"github.com/ecoball/go-ecoball/crypto/secp256k1"
)

//type Algorithm uint8 //算法类型

/*type Keys struct {
	PublicKey  []byte    `json:"Publickey"`
	PrivateKey []byte    `json:"Privatekey"`
}*/

/**
创建账号
*/
func createKey() ([]byte, []byte, error) {
	pri, err := secp256k1.NewECDSAPrivateKey()
	if err != nil {
		return nil, nil, errors.New("NewECDSAPrivateKey error: " + err.Error())
	}
	pridata, err := secp256k1.FromECDSA(pri)
	if err != nil {
		return nil, nil, errors.New("FromECDSAPrivateKey error: " + err.Error())
	}
	pub, err := secp256k1.FromECDSAPub(&pri.PublicKey)
	if err != nil {
		return nil, nil, errors.New("new account error: " + err.Error())
	}

	return pub, pridata, nil
}

func signDigest(digest []byte, privateKey []byte) (signData []byte, err error) {
	return secp256k1.Sign(digest, privateKey)
}

func Verify(data []byte, publicKey []byte, signature []byte) (bool, error) {
	return secp256k1.Verify(data, signature, publicKey)
}

func getPublicKey(privateKey string) ([]byte, error){
	pri, err := secp256k1.ToECDSA([]byte(privateKey))
	if err != nil {
		return nil, errors.New("NewECDSAPrivateKey error: " + err.Error())
	}
	pub, err := secp256k1.FromECDSAPub(&pri.PublicKey)
	if err != nil {
		return nil, errors.New("new account error: " + err.Error())
	}
	return pub, nil
}
