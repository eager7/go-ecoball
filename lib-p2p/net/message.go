package net

import (
	"github.com/ecoball/go-ecoball/common/message/mpb"
	"github.com/ecoball/go-ecoball/common/utils"
	"github.com/ecoball/go-ecoball/core/types"
)

func (i *Instance) NewMessage(msg types.EcoMessage) (*mpb.Message, error) {
	data, err := msg.Serialize()
	if err != nil {
		return nil, err
	}
	nonce := utils.RandomUint64()
	i.MessageMarked(nonce)
	m := &mpb.Message{
		Nonce:    nonce,
		Identify: msg.Identify(),
		Payload:  data,
	}
	return m, nil
}

func (i *Instance) MessageFilter(msg *mpb.Message) bool {
	return i.msgFilter.Contains(msg.Nonce)
}

func (i *Instance) MessageMarked(key interface{}) {
	i.msgFilter.Add(key, struct{}{})
}
