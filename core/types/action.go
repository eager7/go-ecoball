package types

import (
	"github.com/ecoball/go-ecoball/common"
	"fmt"
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

func (act *Action)Print() {
	fmt.Println("account: ", act.ContractAccount)

	invoke, ok := act.Payload.GetObject().(InvokeInfo)
	if !ok {
		fmt.Println("transaction type error[invoke]")
		return
	}

	fmt.Println("action: ", string(invoke.Method))
	fmt.Println("param: ", invoke.Param)
}