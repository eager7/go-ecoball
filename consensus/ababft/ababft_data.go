// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.
//
// The following is the ababft consensus algorithm.
// Author: Xu Wang, 2018.07.16

package ababft

import (
	"github.com/ecoball/go-ecoball/core/types"
	"github.com/ecoball/go-ecoball/core/pb"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

type BlockFirstRound struct {
	BlockFirst types.Block
}

type BlockSecondRound struct {
	BlockSecond types.Block
}

type SignaturePreBlock struct {
	SignPreBlock pb.SignaturePreblock
}

func (sign *SignaturePreBlock) Serialize() ([]byte, error) {
	b, err := sign.SignPreBlock.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (sign *SignaturePreBlock) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := sign.SignPreBlock.Unmarshal(data); err != nil {
		return err
	}
	return nil
}

type REQSyn struct {
	Reqsyn pb.RequestSyn
}

func (reqsyn *REQSyn) Serialize() ([]byte, error) {
	b, err := reqsyn.Reqsyn.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (reqsyn *REQSyn) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := reqsyn.Reqsyn.Unmarshal(data); err != nil {
		return err
	}
	return nil
}

type REQSynSolo struct {
	Reqsyn pb.RequestSyn
}

func (reqsyn *REQSynSolo) Serialize() ([]byte, error) {
	b, err := reqsyn.Reqsyn.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (reqsyn *REQSynSolo) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := reqsyn.Reqsyn.Unmarshal(data); err != nil {
		return err
	}
	return nil
}

type TimeoutMsg struct {
	Toutmsg pb.ToutMsg
}

func (toutmsg *TimeoutMsg) Serialize() ([]byte, error) {
	b, err := toutmsg.Toutmsg.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (toutmsg *TimeoutMsg) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := toutmsg.Toutmsg.Unmarshal(data); err != nil {
		return err
	}
	return nil
}

type SignatureBlkF struct {
	signatureBlkF pb.Signature
}

func (sign *SignatureBlkF) Serialize() ([]byte, error) {
	b, err := sign.signatureBlkF.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (sign *SignatureBlkF) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := sign.signatureBlkF.Unmarshal(data); err != nil {
		return err
	}
	return nil
}

type BlockSyn struct {
	Blksyn pb.BlockSyn
}

func (bls *BlockSyn) Serialize() ([]byte, error) {
	b, err := bls.Blksyn.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (bls *BlockSyn) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := bls.Blksyn.Unmarshal(data); err != nil {
		return err
	}
	return nil
}