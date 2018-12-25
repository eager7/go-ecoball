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

// Define and Implement the gossip pull internal message

package gossip

import (
	"errors"
	"github.com/ecoball/go-ecoball/net/gossip/protos"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
)

type GspPullHello struct {
	MsgType  pb.PullMsgType
	SenderId peer.ID
}

type GspPullDigest struct {
	MsgType  pb.PullMsgType
	SenderId peer.ID
	Digests  []string
}

type GspPullRequest struct {
	MsgType  pb.PullMsgType
	Asker    peer.ID
	ReqItems []string
}

type GspDataEnv struct {
	Data []byte
}

type GspPullReqAck struct {
	MsgType   pb.PullMsgType
	Responser peer.ID
	Payload   []*GspDataEnv
}

type GossipPullMsg struct {
	SubMsg interface{}
}

func (gpm *GossipPullMsg) Serialize() ([]byte, error) {
	switch gpm.SubMsg.(type) {
	case *GspPullHello:
		return gpm.helloSerialize()
	case *GspPullDigest:
		return gpm.digestSerialize()
	case *GspPullRequest:
		return gpm.requestSerialize()
	case *GspPullReqAck:
		return gpm.reqackSerialize()
	}

	return nil, errors.New("serialize an unknown gpssip pull message")
}

func (gpm *GossipPullMsg) helloSerialize() ([]byte, error) {
	ph := gpm.SubMsg.(*GspPullHello)
	p := pb.GossipPullMsg{
		SubMsg: &pb.GossipPullMsg_Hello{
			Hello: &pb.PullHello{
				SenderId: []byte(ph.SenderId),
				MsgType:  ph.MsgType,
			},
		},
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (gpm *GossipPullMsg) digestSerialize() ([]byte, error) {
	pd := gpm.SubMsg.(*GspPullDigest)
	p := pb.GossipPullMsg{
		SubMsg: &pb.GossipPullMsg_Digest{
			Digest: &pb.PullDigest{
				SenderId: []byte(pd.SenderId),
				Digests:  pd.Digests,
				MsgType:  pd.MsgType,
			},
		},
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (gpm *GossipPullMsg) requestSerialize() ([]byte, error) {
	pr := gpm.SubMsg.(*GspPullRequest)
	p := pb.GossipPullMsg{
		SubMsg: &pb.GossipPullMsg_Request{
			Request: &pb.PullRequest{
				Asker:    []byte(pr.Asker),
				ReqItems: pr.ReqItems,
				MsgType:  pr.MsgType,
			},
		},
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (gpm *GossipPullMsg) reqackSerialize() ([]byte, error) {
	pra := gpm.SubMsg.(*GspPullReqAck)

	ack := &pb.PullReqAck{
		MsgType:   pra.MsgType,
		Responser: []byte(pra.Responser),
	}
	for _, denv := range pra.Payload {
		penv := &pb.PullAckDataEnv{
			Data: denv.Data,
		}
		ack.Payload = append(ack.Payload, penv)
	}
	p := pb.GossipPullMsg{
		SubMsg: &pb.GossipPullMsg_ReqAck{
			ReqAck: ack,
		},
	}

	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (gpm *GossipPullMsg) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var pullmsg pb.GossipPullMsg
	if err := pullmsg.Unmarshal(data); err != nil {
		return err
	}

	switch pullmsg.SubMsg.(type) {
	case *pb.GossipPullMsg_Hello:
		return gpm.helloDerialize(pullmsg.GetHello())
	case *pb.GossipPullMsg_Digest:
		return gpm.digestDerialize(pullmsg.GetDigest())
	case *pb.GossipPullMsg_Request:
		return gpm.requestDerialize(pullmsg.GetRequest())
	case *pb.GossipPullMsg_ReqAck:
		return gpm.reqackDerialize(pullmsg.GetReqAck())
	}

	return errors.New("deserialize an unknow gpssip pull message")
}

func (gpm *GossipPullMsg) helloDerialize(ph interface{}) error {
	if ph == nil {
		return errors.New("hello data is nil")
	}
	h := ph.(*pb.PullHello)
	hello := new(GspPullHello)
	hello.SenderId = peer.ID(h.SenderId)
	gpm.SubMsg = hello

	return nil
}

func (gpm *GossipPullMsg) digestDerialize(pd interface{}) error {
	if pd == nil {
		return errors.New("digest data is nil")
	}
	d := pd.(*pb.PullDigest)
	dig := new(GspPullDigest)
	dig.SenderId = peer.ID(d.SenderId)
	dig.Digests = d.Digests
	gpm.SubMsg = dig

	return nil
}

func (gpm *GossipPullMsg) requestDerialize(pr interface{}) error {
	if pr == nil {
		return errors.New("request data is nil")
	}
	r := pr.(*pb.PullRequest)

	req := new(GspPullRequest)
	req.Asker = peer.ID(r.Asker)
	req.ReqItems = r.ReqItems
	gpm.SubMsg = req

	return nil
}

func (gpm *GossipPullMsg) reqackDerialize(pra interface{}) error {
	if pra == nil {
		return errors.New("response data is nil")
	}
	ra := pra.(*pb.PullReqAck)

	rqa := new(GspPullReqAck)
	rqa.Responser = peer.ID(ra.Responser)

	for _, env := range ra.Payload {
		denv := &GspDataEnv{
			Data: env.Data,
		}
		rqa.Payload = append(rqa.Payload, denv)
	}

	gpm.SubMsg = rqa

	return nil
}
