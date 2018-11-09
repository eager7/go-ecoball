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

package http

type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

type PubPriKeyPair struct {
	PrivateKey string
	PublicKey  string
}

type KeyPairs struct {
	Pairs []KeyPair
}

type OneKey struct {
	Key []byte
}

type OnePubKey struct{
	Key string
}

type Keys struct {
	KeyList []OneKey
}

type PubKeys struct {
	KeyList []string
}

type Wallets struct {
	NameList []string
}

type RawTransactionData struct {
	PublicKeys Keys
	RawData    []byte
}

type OneSignTransaction struct {
	PublicKey OneKey
	SignData  []byte
}

type SignTransaction struct {
	Signature []OneSignTransaction
}

type WalletNamePassword struct {
	Name     string
	Password string
}

type WalletName struct {
	Name string
}

type WalletImportKey struct {
	Name   string
	PriKey string
}

type WalletRemoveKey struct {
	NamePassword WalletNamePassword
	PubKey       string
}

type WalletTimeout struct {
	Interval int64
}
