package pb

import pb2 "github.com/ecoball/go-ecoball/core/pb"
import "github.com/syndtr/goleveldb/leveldb/errors"
import (
	"github.com/ecoball/go-ecoball/core/types"
)
type SignaturePreBlockA struct {
	SignPreBlock pb2.SignaturePreblock
}

func (sign *SignaturePreBlockA) Serialize() ([]byte, error) {
	b, err := sign.SignPreBlock.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (sign *SignaturePreBlockA) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := sign.SignPreBlock.Unmarshal(data); err != nil {
		return err
	}
	return nil
}

type BlockFirstRound struct {
	BlockFirst types.Block
}

type REQSynA struct {
	Reqsyn pb2.RequestSyn
}

func (reqsyn *REQSynA) Serialize() ([]byte, error) {
	b, err := reqsyn.Reqsyn.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (reqsyn *REQSynA) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := reqsyn.Reqsyn.Unmarshal(data); err != nil {
		return err
	}
	return nil
}

type REQSynSolo struct {
	Reqsyn pb2.RequestSyn
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
	Toutmsg pb2.ToutMsg
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

type SignatureBlkFA struct {
	SignlkF pb2.Signature
}

func (sign *SignatureBlkFA) Serialize() ([]byte, error) {
	b, err := sign.SignlkF.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (sign *SignatureBlkFA) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := sign.SignlkF.Unmarshal(data); err != nil {
		return err
	}
	return nil
}

type BlockSecondRound struct {
	BlockSecond types.Block
}

type BlockSynA struct {
	Blksyn pb2.BlockSyn
}

func (bls *BlockSynA) Serialize() ([]byte, error) {
	b, err := bls.Blksyn.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (bls *BlockSynA) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}

	if err := bls.Blksyn.Unmarshal(data); err != nil {
		return err
	}
	return nil
}