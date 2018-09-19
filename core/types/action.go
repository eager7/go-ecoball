package types

import (
	"github.com/ecoball/go-ecoball/common"
)

type Action struct {
	ContractAccount       common.AccountName			`json:"account"`
	Payload		  		  Payload						`json:"payload"`
}

func NewAction(tx *Transaction) (*Action, error){
	action := &Action{
		ContractAccount:		tx.Addr,
		Payload:				tx.Payload,
	}

	return action, nil
}

func NewSimpleAction(contract string, payload Payload) (*Action, error){
	action := &Action{
		ContractAccount:		common.NameToIndex(contract),
		Payload:				payload,
	}

	return action, nil
}