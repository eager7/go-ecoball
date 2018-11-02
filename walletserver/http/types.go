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
	PrivateKey []byte
	PublicKey  []byte
}

type KeyPairs struct {
	Pairs []KeyPair
}

type OneKey struct {
	Key []byte
}

type Keys struct {
	KeyList []OneKey
}

type Wallets struct {
	NameList []string
}

type TransactionData struct {
	Data []byte
}

type RawTransactionData struct {
	PublicKeys     Keys
	RawTransaction TransactionData
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
	PriKey OneKey
}

type WalletRemoveKey struct {
	NamePassword WalletNamePassword
	PubKey       OneKey
}

type WalletTimeout struct {
	Interval int64
}
