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

package message

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/net/message/pb"
	"github.com/ecoball/go-ecoball/net/util"
	"gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	pio "gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/io"
	"io"
)

var log = elog.NewLogger("message", elog.DebugLog)

type EcoBallNetMsg interface {
	ChainID() uint32
	Type() pb.MsgType
	Nonce() uint64
	Data() []byte
	Exportable
}

type Exportable interface {
	ToProtoV1() *pb.Message
	ToNetV1(w io.Writer) error
}

type impl struct {
	chainId uint32
	msgType pb.MsgType
	nonce   uint64
	data    []byte
}

func New(msgType pb.MsgType, data []byte) EcoBallNetMsg {
	return newMsg(msgType, data)
}

func newMsg(msgType pb.MsgType, data []byte) *impl {
	return &impl{
		chainId: 1, //TODO
		msgType: msgType,
		nonce:   util.RandomUInt64(),
		data:    data,
	}
}

func NewMessageFromProto(pbm pb.Message) EcoBallNetMsg {
	return &impl{
		chainId: pbm.ChainId,
		msgType: pbm.Type,
		nonce:   pbm.Nonce,
		data:    pbm.Data,
	}
}

func (m *impl) ChainID() uint32 {
	return m.chainId
}

func (m *impl) Type() pb.MsgType {
	return m.msgType
}

func (m *impl) Nonce() uint64 {
	return m.nonce
}

func (m *impl) Data() []byte {
	return m.data
}

func (m *impl) ToProtoV1() *pb.Message {
	return &pb.Message{
		ChainId: m.chainId,
		Type:    m.msgType,
		Nonce:   m.nonce,
		Data:    m.data,
	}
}

func (m *impl) ToNetV1(w io.Writer) error {
	pbw := pio.NewDelimitedWriter(w)
	return pbw.WriteMsg(m.ToProtoV1())
}

func FromNet(r io.Reader) (EcoBallNetMsg, error) {
	pbr := pio.NewDelimitedReader(r, net.MessageSizeMax)
	return FromPBReader(pbr)
}

func FromPBReader(pbr pio.Reader) (EcoBallNetMsg, error) {
	pbMsg := new(pb.Message)
	if err := pbr.ReadMsg(pbMsg); err != nil {
		return nil, errors.New(err.Error())
	}
	return NewMessageFromProto(*pbMsg), nil
}

func NewReader(s net.Stream) pio.ReadCloser {
	return pio.NewDelimitedReader(s, net.MessageSizeMax)
}
