package types

import (
	"errors"
	"github.com/ecoball/go-ecoball/core/pb"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
)

type P2PRTSynMsg struct {
	Req   peer.ID
}

type PeerAddress struct {
	Id       peer.ID
	Ipport   string
}

type P2PRTSynAckMsg struct {
	Resp      peer.ID
	PAddr     []*PeerAddress
}

func (sync *P2PRTSynMsg) Serialize() ([]byte, error) {
	p := pb.P2PRtSyncMsg{
		Req:[]byte(sync.Req),
	}

	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (sync *P2PRTSynMsg) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	var pb pb.P2PRtSyncMsg
	if err := pb.Unmarshal(data); err != nil {
		return err
	}
	sync.Req = peer.ID(pb.Req)

	return nil
}

func (ack *P2PRTSynAckMsg) Serialize() ([]byte, error) {
	p := pb.P2PRtSyncAckMsg{
		Resp:[]byte(ack.Resp),
	}

	var pis []*pb.PeerInfo
	for _, pa := range ack.PAddr {
		pbPI := &pb.PeerInfo{
			Id:     []byte(pa.Id),
			Ipport: pa.Ipport,
		}
		pis = append(pis, pbPI)
	}
	p.Peers = pis

	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (ack *P2PRTSynAckMsg) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	var pb pb.P2PRtSyncAckMsg
	if err := pb.Unmarshal(data); err != nil {
		return err
	}
	ack.Resp = peer.ID(pb.Resp)
	for _, p := range pb.Peers {
		pi := &PeerAddress{
			Id: peer.ID(p.Id),
			Ipport: p.Ipport,
		}
		ack.PAddr = append(ack.PAddr, pi)
	}

	return nil
}